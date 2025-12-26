package tokenguard

import (
	"context"
	"errors"
	"strings"
	"testing"

	birdeye "github.com/Laminar-Bot/birdeye-go"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// ============================================================================
// Mock Implementations
// ============================================================================

// mockSecurityProvider is a mock implementation of TokenSecurityProvider.
type mockSecurityProvider struct {
	security *birdeye.TokenSecurity
	err      error
}

func (m *mockSecurityProvider) GetTokenSecurity(_ context.Context, _ string) (*birdeye.TokenSecurity, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.security, nil
}

// mockOverviewProvider is a mock implementation of TokenOverviewProvider.
type mockOverviewProvider struct {
	overview *birdeye.TokenOverview
	err      error
}

func (m *mockOverviewProvider) GetTokenOverview(_ context.Context, _ string) (*birdeye.TokenOverview, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.overview, nil
}

// ============================================================================
// Test Cases
// ============================================================================

func TestNew(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: Config{
				SecurityProvider: &mockSecurityProvider{},
				OverviewProvider: &mockOverviewProvider{},
				Logger:           logger,
			},
			wantErr: false,
		},
		{
			name: "missing security provider",
			cfg: Config{
				OverviewProvider: &mockOverviewProvider{},
				Logger:           logger,
			},
			wantErr: true,
		},
		{
			name: "missing overview provider",
			cfg: Config{
				SecurityProvider: &mockSecurityProvider{},
				Logger:           logger,
			},
			wantErr: true,
		},
		{
			name: "missing logger",
			cfg: Config{
				SecurityProvider: &mockSecurityProvider{},
				OverviewProvider: &mockOverviewProvider{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScreener_Screen_PassingToken(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	// Create a token that passes all checks
	securityProvider := &mockSecurityProvider{
		security: &birdeye.TokenSecurity{
			MintAuthority:      nil,  // No mint authority - safe
			FreezeAuthority:    nil,  // No freeze authority - safe
			CreatorPercentage:  "5",  // Low creator holdings
			Top10HolderPercent: "30", // Low concentration
			IsToken2022:        false,
			NonTransferable:    false,
		},
	}

	overviewProvider := &mockOverviewProvider{
		overview: &birdeye.TokenOverview{
			Liquidity: decimal.NewFromInt(100000), // $100K liquidity
		},
	}

	screener, err := New(Config{
		SecurityProvider: securityProvider,
		OverviewProvider: overviewProvider,
		Logger:           logger,
	})
	if err != nil {
		t.Fatalf("failed to create screener: %v", err)
	}

	result, err := screener.Screen(ctx, "test-mint", ScreeningLevelNormal)
	if err != nil {
		t.Fatalf("Screen() error = %v", err)
	}

	if !result.Passed {
		t.Errorf("expected token to pass, but got failure reasons: %v", result.FailureReasons)
	}

	if result.Score != 100 {
		t.Errorf("expected score 100, got %d", result.Score)
	}

	if len(result.FailureReasons) != 0 {
		t.Errorf("expected no failure reasons, got: %v", result.FailureReasons)
	}
}

func TestScreener_Screen_FailingMintAuthority(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	mintAuth := "SomeMintAuthority"
	securityProvider := &mockSecurityProvider{
		security: &birdeye.TokenSecurity{
			MintAuthority:      &mintAuth, // Has mint authority - risky
			FreezeAuthority:    nil,
			CreatorPercentage:  "5",
			Top10HolderPercent: "30",
		},
	}

	overviewProvider := &mockOverviewProvider{
		overview: &birdeye.TokenOverview{
			Liquidity: decimal.NewFromInt(100000),
		},
	}

	screener, err := New(Config{
		SecurityProvider: securityProvider,
		OverviewProvider: overviewProvider,
		Logger:           logger,
	})
	if err != nil {
		t.Fatalf("failed to create screener: %v", err)
	}

	// Test with strict level (requires no mint auth)
	result, err := screener.Screen(ctx, "test-mint", ScreeningLevelStrict)
	if err != nil {
		t.Fatalf("Screen() error = %v", err)
	}

	if result.Passed {
		t.Error("expected token to fail due to mint authority")
	}

	if !result.Details.HasMintAuthority {
		t.Error("expected HasMintAuthority to be true")
	}

	if !contains(result.FailureReasons, "has_mint_authority") {
		t.Errorf("expected 'has_mint_authority' in failure reasons, got: %v", result.FailureReasons)
	}
}

func TestScreener_Screen_FailingFreezeAuthority(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	freezeAuth := "SomeFreezeAuthority"
	securityProvider := &mockSecurityProvider{
		security: &birdeye.TokenSecurity{
			MintAuthority:      nil,
			FreezeAuthority:    &freezeAuth, // Has freeze authority - risky
			CreatorPercentage:  "5",
			Top10HolderPercent: "30",
		},
	}

	overviewProvider := &mockOverviewProvider{
		overview: &birdeye.TokenOverview{
			Liquidity: decimal.NewFromInt(100000),
		},
	}

	screener, err := New(Config{
		SecurityProvider: securityProvider,
		OverviewProvider: overviewProvider,
		Logger:           logger,
	})
	if err != nil {
		t.Fatalf("failed to create screener: %v", err)
	}

	result, err := screener.Screen(ctx, "test-mint", ScreeningLevelNormal)
	if err != nil {
		t.Fatalf("Screen() error = %v", err)
	}

	if result.Passed {
		t.Error("expected token to fail due to freeze authority")
	}

	if !result.Details.HasFreezeAuthority {
		t.Error("expected HasFreezeAuthority to be true")
	}

	if !contains(result.FailureReasons, "has_freeze_authority") {
		t.Errorf("expected 'has_freeze_authority' in failure reasons, got: %v", result.FailureReasons)
	}
}

func TestScreener_Screen_FailingLowLiquidity(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	securityProvider := &mockSecurityProvider{
		security: &birdeye.TokenSecurity{
			MintAuthority:      nil,
			FreezeAuthority:    nil,
			CreatorPercentage:  "5",
			Top10HolderPercent: "30",
		},
	}

	overviewProvider := &mockOverviewProvider{
		overview: &birdeye.TokenOverview{
			Liquidity: decimal.NewFromInt(1000), // Only $1K - too low
		},
	}

	screener, err := New(Config{
		SecurityProvider: securityProvider,
		OverviewProvider: overviewProvider,
		Logger:           logger,
	})
	if err != nil {
		t.Fatalf("failed to create screener: %v", err)
	}

	result, err := screener.Screen(ctx, "test-mint", ScreeningLevelRelaxed)
	if err != nil {
		t.Fatalf("Screen() error = %v", err)
	}

	if result.Passed {
		t.Error("expected token to fail due to low liquidity")
	}

	// Should contain a low_liquidity reason
	hasLiquidityReason := false
	for _, reason := range result.FailureReasons {
		if len(reason) > 14 && reason[:14] == "low_liquidity:" {
			hasLiquidityReason = true
			break
		}
	}
	if !hasLiquidityReason {
		t.Errorf("expected 'low_liquidity' in failure reasons, got: %v", result.FailureReasons)
	}
}

func TestScreener_Screen_FailingHighConcentration(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	securityProvider := &mockSecurityProvider{
		security: &birdeye.TokenSecurity{
			MintAuthority:      nil,
			FreezeAuthority:    nil,
			CreatorPercentage:  "50", // Too high
			Top10HolderPercent: "90", // Way too high
		},
	}

	overviewProvider := &mockOverviewProvider{
		overview: &birdeye.TokenOverview{
			Liquidity: decimal.NewFromInt(100000),
		},
	}

	screener, err := New(Config{
		SecurityProvider: securityProvider,
		OverviewProvider: overviewProvider,
		Logger:           logger,
	})
	if err != nil {
		t.Fatalf("failed to create screener: %v", err)
	}

	result, err := screener.Screen(ctx, "test-mint", ScreeningLevelNormal)
	if err != nil {
		t.Fatalf("Screen() error = %v", err)
	}

	if result.Passed {
		t.Error("expected token to fail due to high concentration")
	}

	// Should contain concentration reasons
	hasTop10Reason := false
	hasTopHolderReason := false
	for _, reason := range result.FailureReasons {
		if strings.HasPrefix(reason, "high_top10_concentration") {
			hasTop10Reason = true
		}
		if strings.HasPrefix(reason, "high_single_holder") {
			hasTopHolderReason = true
		}
	}
	if !hasTop10Reason {
		t.Errorf("expected 'high_top10_concentration' in failure reasons, got: %v", result.FailureReasons)
	}
	if !hasTopHolderReason {
		t.Errorf("expected 'high_single_holder' in failure reasons, got: %v", result.FailureReasons)
	}
}

func TestScreener_Screen_NonTransferableToken(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	securityProvider := &mockSecurityProvider{
		security: &birdeye.TokenSecurity{
			MintAuthority:      nil,
			FreezeAuthority:    nil,
			CreatorPercentage:  "5",
			Top10HolderPercent: "30",
			NonTransferable:    true, // Soulbound token
		},
	}

	overviewProvider := &mockOverviewProvider{
		overview: &birdeye.TokenOverview{
			Liquidity: decimal.NewFromInt(100000),
		},
	}

	screener, err := New(Config{
		SecurityProvider: securityProvider,
		OverviewProvider: overviewProvider,
		Logger:           logger,
	})
	if err != nil {
		t.Fatalf("failed to create screener: %v", err)
	}

	result, err := screener.Screen(ctx, "test-mint", ScreeningLevelRelaxed)
	if err != nil {
		t.Fatalf("Screen() error = %v", err)
	}

	if result.Passed {
		t.Error("expected non-transferable token to fail")
	}

	if !contains(result.FailureReasons, "non_transferable") {
		t.Errorf("expected 'non_transferable' in failure reasons, got: %v", result.FailureReasons)
	}
}

func TestScreener_Screen_RelaxedLevelAllowsMintAuth(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	mintAuth := "SomeMintAuthority"
	securityProvider := &mockSecurityProvider{
		security: &birdeye.TokenSecurity{
			MintAuthority:      &mintAuth, // Has mint authority
			FreezeAuthority:    nil,
			CreatorPercentage:  "5",
			Top10HolderPercent: "30",
		},
	}

	overviewProvider := &mockOverviewProvider{
		overview: &birdeye.TokenOverview{
			Liquidity: decimal.NewFromInt(100000),
		},
	}

	screener, err := New(Config{
		SecurityProvider: securityProvider,
		OverviewProvider: overviewProvider,
		Logger:           logger,
	})
	if err != nil {
		t.Fatalf("failed to create screener: %v", err)
	}

	// Relaxed level allows mint authority
	result, err := screener.Screen(ctx, "test-mint", ScreeningLevelRelaxed)
	if err != nil {
		t.Fatalf("Screen() error = %v", err)
	}

	if !result.Passed {
		t.Errorf("expected relaxed level to pass with mint authority, but got failure reasons: %v", result.FailureReasons)
	}
}

func TestScreener_Screen_APIError(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	securityProvider := &mockSecurityProvider{
		err: errors.New("API error"),
	}

	overviewProvider := &mockOverviewProvider{
		overview: &birdeye.TokenOverview{
			Liquidity: decimal.NewFromInt(100000),
		},
	}

	screener, err := New(Config{
		SecurityProvider: securityProvider,
		OverviewProvider: overviewProvider,
		Logger:           logger,
	})
	if err != nil {
		t.Fatalf("failed to create screener: %v", err)
	}

	_, err = screener.Screen(ctx, "test-mint", ScreeningLevelNormal)
	if err == nil {
		t.Error("expected error when API fails")
	}
}

func TestScreener_Screen_InvalidInput(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	screener, err := New(Config{
		SecurityProvider: &mockSecurityProvider{},
		OverviewProvider: &mockOverviewProvider{},
		Logger:           logger,
	})
	if err != nil {
		t.Fatalf("failed to create screener: %v", err)
	}

	// Empty token mint
	_, err = screener.Screen(ctx, "", ScreeningLevelNormal)
	if err == nil {
		t.Error("expected error for empty token mint")
	}

	// Invalid screening level
	_, err = screener.Screen(ctx, "test-mint", ScreeningLevel("invalid"))
	if err == nil {
		t.Error("expected error for invalid screening level")
	}
}

func TestScreener_Screen_WithCache(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	callCount := 0
	securityProvider := &mockSecurityProvider{
		security: &birdeye.TokenSecurity{
			MintAuthority:      nil,
			FreezeAuthority:    nil,
			CreatorPercentage:  "5",
			Top10HolderPercent: "30",
		},
	}

	// Wrap to count calls
	countingProvider := &countingSecurityProvider{
		inner:     securityProvider,
		callCount: &callCount,
	}

	overviewProvider := &mockOverviewProvider{
		overview: &birdeye.TokenOverview{
			Liquidity: decimal.NewFromInt(100000),
		},
	}

	cache := NewInMemoryCache(InMemoryCacheConfig{})

	screener, err := New(Config{
		SecurityProvider: countingProvider,
		OverviewProvider: overviewProvider,
		Cache:            cache,
		Logger:           logger,
	})
	if err != nil {
		t.Fatalf("failed to create screener: %v", err)
	}

	// First call should hit the API
	_, err = screener.Screen(ctx, "test-mint", ScreeningLevelNormal)
	if err != nil {
		t.Fatalf("Screen() error = %v", err)
	}

	if callCount == 0 {
		t.Error("expected API to be called on first request")
	}

	firstCallCount := callCount

	// Second call should use cache
	_, err = screener.Screen(ctx, "test-mint", ScreeningLevelNormal)
	if err != nil {
		t.Fatalf("Screen() error = %v", err)
	}

	if callCount != firstCallCount {
		t.Errorf("expected API call count to remain %d, but got %d (cache should have been used)", firstCallCount, callCount)
	}
}

// countingSecurityProvider wraps a security provider to count calls.
type countingSecurityProvider struct {
	inner     TokenSecurityProvider
	callCount *int
}

func (c *countingSecurityProvider) GetTokenSecurity(ctx context.Context, address string) (*birdeye.TokenSecurity, error) {
	*c.callCount++
	return c.inner.GetTokenSecurity(ctx, address)
}

// Helper function to check if a slice contains a string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
