package auth

import (
	"net/http"
)

type bearerTokenTransport struct {
	rtp http.RoundTripper
	tp  AccessTokenProvider
}

func NewBearerTokenTransport(tsp http.RoundTripper, tp AccessTokenProvider) http.RoundTripper {
	return &bearerTokenTransport{
		rtp: tsp,
		tp:  tp,
	}
}

func (t *bearerTokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	at, err := t.tp.GetAccessToken(req.Context())
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+at)
	return t.rtp.RoundTrip(req)
}
