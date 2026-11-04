package game

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/bogdannbv/armsync/pkg/reforger/api"
	"github.com/bogdannbv/armsync/pkg/reforger/constants"
	"github.com/bogdannbv/armsync/pkg/reforger/httpx"
)

const (
	platformID = "ReforgerSteam"

	searchClientTypePC          = "PLATFORM_PC"
	searchClientFilterAnyCompat = "AnyCompatible"
)

type Server struct {
	ID                       string   `json:"id"`
	ScenarioID               string   `json:"scenarioId"`
	Name                     string   `json:"name"`
	ScenarioName             string   `json:"scenarioName"`
	GameVersion              string   `json:"gameVersion"`
	HostType                 string   `json:"hostType"`
	Official                 bool     `json:"official"`
	Joinable                 bool     `json:"joinable"`
	Visible                  bool     `json:"visible"`
	PasswordProtected        bool     `json:"passwordProtected"`
	Updated                  int      `json:"updated"`
	DetailsUpdatedAt         int      `json:"detailsUpdatedAt"`
	HostAddress              string   `json:"hostAddress"`
	PlayerCountLimit         int      `json:"playerCountLimit"`
	PlayerCount              int      `json:"playerCount"`
	DirectJoinCode           string   `json:"directJoinCode"`
	SupportedGameClientTypes []string `json:"supportedGameClientTypes"`
	Mods                     []Mod    `json:"mods"`
	Favorite                 bool     `json:"favorite"`
	HostedScenarioModID      string   `json:"hostedScenarioModId"`
	PingSiteID               string   `json:"pingSiteId"`
	PlatformName             string   `json:"platformName"`
	SessionID                string   `json:"sessionId"`
	JoinQueue                Queue    `json:"joinQueue"`
	Flags                    int      `json:"flags"`
	LastJoinedAt             int      `json:"lastJoinedAt"`
	//BattlEye                 uint      `json:"battlEye"` // this is cooked, on /search it's returned as bool but on /getRoomByIds it's int
	//RuntimeStats             struct{} `json:"runtimeStats"`
}

type Queue struct {
	Type                string `json:"type"`
	MaxSize             int    `json:"maxSize"`
	Size                int    `json:"size"`
	PositionAvgWaitTime int    `json:"positionAvgWaitTime"`
}

type Mod struct {
	ModID   string `json:"modId"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

func (c *Client) GetServersByIDs(ctx context.Context, ids []string) ([]Server, error) {
	reqBody := defaultSearchReqBody
	reqBody.Fields = []string{
		"id",
		"playerCount",
		"joinQueue",
		"joinable",
		"updated",
		"detailsUpdatedAt",
	}
	for _, id := range ids {
		reqBody.RoomIDTimestamps = append(reqBody.RoomIDTimestamps, roomIDTimestamp{
			ID:               id,
			UpdatedAt:        time.Now().Unix(),
			DetailsUpdatedAt: time.Now().Unix(),
		})
	}

	return c.searchServers(ctx, "lobby/rooms/getRoomsByIds", &reqBody)
}

func (c *Client) GetFavoriteServers(ctx context.Context) ([]Server, error) {
	reqBody := defaultSearchReqBody
	reqBody.Favorites = true
	return c.searchServers(ctx, "lobby/rooms/search", &reqBody)
}

func (c *Client) SearchServers(ctx context.Context, term string) ([]Server, error) {
	reqBody := defaultSearchReqBody
	reqBody.Text = term

	return c.searchServers(ctx, "lobby/rooms/search", &reqBody)
}

func (c *Client) searchServers(ctx context.Context, path string, rb *searchReqBody) ([]Server, error) {
	t, err := c.tp.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	rb.AccessToken = t

	buf := new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(rb)
	if err != nil {
		return nil, err
	}

	resp, err := c.h.Post(ctx, path, httpx.MIMEApplicationJSON, buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, api.GetError(resp.Body)
	}

	var rbody struct {
		Rooms []Server `json:"rooms"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&rbody); err != nil {
		return nil, err
	}

	return rbody.Rooms, nil
}

var defaultSearchReqBody = searchReqBody{
	ClientVersion:    constants.Version,
	PlatformID:       platformID,
	Order:            "PlayerCount",
	Ascendent:        false,
	GameClientType:   searchClientTypePC,
	GameClientFilter: searchClientFilterAnyCompat,
	Locked:           false,
	Limit:            50,
}

type searchReqBody struct {
	AccessToken      string            `json:"accessToken"`
	ClientVersion    string            `json:"clientVersion"`
	PlatformID       string            `json:"platformId"`
	Order            string            `json:"order"`
	Ascendent        bool              `json:"ascendent"`
	Favorites        bool              `json:"favorites,omitempty"`
	DirectJoinCode   string            `json:"directJoinCode,omitempty"`
	HostAddress      string            `json:"hostAddress,omitempty"`
	IncludePing      int               `json:"includePing,omitempty"`
	ScenarioID       string            `json:"scenarioId,omitempty"`
	Text             string            `json:"text,omitempty"`
	PingLimit        float64           `json:"pingLimit,omitempty"`
	Locked           bool              `json:"locked,omitempty"`
	Modded           bool              `json:"modded,omitempty"`
	BattlEye         bool              `json:"battlEye,omitempty"`
	GameClientFilter string            `json:"gameClientFilter,omitempty"`
	GameClientType   string            `json:"gameClientType,omitempty"`
	Lightweight      bool              `json:"lightweight,omitempty"`
	From             int               `json:"from,omitempty"`
	Limit            int               `json:"limit,omitempty"`
	PingValues       []pingValue       `json:"pingValues"`
	Fields           []string          `json:"fields"`
	RoomIDTimestamps []roomIDTimestamp `json:"roomIdTimestamps"`
}

type pingValue struct {
	PingSiteID string  `json:"pingSiteId"`
	Value      float64 `json:"value"`
}

type roomIDTimestamp struct {
	ID               string `json:"id"`
	UpdatedAt        int64  `json:"uAt"`
	DetailsUpdatedAt int64  `json:"duAt"`
}
