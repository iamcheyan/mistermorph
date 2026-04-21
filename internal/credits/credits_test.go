package credits

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadFS(t *testing.T) {
	fsys := fstest.MapFS{
		DataPath: {
			Data: []byte(`{
				"open_source": [
					{
						"id": "cobra",
						"name": "Cobra",
						"link": "https://github.com/spf13/cobra",
						"license": "Apache-2.0",
						"summary": "CLI command tree."
					}
				],
				"contributors": [
					{
						"id": "quaily",
						"name": "Quaily",
						"link": "https://github.com/quailyquaily",
						"summary": "Maintains the project."
					}
				]
			}`),
		},
	}

	data, err := LoadFS(fsys, DataPath)
	if err != nil {
		t.Fatalf("LoadFS() error = %v", err)
	}
	if len(data.OpenSource) != 1 || data.OpenSource[0].ID != "cobra" {
		t.Fatalf("OpenSource = %+v", data.OpenSource)
	}
	if len(data.Contributors) != 1 || data.Contributors[0].ID != "quaily" {
		t.Fatalf("Contributors = %+v", data.Contributors)
	}
}

func TestLoadFSRejectsDuplicateIDs(t *testing.T) {
	fsys := fstest.MapFS{
		DataPath: {
			Data: []byte(`{
				"open_source": [
					{
						"id": "cobra",
						"name": "Cobra",
						"link": "https://github.com/spf13/cobra",
						"license": "Apache-2.0",
						"summary": "CLI command tree."
					},
					{
						"id": "cobra",
						"name": "Viper",
						"link": "https://github.com/spf13/viper",
						"license": "MIT",
						"summary": "Configuration loader."
					}
				],
				"contributors": []
			}`),
		},
	}

	_, err := LoadFS(fsys, DataPath)
	if err == nil {
		t.Fatalf("LoadFS() error = nil, want duplicate id error")
	}
	if !strings.Contains(err.Error(), `duplicate id "cobra"`) {
		t.Fatalf("LoadFS() error = %v, want duplicate id", err)
	}
}

func TestLoadFSRejectsMissingContributorSummary(t *testing.T) {
	fsys := fstest.MapFS{
		DataPath: {
			Data: []byte(`{
				"open_source": [],
				"contributors": [
					{
						"id": "quaily",
						"name": "Quaily",
						"link": "https://github.com/quailyquaily",
						"summary": ""
					}
				]
			}`),
		},
	}

	_, err := LoadFS(fsys, DataPath)
	if err == nil {
		t.Fatalf("LoadFS() error = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "contributors[0]: summary is required") {
		t.Fatalf("LoadFS() error = %v, want summary validation error", err)
	}
}
