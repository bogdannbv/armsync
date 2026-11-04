package workshop

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"unicode"

	"github.com/bogdannbv/armsync/internal/utils"
	"github.com/bogdannbv/armsync/pkg/reforger/api"
	"github.com/bogdannbv/armsync/pkg/reforger/auth"
	"github.com/bogdannbv/armsync/pkg/reforger/constants"
	"github.com/bogdannbv/armsync/pkg/reforger/httpx"
	"golang.org/x/sync/errgroup"
)

const (
	workshopBaseUrl = "https://api-ar-workshop.bistudio.com/workshop-api/api/v3.0"
	cdnBaseUrl      = "https://ar-gcp-cdn.bistudio.com"
)

type Client struct {
	ws  *httpx.Client
	cdn *httpx.Client
}

type ClientOpt func(*Client)

func New(tp auth.AccessTokenProvider, opts ...ClientOpt) *Client {
	c := &Client{
		ws: httpx.NewClient(
			httpx.WithBaseURL(workshopBaseUrl),
			httpx.WithHeader("User-Agent", constants.UserAgent),
			httpx.WithTransport(auth.NewBearerTokenTransport(http.DefaultTransport, tp)),
		),
		cdn: httpx.NewClient(httpx.WithBaseURL(cdnBaseUrl)),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) GetAssetsDifferences(ctx context.Context, assDiffs map[string]AssetDiff) (map[string]AssetDiff, error) {
	var mux sync.RWMutex
	dih := make(map[string]AssetDiff, len(assDiffs))
	eg, ectx := errgroup.WithContext(ctx)
	// TODO: make this limit configurable
	eg.SetLimit(3)
	for chunk := range slices.Chunk(utils.ToSlice(assDiffs), 50) {
		eg.Go(func() error {
			ad, err := c.getAssetsDifferences(ectx, chunk)
			if err != nil {
				return fmt.Errorf("get assets differences: %w", err)
			}

			mux.Lock()
			defer mux.Unlock()
			for _, d := range ad {
				dih[d.ID] = d
			}

			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return dih, nil
}

func (c *Client) getAssetsDifferences(ctx context.Context, assDiffs []AssetDiff) ([]AssetDiff, error) {
	body := struct {
		Assets []AssetDiff `json:"assets"`
	}{
		Assets: assDiffs,
	}
	b, err := json.Marshal(&body)
	if err != nil {
		return nil, err
	}
	resp, err := c.ws.Post(ctx, "/assets/difference/size", httpx.MIMEApplicationJSON, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, api.GetError(resp.Body)
	}

	var rbody struct {
		Assets []AssetDiff `json:"assets"`
	}
	err = json.NewDecoder(resp.Body).Decode(&rbody)
	if err != nil {
		return nil, err
	}

	return rbody.Assets, nil
}

func (c *Client) GetAssetScenarios(ctx context.Context, id string) ([]Scenario, error) {
	resp, err := c.ws.Get(ctx, path.Join("assets", id, "scenarios"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, api.GetError(resp.Body)
	}

	s := make([]Scenario, 0)
	if err = json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, err
	}
	return s, nil
}

func (c *Client) GetAssetVersionDependencies(ctx context.Context, id, version string) ([]Dependency, error) {
	resp, err := c.ws.Get(ctx, path.Join("assets", id, "versions", version, "dependencies"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, api.GetError(resp.Body)
	}

	deps := make([]Dependency, 0)
	if err = json.NewDecoder(resp.Body).Decode(&deps); err != nil {
		return nil, err
	}
	return deps, nil
}

func (c *Client) GetAssetVersionChangelog(ctx context.Context, id, version string) (Changelog, error) {
	resp, err := c.ws.Get(ctx, path.Join("assets", id, "changelog?="+version))
	if err != nil {
		return Changelog{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Changelog{}, api.GetError(resp.Body)
	}

	var clog Changelog
	if err = json.NewDecoder(resp.Body).Decode(&clog); err != nil {
		return Changelog{}, err
	}
	return clog, nil
}

func (c *Client) GetAsset(ctx context.Context, id string) (Asset, error) {
	resp, err := c.ws.Get(ctx, path.Join("assets", id))
	if err != nil {
		return Asset{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Asset{}, api.GetError(resp.Body)
	}

	var asset Asset
	if err = json.NewDecoder(resp.Body).Decode(&asset); err != nil {
		return Asset{}, err
	}
	return asset, nil
}

func (c *Client) GetAssetsDownloadList(ctx context.Context, assetsVersions map[string]string) ([]DownloadItem, error) {
	body := struct {
		Assets map[string]string `json:"assets"`
	}{
		Assets: assetsVersions,
	}
	b, err := json.Marshal(&body)
	if err != nil {
		return nil, err
	}
	resp, err := c.ws.Post(ctx, "/assets/download-list", httpx.MIMEApplicationJSON, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, api.GetError(resp.Body)
	}

	di := make([]DownloadItem, 0)
	err = json.NewDecoder(resp.Body).Decode(&di)
	if err != nil {
		return nil, err
	}

	return di, nil
}

func (c *Client) GetManifest(ctx context.Context, mp string) (Manifest, error) {
	resp, err := c.cdn.Get(ctx, path.Join("manifest", mp))
	if err != nil {
		return Manifest{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Manifest{}, api.GetError(resp.Body)
	}

	var manifest Manifest
	err = json.NewDecoder(resp.Body).Decode(&manifest)
	if err != nil {
		return Manifest{}, err
	}

	return manifest, nil
}

func (c *Client) DownloadFile(ctx context.Context, f *File, dir string) error {
	m, err := c.GetManifest(ctx, f.ManifestPath)
	if err != nil {
		return fmt.Errorf("could not get file manifest: %w", err)
	}

	fp := filepath.Join(dir, f.Name)
	tfp := fp + ".incomplete"
	of, err := os.Create(tfp)
	if err != nil {
		return err
	}
	defer of.Close()

	if err = c.DownloadManifest(ctx, &m, of); err != nil {
		return err
	}
	if err = ValidateFile(&m, of); err != nil {
		return err
	}

	return os.Rename(tfp, fp)
}

func (c *Client) DownloadManifest(ctx context.Context, m *Manifest, wa io.WriterAt) error {
	frags := make([]frag, 0)
	for _, fg := range m.Fragments {
		for _, off := range fg.Offsets {
			frags = append(frags, frag{
				Offset: off,
				Size:   fg.Size,
				Sha512: fg.Sha512,
			})
		}
	}

	// sort fragments by offset to make the download sequential
	slices.SortFunc(frags, func(a, b frag) int { return a.Offset - b.Offset })

	// to avoid downloading fragments that are duplicated multiple times
	fragsHash := make(map[string][]int, len(frags))
	for i, fg := range frags {
		fragsHash[fg.Sha512] = append(fragsHash[fg.Sha512], i)
	}

	ols := findFragOverlaps(frags)
	if len(ols) != 0 {
		return fmt.Errorf("overlapping fragments: %v", ols)
	}

	ow := io.NewOffsetWriter(wa, 0)

	eg, ectx := errgroup.WithContext(ctx)
	// TODO: make this configurable
	eg.SetLimit(5)
	var n int
	for sha, idxs := range fragsHash {
		eg.Go(func() error {
			rsp, err := c.cdn.Get(ectx, sha512ToPath(sha, frags[idxs[0]].Size))
			if err != nil {
				return fmt.Errorf("could not download fragment %s: %w", sha, err)
			}
			buf, err := io.ReadAll(rsp.Body)
			rsp.Body.Close()
			if err != nil {
				return fmt.Errorf("could not read body %s: %w", sha, err)
			}

			for _, idx := range idxs {
				n, err = ow.WriteAt(buf, int64(frags[idx].Offset))
				if err != nil {
					return fmt.Errorf("could not write fragment %s: %w", sha, err)
				}
				if n != len(buf) {
					return fmt.Errorf("could not write fragment %s, expected %d, wrote %d", sha, frags[idx].Size, n)
				}
			}

			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}

	if m.Remainder == nil || m.Remainder.Size <= 0 {
		return nil
	}

	r := m.Remainder
	// find gaps to fill with data from remainder
	pos := 0
	gaps := make([]frag, 0)
	gapsSz := 0
	for _, fg := range frags {
		if fg.Offset > pos {
			gf := frag{
				Offset: pos,
				Size:   fg.Offset - pos,
			}
			gapsSz += gf.Size
			gaps = append(gaps, gf)
		}
		pos = fg.Offset + fg.Size
	}

	if gapsSz > m.Remainder.Size {
		return fmt.Errorf("remainder size smaller than gaps size, expected %d >= %d", gapsSz, m.Remainder.Size)
	}

	// download remainder
	rsp, err := c.cdn.Get(ctx, sha512ToPath(r.Sha512, r.Size))
	if err != nil {
		return fmt.Errorf("could not download remainder: %w", err)
	}
	defer rsp.Body.Close()

	for _, g := range gaps {
		buf := make([]byte, g.Size)
		n, err = io.ReadAtLeast(rsp.Body, buf, g.Size)
		if err != nil {
			return fmt.Errorf("could not download remainder: %w", err)
		}
		if n != g.Size {
			return fmt.Errorf("could not download remainder, expected %d, got %d", g.Size, n)
		}
		n, err = ow.WriteAt(buf, int64(g.Offset))
		if err != nil {
			return fmt.Errorf("could not write remainder: %w", err)
		}
		if n != g.Size {
			return fmt.Errorf("could not write remainder, expected %d, wrote %d", g.Size, n)
		}
	}

	lastOfRem := r.Size - gapsSz
	buf := make([]byte, lastOfRem)
	n, err = io.ReadFull(rsp.Body, buf)
	if err != nil {
		return fmt.Errorf("could not read last of remainder: %w", err)
	}
	if n != lastOfRem {
		return fmt.Errorf("could not read last of remainder, expected: %d, got %d", lastOfRem, n)
	}

	n, err = ow.WriteAt(buf, int64(m.Size-lastOfRem))
	if err != nil {
		return fmt.Errorf("could not write last of remainder: %w", err)
	}
	if n != lastOfRem {
		return fmt.Errorf("could not write last of remainder, expected %d, wrote %d", lastOfRem, n)
	}

	return nil
}

func ValidateFile(m *Manifest, f io.Reader) error {
	h := sha512.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	if sum := hex.EncodeToString(h.Sum(nil)); sum != m.Sha512 {
		return fmt.Errorf("could not validate sha512 sum, expected: %s, got %s", m.Sha512, sum)
	}
	return nil
}

func sha512ToPath(sha512 string, sz int) string {
	return path.Join(
		"fragment",
		sha512[0:4],
		sha512[4:8],
		sha512[8:],
		fmt.Sprintf("%d", sz),
	)
}

func findFragOverlaps(frags []frag) [][2]int {
	var overlaps [][2]int

	for i := 0; i < len(frags); i++ {
		for j := i + 1; j < len(frags); j++ {
			if frags[j].Offset >= frags[i].Offset+frags[i].Size {
				break
			}
			overlaps = append(overlaps, [2]int{i, j})
		}
	}

	return overlaps
}

func GetGameAssetDirName(ass *Asset) string {
	name := replaceSpaces(ass.Name, "")
	return fmt.Sprintf("%s_%s", name, ass.ID)
}

func GetServerAssetDirName(ass *Asset) string {
	name := replaceSpaces(ass.Name, "_")
	return fmt.Sprintf("%s_%s", name, ass.ID)
}

func replaceSpaces(str, r string) string {
	var b strings.Builder
	b.Grow(len(str))
	for _, ch := range str {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
			continue
		}
		b.WriteString(r)
	}
	return b.String()
}
