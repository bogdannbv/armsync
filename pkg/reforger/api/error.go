package api

import (
	"encoding/json"
	"fmt"
	"io"
)

const (
	ErrCodeUnauthorizedToken    = "UnauthorizedToken"
	ErrCodeTokenExpired         = "TokenExpired"
	ErrCodeUnknownIdentity      = "UnknownIdentity"
	ErrCodeSteamConnectionError = "SteamConnectionError"
)

type Error struct {
	code    string
	message string
}

type errResponseBody struct {
	APICode string `json:"apiCode"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return fmt.Sprintf("%s (%s)", e.code, e.message)
}

func GetError(body io.Reader) error {
	errBody := errResponseBody{}
	if err := json.NewDecoder(body).Decode(&errBody); err != nil {
		return err
	}

	return Error{
		code:    errBody.APICode,
		message: errBody.Message,
	}
}
