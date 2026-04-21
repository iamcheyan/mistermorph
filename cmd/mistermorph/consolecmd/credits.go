package consolecmd

import (
	"net/http"

	sharedcredits "github.com/quailyquaily/mistermorph/internal/credits"
)

type creditsOpenSourcePayload struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Link    string `json:"link"`
	License string `json:"license"`
	Summary string `json:"summary"`
}

type creditsContributorPayload struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Link    string `json:"link"`
	Summary string `json:"summary"`
}

type creditsPayload struct {
	OpenSource   []creditsOpenSourcePayload  `json:"open_source"`
	Contributors []creditsContributorPayload `json:"contributors"`
}

func (s *server) handleCredits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	data, err := sharedcredits.Load()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, buildCreditsPayload(data))
}

func buildCreditsPayload(data sharedcredits.Data) creditsPayload {
	out := creditsPayload{
		OpenSource:   make([]creditsOpenSourcePayload, 0, len(data.OpenSource)),
		Contributors: make([]creditsContributorPayload, 0, len(data.Contributors)),
	}
	for _, item := range data.OpenSource {
		out.OpenSource = append(out.OpenSource, creditsOpenSourcePayload{
			ID:      item.ID,
			Name:    item.Name,
			Link:    item.Link,
			License: item.License,
			Summary: item.Summary,
		})
	}
	for _, item := range data.Contributors {
		out.Contributors = append(out.Contributors, creditsContributorPayload{
			ID:      item.ID,
			Name:    item.Name,
			Link:    item.Link,
			Summary: item.Summary,
		})
	}
	return out
}
