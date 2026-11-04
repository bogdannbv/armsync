package manager

import (
	"context"
	"fmt"
	"time"

	"github.com/bogdannbv/armsync/pkg/reforger/workshop"
)

const (
	MetaTypeAddon = 1
)

func (mgr *Manager) getMeta(ctx context.Context, id string, ver string) (Meta, error) {
	a, err := mgr.ws.GetAsset(ctx, id)
	if err != nil {
		return Meta{}, fmt.Errorf("could not get asset %s: %w", id, err)
	}
	var wv *workshop.Version
	for _, av := range a.Versions {
		if av.Version == ver {
			wv = &av
			break
		}
	}
	if wv == nil {
		return Meta{}, fmt.Errorf("could not find version in asset %s", ver)
	}

	typ := 0
	switch a.Type {
	case "addon":
		typ = MetaTypeAddon
	}

	m := Meta{
		ID:                a.ID,
		Name:              a.Name,
		Description:       a.Description,
		Type:              typ,
		CreatedAt:         a.CreatedAt,
		UpdatedAt:         a.UpdatedAt,
		Versions:          make([]Version, 0, 1),
		SelectedRev:       0,
		Previews:          []Image{},
		AverageRating:     a.AverageRating,
		SubscriberCount:   a.SubscriberCount,
		RatingCount:       a.RatingCount,
		Summary:           a.Summary,
		License:           a.License,
		LicenseText:       a.LicenseText,
		Access:            false,
		Unlisted:          false,
		Author:            a.Author,
		Tags:              a.Tags,
		Screenshots:       []Image{},
		Deleted:           false,
		TimeLastPlay:      0,
		TimeFirstDownload: 0,
	}

	return m, nil
}

func (mgr *Manager) getMetaVersion(dli *workshop.DownloadItem, dir string) (Version, error) {
	//var clog workshop.Changelog
	//var deps []workshop.Dependency
	//var scros []workshop.Scenario

	//eg, ectx := errgroup.WithContext(ctx)
	//eg.SetLimit(3)
	//eg.Go(func() (err error) {
	//	clog, err = mgr.ws.GetAssetVersionChangelog(ectx, dli.Asset.ID, dli.Version)
	//	if err != nil {
	//		log.Warn("could not get version changelog", "err", err)
	//	}
	//	return nil
	//})
	//eg.Go(func() (err error) {
	//	deps, err = mgr.ws.GetAssetVersionDependencies(ectx, dli.Asset.ID, dli.Version)
	//	if err != nil {
	//		log.Warn("could not get version dependencies", "err", err)
	//	}
	//	return nil
	//})
	//eg.Go(func() (err error) {
	//	scros, err = mgr.ws.GetAssetScenarios(ectx, dli.Asset.ID)
	//	if err != nil {
	//		log.Warn("could not get scenarios", "err", err)
	//	}
	//	return nil
	//})
	//if err := eg.Wait(); err != nil {
	//	return Version{}, err
	//}
	//
	//d := make([]Dependency, 0, len(deps))
	//for _, dep := range deps {
	//	d = append(d, Dependency{
	//		AssetID:   dep.Asset.ID,
	//		AssetName: dep.Asset.Name,
	//		Version:   dep.Version,
	//	})
	//}
	//
	//s := make([]Scenario, 0, len(scros))
	//for _, sc := range scros {
	//	s = append(s, Scenario{
	//		Name:        sc.Name,
	//		Description: sc.Description,
	//		AuthorName:  sc.AuthorName,
	//		OwnerID:     dli.Asset.ID,
	//		Image:       Image{},
	//		GameID:      sc.GameID,
	//		PlayerCount: sc.PlayerCount,
	//	})
	//}

	pfs := make([]File, 0, len(dli.Files))
	for _, f := range dli.Files {
		pfs = append(pfs, File{
			IsCorrupted:    false,
			IsDownloaded:   true,
			Size:           f.Size,
			DownloadedSize: f.Size,
			FilePath:       getPackageFilePath(dir, f.Name),
			UpdatedAt:      time.Now().Format(time.RFC3339),
			Name:           f.Name,
		})
	}

	return Version{
		Version:      dli.Version,
		GameVersion:  "", // not being set
		Changelog:    "", // not needed with latest 1.7 and 1.8 updates
		Corrupted:    false,
		Dependencies: nil, // not needed with latest 1.7 and 1.8 updates
		Scenarios:    nil, // not needed with latest 1.7 and 1.8 updates
		Env:          7,   // TODO: find out what does this mean??
		Package: Package{
			TotalSize: dli.TotalFileSize,
			Files:     pfs,
		},
	}, nil
}
