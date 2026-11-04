package manager

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"charm.land/log/v2"
	"github.com/bogdannbv/armsync/internal/progress"
	"github.com/bogdannbv/armsync/internal/utils"
	"github.com/bogdannbv/armsync/pkg/reforger/game"
	"github.com/bogdannbv/armsync/pkg/reforger/workshop"
	"golang.org/x/sync/errgroup"
)

const (
	ActionInstall = "install"
	ActionUpdate  = "update"
)

type Manager struct {
	addonsDir string
	ws        *workshop.Client
	mux       sync.RWMutex
	installed map[string]Asset
}

func New(addonsDir string, ws *workshop.Client) (*Manager, error) {
	m := &Manager{
		addonsDir: addonsDir,
		ws:        ws,
		mux:       sync.RWMutex{},
		installed: make(map[string]Asset),
	}
	if err := m.LoadInstalledAssets(); err != nil {
		return nil, err
	}

	return m, nil
}

func (mgr *Manager) Init() error {
	return mgr.LoadInstalledAssets()
}

type Diffs struct {
	ToInstall   map[string]game.Mod
	ToUpgrade   map[string]AssetDiff
	ToDowngrade map[string]AssetDiff
}

//type AssetDiff struct {
//	workshop.AssetDiff
//	Name string
//}

type AssetDiff = workshop.AssetDiff

type Diff struct {
	ToInstall map[string]game.Mod
	ToUpdate  map[string]AssetDiff
}

func (mgr *Manager) DiffInstalled(mods []game.Mod) Diff {
	mgr.mux.RLock()
	defer mgr.mux.RUnlock()
	ti := make(map[string]game.Mod)
	diffs := make(map[string]AssetDiff)
	for _, mod := range mods {
		id := mod.ModID
		ass, installed := mgr.installed[id]
		if !installed {
			ti[id] = mod
			continue
		}
		if len(ass.Versions) != 1 {
			panic(fmt.Sprintf("mod %s has zero or more than one versions", id))
		}
		if ass.Versions[0].Version == mod.Version {
			continue
		}

		diffs[id] = AssetDiff{
			ID:            id,
			Name:          mod.Name,
			Version:       ass.Versions[0].Version,
			TargetVersion: mod.Version,
		}

		//diffs[id] = AssetDiff{
		//	Name: mod.Name,
		//	AssetDiff: workshop.AssetDiff{
		//		ID:            id,
		//		Version:       ass.Versions[0].Version,
		//		TargetVersion: mod.Version,
		//	},
		//}
	}
	return Diff{
		ToInstall: ti,
		ToUpdate:  diffs,
	}
}

func (mgr *Manager) Verify(ids ...string) error {
	if len(ids) == 0 {
		ids = utils.ToKeysSlice(mgr.installed)
	}
	fmt.Printf("Verifying %d ids: %v\n", len(ids), ids)
	var eg errgroup.Group
	eg.SetLimit(10)
	var mux sync.Mutex
	errs := make(map[string]map[string]error, len(ids))
	for _, id := range ids {
		meta, ok := mgr.installed[id]
		if !ok {
			return fmt.Errorf("mod %s not found", id)
		}
		if len(meta.Versions) != 1 {
			return fmt.Errorf("mod %s has zero or more than one versions", id)
		}
		for _, file := range meta.Versions[0].Package.Files {
			eg.Go(func() error {
				fmt.Printf("Verifying file %s\n", file.Name)
				mfp := filepath.Join(meta.Dir, getManifestFileName(file.Name, meta.Versions[0].Version))
				mf, err := os.Open(mfp)
				if err != nil {
					return err
				}
				defer mf.Close()

				var man workshop.Manifest
				if err = json.NewDecoder(mf).Decode(&man); err != nil {
					return err
				}

				fp := filepath.Join(meta.Dir, file.Name)
				f, err := os.Open(fp)
				if err != nil {
					return err
				}
				defer f.Close()

				mux.Lock()
				defer mux.Unlock()
				errs[id][file.Name] = workshop.ValidateFile(&man, f)

				return nil
			})
		}
	}
	return eg.Wait()
}

type Event struct {
	ModID      string
	Name       string
	Downloaded int
	Total      int
	Done       bool
	Err        error
}

func (mgr *Manager) Install(ctx context.Context, mods []game.Mod, report func(Event)) error {
	if report == nil {
		report = func(Event) {}
	}
	av := make(map[string]string, len(mods))
	order := make(map[string]int, len(mods))
	for i, mod := range mods {
		av[mod.ModID] = mod.Version
		order[mod.ModID] = i
	}
	dll, err := mgr.ws.GetAssetsDownloadList(ctx, av)
	if err != nil {
		return err
	}
	slices.SortFunc(dll, func(a, b workshop.DownloadItem) int {
		return cmp.Compare(order[a.Asset.ID], order[b.Asset.ID])
	})
	for _, dli := range dll {
		dir := filepath.Join(mgr.addonsDir, workshop.GetGameAssetDirName(&dli.Asset))
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
		tmpDir := filepath.Join(dir, "armsync")
		if err = os.MkdirAll(tmpDir, os.ModePerm); err != nil {
			return err
		}
		ev := Event{ModID: dli.Asset.ID, Name: dli.Asset.Name, Total: dli.TotalFileSize}
		report(ev)
		pct := 0
		var ferr error
		m := workshop.Manifest{}
		for _, dlf := range dli.Files {
			m, ferr = mgr.ws.GetManifest(ctx, dlf.ManifestPath)
			if ferr != nil {
				log.Error("failed to get manifest file", "err", ferr)
				break
			}
			done := ev.Downloaded
			ferr = mgr.installFile(ctx, &dlf, &m, tmpDir, dir, func(p float64) {
				ev.Downloaded = done + int(p*float64(dlf.Size))
				if ev.Total > 0 && ev.Downloaded*100/ev.Total > pct {
					pct = ev.Downloaded * 100 / ev.Total
					report(ev)
				}
			})
			if ferr != nil {
				log.Error("failed to install file", "err", ferr)
				break
			}
			ev.Downloaded = done + dlf.Size
			if ferr = mgr.syncManifestFile(&dlf, &m, dli.Version, dir); ferr != nil {
				log.Error("failed to sync manifest file", "err", ferr)
				break
			}
		}
		if err = removeDir(tmpDir); err != nil {
			log.Warn("failed cleanup temp contents", "err", err)
		}
		if ferr == nil {
			ferr = mgr.syncMetaFile(ctx, &dli, dir)
		}
		if ferr != nil {
			if err = removeDir(dir); err != nil {
				log.Warn("failed cleanup contents", "err", err)
			}
			mgr.mux.Lock()
			delete(mgr.installed, dli.Asset.ID)
			mgr.mux.Unlock()
		}
		ev.Done, ev.Err = true, ferr
		report(ev)
	}

	return nil
}

func (mgr *Manager) syncMetaFile(ctx context.Context, dli *workshop.DownloadItem, dir string) error {
	var err error
	mgr.mux.RLock()
	asset, ok := mgr.installed[dli.Asset.ID]
	mgr.mux.RUnlock()
	if !ok {
		asset.Meta, err = mgr.getMeta(ctx, dli.Asset.ID, dli.Version)
		if err != nil {
			return fmt.Errorf("could not get meta: %w", err)
		}
	}

	v, err := mgr.getMetaVersion(dli, dir)
	if err != nil {
		return fmt.Errorf("could not get meta version: %w", err)
	}
	// swap version
	asset.Versions = []Version{v}
	mf := MetaFile{
		Meta: asset.Meta,
	}
	mfp := filepath.Join(dir, "meta")
	f, err := os.Create(mfp)
	if err != nil {
		return fmt.Errorf("could not create meta file %s: %w", mfp, err)
	}
	defer f.Close()
	if err = json.NewEncoder(f).Encode(&mf); err != nil {
		return fmt.Errorf("could not encode meta file %s: %w", mfp, err)
	}

	asset.Dir = dir
	mgr.mux.Lock()
	mgr.installed[dli.Asset.ID] = asset
	mgr.mux.Unlock()

	return nil
}

func (mgr *Manager) installFile(ctx context.Context, dlf *workshop.File, m *workshop.Manifest, tmpDir, dir string, onProgress func(float64)) error {
	tfp := filepath.Join(tmpDir, dlf.Name)

	tf, err := os.Create(tfp)
	if err != nil {
		return err
	}
	defer tf.Close()

	pf := progress.NewFile(dlf.Size, tf, onProgress)

	if err = mgr.ws.DownloadManifest(ctx, m, pf); err != nil {
		return err
	}

	if err = workshop.ValidateFile(m, tf); err != nil {
		return err
	}

	nfp := filepath.Join(dir, dlf.Name)
	return os.Rename(tfp, nfp)
}

func (mgr *Manager) syncManifestFile(f *workshop.File, m *workshop.Manifest, version, dir string) error {
	emfs, err := filepath.Glob(filepath.Join(dir, fmt.Sprintf("%s_*_manifest.json", f.Name)))
	if err != nil {
		fmt.Printf("could not find manifest files in %s: %v\n", dir, err)
		return err
	}
	mfp := filepath.Join(dir, getManifestFileName(f.Name, version))
	mf, err := os.Create(mfp)
	if err != nil {
		return err
	}
	defer mf.Close()

	if err = json.NewEncoder(mf).Encode(m); err != nil {
		return err
	}
	for _, ef := range emfs {
		os.Remove(ef)
	}
	return nil
}

//func (m *Manager) installMod(ctx context.Context, mod game.Mod) error {
//	dir := filepath.Join(m.addonsDir, fmt.Sprintf("%s_%s", mod.Name, mod.ModID))
//	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
//		return err
//	}
//	tmpDir := filepath.Join(dir, "temp")
//	if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
//		return err
//	}
//	downloadList := m.ws.GetAssetsDownloadList(ctx)
//}

func (mgr *Manager) LoadInstalledAssets() error {
	mfps, err := filepath.Glob(filepath.Join(mgr.addonsDir, "*", "meta"))
	if err != nil {
		return fmt.Errorf("could not glob addons dir %s: %w", mgr.addonsDir, err)
	}

	var eg errgroup.Group
	eg.SetLimit(10)
	for _, mfp := range mfps {
		eg.Go(func() error {
			var ferr error

			fi, ferr := os.Stat(mfp)
			if ferr != nil || fi.IsDir() {
				return err
			}

			f, ferr := os.Open(mfp)
			if ferr != nil {
				return err
			}
			defer f.Close()

			b, ferr := io.ReadAll(f)
			if ferr != nil {
				return err
			}
			b = bytes.TrimPrefix(b, []byte("\xef\xbb\xbf"))

			var mf MetaFile
			if ferr = json.Unmarshal(b, &mf); err != nil {
				return err
			}
			mgr.mux.Lock()
			defer mgr.mux.Unlock()

			mgr.installed[mf.Meta.ID] = Asset{
				Meta: mf.Meta,
				Dir:  filepath.Dir(mfp),
			}

			return nil
		})
	}

	return eg.Wait()
}
