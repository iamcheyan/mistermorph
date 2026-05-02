package skillsutil

import (
	"fmt"
	"strings"

	"github.com/quailyquaily/mistermorph/skills"
)

type SkillStatus struct {
	Enabled   bool
	Loaded    []SkillStatusItem
	NotLoaded []SkillStatusItem
	Missing   []string
}

type SkillStatusItem struct {
	ID   string
	Name string
}

func BuildSkillStatus(cfg SkillsConfig, currentLoaded []string) (SkillStatus, error) {
	status := SkillStatus{Enabled: cfg.Enabled}
	discovered, err := skills.Discover(skills.DiscoverOptions{Roots: cfg.Roots})
	if err != nil {
		return status, err
	}

	loadedIDs := map[string]bool{}
	if cfg.Enabled {
		requested := append([]string{}, cfg.Requested...)
		requested = append(requested, currentLoaded...)
		finalReq, loadAll := normalizeSkillStatusRequests(requested)
		if len(finalReq) == 0 {
			loadAll = true
		}
		if loadAll {
			for _, sk := range discovered {
				loadedIDs[strings.ToLower(strings.TrimSpace(sk.ID))] = true
			}
		} else {
			for _, query := range finalReq {
				sk, err := skills.Resolve(discovered, query)
				if err != nil {
					status.Missing = append(status.Missing, query)
					continue
				}
				loadedIDs[strings.ToLower(strings.TrimSpace(sk.ID))] = true
			}
		}
	}

	for _, sk := range discovered {
		item := SkillStatusItem{
			ID:   strings.TrimSpace(sk.ID),
			Name: strings.TrimSpace(sk.Name),
		}
		key := strings.ToLower(item.ID)
		if loadedIDs[key] {
			status.Loaded = append(status.Loaded, item)
		} else {
			status.NotLoaded = append(status.NotLoaded, item)
		}
	}
	return status, nil
}

func RenderSkillStatus(cfg SkillsConfig, currentLoaded []string) (string, error) {
	status, err := BuildSkillStatus(cfg, currentLoaded)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	if status.Enabled {
		b.WriteString("Skills: enabled")
	} else {
		b.WriteString("Skills: disabled")
	}
	writeSkillStatusItems(&b, "Loaded", status.Loaded)
	writeSkillStatusItems(&b, "Not loaded", status.NotLoaded)
	if len(status.Missing) > 0 {
		b.WriteString("\nMissing requested:")
		for _, name := range status.Missing {
			b.WriteString("\n- ")
			b.WriteString(name)
		}
	}
	if len(status.Loaded) == 0 && len(status.NotLoaded) == 0 {
		b.WriteString("\nNo skills discovered.")
	}
	return b.String(), nil
}

func normalizeSkillStatusRequests(requested []string) ([]string, bool) {
	seen := map[string]bool{}
	var out []string
	loadAll := false
	for _, raw := range requested {
		item := strings.TrimSpace(raw)
		if item == "" {
			continue
		}
		if item == "*" {
			loadAll = true
		}
		key := strings.ToLower(item)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, item)
	}
	return out, loadAll
}

func writeSkillStatusItems(b *strings.Builder, title string, items []SkillStatusItem) {
	if b == nil {
		return
	}
	if len(items) == 0 {
		fmt.Fprintf(b, "\n%s: none", title)
		return
	}
	fmt.Fprintf(b, "\n%s (%d):", title, len(items))
	for _, item := range items {
		label := strings.TrimSpace(item.ID)
		if label == "" {
			label = strings.TrimSpace(item.Name)
		}
		if label == "" {
			continue
		}
		b.WriteString("\n- ")
		b.WriteString(label)
	}
}
