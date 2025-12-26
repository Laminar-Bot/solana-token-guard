# LAMINAR BOT - TECHNICAL REFERENCE DATA

This file contains locked-in technical decisions and code patterns for reference when expanding specs.

---

## GO MODULE

```go
module github.com/laminar/trading-engine

go 1.22

require (
    github.com/go-chi/chi/v5 v5.0.12
    github.com/go-chi/cors v1.2.1
    github.com/jackc/pgx/v5 v5.5.5
    github.com/valkey-io/valkey-go v1.0.0
    github.com/gagliardetto/solana-go v1.10.0
    github.com/rs/zerolog v1.32.0
    github.com/prometheus/client_golang v1.19.0
    github.com/spf13/viper v1.18.2
    github.com/go-playground/validator/v10 v10.19.0
    github.com/google/uuid v1.6.0
    github.com/shopspring/decimal v1.3.1
    cloud.google.com/go/secretmanager v1.11.5
    github.com/go-telegram/bot v1.1.0
    github.com/stretchr/testify v1.9.0
    github.com/testcontainers/testcontainers-go v0.29.1
)
```

---

## PACKAGE STRUCTURE

```
trading-engine/
├── cmd/engine/main.go
├── internal/
│   ├── config/
│   ├── server/
│   ├── webhook/
│   ├── queue/
│   ├── worker/
│   ├── domain/{user,wallet,position,trade,token}/
│   ├── screener/
│   ├── risk/
│   ├── exit/
│   ├── executor/
│   ├── pricer/
│   ├── keystore/
│   ├── external/{helius,jupiter,birdeye}/
│   ├── notification/
│   └── observability/
├── pkg/{money,solana,idempotency,errors}/
├── db/migrations/
├── sqlc/
├── deployments/
├── configs/
└── Makefile
```

---

## VALKEY QUEUE KEYS

```go
const (
    QueueCopyTrade   = "laminar:queue:copy_trade"
    QueueExitCheck   = "laminar:queue:exit_check"
    QueueCopyExit    = "laminar:queue:copy_exit"
    
    ChannelNotifications = "laminar:notifications"
    ChannelConfigUpdates = "laminar:config_updates"
    
    CachePricePrefix = "laminar:cache:price:"
    CacheTokenPrefix = "laminar:cache:token:"
)
```

---

## API ENDPOINTS SUMMARY

```
# Auth
POST /auth/nonce
POST /auth/wallet
POST /auth/refresh
POST /auth/logout

# Users
GET    /users/me
PATCH  /users/me
GET    /users/me/settings
PATCH  /users/me/settings
POST   /users/me/pause
POST   /users/me/resume

# Wallets
GET    /wallets
POST   /wallets
GET    /wallets/:id
PATCH  /wallets/:id
DELETE /wallets/:id
POST   /wallets/:id/pause
POST   /wallets/:id/resume
GET    /wallets/:id/stats

# Positions
GET    /positions
GET    /positions/:id
POST   /positions/:id/close
GET    /positions/open
GET    /positions/closed

# Trades
GET    /trades
GET    /trades/:id

# Analytics
GET    /analytics/summary
GET    /analytics/daily
GET    /analytics/wallets
GET    /analytics/tokens

# Bot Wallet
GET    /bot-wallet
GET    /bot-wallet/balance
POST   /bot-wallet/generate

# Trading Engine (internal)
POST   /webhooks/helius
GET    /health
GET    /health/ready
GET    /health/live
GET    /metrics
```

---

## EXTERNAL API BASE URLS

```
Helius RPC: https://mainnet.helius-rpc.com/?api-key={key}
Helius API: https://api.helius.xyz/v0
Jupiter:    https://quote-api.jup.ag/v6
Jupiter Price: https://price.jup.ag/v6
Birdeye:    https://public-api.birdeye.so
```

---

## SOLANA CONSTANTS

```go
const (
    NativeSOLMint = "So11111111111111111111111111111111111111112"
    USDCMint      = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
    LamportsPerSOL = 1_000_000_000
)
```

---

## EXIT RULE PRIORITIES

```
Stop Loss:      100 (highest - always executes first)
Trailing Stop:   90
Copy Exit:       80
Take Profit:     50+ (higher TP levels get higher priority)
Time Stop:       30 (lowest - only if not winning)
```

---

## SCREENING THRESHOLDS

### Strict
- Mint authority revoked: required
- Freeze authority revoked: required
- Min LP: 50 SOL
- LP locked: required, 80%+
- Top 10 holders: <30%
- Max single holder: <10%
- Position % of LP: <1%
- Honeypot check: enabled

### Normal
- Mint authority revoked: required
- Freeze authority revoked: required
- Min LP: 20 SOL
- LP locked: not required
- Top 10 holders: <50%
- Max single holder: <20%
- Position % of LP: <2%
- Honeypot check: disabled

### Relaxed
- Mint authority revoked: not required
- Freeze authority revoked: not required
- Min LP: 5 SOL
- Top 10 holders: <70%
- Max single holder: <30%
- Position % of LP: <5%
- Honeypot check: disabled

---

## DEFAULT USER SETTINGS

```json
{
  "defaultPositionSizeSOL": "0.5",
  "defaultPositionSizePct": "5",
  "usePercentageSizing": false,
  "maxPositions": 10,
  "maxPositionPerToken": "2",
  "dailyLossLimitSOL": "5",
  "dailyLossLimitPct": "20",
  "exitRules": {
    "stopLossPct": "-35",
    "takeProfitLevels": [
      {"triggerPct": "50", "sellPct": "25"},
      {"triggerPct": "100", "sellPct": "25"},
      {"triggerPct": "200", "sellPct": "25"}
    ],
    "trailingStopEnabled": true,
    "trailingStopPct": "20",
    "trailingActivationPct": "50",
    "copyExitEnabled": true,
    "timeStopHours": 0,
    "moonBagPct": "25"
  },
  "screeningLevel": "normal",
  "notifyOnBuy": true,
  "notifyOnSell": true,
  "notifyOnError": true
}
```

---

## ENUM TYPES (PostgreSQL)

```sql
CREATE TYPE user_status AS ENUM ('active', 'paused', 'stopped', 'disabled');
CREATE TYPE wallet_status AS ENUM ('active', 'paused', 'removed');
CREATE TYPE position_status AS ENUM ('open', 'closed');
CREATE TYPE trade_side AS ENUM ('buy', 'sell');
CREATE TYPE tx_status AS ENUM ('pending', 'confirmed', 'failed');
CREATE TYPE exit_reason AS ENUM (
    'stop_loss', 
    'take_profit', 
    'trailing_stop', 
    'copy_exit', 
    'time_stop', 
    'manual', 
    'risk_limit'
);
CREATE TYPE screening_level AS ENUM ('strict', 'normal', 'relaxed');
```

---

## NOTIFICATION EVENT TYPES

```go
const (
    NotifyBuyExecuted   = "buy_executed"
    NotifySellExecuted  = "sell_executed"
    NotifyExitTriggered = "exit_triggered"
    NotifyRiskLimitHit  = "risk_limit_hit"
    NotifyError         = "error"
    NotifyWalletAlert   = "wallet_alert"
    NotifyDailySummary  = "daily_summary"
)
```

---

## TELEGRAM COMMANDS

```
/start    - Initialize bot, link Telegram account
/status   - Show trading status, balance, open positions
/positions - List open positions with P&L
/wallets  - List watched wallets
/pause    - Pause all trading
/resume   - Resume trading
/settings - Show/edit settings
/help     - Show help message
```

---

## HELIUS WEBHOOK PAYLOAD (Enhanced Transaction)

```json
{
  "signature": "5xyz...",
  "slot": 123456789,
  "timestamp": 1703289600,
  "type": "SWAP",
  "source": "JUPITER",
  "fee": 5000,
  "feePayer": "DYw8j...",
  "nativeTransfers": [
    {
      "fromUserAccount": "DYw8j...",
      "toUserAccount": "...",
      "amount": 500000000
    }
  ],
  "tokenTransfers": [
    {
      "fromUserAccount": "...",
      "toUserAccount": "DYw8j...",
      "mint": "DezXA...",
      "tokenAmount": 1000000000,
      "tokenStandard": "Fungible"
    }
  ],
  "events": {
    "swap": {
      "nativeInput": {"account": "...", "amount": "500000000"},
      "tokenOutputs": [{"mint": "...", "rawTokenAmount": "..."}]
    }
  }
}
```

---

## JUPITER QUOTE RESPONSE

```json
{
  "inputMint": "So111...",
  "inAmount": "500000000",
  "outputMint": "DezXA...",
  "outAmount": "1000000000000",
  "otherAmountThreshold": "990000000000",
  "swapMode": "ExactIn",
  "slippageBps": 100,
  "priceImpactPct": "0.15",
  "routePlan": [...],
  "contextSlot": 123456789,
  "timeTaken": 0.05
}
```

---

## JUPITER SWAP RESPONSE

```json
{
  "swapTransaction": "base64-encoded-transaction...",
  "lastValidBlockHeight": 123456789,
  "prioritizationFeeLamports": 50000
}
```

---

## BIRDEYE TOKEN SECURITY RESPONSE

```json
{
  "success": true,
  "data": {
    "ownerAddress": "...",
    "creatorAddress": "...",
    "mintAuthority": null,
    "freezeAuthority": null,
    "totalLiquidity": 50000,
    "totalLiquidityUSD": 10000000,
    "lpBurned": true,
    "lpBurnedPct": 95.5,
    "top10HolderPercent": 25.3,
    "isToken2022": false,
    "isMutable": false
  }
}
```

---

## METRICS NAMES

```
laminar_webhooks_received_total{status}
laminar_webhook_processing_seconds
laminar_copy_trades_total{status}
laminar_copy_trade_latency_seconds
laminar_exits_triggered_total{reason}
laminar_swaps_executed_total{side}
laminar_swap_latency_seconds{side}
laminar_external_api_latency_seconds{service,operation}
laminar_external_api_errors_total{service,error_type}
laminar_open_positions{user_id}
laminar_queue_depth{queue}
laminar_screening_duration_seconds
laminar_http_request_duration_seconds{method,path,status}
laminar_http_requests_total{method,path,status}
```

---

## ERROR CODES

```go
const (
    ErrCodeInvalidInput     = "INVALID_INPUT"
    ErrCodeUnauthorized     = "UNAUTHORIZED"
    ErrCodeForbidden        = "FORBIDDEN"
    ErrCodeNotFound         = "NOT_FOUND"
    ErrCodeConflict         = "CONFLICT"
    ErrCodeRateLimited      = "RATE_LIMITED"
    ErrCodeInternalError    = "INTERNAL_ERROR"
    ErrCodeExternalService  = "EXTERNAL_SERVICE_ERROR"
    ErrCodeInsufficientFunds = "INSUFFICIENT_FUNDS"
    ErrCodeRiskLimitExceeded = "RISK_LIMIT_EXCEEDED"
    ErrCodeScreeningFailed  = "SCREENING_FAILED"
    ErrCodeTransactionFailed = "TRANSACTION_FAILED"
)
```

---

## FRONTEND ROUTES

```
/               - Dashboard
/auth           - Login with wallet
/auth/callback  - OAuth callback (if needed)
/positions      - Positions list
/positions/[id] - Position detail
/wallets        - Wallets list
/wallets/new    - Add wallet
/wallets/[id]   - Wallet detail
/trades         - Trade history
/analytics      - Analytics dashboard
/settings       - User settings
```

---

## FRONTEND DEPENDENCIES

```json
{
  "dependencies": {
    "@tanstack/svelte-query": "^5.24.0",
    "bits-ui": "^0.21.0",
    "chart.js": "^4.4.2",
    "clsx": "^2.1.0",
    "lightweight-charts": "^4.1.3",
    "lucide-svelte": "^0.344.0",
    "mode-watcher": "^0.3.0",
    "svelte-sonner": "^0.3.19",
    "tailwind-merge": "^2.2.1",
    "tailwind-variants": "^0.2.0",
    "zod": "^3.22.4"
  }
}
```

---

## GCP RESOURCES

```
Cloud Run:
  - laminar-trading-engine
  - laminar-user-api
  - laminar-telegram-bot
  - laminar-api-gateway

Cloud SQL:
  - laminar-postgres (PostgreSQL 16)

Memorystore:
  - laminar-valkey (Valkey 7)

Secret Manager:
  - user-{userID}-bot-wallet (per user)
  - helius-api-key
  - helius-webhook-secret
  - birdeye-api-key
  - telegram-bot-token
  - jwt-signing-key

Cloud Storage:
  - laminar-frontend (static hosting)
```