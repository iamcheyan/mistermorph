package llm

import "testing"

func TestUsageCostField(t *testing.T) {
	u := Usage{
		InputTokens:  100,
		OutputTokens: 50,
		TotalTokens:  150,
		Cost: &UsageCost{
			Currency: "USD",
			Total:    0.05,
		},
	}
	if u.Cost == nil || u.Cost.Total != 0.05 {
		t.Fatalf("expected Cost.Total 0.05, got %#v", u.Cost)
	}
}

func TestUsageCostDefaultsToNil(t *testing.T) {
	u := Usage{}
	if u.Cost != nil {
		t.Fatalf("expected Cost to default to nil, got %#v", u.Cost)
	}
}

func TestUsageCacheDefaultsToZeroValues(t *testing.T) {
	u := Usage{}
	if u.Cache.CachedInputTokens != 0 {
		t.Fatalf("expected CachedInputTokens=0, got %d", u.Cache.CachedInputTokens)
	}
	if u.Cache.CacheCreationInputTokens != 0 {
		t.Fatalf("expected CacheCreationInputTokens=0, got %d", u.Cache.CacheCreationInputTokens)
	}
	if len(u.Cache.Details) != 0 {
		t.Fatalf("expected Details empty, got %#v", u.Cache.Details)
	}
}
