package credits

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"

	"github.com/quailyquaily/mistermorph/assets"
)

const DataPath = "credits/data.json"

type OpenSourceEntry struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Link    string `json:"link"`
	License string `json:"license"`
	Summary string `json:"summary"`
}

type ContributorEntry struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Link    string `json:"link"`
	Summary string `json:"summary"`
}

type Data struct {
	OpenSource   []OpenSourceEntry  `json:"open_source"`
	Contributors []ContributorEntry `json:"contributors"`
}

func Load() (Data, error) {
	return LoadFS(assets.CreditsFS, DataPath)
}

func LoadFS(fsys fs.FS, path string) (Data, error) {
	if fsys == nil {
		return Data{}, fmt.Errorf("credits fs is nil")
	}
	dataPath := strings.TrimSpace(path)
	if dataPath == "" {
		dataPath = DataPath
	}
	raw, err := fs.ReadFile(fsys, dataPath)
	if err != nil {
		return Data{}, fmt.Errorf("read credits data: %w", err)
	}
	var out Data
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&out); err != nil {
		return Data{}, fmt.Errorf("decode credits data: %w", err)
	}
	if err := validate(out); err != nil {
		return Data{}, err
	}
	if out.OpenSource == nil {
		out.OpenSource = []OpenSourceEntry{}
	}
	if out.Contributors == nil {
		out.Contributors = []ContributorEntry{}
	}
	return out, nil
}

func validate(data Data) error {
	openSourceIDs := make(map[string]struct{}, len(data.OpenSource))
	for i, item := range data.OpenSource {
		if err := validateOpenSourceEntry(item, i, openSourceIDs); err != nil {
			return err
		}
	}
	contributorIDs := make(map[string]struct{}, len(data.Contributors))
	for i, item := range data.Contributors {
		if err := validateContributorEntry(item, i, contributorIDs); err != nil {
			return err
		}
	}
	return nil
}

func validateOpenSourceEntry(item OpenSourceEntry, index int, seen map[string]struct{}) error {
	if err := validateID("open_source", item.ID, index, seen); err != nil {
		return err
	}
	if strings.TrimSpace(item.Name) == "" {
		return fmt.Errorf("credits open_source[%d]: name is required", index)
	}
	if strings.TrimSpace(item.Link) == "" {
		return fmt.Errorf("credits open_source[%d]: link is required", index)
	}
	if strings.TrimSpace(item.License) == "" {
		return fmt.Errorf("credits open_source[%d]: license is required", index)
	}
	if strings.TrimSpace(item.Summary) == "" {
		return fmt.Errorf("credits open_source[%d]: summary is required", index)
	}
	return nil
}

func validateContributorEntry(item ContributorEntry, index int, seen map[string]struct{}) error {
	if err := validateID("contributors", item.ID, index, seen); err != nil {
		return err
	}
	if strings.TrimSpace(item.Name) == "" {
		return fmt.Errorf("credits contributors[%d]: name is required", index)
	}
	if strings.TrimSpace(item.Link) == "" {
		return fmt.Errorf("credits contributors[%d]: link is required", index)
	}
	if strings.TrimSpace(item.Summary) == "" {
		return fmt.Errorf("credits contributors[%d]: summary is required", index)
	}
	return nil
}

func validateID(section, id string, index int, seen map[string]struct{}) error {
	key := strings.TrimSpace(id)
	if key == "" {
		return fmt.Errorf("credits %s[%d]: id is required", section, index)
	}
	if _, exists := seen[key]; exists {
		return fmt.Errorf("credits %s[%d]: duplicate id %q", section, index, key)
	}
	seen[key] = struct{}{}
	return nil
}
