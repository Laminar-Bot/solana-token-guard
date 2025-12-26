# LAMINAR BOT - SPECIFICATION MASTER DOCUMENT

## HOW TO USE THIS DOCUMENT

This is a **master outline** for Claude Code to expand into full specifications. Each section has:
1. A summary of what needs to be specified
2. Key decisions already made
3. Expansion instructions for Claude Code

**For Claude Code:** When asked to expand a section, generate complete, production-ready specifications. No stubs. No placeholders. Full implementations.

---

# SYSTEM OVERVIEW

## Product Summary
Laminar Bot is a multi-tenant Solana copy-trading platform. Users watch successful wallets and automatically copy their trades with risk management and automated exits.

## Core Value Props
- **Copy Trading:** Auto-replicate trades from watched wallets
- **Risk Management:** Position limits, daily loss limits, exposure controls
- **Automated Exits:** Stop-loss, take-profit ladders, trailing stops, time stops, copy exits
- **Token Screening:** Reject rugs, honeypots, concentrated holdings
- **Analytics:** P&L tracking, wallet performance, win rates

## Target Scale
- V1: 10-50 invite-only users
- V2: 500+ users
- Trust Model: Users create dedicated bot wallets; Laminar holds keys securely in GCP Secret Manager

---

# ARCHITECTURE DECISIONS (LOCKED)

## Technology Stack

| Component | Technology | Rationale |
|-----------|------------|-----------|
| Trading Engine | Go 1.22 | Performance, concurrency, Solana SDK |
| User API | Go 1.22 (Chi) | Consistency with engine |
| Telegram Bot | Go 1.22 | Consistency, good Telegram libs |
| Frontend | SvelteKit 2 + TypeScript | Fast, simple, great DX |
| Database | PostgreSQL 16 | Reliable, feature-rich |
| Queue/Cache | Valkey 7 | Redis-compatible, open license |
| API Gateway | KrakenD | Fast, declarative JWT handling |
| Secret Storage | GCP Secret Manager | Secure key management |
| Hosting | GCP (Cloud Run + Cloud SQL) | Team familiarity |

## External Services

| Service | Purpose | API Type |
|---------|---------|----------|
| Helius | RPC, Webhooks, DAS API, Token Holders | REST + Webhooks |
| Jupiter | Swap quotes, swap execution, prices | REST |
| Birdeye | Token prices, liquidity info, wallet analysis | REST |

## Service Boundaries

```
┌─────────────────────────────────────────────────────────────────┐
│                         SERVICES                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │  FRONTEND   │  │  USER API   │  │  TELEGRAM   │             │
│  │  (SvelteKit)│  │    (Go)     │  │    BOT      │             │
│  │             │  │             │  │    (Go)     │             │
│  │ • Dashboard │  │ • Auth      │  │             │             │
│  │ • Settings  │  │ • CRUD      │  │ • Alerts    │             │
│  │ • Analytics │  │ • Analytics │  │ • Commands  │             │
│  └─────────────┘  └──────┬──────┘  └──────┬──────┘             │
│                          │                │                     │
│                          ▼                ▼                     │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              TRADING ENGINE (Go) - CORE SERVICE          │  │
│  ├──────────────────────────────────────────────────────────┤  │
│  │ • Webhook Handler (receives Helius webhooks)             │  │
│  │ • Copy Trade Worker (executes copy trades)               │  │
│  │ • Exit Check Worker (monitors positions, triggers exits) │  │
│  │ • Screener (validates tokens before buying)              │  │
│  │ • Risk Engine (enforces limits before trades)            │  │
│  │ • Executor (builds and sends Solana transactions)        │  │
│  └──────────────────────────────────────────────────────────┘  │
│                          │                                      │
│         ┌────────────────┼────────────────┐                    │
│         ▼                ▼                ▼                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │ PostgreSQL  │  │   Valkey    │  │ GCP Secret  │            │
│  │             │  │             │  │   Manager   │            │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

# SPECIFICATION SECTIONS

Each section below should be expanded by Claude Code into a complete specification file.

---

## SECTION 1: DOMAIN MODEL

**File:** `specs/01-domain-model.md`

**What to specify:**
- All entities with complete field definitions
- Value objects (UserID, WalletID, PositionID, etc.)
- Enums (Status types, ExitReason, Side, etc.)
- Business rules encoded in entities
- Validation rules

**Entities to define:**

### User
- ID, Email (optional), TelegramID (optional), BotWalletAddress
- Settings (nested): position sizing, risk limits, exit rules, screening level, notifications
- Status: active, paused, stopped (hit daily loss), disabled

### WatchedWallet
- ID, UserID, Address, Label
- Settings (overrides): position size, exit rules, copy delay, token filters
- Stats: trade counts, P&L, win rate, last trade timestamp
- Status: active, paused, removed

### Position
- ID, UserID, TokenAddress, TokenSymbol, TokenDecimals
- Entry: amount SOL, amount token, price, tx signature, timestamp
- Copy source: wallet ID, tx signature (nullable for manual trades)
- Current: amount token, price, value, unrealized P&L (SOL and %), high water mark
- Realized: total exited SOL, realized P&L
- TP levels hit tracking
- Status: open, closed
- Timestamps: created, updated, closed

### Trade (Immutable audit log)
- ID, UserID, PositionID
- Side: buy, sell
- Token info: address, symbol
- Amounts: token, SOL, price
- Fees: network fee, slippage
- Transaction: signature, status (pending, confirmed, failed)
- Copy source: wallet ID, tx signature (nullable)
- Exit reason (for sells): stop_loss, take_profit, trailing_stop, copy_exit, time_stop, manual, risk_limit
- Idempotency key
- Timestamp

### Token (Cached metadata)
- Address (PK), Symbol, Name, Decimals, LogoURL
- Screening results (cached): authorities, LP, holders, score, passed
- Last screened timestamp

### DailyUserStats
- UserID, Date (composite PK)
- Trade counts, buy/sell counts
- Volume bought/sold
- Realized P&L, fees
- Max positions held

**Exit Rules Structure:**
```
ExitRules {
  stopLossPct: decimal (e.g., -35)
  takeProfitLevels: [{triggerPct, sellPct}, ...]
  trailingStopEnabled: bool
  trailingStopPct: decimal
  trailingActivationPct: decimal
  copyExitEnabled: bool
  timeStopHours: int (0 = disabled)
  moonBagPct: decimal
}
```

---

## SECTION 2: DATABASE SCHEMA

**File:** `specs/02-database-schema.md`

**What to specify:**
- Complete PostgreSQL DDL (CREATE TABLE statements)
- All indexes
- Triggers for updated_at
- Wallet subscription sync trigger
- JSONB structures for settings, stats
- Views for analytics
- Enum types
- Constraints and foreign keys

**Tables:**
- users
- watched_wallets
- wallet_subscriptions (denormalized lookup)
- positions
- trades
- daily_user_stats
- token_cache
- processed_webhooks (idempotency)
- helius_subscriptions (webhook management)

---

## SECTION 3: SQLC QUERIES

**File:** `specs/03-sqlc-queries.md`

**What to specify:**
- sqlc.yaml configuration
- All SQL queries with sqlc annotations
- Query files organized by entity

**Query categories:**
- Users: CRUD, status updates, get by telegram/wallet
- Wallets: CRUD, get subscribers by address, stats updates
- Positions: CRUD, get open, get by token, update price, record partial exit
- Trades: create, get by idempotency key, get by position/user
- Daily Stats: upsert, increment buy/sell
- Tokens: upsert, get by address, update screening
- Webhooks: check processed, mark processed, cleanup

---

## SECTION 4: EXTERNAL API CLIENTS

**File:** `specs/04-external-clients.md`

**What to specify:**
- Complete Go implementations for each client
- All API methods with request/response types
- Error handling and retries
- Rate limiting awareness
- Metrics integration

### Helius Client
- Webhook signature validation (HMAC-SHA256)
- RPC methods: getBalance, getTokenAccountsByOwner, sendTransaction, getTransaction
- DAS API: getAsset (token metadata)
- Token holders endpoint
- Webhook management: create, update, delete

### Jupiter Client
- Quote API: get swap quotes
- Swap API: build swap transactions
- Price API: get token prices (single and batch)
- Helper functions: BuildBuyQuote, BuildSellQuote

### Birdeye Client
- Price API: single and multi-token prices
- Token info/overview
- Token security/liquidity info
- Wallet portfolio
- Top traders
- OHLCV data

---

## SECTION 5: TRADING ENGINE CORE

**File:** `specs/05-trading-engine.md`

**What to specify:**
- Complete package structure
- Main application entrypoint
- Dependency injection / initialization
- Graceful shutdown
- Health checks

### Webhook Handler
- HTTP handler for POST /webhooks/helius
- Signature validation
- Payload parsing (Helius enhanced transaction format)
- Swap detection and extraction
- Idempotency checking
- Subscriber lookup
- Job enqueueing

### Queue System
- Valkey client wrapper
- Queue definitions (copy_trade_jobs, exit_check_jobs, copy_exit_jobs)
- Producer: enqueue methods for each job type
- Consumer: dequeue with blocking
- Pub/Sub: notification events, config updates
- Cache: price caching with TTL

### Copy Trade Worker
- Concurrent job processing
- Full execution flow:
  1. Idempotency check
  2. Load user and wallet settings
  3. Check token filters
  4. Calculate position size
  5. Risk check
  6. Screen token
  7. Apply copy delay
  8. Execute swap
  9. Create position record
  10. Create trade record
  11. Update daily stats
  12. Publish notification
- Retry logic with backoff
- Error handling and notifications

### Exit Check Worker
- Scheduled loop (configurable interval)
- Batch load open positions
- Batch fetch prices
- Update position P&L and high water marks
- Evaluate exit rules for each position
- Execute exits when triggered
- Notification publishing

---

## SECTION 6: SCREENER

**File:** `specs/06-screener.md`

**What to specify:**
- Complete Screener implementation
- All check functions
- Configurable thresholds by level (strict, normal, relaxed)
- Caching strategy
- Result structure with scores

### Checks
1. Mint Authority: revoked?
2. Freeze Authority: revoked?
3. Liquidity: min LP value, position % of LP, LP locked %
4. Holder Concentration: top 10 %, max single holder %
5. Honeypot: simulation or heuristics

### Threshold Defaults
```
Strict:
  - Mint/freeze revoked: required
  - Min LP: 50 SOL
  - LP locked: required, 80%+
  - Top 10 holders: <30%
  - Max single holder: <10%
  - Position % of LP: <1%
  - Honeypot check: enabled

Normal:
  - Mint/freeze revoked: required
  - Min LP: 20 SOL
  - LP locked: not required
  - Top 10 holders: <50%
  - Max single holder: <20%
  - Position % of LP: <2%

Relaxed:
  - Mint/freeze revoked: not required
  - Min LP: 5 SOL
  - Top 10 holders: <70%
  - Max single holder: <30%
  - Position % of LP: <5%
```

---

## SECTION 7: RISK ENGINE

**File:** `specs/07-risk-engine.md`

**What to specify:**
- Complete RiskEngine implementation
- All check functions
- Position adjustment logic
- Daily loss tracking

### Pre-Buy Checks
1. User status: must be active
2. Daily loss limit: check and update status to stopped if hit
3. Max positions: count open positions
4. Existing token position: check max per token
5. Balance check: ensure sufficient SOL with reserve

### Result Structure
```
CheckResult {
  Approved: bool
  Reason: string (if rejected)
  Warnings: []string
  Adjustments: {
    AdjustedSizeSOL: decimal
    Reason: string
  }
}
```

---

## SECTION 8: EXIT ENGINE

**File:** `specs/08-exit-engine.md`

**What to specify:**
- Exit rule engine with priority system
- Complete implementation of each rule
- Moon bag handling
- Signal structure

### Exit Rules (by priority)
1. **Stop Loss (100):** Full exit when loss exceeds threshold
2. **Trailing Stop (90):** Exit on drawdown from high (respects moon bag)
3. **Take Profit (50+):** Partial exits at profit levels (respects moon bag)
4. **Time Stop (30):** Exit if held too long AND not winning
5. **Copy Exit (80):** Exit when copied wallet sells (separate worker)

### Signal Structure
```
Signal {
  ShouldExit: bool
  Reason: string
  SellPct: float64
  Priority: int
  Message: string
}
```

---

## SECTION 9: EXECUTOR

**File:** `specs/09-executor.md`

**What to specify:**
- Complete swap execution implementation
- Transaction building with solana-go
- Signing with keystore
- Confirmation waiting
- Slippage and priority fee handling

### ExecuteBuy Flow
1. Get user (for bot wallet address)
2. Get quote from Jupiter
3. Build swap transaction
4. Decode transaction bytes
5. Get private key from keystore
6. Sign transaction
7. Send via Helius RPC
8. Wait for confirmation
9. Calculate result (amounts, price, fees)

### ExecuteSell Flow
- Same as buy but:
  - Higher default slippage
  - Higher priority fee
  - Input is token, output is SOL

---

## SECTION 10: KEYSTORE

**File:** `specs/10-keystore.md`

**What to specify:**
- Keystore interface
- GCP Secret Manager implementation
- Key generation
- Security considerations

### Interface
```go
type Keystore interface {
  GetPrivateKey(ctx, userID) (PrivateKey, error)
  StorePrivateKey(ctx, userID, key) error
  DeletePrivateKey(ctx, userID) error
  GenerateWallet(ctx, userID) (publicKey string, error)
}
```

### Secret Naming
`projects/{project}/secrets/user-{userID}-bot-wallet`

---

## SECTION 11: USER API

**File:** `specs/11-user-api.md`

**What to specify:**
- Complete REST API specification
- All endpoints with request/response schemas
- Authentication (wallet signature + JWT)
- Error responses
- Rate limiting

### Endpoints

**Auth:**
- POST /auth/nonce - Get signing nonce
- POST /auth/wallet - Authenticate with signature
- POST /auth/refresh - Refresh JWT
- POST /auth/logout - Invalidate token

**Users:**
- GET /users/me
- PATCH /users/me
- GET /users/me/settings
- PATCH /users/me/settings
- POST /users/me/pause
- POST /users/me/resume

**Wallets:**
- GET /wallets
- POST /wallets
- GET /wallets/:id
- PATCH /wallets/:id
- DELETE /wallets/:id
- POST /wallets/:id/pause
- POST /wallets/:id/resume
- GET /wallets/:id/stats

**Positions:**
- GET /positions
- GET /positions/:id
- POST /positions/:id/close
- GET /positions/open
- GET /positions/closed

**Trades:**
- GET /trades
- GET /trades/:id

**Analytics:**
- GET /analytics/summary
- GET /analytics/daily
- GET /analytics/wallets
- GET /analytics/tokens

**Bot Wallet:**
- GET /bot-wallet
- GET /bot-wallet/balance
- POST /bot-wallet/generate

---

## SECTION 12: TELEGRAM BOT

**File:** `specs/12-telegram-bot.md`

**What to specify:**
- Bot setup and lifecycle
- Command handlers
- Notification subscriber
- Message templates
- Inline keyboards

### Commands
- /start - Initialize, link account
- /status - Show trading status, balance, positions summary
- /positions - List open positions with P&L
- /wallets - List watched wallets
- /pause - Pause all trading
- /resume - Resume trading
- /settings - Show/edit settings
- /help - Help message

### Notification Types
- Buy executed
- Sell executed / exit triggered
- Risk limit hit
- Error
- Daily summary

### Message Templates
(Provide complete formatted templates for each notification type)

---

## SECTION 13: FRONTEND

**File:** `specs/13-frontend.md`

**What to specify:**
- Complete SvelteKit project structure
- All routes and pages
- Key components
- API client
- Auth flow (wallet connect)
- Stores (auth, positions, wallets)
- Type definitions

### Pages
- / (Dashboard)
- /auth (Login)
- /positions
- /positions/[id]
- /wallets
- /wallets/new
- /wallets/[id]
- /trades
- /analytics
- /settings

### Key Components
- PortfolioSummary
- PositionCard / PositionTable
- WalletCard / WalletForm
- PnLChart
- ExitRulesForm
- TokenIcon
- PnLBadge

---

## SECTION 14: OBSERVABILITY

**File:** `specs/14-observability.md`

**What to specify:**
- Structured logging (zerolog)
- Prometheus metrics
- Health checks
- Tracing (OpenTelemetry basics)

### Metrics
- Webhooks: received (by status), processing duration
- Copy trades: attempted (by status), latency
- Exits: triggered (by reason)
- Swaps: executed (by side), latency
- External APIs: latency, errors
- Positions: open count
- Queue: depth by queue

### Log Context Builders
- WithUserID
- WithTx
- WithPosition
- WithTraceID

### Health Checks
- Database ping
- Valkey ping
- Helius RPC check
- Returns: healthy, degraded, unhealthy with per-component status

---

## SECTION 15: SECURITY

**File:** `specs/15-security.md`

**What to specify:**
- STRIDE threat model
- Authentication design
- Key management
- Tenant isolation
- Input validation
- What to never log

### Key Security Rules
1. Private keys: GCP Secret Manager only, never in logs/errors
2. Tenant isolation: user_id in every query
3. Webhook validation: HMAC-SHA256
4. JWT: short expiry, refresh tokens
5. Input validation: all external inputs

---

## SECTION 16: TESTING

**File:** `specs/16-testing.md`

**What to specify:**
- Test pyramid strategy
- Unit test examples for key components
- Integration test setup (testcontainers)
- Mock strategies for external APIs
- Key test cases for each component

### Coverage Priorities
1. Exit rules: all trigger conditions, edge cases
2. Risk engine: all limit types, adjustments
3. Screener: each check independently
4. P&L calculations: accuracy
5. Webhook parsing: various transaction formats

---

## SECTION 17: DEPLOYMENT

**File:** `specs/17-deployment.md`

**What to specify:**
- Dockerfile (multi-stage)
- docker-compose.yml (local dev)
- Makefile
- Configuration files
- Environment variables
- GCP deployment (Cloud Run)
- CI/CD pipeline (GitHub Actions)

---

## SECTION 18: ADRS

**File:** `specs/18-adrs.md`

**What to specify:**
- Architecture Decision Records for key choices

### ADRs to Document
1. Go for Trading Engine
2. Valkey for Queues (vs RabbitMQ, Kafka)
3. Webhook-Driven Copy Trading (vs Polling)
4. Multi-Tenant Single Database
5. GCP Secret Manager for Keys
6. JWT + Wallet Signature Auth
7. sqlc over ORM
8. Chi HTTP Framework

---

## SECTION 19: IMPLEMENTATION PHASES

**File:** `specs/19-phases.md`

**What to specify:**
- Phased implementation plan
- MVP scope
- Timeline estimates
- Dependencies between phases

### Suggested Phases

**Phase 1: Minimal Copy Trading (2 weeks)**
- Webhook handler + basic copy trade
- No screening, no exits
- Single user testing

**Phase 2: Screening + Risk (2-3 weeks)**
- Token screening (all checks)
- Risk engine
- Multi-user support

**Phase 3: Exit Rules (1-2 weeks)**
- Position monitoring
- All exit rules
- Notifications

**Phase 4: User API + Frontend (2-3 weeks)**
- Complete REST API
- SvelteKit frontend
- Wallet auth

**Phase 5: Telegram Bot (1 week)**
- Notifications
- Commands

**Phase 6: Polish (1-2 weeks)**
- Analytics
- Performance optimization
- Documentation

---

# INSTRUCTIONS FOR CLAUDE CODE

When asked to expand any section:

1. **Read this master document first** to understand context and decisions
2. **Generate complete, production-ready specifications** - no stubs, no "TODO"
3. **Include full code implementations** where specified
4. **Use consistent naming** across all specs
5. **Cross-reference** other sections when relevant
6. **Include edge cases** and error handling

Example prompt: "Expand SECTION 5: TRADING ENGINE CORE into a complete specification"

The output should be a complete markdown file that could be handed to a developer to implement.