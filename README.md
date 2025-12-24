# solana-token-guard

[![Go Reference](https://pkg.go.dev/badge/github.com/Laminar-Bot/solana-token-guard.svg)](https://pkg.go.dev/github.com/Laminar-Bot/solana-token-guard)
[![Go Report Card](https://goreportcard.com/badge/github.com/Laminar-Bot/solana-token-guard)](https://goreportcard.com/report/github.com/Laminar-Bot/solana-token-guard)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A token safety screening library for Solana. Check tokens for common red flags before trading.

> ⚠️ **Disclaimer**: This library helps identify common risk patterns but cannot guarantee token safety. Always do your own research.

## Features

- 🔐 **Authority Checks** - Mint and freeze authority status
- 💧 **Liquidity Analysis** - LP size, locked percentage
- 👥 **Holder Concentration** - Top holder distribution
- 🍯 **Honeypot Detection** - Basic sellability checks
- 📊 **Safety Score** - 0-100 risk score
- ⚙️ **Configurable Thresholds** - Strict, normal, relaxed presets

## Installation
```bash
go get github.com/Laminar-Bot/solana-token-guard
```

## Quick Start
```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Laminar-Bot/solana-token-guard"
    "github.com/Laminar-Bot/helius-go"
    "github.com/Laminar-Bot/birdeye-go"
)

func main() {
    // Initialize with data providers
    guard := tokenguard.New(tokenguard.Config{
        Helius:  helius.NewClient(helius.Config{APIKey: "..."}),
        Birdeye: birdeye.NewClient(birdeye.Config{APIKey: "..."}),
    })

    // Screen a token
    result, err := guard.Screen(context.Background(), "TokenMintAddress...", tokenguard.LevelNormal)
    if err != nil {
        log.Fatal(err)
    }

    if result.Passed {
        fmt.Printf("✅ Token passed screening (score: %d/100)\n", result.Score)
    } else {
        fmt.Printf("❌ Token failed screening\n")
        for _, check := range result.FailedChecks {
            fmt.Printf("   - %s: %s\n", check.Name, check.Message)
        }
    }
}
```

## Screening Checks

| Check | Description |
|-------|-------------|
| `mint_authority` | Is mint authority revoked? (can't print more tokens) |
| `freeze_authority` | Is freeze authority revoked? (can't freeze wallets) |
| `liquidity` | Is there enough LP? Is your trade size safe vs LP? |
| `lp_locked` | Is liquidity locked/burned? |
| `holder_concentration` | Are tokens distributed or concentrated? |
| `honeypot` | Basic sellability heuristics |

## Preset Levels

### Strict
Best for automated trading, rejects more tokens.
```go
result, _ := guard.Screen(ctx, token, tokenguard.LevelStrict)
```
- Mint/freeze authority: **must be revoked**
- Min LP: **50 SOL**
- LP locked: **required, 80%+**
- Top 10 holders: **<30%**
- Max single holder: **<10%**

### Normal (Default)
Balanced for most use cases.
```go
result, _ := guard.Screen(ctx, token, tokenguard.LevelNormal)
```
- Mint/freeze authority: **must be revoked**
- Min LP: **20 SOL**
- LP locked: **not required**
- Top 10 holders: **<50%**
- Max single holder: **<20%**

### Relaxed
For experienced traders, accepts more risk.
```go
result, _ := guard.Screen(ctx, token, tokenguard.LevelRelaxed)
```
- Mint/freeze authority: **not required**
- Min LP: **5 SOL**
- Top 10 holders: **<70%**
- Max single holder: **<30%**

## Custom Thresholds
```go
result, _ := guard.ScreenWithThresholds(ctx, token, tokenguard.Thresholds{
    RequireMintRevoked:   true,
    RequireFreezeRevoked: true,
    MinLPValueSOL:        decimal.NewFromFloat(100),
    RequireLPLocked:      true,
    MinLPLockedPct:       decimal.NewFromFloat(90),
    MaxTop10HolderPct:    decimal.NewFromFloat(25),
    MaxSingleHolderPct:   decimal.NewFromFloat(5),
    MaxPositionPctOfLP:   decimal.NewFromFloat(0.5),
})
```

## Position Size Check

Check if your trade size is safe relative to liquidity:
```go
result, _ := guard.ScreenWithPositionSize(ctx, token, tokenguard.LevelNormal, 
    decimal.NewFromFloat(2.0)) // 2 SOL position

// Will fail if 2 SOL > threshold % of LP
```

## Result Structure
```go
type Result struct {
    Passed       bool      // Overall pass/fail
    Score        int       // 0-100 safety score
    Checks       []Check   // All checks performed
    FailedChecks []Check   // Only failed checks
    Warnings     []string  // Non-fatal warnings
}

type Check struct {
    Name    string      // e.g., "mint_authority"
    Passed  bool
    Value   interface{} // Actual value found
    Message string      // Human-readable result
}
```

## Caching

Results are cached to avoid redundant API calls:
```go
guard := tokenguard.New(tokenguard.Config{
    Helius:   heliusClient,
    Birdeye:  birdeyeClient,
    CacheTTL: 30 * time.Minute, // default
})

// Force fresh data
result, _ := guard.Screen(ctx, token, level, tokenguard.WithNoCache())
```

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) first.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Related Projects

- [helius-go](https://github.com/Laminar-Bot/helius-go) - Helius API client
- [birdeye-go](https://github.com/Laminar-Bot/birdeye-go) - Birdeye API client
- [jupiter-go](https://github.com/Laminar-Bot/jupiter-go) - Jupiter API client
