package steam

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

const childEnv = "ARMSYNC_STEAM_TICKET_CHILD"

func GetAuthSessionTicketDetached(gameDir string) (string, func(), error) {
	exe, err := os.Executable()
	if err != nil {
		return "", nil, err
	}
	cmd := exec.Command(exe)
	cmd.Env = append(os.Environ(), childEnv+"="+gameDir)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", nil, err
	}
	if err = cmd.Start(); err != nil {
		return "", nil, err
	}
	release := func() {
		stdin.Close()
		cmd.Wait()
	}
	line, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil {
		release()
		return "", nil, fmt.Errorf("steam ticket child process failed: %w", err)
	}
	return strings.TrimSpace(line), release, nil
}

func RunTicketChild() {
	gd := os.Getenv(childEnv)
	if gd == "" {
		return
	}
	if err := Init(gd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ticket, err := GetAuthSessionTicket()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(ticket)
	io.Copy(io.Discard, os.Stdin)
	Shutdown()
	os.Exit(0)
}
