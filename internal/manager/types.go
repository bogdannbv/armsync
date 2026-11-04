package manager

import "github.com/bogdannbv/armsync/pkg/reforger/workshop"

type Asset struct {
	Meta
	Dir string
}

type Meta struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Type              int       `json:"type"`
	CreatedAt         string    `json:"createdAt"`
	UpdatedAt         string    `json:"updatedAt"`
	Versions          []Version `json:"versions"`
	SelectedRev       int       `json:"selectedRev"`
	Previews          []Image   `json:"previews"`
	AverageRating     float64   `json:"averageRating"`
	SubscriberCount   int       `json:"subscriberCount"`
	RatingCount       int       `json:"ratingCount"`
	Summary           string    `json:"summary"`
	License           string    `json:"license"`
	LicenseText       *string   `json:"licenseText"`
	Access            bool      `json:"access"`
	Unlisted          bool      `json:"unlisted"`
	Author            Author    `json:"author"`
	Tags              []Tag     `json:"tags"`
	Screenshots       []Image   `json:"screenshots"`
	Deleted           bool      `json:"deleted"`
	TimeLastPlay      uint64    `json:"timeLastPlay"`
	TimeFirstDownload uint64    `json:"timeFirstDownload"`
}

type Thumbnails struct {
	ImageJpeg []Image `json:"image/jpeg"`
}

type Image struct {
	URL         string      `json:"url"`
	Width       int         `json:"width"`
	Height      int         `json:"height"`
	Thumbnails  *Thumbnails `json:"thumbnails,omitempty"`
	ContentType string      `json:"contentType"`
	FilePath    string      `json:"filePath"`
	Checksum    string      `json:"checksum"`
	URLHash     string      `json:"urlHash"`
}

type Author = workshop.Author

type Tag = workshop.Tag

type Dependency struct {
	AssetID   string `json:"assetId"`
	AssetName string `json:"assetName"`
	Version   string `json:"version"`
}

type Scenario struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	AuthorName  string `json:"authorName"`
	OwnerID     string `json:"ownerId"`
	Image       Image  `json:"image"`
	GameID      string `json:"gameId"`
	PlayerCount int    `json:"playerCount"`
}

type Version struct {
	Version      string       `json:"version"`
	GameVersion  string       `json:"gameVersion"`
	Changelog    string       `json:"changelog"`
	Corrupted    bool         `json:"corrupted"`
	Dependencies []Dependency `json:"dependencies"`
	Scenarios    []Scenario   `json:"scenarios"`
	Env          int          `json:"env"`
	Package      Package      `json:"package"`
}

type Package struct {
	TotalSize int    `json:"totalSize"`
	Files     []File `json:"files"`
}

type File struct {
	IsCorrupted    bool   `json:"isCorrupted"`
	IsDownloaded   bool   `json:"isDownloaded"`
	Size           int    `json:"size"`
	DownloadedSize int    `json:"downloadedSize"`
	FilePath       string `json:"filePath"`
	UpdatedAt      string `json:"updatedAt"`
	Name           string `json:"name"`
}

type MetaFile struct {
	Meta Meta `json:"meta"`
}
