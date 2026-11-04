package game

import (
	"github.com/bogdannbv/armsync/pkg/reforger/auth"
	"github.com/bogdannbv/armsync/pkg/reforger/constants"
	"github.com/bogdannbv/armsync/pkg/reforger/httpx"
)

const (
	userAgent = "Arma Reforger/" + constants.Version + " (Client; Windows)"
	baseUrl   = "https://api-ar-game.bistudio.com/game-api/api/v1.0"
)

type Client struct {
	tp auth.AccessTokenProvider
	h  *httpx.Client
}

type ClientOpt func(*Client)

func New(tp auth.AccessTokenProvider, opts ...ClientOpt) *Client {
	c := &Client{
		tp: tp,
		h: httpx.NewClient(
			httpx.WithBaseURL(baseUrl),
			httpx.WithHeader("User-Agent", constants.UserAgent),
		),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}
