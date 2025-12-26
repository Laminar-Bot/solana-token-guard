package tokenguard

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestInMemoryCache_SetGet(t *testing.T) {
	ctx := context.Background()
	cache := NewInMemoryCache(InMemoryCacheConfig{TTL: time.Minute})
	defer func() { _ = cache.Close() }()

	result := &TokenScreeningResult{
		TokenMint:      "test-mint",
		Passed:         true,
		Score:          100,
		Level:          ScreeningLevelNormal,
		FailureReasons: []string{},
		ScreenedAt:     time.Now(),
	}

	// Set
	err := cache.Set(ctx, result)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Get
	got, ok := cache.Get(ctx, "test-mint")
	if !ok {
		t.Fatal("expected cache hit")
	}

	if got.TokenMint != result.TokenMint {
		t.Errorf("expected TokenMint %s, got %s", result.TokenMint, got.TokenMint)
	}
	if got.Passed != result.Passed {
		t.Errorf("expected Passed %v, got %v", result.Passed, got.Passed)
	}
	if got.Score != result.Score {
		t.Errorf("expected Score %d, got %d", result.Score, got.Score)
	}
}

func TestInMemoryCache_CacheMiss(t *testing.T) {
	ctx := context.Background()
	cache := NewInMemoryCache(InMemoryCacheConfig{TTL: time.Minute})
	defer func() { _ = cache.Close() }()

	got, ok := cache.Get(ctx, "nonexistent")
	if ok {
		t.Error("expected cache miss")
	}
	if got != nil {
		t.Error("expected nil result")
	}
}

func TestInMemoryCache_Expiration(t *testing.T) {
	ctx := context.Background()
	cache := NewInMemoryCache(InMemoryCacheConfig{TTL: 10 * time.Millisecond})
	defer func() { _ = cache.Close() }()

	result := &TokenScreeningResult{
		TokenMint:  "test-mint",
		Passed:     true,
		Score:      100,
		ScreenedAt: time.Now(),
	}

	err := cache.Set(ctx, result)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Should exist immediately
	_, ok := cache.Get(ctx, "test-mint")
	if !ok {
		t.Fatal("expected cache hit immediately after set")
	}

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Should be expired
	_, ok = cache.Get(ctx, "test-mint")
	if ok {
		t.Error("expected cache miss after expiration")
	}
}

func TestInMemoryCache_Delete(t *testing.T) {
	ctx := context.Background()
	cache := NewInMemoryCache(InMemoryCacheConfig{TTL: time.Minute})
	defer func() { _ = cache.Close() }()

	result := &TokenScreeningResult{
		TokenMint:  "test-mint",
		Passed:     true,
		Score:      100,
		ScreenedAt: time.Now(),
	}

	err := cache.Set(ctx, result)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Verify it exists
	_, ok := cache.Get(ctx, "test-mint")
	if !ok {
		t.Fatal("expected cache hit")
	}

	// Delete
	cache.Delete(ctx, "test-mint")

	// Verify it's gone
	_, ok = cache.Get(ctx, "test-mint")
	if ok {
		t.Error("expected cache miss after delete")
	}
}

func TestInMemoryCache_Clear(t *testing.T) {
	ctx := context.Background()
	cache := NewInMemoryCache(InMemoryCacheConfig{TTL: time.Minute})
	defer func() { _ = cache.Close() }()

	// Add multiple entries
	for i := 0; i < 5; i++ {
		result := &TokenScreeningResult{
			TokenMint:  "mint-" + string(rune('A'+i)),
			Passed:     true,
			Score:      100,
			ScreenedAt: time.Now(),
		}
		err := cache.Set(ctx, result)
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}
	}

	if cache.Size() != 5 {
		t.Errorf("expected size 5, got %d", cache.Size())
	}

	// Clear
	cache.Clear(ctx)

	if cache.Size() != 0 {
		t.Errorf("expected size 0 after clear, got %d", cache.Size())
	}
}

func TestInMemoryCache_Overwrite(t *testing.T) {
	ctx := context.Background()
	cache := NewInMemoryCache(InMemoryCacheConfig{TTL: time.Minute})
	defer func() { _ = cache.Close() }()

	// First value
	result1 := &TokenScreeningResult{
		TokenMint:  "test-mint",
		Passed:     true,
		Score:      100,
		ScreenedAt: time.Now(),
	}
	err := cache.Set(ctx, result1)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Overwrite with different value
	result2 := &TokenScreeningResult{
		TokenMint:  "test-mint",
		Passed:     false,
		Score:      50,
		ScreenedAt: time.Now(),
	}
	err = cache.Set(ctx, result2)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Should get the new value
	got, ok := cache.Get(ctx, "test-mint")
	if !ok {
		t.Fatal("expected cache hit")
	}

	if got.Passed != false {
		t.Error("expected Passed to be false (new value)")
	}
	if got.Score != 50 {
		t.Errorf("expected Score 50, got %d", got.Score)
	}
}

func TestInMemoryCache_PreservesDetails(t *testing.T) {
	ctx := context.Background()
	cache := NewInMemoryCache(InMemoryCacheConfig{TTL: time.Minute})
	defer func() { _ = cache.Close() }()

	result := &TokenScreeningResult{
		TokenMint: "test-mint",
		Passed:    true,
		Score:     85,
		Level:     ScreeningLevelStrict,
		Details: ScreeningDetails{
			HasMintAuthority:   false,
			HasFreezeAuthority: true,
			LiquidityUSD:       decimal.NewFromInt(75000),
			LPLockedPct:        decimal.NewFromFloat(82.5),
			Top10HoldersPct:    decimal.NewFromFloat(35.2),
			TopHolderPct:       decimal.NewFromFloat(12.1),
			IsToken2022:        true,
			HasTransferFee:     false,
			NonTransferable:    false,
			MutableMetadata:    true,
		},
		FailureReasons: []string{"has_freeze_authority"},
		ScreenedAt:     time.Now(),
	}

	err := cache.Set(ctx, result)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	got, ok := cache.Get(ctx, "test-mint")
	if !ok {
		t.Fatal("expected cache hit")
	}

	// Verify all fields are preserved
	if got.Details.HasMintAuthority != false {
		t.Error("HasMintAuthority mismatch")
	}
	if got.Details.HasFreezeAuthority != true {
		t.Error("HasFreezeAuthority mismatch")
	}
	if !got.Details.LiquidityUSD.Equal(decimal.NewFromInt(75000)) {
		t.Errorf("LiquidityUSD mismatch: got %s", got.Details.LiquidityUSD)
	}
	if !got.Details.LPLockedPct.Equal(decimal.NewFromFloat(82.5)) {
		t.Errorf("LPLockedPct mismatch: got %s", got.Details.LPLockedPct)
	}
	if got.Details.IsToken2022 != true {
		t.Error("IsToken2022 mismatch")
	}
	if len(got.FailureReasons) != 1 || got.FailureReasons[0] != "has_freeze_authority" {
		t.Errorf("FailureReasons mismatch: got %v", got.FailureReasons)
	}
}

func TestNoOpCache(t *testing.T) {
	ctx := context.Background()
	cache := NewNoOpCache()

	result := &TokenScreeningResult{
		TokenMint:  "test-mint",
		Passed:     true,
		Score:      100,
		ScreenedAt: time.Now(),
	}

	// Set should not error
	err := cache.Set(ctx, result)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Get should always miss
	got, ok := cache.Get(ctx, "test-mint")
	if ok {
		t.Error("NoOpCache should always return cache miss")
	}
	if got != nil {
		t.Error("NoOpCache should always return nil")
	}
}

func TestInMemoryCache_BackgroundCleanup(t *testing.T) {
	ctx := context.Background()
	cache := NewInMemoryCache(InMemoryCacheConfig{
		TTL:             10 * time.Millisecond,
		CleanupInterval: 5 * time.Millisecond,
	})
	defer func() { _ = cache.Close() }()

	// Add entries
	for i := 0; i < 3; i++ {
		result := &TokenScreeningResult{
			TokenMint:  "mint-" + string(rune('A'+i)),
			Passed:     true,
			Score:      100,
			ScreenedAt: time.Now(),
		}
		err := cache.Set(ctx, result)
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}
	}

	if cache.Size() != 3 {
		t.Errorf("expected size 3, got %d", cache.Size())
	}

	// Wait for expiration and cleanup
	time.Sleep(30 * time.Millisecond)

	// Background cleanup should have removed entries
	if cache.Size() != 0 {
		t.Errorf("expected size 0 after cleanup, got %d", cache.Size())
	}
}

func TestInMemoryCache_DefaultTTL(t *testing.T) {
	cache := NewInMemoryCache(InMemoryCacheConfig{})
	defer func() { _ = cache.Close() }()

	if cache.ttl != DefaultCacheTTL {
		t.Errorf("expected default TTL %v, got %v", DefaultCacheTTL, cache.ttl)
	}
}
