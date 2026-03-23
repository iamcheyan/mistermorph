//go:build wailsdesktop

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const desktopLinuxWebviewGPUEnv = "MISTERMORPH_DESKTOP_WEBVIEW_GPU_POLICY"

func main() {
	cfgPath, explicit := resolveDesktopConfigPath(os.Args[1:])
	printDesktopConfigPath("desktop app", cfgPath, explicit)

	if handled, err := maybeRunDesktopConsoleServe(os.Args[1:]); handled {
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "desktop console host failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	host := NewDesktopHost(DesktopHostConfig{
		ConsoleBasePath: defaultConsoleBasePath,
		ConfigPath:      cfgPath,
	})
	if err := host.Start(context.Background()); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to start desktop host: %v\n", err)
		os.Exit(1)
	}
	defer host.Stop()

	appBinding := NewApp()
	app := application.New(buildDesktopAppOptions(host, appBinding))
	appBinding.Attach(app)
	app.Window.NewWithOptions(buildDesktopWindowOptions(host.ConsoleURL()))

	err := app.Run()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "desktop app exited with error: %v\n", err)
		os.Exit(1)
	}
}

func buildDesktopAppOptions(host *DesktopHost, appBinding *App) application.Options {
	return application.Options{
		Name:        "MisterMorph",
		Description: "MisterMorph Desktop",
		Icon:        desktopAppIconPNG,
		Linux: application.LinuxOptions{
			ProgramName: "MisterMorph",
		},
		Assets: application.AssetOptions{
			// Linux custom-scheme requests can lose JSON bodies; load the console over
			// the local HTTP host instead of proxying the UI through the asset handler.
			Handler: http.NotFoundHandler(),
		},
		OnShutdown: host.Stop,
		Services: []application.Service{
			application.NewService(appBinding),
		},
	}
}

func buildDesktopWindowOptions(consoleURL string) application.WebviewWindowOptions {
	return application.WebviewWindowOptions{
		Title:     "MisterMorph",
		Width:     1360,
		Height:    860,
		MinWidth:  1000,
		MinHeight: 680,
		URL:       consoleURL,
		Linux: application.LinuxWindow{
			WebviewGpuPolicy: resolveLinuxWebviewGPUPolicy(),
		},
	}
}

func resolveLinuxWebviewGPUPolicy() application.WebviewGpuPolicy {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(desktopLinuxWebviewGPUEnv))) {
	case "", "ondemand", "on_demand", "on-demand":
		return application.WebviewGpuPolicyOnDemand
	case "always":
		return application.WebviewGpuPolicyAlways
	case "never", "off", "disabled":
		return application.WebviewGpuPolicyNever
	default:
		return application.WebviewGpuPolicyOnDemand
	}
}
