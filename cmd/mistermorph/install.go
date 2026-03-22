package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/quailyquaily/mistermorph/assets"
	"github.com/quailyquaily/mistermorph/internal/clifmt"
	"github.com/quailyquaily/mistermorph/internal/pathutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newInstallCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "install [dir]",
		Short: "Install config.yaml plus the core onboarding markdown files",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := resolveInstallDir(args)
			if err != nil {
				return err
			}

			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}

			cfgPath := filepath.Join(dir, "config.yaml")
			writeConfig := true
			if _, err := os.Stat(cfgPath); err == nil {
				fmt.Fprintf(os.Stderr, "warn: config already exists, skipping: %s\n", cfgPath)
				writeConfig = false
			}
			var cfgSetup *installConfigSetup
			if writeConfig {
				if source, ok := findReadableInstallConfig(cmd, dir); ok {
					fmt.Printf("[i] found config.yaml, skip interactive setup: %s\n", source)
				} else {
					cfgSetup, err = maybeCollectInstallConfigSetup(cmd, yes)
					if err != nil {
						return err
					}
				}
			}

			identityPath := filepath.Join(dir, "IDENTITY.md")
			writeIdentity := true
			if _, err := os.Stat(identityPath); err == nil {
				writeIdentity = false
			}
			soulPath := filepath.Join(dir, "SOUL.md")
			writeSoul := true
			if _, err := os.Stat(soulPath); err == nil {
				writeSoul = false
			}
			hbPath := filepath.Join(dir, "HEARTBEAT.md")
			writeHeartbeat := true
			if _, err := os.Stat(hbPath); err == nil {
				writeHeartbeat = false
			}
			toolsPath := filepath.Join(dir, "SCRIPTS.md")
			writeTools := true
			if _, err := os.Stat(toolsPath); err == nil {
				writeTools = false
			}
			todoWIPPath := filepath.Join(dir, "TODO.md")
			writeTodoWIP := true
			if _, err := os.Stat(todoWIPPath); err == nil {
				writeTodoWIP = false
			}
			todoDonePath := filepath.Join(dir, "TODO.DONE.md")
			writeTodoDone := true
			if _, err := os.Stat(todoDonePath); err == nil {
				writeTodoDone = false
			}

			initialSteps := []installStep{
				{
					Name:  "config.yaml",
					Path:  cfgPath,
					Write: writeConfig,
					Loader: func() (string, error) {
						body, err := loadConfigExample()
						if err != nil {
							return "", err
						}
						return patchInitConfigWithSetup(body, dir, cfgSetup), nil
					},
				},
				{
					Name:   "IDENTITY.md",
					Path:   identityPath,
					Write:  writeIdentity,
					Loader: loadIdentityTemplate,
				},
				{
					Name:   "SOUL.md",
					Path:   soulPath,
					Write:  writeSoul,
					Loader: loadSoulTemplate,
				},
			}
			fmt.Println(clifmt.Headerf("==> Installing setup flow (%d steps)", len(initialSteps)))
			for i, step := range initialSteps {
				if err := writeInstallStepFile(i+1, len(initialSteps), step); err != nil {
					return err
				}
				if !step.Write {
					continue
				}
				if yes || !supportsInteractivePrompts(cmd) {
					continue
				}
				switch step.Name {
				case "IDENTITY.md":
					if err := runInstallIdentitySetup(cmd.InOrStdin(), cmd.OutOrStdout(), step.Path); err != nil {
						return err
					}
				case "SOUL.md":
					if err := runInstallSoulSetup(cmd.InOrStdin(), cmd.OutOrStdout(), step.Path); err != nil {
						return err
					}
				}
			}

			deferredSteps := []installStep{
				{
					Name:   "HEARTBEAT.md",
					Path:   hbPath,
					Write:  writeHeartbeat,
					Loader: loadHeartbeatTemplate,
				},
				{
					Name:   "SCRIPTS.md",
					Path:   toolsPath,
					Write:  writeTools,
					Loader: loadToolsTemplate,
				},
				{
					Name:   "TODO.md",
					Path:   todoWIPPath,
					Write:  writeTodoWIP,
					Loader: loadTodoWIPTemplate,
				},
				{
					Name:   "TODO.DONE.md",
					Path:   todoDonePath,
					Write:  writeTodoDone,
					Loader: loadTodoDoneTemplate,
				},
			}
			fmt.Println(clifmt.Headerf("==> Installing deferred markdown files (%d files)", len(deferredSteps)))
			for i, step := range deferredSteps {
				if err := writeInstallStepFile(i+1, len(deferredSteps), step); err != nil {
					return err
				}
			}

			fmt.Printf("[i] initialized %s\n", dir)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompts (dangerous)")

	return cmd
}

type installStep struct {
	Name   string
	Path   string
	Write  bool
	Loader func() (string, error)
}

func writeInstallStepFile(index int, total int, step installStep) error {
	fmt.Printf("[%d/%d] %s (1 file) ... ", index, total, step.Name)
	if !step.Write {
		fmt.Printf("%s %s\n", clifmt.Success("done"), clifmt.Warn("(skipped)"))
		fmt.Printf("    %s %s\n", step.Path, clifmt.Warn("(skipped)"))
		return nil
	}
	body, err := step.Loader()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(step.Path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(step.Path, []byte(body), 0o644); err != nil {
		return err
	}
	fmt.Println(clifmt.Success("done"))
	fmt.Printf("    %s\n", step.Path)
	return nil
}

func resolveInstallDir(args []string) (string, error) {
	var dir string
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		dir = strings.TrimSpace(args[0])
	} else {
		dir = strings.TrimSpace(viper.GetString("file_state_dir"))
		if dir == "" {
			dir = "~/.morph"
		}
	}
	dir = pathutil.ExpandHomePath(dir)
	if strings.TrimSpace(dir) == "" {
		return "", fmt.Errorf("invalid dir")
	}
	return filepath.Clean(dir), nil
}

func loadConfigExample() (string, error) {
	data, err := assets.ConfigFS.ReadFile("config/config.example.yaml")
	if err != nil {
		return "", fmt.Errorf("read embedded config.example.yaml: %w", err)
	}
	return string(data), nil
}

func loadHeartbeatTemplate() (string, error) {
	data, err := assets.ConfigFS.ReadFile("config/HEARTBEAT.md")
	if err != nil {
		return "", fmt.Errorf("read embedded HEARTBEAT.md: %w", err)
	}
	return string(data), nil
}

func loadIdentityTemplate() (string, error) {
	data, err := assets.ConfigFS.ReadFile("config/IDENTITY.md")
	if err != nil {
		return "", fmt.Errorf("read embedded IDENTITY.md: %w", err)
	}
	return string(data), nil
}

func loadToolsTemplate() (string, error) {
	data, err := assets.ConfigFS.ReadFile("config/SCRIPTS.md")
	if err != nil {
		return "", fmt.Errorf("read embedded SCRIPTS.md: %w", err)
	}
	return string(data), nil
}

func loadTodoWIPTemplate() (string, error) {
	data, err := assets.ConfigFS.ReadFile("config/TODO.md")
	if err != nil {
		return "", fmt.Errorf("read embedded TODO.md: %w", err)
	}
	return string(data), nil
}

func loadTodoDoneTemplate() (string, error) {
	data, err := assets.ConfigFS.ReadFile("config/TODO.DONE.md")
	if err != nil {
		return "", fmt.Errorf("read embedded TODO.DONE.md: %w", err)
	}
	return string(data), nil
}

func loadContactsActiveTemplate() (string, error) {
	data, err := assets.ConfigFS.ReadFile("config/contacts/ACTIVE.md")
	if err != nil {
		return "", fmt.Errorf("read embedded contacts/ACTIVE.md: %w", err)
	}
	return string(data), nil
}

func loadContactsInactiveTemplate() (string, error) {
	data, err := assets.ConfigFS.ReadFile("config/contacts/INACTIVE.md")
	if err != nil {
		return "", fmt.Errorf("read embedded contacts/INACTIVE.md: %w", err)
	}
	return string(data), nil
}

func loadMemoryIndexTemplate() (string, error) {
	data, err := assets.ConfigFS.ReadFile("config/memory/index.md")
	if err != nil {
		return "", fmt.Errorf("read embedded memory/index.md: %w", err)
	}
	return string(data), nil
}

func loadSoulTemplate() (string, error) {
	data, err := assets.ConfigFS.ReadFile("config/SOUL.md")
	if err != nil {
		return "", fmt.Errorf("read embedded SOUL.md: %w", err)
	}
	return string(data), nil
}

func patchInitConfigWithSetup(cfg string, dir string, setup *installConfigSetup) string {
	if strings.TrimSpace(cfg) == "" {
		return cfg
	}
	dir = filepath.Clean(dir)
	dir = filepath.ToSlash(dir)
	cfg = strings.ReplaceAll(cfg, `file_state_dir: "~/.morph"`, fmt.Sprintf(`file_state_dir: "%s"`, dir))
	cfg = applyInstallConfigSetupOverrides(cfg, setup)
	return cfg
}
