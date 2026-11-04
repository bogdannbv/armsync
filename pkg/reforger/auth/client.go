package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/bogdannbv/armsync/pkg/reforger/api"
	"github.com/bogdannbv/armsync/pkg/reforger/constants"
	"github.com/bogdannbv/armsync/pkg/reforger/httpx"
)

const (
	idBaseUrl     = "https://api-ar-id.bistudio.com/game-identity/api/v1.1"
	platformSteam = "steam"
)

type AccessTokenProvider interface {
	GetAccessToken(context.Context) (string, error)
}

type Client struct {
	h         *httpx.Client
	mux       sync.RWMutex
	at        accessToken
	ticket    string
	delayOnce sync.Once
}

func New(steamTicket string) *Client {
	return &Client{
		h:      httpx.NewClient(httpx.WithBaseURL(idBaseUrl)),
		ticket: steamTicket,
	}
}

func (c *Client) GetAccessToken(ctx context.Context) (string, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if !c.at.expired() {
		return c.at.token, nil
	}

	b, err := json.Marshal(&authReqBody{
		Platform: platformSteam,
		Token:    c.ticket,
		PlatformOpts: struct {
			AppID string `json:"appId"`
		}{
			AppID: constants.AppID,
		},
	})
	if err != nil {
		return "", err
	}
	resp, err := c.h.Post(ctx, "identities/reforger/auth", httpx.MIMEApplicationJSON, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", api.GetError(resp.Body)
	}

	var respBody authRespBody
	if err = json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", err
	}

	c.at = accessToken{
		token:     respBody.AccessToken,
		expiresAt: time.Time(respBody.AccessTokenExp),
	}

	return c.at.token, nil
}

type accessToken struct {
	token     string
	expiresAt time.Time
}

func (a *accessToken) expired() bool {
	return time.Now().After(a.expiresAt)
}

type authReqBody struct {
	Platform     string `json:"platform"`
	Token        string `json:"token"`
	PlatformOpts struct {
		AppID string `json:"appId"`
	} `json:"platformOpts"`
}

type authRespBody struct {
	IdentityID     string   `json:"identityId"`
	AccessToken    string   `json:"accessToken"`
	AccessTokenExp jsonTime `json:"accessTokenExp"`
}

type jsonTime time.Time

func (t *jsonTime) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(time.Time(*t).Unix(), 10)), nil
}

func (t *jsonTime) UnmarshalJSON(s []byte) (err error) {
	q, err := strconv.ParseInt(string(s), 10, 64)

	if err != nil {
		return err
	}
	*t = jsonTime(time.Unix(q, 0))
	return
}

func (t *jsonTime) String() string { return time.Time(*t).String() }
