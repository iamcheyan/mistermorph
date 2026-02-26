package entryutil

import (
	"context"
	"reflect"
	"testing"
)

type stubSemanticResolver struct {
	keep []int
	err  error
}

func (s stubSemanticResolver) SelectDedupKeepIndices(_ context.Context, _ []SemanticItem) ([]int, error) {
	if s.err != nil {
		return nil, s.err
	}
	return append([]int{}, s.keep...), nil
}

func TestResolveKeepIndicesAddsNewestWhenMissing(t *testing.T) {
	items := []SemanticItem{
		{CreatedAt: "2026-02-09 10:02", Content: "newest"},
		{CreatedAt: "2026-02-09 10:01", Content: "a"},
		{CreatedAt: "2026-02-09 10:00", Content: "b"},
	}

	got, err := ResolveKeepIndices(context.Background(), items, stubSemanticResolver{keep: []int{2, 1}})
	if err != nil {
		t.Fatalf("ResolveKeepIndices() error = %v", err)
	}
	want := []int{0, 1, 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("keep indices = %#v, want %#v", got, want)
	}
}

func TestResolveKeepIndicesDropsOutOfRangeIndices(t *testing.T) {
	items := []SemanticItem{
		{CreatedAt: "2026-02-09 10:02", Content: "newest"},
		{CreatedAt: "2026-02-09 10:01", Content: "a"},
		{CreatedAt: "2026-02-09 10:00", Content: "b"},
	}

	got, err := ResolveKeepIndices(context.Background(), items, stubSemanticResolver{keep: []int{0, 2, 9, -1, 2}})
	if err != nil {
		t.Fatalf("ResolveKeepIndices() error = %v", err)
	}
	want := []int{0, 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("keep indices = %#v, want %#v", got, want)
	}
}

func TestResolveKeepIndicesFallsBackToNewestWhenAllInvalid(t *testing.T) {
	items := []SemanticItem{
		{CreatedAt: "2026-02-09 10:02", Content: "newest"},
		{CreatedAt: "2026-02-09 10:01", Content: "a"},
		{CreatedAt: "2026-02-09 10:00", Content: "b"},
	}

	got, err := ResolveKeepIndices(context.Background(), items, stubSemanticResolver{keep: []int{8, -3}})
	if err != nil {
		t.Fatalf("ResolveKeepIndices() error = %v", err)
	}
	want := []int{0}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("keep indices = %#v, want %#v", got, want)
	}
}
