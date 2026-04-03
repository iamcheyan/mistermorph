package llmselect

import (
	"strings"
	"sync"
)

const (
	ModeAuto   = "auto"
	ModeManual = "manual"
)

type MainSelection struct {
	Mode          string
	ManualProfile string
}

type Store struct {
	mu        sync.RWMutex
	selection MainSelection
}

var processStore = NewStore()

func NewStore() *Store {
	return &Store{selection: MainSelection{Mode: ModeAuto}}
}

func ProcessStore() *Store {
	return processStore
}

func NormalizeSelection(sel MainSelection) MainSelection {
	sel.ManualProfile = strings.TrimSpace(sel.ManualProfile)
	if sel.Mode != ModeManual {
		return MainSelection{Mode: ModeAuto}
	}
	if sel.ManualProfile == "" {
		return MainSelection{Mode: ModeAuto}
	}
	return MainSelection{
		Mode:          ModeManual,
		ManualProfile: sel.ManualProfile,
	}
}

func (s *Store) Get() MainSelection {
	if s == nil {
		return MainSelection{Mode: ModeAuto}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return NormalizeSelection(s.selection)
}

func (s *Store) SetProfile(profileName string) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.selection = NormalizeSelection(MainSelection{
		Mode:          ModeManual,
		ManualProfile: profileName,
	})
}

func (s *Store) Reset() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.selection = MainSelection{Mode: ModeAuto}
}
