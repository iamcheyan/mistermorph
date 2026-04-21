package consolecmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleCreditsGet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/settings/credits", nil)
	rec := httptest.NewRecorder()

	(&server{}).handleCredits(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (%s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body creditsPayload
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(body.OpenSource) == 0 {
		t.Fatalf("OpenSource = empty")
	}
	if len(body.Contributors) == 0 {
		t.Fatalf("Contributors = empty")
	}
	if body.OpenSource[0].ID == "" || body.OpenSource[0].Name == "" {
		t.Fatalf("OpenSource[0] = %+v", body.OpenSource[0])
	}
}

func TestHandleCreditsMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/settings/credits", nil)
	rec := httptest.NewRecorder()

	(&server{}).handleCredits(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}
