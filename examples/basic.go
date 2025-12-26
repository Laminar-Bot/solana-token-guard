package main

import (
	"context"
	"fmt"
	"log"
	"time"

	birdeye "github.com/Laminar-Bot/birdeye-go"
	tokenguard "github.com/Laminar-Bot/solana-token-guard"
	"go.uber.org/zap"
)

func main() {
	// Set up logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Create Birdeye client
	birdeyeClient, err := birdeye.NewClient(birdeye.Config{
		APIKey: "your-birdeye-api-key-here",
		Logger: logger,
	})
	if err != nil {
		log.Fatalf("failed to create birdeye client: %v", err)
	}

	// Create token screener with in-memory cache
	screener, err := tokenguard.New(tokenguard.Config{
		SecurityProvider: birdeyeClient,
		OverviewProvider: birdeyeClient,
		Cache: tokenguard.NewInMemoryCache(tokenguard.InMemoryCacheConfig{
			TTL:             5 * time.Minute,
			MaxSize:         10000,
			CleanupInterval: 1 * time.Minute,
		}),
		Logger: logger,
	})
	if err != nil {
		log.Fatalf("failed to create screener: %v", err)
	}

	// Screen a token (example: USDC)
	ctx := context.Background()
	tokenMint := "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"

	result, err := screener.Screen(ctx, tokenMint, tokenguard.ScreeningLevelNormal)
	if err != nil {
		log.Fatalf("screening failed: %v", err)
	}

	// Print results
	fmt.Printf("Token Screening Results for %s\n", result.TokenMint)
	fmt.Printf("=====================================\n")
	fmt.Printf("Passed: %v\n", result.Passed)
	fmt.Printf("Score: %d/100\n", result.Score)
	fmt.Printf("Level: %s\n", result.Level)
	fmt.Printf("\n")

	fmt.Printf("Details:\n")
	fmt.Printf("  Has Mint Authority: %v\n", result.Details.HasMintAuthority)
	fmt.Printf("  Has Freeze Authority: %v\n", result.Details.HasFreezeAuthority)
	fmt.Printf("  Liquidity USD: $%s\n", result.Details.LiquidityUSD.StringFixed(2))
	fmt.Printf("  LP Locked: %s%%\n", result.Details.LPLockedPct.StringFixed(2))
	fmt.Printf("  Top 10 Holders: %s%%\n", result.Details.Top10HoldersPct.StringFixed(2))
	fmt.Printf("  Top Holder: %s%%\n", result.Details.TopHolderPct.StringFixed(2))
	fmt.Printf("\n")

	if len(result.FailureReasons) > 0 {
		fmt.Printf("Failure Reasons:\n")
		for _, reason := range result.FailureReasons {
			fmt.Printf("  - %s\n", reason)
		}
	}
}
