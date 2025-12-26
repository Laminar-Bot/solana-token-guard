// Package tokenguard provides token screening and safety analysis for Solana tokens.
//
// This library performs security checks on Solana tokens before trading, including:
//   - Authority validation (mint/freeze authority status)
//   - Liquidity analysis (minimum LP thresholds)
//   - Holder concentration analysis (top holder distribution)
//   - LP lock verification (estimated based on creator holdings)
//
// The library supports three screening levels (Strict, Normal, Relaxed) with
// configurable thresholds for each check.
//
// Example usage:
//
//	guard := tokenguard.New(tokenguard.Config{
//	    BirdeyeAPIKey: "your-api-key",
//	})
//
//	result, err := guard.Screen(ctx, tokenMint, tokenguard.ScreeningLevelNormal)
//	if err != nil {
//	    return err
//	}
//
//	if !result.Passed {
//	    log.Printf("Token failed screening: %v", result.FailureReasons)
//	}
package tokenguard

import (
	"time"

	"github.com/shopspring/decimal"
)

// ScreeningLevel defines how strict token screening should be.
type ScreeningLevel string

// Screening level constants.
const (
	// ScreeningLevelStrict requires:
	//   - No mint authority
	//   - No freeze authority
	//   - Min liquidity: $50,000
	//   - Min LP locked: 80%
	//   - Max top 10 holders: 40%
	//   - Max single holder: 15%
	ScreeningLevelStrict ScreeningLevel = "strict"

	// ScreeningLevelNormal requires:
	//   - No mint authority
	//   - No freeze authority
	//   - Min liquidity: $20,000
	//   - Min LP locked: 50%
	//   - Max top 10 holders: 60%
	//   - Max single holder: 25%
	ScreeningLevelNormal ScreeningLevel = "normal"

	// ScreeningLevelRelaxed requires:
	//   - Mint authority allowed
	//   - No freeze authority
	//   - Min liquidity: $5,000
	//   - Min LP locked: 25%
	//   - Max top 10 holders: 75%
	//   - Max single holder: 35%
	ScreeningLevelRelaxed ScreeningLevel = "relaxed"
)

// ValidScreeningLevel checks if a screening level is valid.
func ValidScreeningLevel(level ScreeningLevel) bool {
	switch level {
	case ScreeningLevelStrict, ScreeningLevelNormal, ScreeningLevelRelaxed:
		return true
	default:
		return false
	}
}

// TokenScreeningResult contains the outcome of token analysis.
type TokenScreeningResult struct {
	// TokenMint is the token's mint address that was screened.
	TokenMint string `json:"tokenMint"`

	// Passed indicates whether the token passed all checks for the given level.
	Passed bool `json:"passed"`

	// Score is a 0-100 safety score, where higher = safer.
	// Starts at 100 and deducts points for each failed check.
	Score int `json:"score"`

	// Level is the screening level that was applied.
	Level ScreeningLevel `json:"level"`

	// Details contains detailed information about each check performed.
	Details ScreeningDetails `json:"details"`

	// FailureReasons lists human-readable reasons why checks failed.
	// Empty if Passed is true.
	FailureReasons []string `json:"failureReasons,omitempty"`

	// ScreenedAt is when the screening was performed.
	ScreenedAt time.Time `json:"screenedAt"`
}

// ScreeningDetails contains detailed information about each check performed.
type ScreeningDetails struct {
	// Authority checks
	HasMintAuthority   bool `json:"hasMintAuthority"`   // Token has active mint authority
	HasFreezeAuthority bool `json:"hasFreezeAuthority"` // Token has active freeze authority

	// Liquidity check
	LiquidityUSD decimal.Decimal `json:"liquidityUsd"` // Total liquidity in USD

	// LP lock check (estimated from creator holdings)
	LPLockedPct decimal.Decimal `json:"lpLockedPct"` // Estimated LP locked percentage

	// Holder concentration checks
	Top10HoldersPct decimal.Decimal `json:"top10HoldersPct"` // % held by top 10 holders
	TopHolderPct    decimal.Decimal `json:"topHolderPct"`    // % held by single top holder

	// Token-2022 specific features
	IsToken2022     bool `json:"isToken2022"`     // Uses Token-2022 program
	HasTransferFee  bool `json:"hasTransferFee"`  // Has transfer fee enabled
	NonTransferable bool `json:"nonTransferable"` // Token is non-transferable (soulbound)
	MutableMetadata bool `json:"mutableMetadata"` // Metadata can be changed
}

// ScreeningThresholds defines thresholds for each screening level.
//
// These can be customized per-screening call using ScreenWithThresholds,
// or use the preset levels (Strict/Normal/Relaxed).
type ScreeningThresholds struct {
	// RequireNoMintAuth requires that mint authority be revoked/disabled.
	RequireNoMintAuth bool

	// RequireNoFreezeAuth requires that freeze authority be revoked/disabled.
	RequireNoFreezeAuth bool

	// MinLiquidityUSD is the minimum required liquidity in USD.
	MinLiquidityUSD decimal.Decimal

	// MinLPLockedPct is the minimum estimated LP locked percentage.
	// This is estimated as 100% - creator percentage.
	MinLPLockedPct decimal.Decimal

	// MaxTop10HoldersPct is the maximum percentage that can be held by top 10 holders.
	MaxTop10HoldersPct decimal.Decimal

	// MaxTopHolderPct is the maximum percentage that can be held by a single holder.
	MaxTopHolderPct decimal.Decimal
}
