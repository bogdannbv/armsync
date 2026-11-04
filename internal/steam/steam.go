package steam

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/bogdannbv/armsync/pkg/reforger/constants"
	"github.com/bogdannbv/go-steamworks"
)

var initialized atomic.Bool

func Init(path string) error {
	if err := ValidateSteamAppIDPresence(path); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get cwd: %v", err)
	}
	defer os.Chdir(cwd)

	err = os.Chdir(path)
	if err != nil {
		return fmt.Errorf("unable to chdir to %s: %w", path, err)
	}

	var serr error
	err = suppressStderr(func() {
		if serr = steamworks.Load(); serr != nil {
			return
		}
		serr = steamworks.Init()
	})
	if err != nil {
		panic(fmt.Errorf("unable to suppress stderr: %w", err))
	}
	if serr == nil {
		initialized.Store(true)
		return nil
	}

	return fmt.Errorf("failed to initialize steamworks: %w", serr)
}

func Initialized() bool {
	return initialized.Load()
}

func GetAuthSessionTicket() (string, error) {
	if !Initialized() {
		return "", errors.New("not initialized")
	}

	token := make([]byte, 2048) // according to docs, 1024 should be fine but to be safe double it and give it to the next person
	h, l := steamworks.SteamUser().GetAuthSessionTicket(token, nil)
	if h == 0 || l == 0 {
		return "", errors.New("failed to get auth session ticket")
	}

	return base64.StdEncoding.EncodeToString(token[:l]), nil
}

func Shutdown() {
	steamworks.Shutdown()
}

func ValidateSteamAppIDPresence(path string) error {
	path = filepath.Join(path, "steam_appid.txt")

	i, err := os.Stat(path)
	if err != nil {
		return err
	}
	if i.IsDir() {
		return fmt.Errorf("%s is a directory", path)
	}

	c, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if string(c) != constants.AppID {
		return fmt.Errorf("wrong app id, expected %s, got %s", constants.AppID, string(c))
	}
	return nil
}
