//go:build wailsdesktop

package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type App struct {
	wailsApp   *application.App
	restartMu  sync.Mutex
	restarting bool
}

func NewApp() *App {
	return &App{}
}

func (a *App) Attach(wailsApp *application.App) {
	a.wailsApp = wailsApp
}

// RestartApp relaunches the current executable and quits the current process.
func (a *App) RestartApp() error {
	a.restartMu.Lock()
	if a.restarting {
		a.restartMu.Unlock()
		return nil
	}
	a.restarting = true
	a.restartMu.Unlock()

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}

	cmd := exec.Command(exePath, os.Args[1:]...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if wd, wdErr := os.Getwd(); wdErr == nil {
		cmd.Dir = wd
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start new app process: %w", err)
	}

	if a.wailsApp != nil {
		a.wailsApp.Quit()
	}
	return nil
}
