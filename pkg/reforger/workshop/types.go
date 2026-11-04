package workshop

import "time"

type DownloadItem struct {
	ID                         int        `json:"id"`
	Blocked                    bool       `json:"blocked"`
	Private                    bool       `json:"private"`
	GameVersion                string     `json:"gameVersion"`
	PlatformCompatibility      string     `json:"platformCompatibility"`
	PlatformCompatibilityValue int        `json:"platformCompatibilityValue"`
	Accessible                 bool       `json:"accessible"`
	Version                    string     `json:"version"`
	TotalFileSize              int        `json:"totalFileSize"`
	Deleted                    bool       `json:"deleted"`
	AccessValue                int        `json:"accessValue"`
	Files                      []File     `json:"files"`
	Scenarios                  []Scenario `json:"scenarios"`
	Asset                      Asset      `json:"asset"`
	//DeletedAt                  any    `json:"deletedAt"`
	//Dependencies               []any  `json:"dependencies"`
}

type AssetDiff struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Version          string            `json:"version"`
	TargetVersion    string            `json:"targetVersion"`
	TotalSize        int               `json:"totalSize,omitempty"`
	TotalSizeDiff    int               `json:"totalSizeDiff,omitempty"`
	ChangedManifests []ChangedManifest `json:"changedManifests,omitempty"`
}

type ChangedManifest struct {
	Filename     string `json:"filename"`
	ManifestPath string `json:"manifestPath"`
	SizeDiff     int    `json:"sizeDiff"`
}

type Asset struct {
	ID                        string         `json:"id"`
	Name                      string         `json:"name"`
	Type                      string         `json:"type"`
	Summary                   string         `json:"summary"`
	Description               string         `json:"description"`
	License                   string         `json:"license"`
	LicenseText               *string        `json:"licenseText"`
	Owned                     bool           `json:"owned"`
	Unlisted                  bool           `json:"unlisted"`
	Private                   bool           `json:"private"`
	Blocked                   bool           `json:"blocked"`
	DmcaReported              bool           `json:"dmcaReported"`
	RatingCount               int            `json:"ratingCount"`
	AverageRating             float64        `json:"averageRating"`
	SubscriberCount           int            `json:"subscriberCount"`
	CurrentVersionNumber      string         `json:"currentVersionNumber"`
	CurrentVersionID          int            `json:"currentVersionId"`
	CurrentVersionSize        int            `json:"currentVersionSize"`
	Previews                  []Image        `json:"previews"`
	Screenshots               []Image        `json:"screenshots"`
	Author                    Author         `json:"author"`
	Tags                      []Tag          `json:"tags"`
	AuthorID                  string         `json:"authorId"`
	DependencyTree            DependencyTree `json:"dependencyTree"`
	Reported                  bool           `json:"reported"`
	Ratings                   Ratings        `json:"ratings"`
	EnforceVersionRule        int            `json:"enforceVersionRule"`
	LatestObsolete            int            `json:"latest_obsolete"`
	LatestPlatformCompatible  int            `json:"latest_platformCompatible"`
	CurrentPlatformCompatible int            `json:"current_platformCompatible"`
	Updatable                 bool           `json:"updatable"`
	GameVersion               string         `json:"gameVersion"`
	Obsolete                  bool           `json:"obsolete"`
	Available                 bool           `json:"available"`
	Versions                  []Version      `json:"versions"`
	Scenarios                 []Scenario     `json:"scenarios"`
	CreatedAt                 string         `json:"createdAt"`
	UpdatedAt                 string         `json:"updatedAt"`
	DeletedAt                 *string        `json:"deletedAt"`
	//Dependencies              []Dependency   `json:"dependencies"`
	//Contributors              []interface{}  `json:"contributors"`
	//Meta                      any            `json:"meta"`
}

type Scenario struct {
	Name        string `json:"name"`
	Image       Image  `json:"image"`
	GameID      string `json:"gameId"`
	GameMode    string `json:"gameMode"`
	AuthorName  string `json:"authorName"`
	Description string `json:"description"`
	PlayerCount int    `json:"playerCount"`
}

type Dependency struct {
	//Dependencies  []interface{} `json:"dependencies"` // Probably a Dependency slice as well
	Version       string `json:"version"`
	Blocked       bool   `json:"blocked"`
	Private       bool   `json:"private"`
	Published     bool   `json:"published"`
	TotalFileSize int64  `json:"totalFileSize"`
	Asset         struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"asset"`
}

type Changelog struct {
	ID          string `json:"id"`
	AssetID     string `json:"assetId"`
	Changelog   string `json:"changelog"`
	Version     string `json:"version"`
	GameVersion string `json:"gameVersion"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type Image struct {
	URL         string      `json:"url"`
	Width       int         `json:"width"`
	Height      int         `json:"height"`
	Thumbnails  *Thumbnails `json:"thumbnails,omitempty"`
	ContentType string      `json:"contentType"`
}

type Thumbnails struct {
	ImageJpeg []Image `json:"image/jpeg"`
}

type Tag struct {
	Name     string      `json:"name"`
	Category interface{} `json:"category,omitempty"`
}

type Ratings struct {
	Likes    int     `json:"likes"`
	Dislikes int     `json:"dislikes"`
	Rating   float64 `json:"rating"`
}

type Author struct {
	Roles           interface{} `json:"roles,omitempty"`
	ID              string      `json:"id"`
	Username        string      `json:"username"`
	PersonalBlocked bool        `json:"personalBlocked"`
}

type Version struct {
	Version               string      `json:"version"`
	Approved              bool        `json:"approved"`
	Published             bool        `json:"published"`
	GameVersion           string      `json:"gameVersion"`
	MinGameVersion        int         `json:"minGameVersion"`
	PlatformCompatibility int         `json:"platformCompatibility"`
	TotalFileSize         int         `json:"totalFileSize"`
	Meta                  interface{} `json:"meta"`
	Milestone             bool        `json:"milestone"`
	CreatedAt             time.Time   `json:"createdAt"`
	UpdatedAt             time.Time   `json:"updatedAt"`
	AssetID               string      `json:"assetId"`
	ScenariosCount        int         `json:"scenariosCount"`
	DependenciesCount     int         `json:"dependenciesCount"`
}

type DependencyTree struct {
	Access                     string `json:"access"`
	Blocked                    bool   `json:"blocked"`
	Private                    bool   `json:"private"`
	Unlisted                   bool   `json:"unlisted"`
	AccessValue                int    `json:"accessValue"`
	GameVersion                string `json:"gameVersion"`
	PlatformCompatibility      string `json:"platformCompatibility"`
	PlatformCompatibilityValue int    `json:"platformCompatibilityValue"`
	//DeletedAt                  any `json:"deletedAt"`
}

type File struct {
	Name         string `json:"name"`
	Size         int    `json:"size"`
	ContentType  string `json:"contentType"`
	ManifestPath string `json:"manifestPath"`
}

type Manifest struct {
	Version   int        `json:"version"`
	Size      int        `json:"size"`
	Sha512    string     `json:"sha512"`
	Remainder *Fragment  `json:"remainder"`
	Fragments []Fragment `json:"fragments"`
}

type Fragment struct {
	Sha512  string `json:"sha512"`
	Size    int    `json:"size"`
	Offsets []int  `json:"offsets"`
}

type frag struct {
	Offset int
	Size   int
	Sha512 string
}
