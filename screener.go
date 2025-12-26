// Package tokenguard provides token screening and safety analysis for Solana tokens.
//
// The screener performs security checks on Solana tokens before trading, including:
//   - Mint/freeze authority status (rug pull risk)
//   - Liquidity depth (slippage risk)
//   - LP lock percentage (rug pull risk)
//   - Holder concentration (manipulation risk)
//
// Results are cached to avoid redundant API calls.
package tokenguard

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Laminar-Bot/birdeye-go"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// ============================================================================
// Interfaces for dependency injection and testing
// ============================================================================

// TokenSecurityProvider provides token security data (authorities, holder concentration).
// This interface allows mocking in tests.
type TokenSecurityProvider interface {
	GetTokenSecurity(ctx context.Context, address string) (*birdeye.TokenSecurity, error)
}

// TokenOverviewProvider provides token market data (liquidity, price, volume).
// This interface allows mocking in tests.
type TokenOverviewProvider interface {
	GetTokenOverview(ctx context.Context, address string) (*birdeye.TokenOverview, error)
}

// Cache provides caching for screening results.
// This interface allows swapping cache implementations.
type Cache interface {
	// Get retrieves a cached screening result if available and not expired.
	Get(ctx context.Context, tokenMint string) (*TokenScreeningResult, bool)

	// Set stores a screening result in the cache.
	Set(ctx context.Context, result *TokenScreeningResult) error
}

// ============================================================================
// Screener Implementation
// ============================================================================

// Screener performs token security analysis before trades.
//
// It checks:
//   - Mint/freeze authority status (rug pull risk)
//   - Liquidity depth (slippage risk)
//   - LP lock percentage (rug pull risk)
//   - Holder concentration (manipulation risk)
//
// Results are cached to avoid redundant API calls.
type Screener struct {
	security TokenSecurityProvider
	overview TokenOverviewProvider
	cache    Cache // Optional; nil disables caching
	logger   *zap.Logger

	// Thresholds for each screening level
	thresholds map[ScreeningLevel]ScreeningThresholds
}

// Config holds configuration for creating a new Screener.
type Config struct {
	// SecurityProvider is used to fetch token security data (required).
	SecurityProvider TokenSecurityProvider

	// OverviewProvider is used to fetch token market data (required).
	OverviewProvider TokenOverviewProvider

	// Cache stores screening results (optional; nil disables caching).
	Cache Cache

	// Logger for structured logging (required).
	Logger *zap.Logger
}

// New creates a new token screener.
func New(cfg Config) (*Screener, error) {
	if cfg.SecurityProvider == nil {
		return nil, fmt.Errorf("security provider is required")
	}
	if cfg.OverviewProvider == nil {
		return nil, fmt.Errorf("overview provider is required")
	}
	if cfg.Logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	return &Screener{
		security:   cfg.SecurityProvider,
		overview:   cfg.OverviewProvider,
		cache:      cfg.Cache,
		logger:     cfg.Logger,
		thresholds: defaultThresholds(),
	}, nil
}

// defaultThresholds returns the standard screening thresholds for each level.
//
// These values are based on industry best practices:
//   - Strict: High security requirements, lower risk tolerance
//   - Normal: Balanced requirements for most users
//   - Relaxed: Lower requirements, higher risk tolerance
func defaultThresholds() map[ScreeningLevel]ScreeningThresholds {
	return map[ScreeningLevel]ScreeningThresholds{
		ScreeningLevelStrict: {
			RequireNoMintAuth:   true,
			RequireNoFreezeAuth: true,
			MinLiquidityUSD:     decimal.NewFromInt(50000), // $50K minimum
			MinLPLockedPct:      decimal.NewFromInt(80),    // 80% LP locked
			MaxTop10HoldersPct:  decimal.NewFromInt(40),    // Top 10 hold max 40%
			MaxTopHolderPct:     decimal.NewFromInt(15),    // Single holder max 15%
		},
		ScreeningLevelNormal: {
			RequireNoMintAuth:   true,
			RequireNoFreezeAuth: true,
			MinLiquidityUSD:     decimal.NewFromInt(20000), // $20K minimum
			MinLPLockedPct:      decimal.NewFromInt(50),    // 50% LP locked
			MaxTop10HoldersPct:  decimal.NewFromInt(60),    // Top 10 hold max 60%
			MaxTopHolderPct:     decimal.NewFromInt(25),    // Single holder max 25%
		},
		ScreeningLevelRelaxed: {
			RequireNoMintAuth:   false, // Allows mint authority
			RequireNoFreezeAuth: true,
			MinLiquidityUSD:     decimal.NewFromInt(5000), // $5K minimum
			MinLPLockedPct:      decimal.NewFromInt(25),   // 25% LP locked
			MaxTop10HoldersPct:  decimal.NewFromInt(75),   // Top 10 hold max 75%
			MaxTopHolderPct:     decimal.NewFromInt(35),   // Single holder max 35%
		},
	}
}

// Screen performs security analysis on a token.
//
// It runs multiple checks based on the specified screening level and returns
// a result indicating whether the token passes screening.
//
// The screening process:
//  1. Check cache for recent results
//  2. Fetch token security data (authorities, holders)
//  3. Fetch token market data (liquidity)
//  4. Run checks against thresholds
//  5. Cache and return result
//
// Example:
//
//	result, err := screener.Screen(ctx, "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", tokenguard.ScreeningLevelNormal)
//	if err != nil {
//	    return err
//	}
//	if !result.Passed {
//	    log.Warn("token failed screening", zap.Strings("reasons", result.FailureReasons))
//	}
func (s *Screener) Screen(ctx context.Context, tokenMint string, level ScreeningLevel) (*TokenScreeningResult, error) {
	if tokenMint == "" {
		return nil, fmt.Errorf("token mint is required")
	}
	if !ValidScreeningLevel(level) {
		return nil, fmt.Errorf("invalid screening level: %s", level)
	}

	// Check cache first (if enabled)
	if s.cache != nil {
		if cached, ok := s.cache.Get(ctx, tokenMint); ok {
			s.logger.Debug("using cached screening result",
				zap.String("token_mint", tokenMint),
				zap.Bool("passed", cached.Passed),
			)
			return cached, nil
		}
	}

	// Get thresholds for this level
	threshold, ok := s.thresholds[level]
	if !ok {
		return nil, fmt.Errorf("unknown screening level: %s", level)
	}

	// Initialize result
	result := &TokenScreeningResult{
		TokenMint:      tokenMint,
		Passed:         true,
		Score:          100,
		Level:          level,
		FailureReasons: []string{},
		ScreenedAt:     time.Now(),
	}

	// Run all checks, collecting failures
	if err := s.checkAuthorities(ctx, tokenMint, threshold, result); err != nil {
		return nil, fmt.Errorf("authority check failed: %w", err)
	}

	if err := s.checkLiquidity(ctx, tokenMint, threshold, result); err != nil {
		return nil, fmt.Errorf("liquidity check failed: %w", err)
	}

	if err := s.checkHolderConcentration(ctx, tokenMint, threshold, result); err != nil {
		return nil, fmt.Errorf("holder concentration check failed: %w", err)
	}

	if err := s.checkLPLock(ctx, tokenMint, threshold, result); err != nil {
		return nil, fmt.Errorf("LP lock check failed: %w", err)
	}

	// Ensure score doesn't go negative
	if result.Score < 0 {
		result.Score = 0
	}

	// Cache result (if caching enabled)
	if s.cache != nil {
		if err := s.cache.Set(ctx, result); err != nil {
			s.logger.Warn("failed to cache screening result",
				zap.String("token_mint", tokenMint),
				zap.Error(err),
			)
			// Don't fail screening because of cache error
		}
	}

	s.logger.Info("token screening complete",
		zap.String("token_mint", tokenMint),
		zap.String("level", string(level)),
		zap.Bool("passed", result.Passed),
		zap.Int("score", result.Score),
		zap.Strings("failure_reasons", result.FailureReasons),
	)

	return result, nil
}

// checkAuthorities checks mint and freeze authority status.
//
// Tokens with active mint authority can have supply inflated (rug pull risk).
// Tokens with freeze authority can have user accounts frozen (loss of funds risk).
func (s *Screener) checkAuthorities(
	ctx context.Context,
	tokenMint string,
	threshold ScreeningThresholds,
	result *TokenScreeningResult,
) error {
	security, err := s.security.GetTokenSecurity(ctx, tokenMint)
	if err != nil {
		return fmt.Errorf("get token security: %w", err)
	}

	// Check mint authority
	hasMintAuth := security.HasMintAuthority()
	result.Details.HasMintAuthority = hasMintAuth

	if threshold.RequireNoMintAuth && hasMintAuth {
		result.Passed = false
		result.Score -= 30
		result.FailureReasons = append(result.FailureReasons, "has_mint_authority")
		s.logger.Debug("token has mint authority",
			zap.String("token_mint", tokenMint),
		)
	}

	// Check freeze authority
	hasFreezeAuth := security.HasFreezeAuthority()
	result.Details.HasFreezeAuthority = hasFreezeAuth

	if threshold.RequireNoFreezeAuth && hasFreezeAuth {
		result.Passed = false
		result.Score -= 20
		result.FailureReasons = append(result.FailureReasons, "has_freeze_authority")
		s.logger.Debug("token has freeze authority",
			zap.String("token_mint", tokenMint),
		)
	}

	// Check Token-2022 specific features
	result.Details.IsToken2022 = security.IsToken2022
	result.Details.HasTransferFee = security.TransferFeeEnable
	result.Details.NonTransferable = security.NonTransferable
	result.Details.MutableMetadata = security.MutableMetadata

	// Non-transferable tokens are always rejected
	if security.NonTransferable {
		result.Passed = false
		result.Score -= 50
		result.FailureReasons = append(result.FailureReasons, "non_transferable")
	}

	return nil
}

// checkLiquidity verifies the token has sufficient liquidity.
//
// Low liquidity means high slippage risk and potential for market manipulation.
func (s *Screener) checkLiquidity(
	ctx context.Context,
	tokenMint string,
	threshold ScreeningThresholds,
	result *TokenScreeningResult,
) error {
	overview, err := s.overview.GetTokenOverview(ctx, tokenMint)
	if err != nil {
		return fmt.Errorf("get token overview: %w", err)
	}

	result.Details.LiquidityUSD = overview.Liquidity

	if overview.Liquidity.LessThan(threshold.MinLiquidityUSD) {
		result.Passed = false
		result.Score -= 25
		result.FailureReasons = append(result.FailureReasons,
			fmt.Sprintf("low_liquidity:$%s", overview.Liquidity.StringFixed(2)))
		s.logger.Debug("token has low liquidity",
			zap.String("token_mint", tokenMint),
			zap.String("liquidity", overview.Liquidity.String()),
			zap.String("required", threshold.MinLiquidityUSD.String()),
		)
	}

	return nil
}

// checkHolderConcentration analyzes token distribution.
//
// High concentration in few wallets indicates manipulation risk.
func (s *Screener) checkHolderConcentration(
	ctx context.Context,
	tokenMint string,
	threshold ScreeningThresholds,
	result *TokenScreeningResult,
) error {
	security, err := s.security.GetTokenSecurity(ctx, tokenMint)
	if err != nil {
		return fmt.Errorf("get token security: %w", err)
	}

	// Parse top 10 holder percentage
	top10Pct := parsePercentage(security.Top10HolderPercent)
	result.Details.Top10HoldersPct = top10Pct

	if top10Pct.GreaterThan(threshold.MaxTop10HoldersPct) {
		result.Passed = false
		result.Score -= 15
		result.FailureReasons = append(result.FailureReasons,
			fmt.Sprintf("high_top10_concentration:%s%%", top10Pct.StringFixed(2)))
		s.logger.Debug("high top 10 holder concentration",
			zap.String("token_mint", tokenMint),
			zap.String("top10_pct", top10Pct.String()),
			zap.String("max_allowed", threshold.MaxTop10HoldersPct.String()),
		)
	}

	// Parse creator/top holder percentage
	topHolderPct := parsePercentage(security.CreatorPercentage)
	result.Details.TopHolderPct = topHolderPct

	if topHolderPct.GreaterThan(threshold.MaxTopHolderPct) {
		result.Passed = false
		result.Score -= 10
		result.FailureReasons = append(result.FailureReasons,
			fmt.Sprintf("high_single_holder:%s%%", topHolderPct.StringFixed(2)))
		s.logger.Debug("high single holder concentration",
			zap.String("token_mint", tokenMint),
			zap.String("top_holder_pct", topHolderPct.String()),
			zap.String("max_allowed", threshold.MaxTopHolderPct.String()),
		)
	}

	return nil
}

// checkLPLock estimates LP lock percentage.
//
// LP lock prevents the creator from pulling liquidity (rug pull).
// Note: This is an estimation based on creator holdings.
// A more accurate check would verify actual lock contracts.
func (s *Screener) checkLPLock(
	ctx context.Context,
	tokenMint string,
	threshold ScreeningThresholds,
	result *TokenScreeningResult,
) error {
	security, err := s.security.GetTokenSecurity(ctx, tokenMint)
	if err != nil {
		return fmt.Errorf("get token security: %w", err)
	}

	// Estimate LP lock percentage based on creator holdings.
	// If creator holds a small percentage, it suggests LP is locked.
	// This is a simplified heuristic; production should verify lock contracts.
	creatorPct := parsePercentage(security.CreatorPercentage)

	// Rough estimate: 100% - creator% gives an upper bound on locked LP.
	lpLockedPct := decimal.NewFromInt(100).Sub(creatorPct)
	if lpLockedPct.IsNegative() {
		lpLockedPct = decimal.Zero
	}
	if lpLockedPct.GreaterThan(decimal.NewFromInt(100)) {
		lpLockedPct = decimal.NewFromInt(100)
	}

	result.Details.LPLockedPct = lpLockedPct

	if lpLockedPct.LessThan(threshold.MinLPLockedPct) {
		result.Passed = false
		result.Score -= 15
		result.FailureReasons = append(result.FailureReasons,
			fmt.Sprintf("low_lp_locked:%s%%", lpLockedPct.StringFixed(2)))
		s.logger.Debug("low LP lock percentage",
			zap.String("token_mint", tokenMint),
			zap.String("lp_locked_pct", lpLockedPct.String()),
			zap.String("min_required", threshold.MinLPLockedPct.String()),
		)
	}

	return nil
}

// parsePercentage safely parses a percentage string to decimal.
// Returns zero if parsing fails.
func parsePercentage(s string) decimal.Decimal {
	if s == "" {
		return decimal.Zero
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return decimal.Zero
	}

	return decimal.NewFromFloat(val)
}
