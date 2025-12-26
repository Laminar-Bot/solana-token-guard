# CryptoRugMunch Implementation Guide

**Status**: ✅ Complete
**Last Updated**: 2025-01-19

This reference provides step-by-step workflows for implementing CryptoRugMunch from local development to production deployment.

---

## Table of Contents

1. [Environment Setup](#environment-setup)
2. [Core Implementation Workflow](#core-implementation-workflow)
3. [Testing Strategy](#testing-strategy)
4. [Deployment Process](#deployment-process)
5. [Monitoring & Operations](#monitoring--operations)
6. [Common Implementation Patterns](#common-implementation-patterns)

---

## Environment Setup

### Prerequisites Installation

```bash
# Required versions
node --version    # v20.x or higher
psql --version    # PostgreSQL 15+
redis-cli --version  # Redis 7+
docker --version  # Docker 24+

# Install dependencies
npm install -g pnpm  # Package manager
```

### Initial Project Setup

**Step 1: Clone and Install**

```bash
git clone https://github.com/Laminar-Bot/rug-muncher.git
cd rug-muncher
pnpm install
```

**Step 2: Environment Configuration**

```bash
# Copy environment template
cp .env.example .env

# Edit with your values
nano .env
```

**Required environment variables** (see `docs/03-TECHNICAL/operations/environment-variables.md` for complete list):

```bash
# Database
DATABASE_URL="postgresql://user:password@localhost:5432/rugmunch_dev"

# Redis
REDIS_URL="redis://localhost:6379"

# Telegram Bot
TELEGRAM_BOT_TOKEN="your_bot_token_from_BotFather"

# Blockchain APIs
HELIUS_API_KEY="your_helius_key"
BIRDEYE_API_KEY="your_birdeye_key"
RUGCHECK_API_KEY="your_rugcheck_key"

# Monitoring (optional for local dev)
DATADOG_API_KEY="your_datadog_key"
SENTRY_DSN="your_sentry_dsn"
```

**Step 3: Infrastructure Setup**

```bash
# Start PostgreSQL and Redis via Docker
docker-compose up -d postgres redis

# Verify services are running
docker ps

# Expected output:
# CONTAINER ID   IMAGE         PORTS
# abc123         postgres:15   0.0.0.0:5432->5432/tcp
# def456         redis:7       0.0.0.0:6379->6379/tcp
```

**Step 4: Database Initialization**

```bash
# Run Prisma migrations
npx prisma migrate dev

# Seed database with test data (optional)
npx prisma db seed

# Open Prisma Studio to verify
npx prisma studio
# Opens at http://localhost:5555
```

**Step 5: Verify Setup**

```bash
# Run tests to verify everything works
pnpm test:unit

# Start development server
pnpm dev

# In another terminal, start worker
pnpm worker:dev

# Expected output:
# ✓ Fastify server listening on http://localhost:3000
# ✓ Worker connected to Redis
# ✓ BullMQ queue initialized
```

---

## Core Implementation Workflow

### Phase 1: Core Infrastructure (Week 1-2)

#### 1.1 Database Schema Implementation

**File**: `prisma/schema.prisma`

```bash
# Already exists from docs, verify it matches:
# - User model
# - Scan model
# - Subscription model
# - Alert model
# - Badge model

# Run migration
npx prisma migrate dev --name init_schema

# Generate Prisma Client
npx prisma generate
```

**Verification**:
```bash
npx prisma studio
# Verify all tables exist: User, Scan, Subscription, Alert, Badge
```

#### 1.2 Fastify API Skeleton

**File**: `src/app.ts`

```typescript
import Fastify from 'fastify';
import { registerRoutes } from './routes';
import { setupMonitoring } from './config/monitoring';
import { setupErrorHandling } from './config/errors';

export async function createApp() {
  const app = Fastify({
    logger: {
      level: process.env.LOG_LEVEL || 'info',
    },
  });

  // Register plugins
  await app.register(import('@fastify/cors'));
  await app.register(import('@fastify/helmet'));
  await app.register(import('@fastify/rate-limit'), {
    max: 100,
    timeWindow: '1 minute',
  });

  // Setup monitoring
  setupMonitoring(app);

  // Setup error handling
  setupErrorHandling(app);

  // Register routes
  await registerRoutes(app);

  return app;
}

// src/server.ts
import { createApp } from './app';

async function start() {
  const app = await createApp();

  try {
    await app.listen({
      port: Number(process.env.PORT) || 3000,
      host: '0.0.0.0',
    });
  } catch (err) {
    app.log.error(err);
    process.exit(1);
  }
}

start();
```

**Verification**:
```bash
pnpm dev
curl http://localhost:3000/health
# Expected: {"status":"ok"}
```

#### 1.3 Redis & BullMQ Setup

**File**: `src/config/queue.ts`

```typescript
import { Queue, Worker } from 'bullmq';
import IORedis from 'ioredis';

export const connection = new IORedis(process.env.REDIS_URL!, {
  maxRetriesPerRequest: null,
});

export const scanQueue = new Queue('token-scan', {
  connection,
  defaultJobOptions: {
    attempts: 3,
    backoff: {
      type: 'exponential',
      delay: 2000,
    },
    removeOnComplete: {
      count: 1000,
      age: 24 * 3600, // 24 hours
    },
    removeOnFail: {
      count: 5000,
      age: 7 * 24 * 3600, // 7 days
    },
  },
});
```

**File**: `src/workers/scan-worker.ts`

```typescript
import { Worker } from 'bullmq';
import { connection } from '../config/queue';
import { processScan } from '../modules/scan/scan.processor';

export const scanWorker = new Worker('token-scan', processScan, {
  connection,
  concurrency: Number(process.env.WORKER_CONCURRENCY) || 4,
});

scanWorker.on('completed', (job) => {
  console.log(`Job ${job.id} completed`);
});

scanWorker.on('failed', (job, err) => {
  console.error(`Job ${job?.id} failed:`, err);
});

// src/workers/index.ts
import { scanWorker } from './scan-worker';

process.on('SIGTERM', async () => {
  await scanWorker.close();
  process.exit(0);
});
```

**Verification**:
```bash
# Start worker
pnpm worker:dev

# In another terminal, add test job
node -e "
const { Queue } = require('bullmq');
const queue = new Queue('token-scan', {
  connection: { host: 'localhost', port: 6379 }
});
queue.add('test', { tokenAddress: 'test123' });
"

# Check worker logs for job processing
```

---

### Phase 2: Risk Scoring Engine (Week 3-4)

#### 2.1 Blockchain API Providers

**File**: `src/modules/scan/providers/helius.provider.ts`

```typescript
import axios from 'axios';
import { TokenMetadata } from '../scan.types';

export class HeliusProvider {
  private apiKey: string;
  private baseUrl = 'https://api.helius.xyz/v0';

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async getTokenMetadata(tokenAddress: string): Promise<TokenMetadata> {
    try {
      const response = await axios.get(
        `${this.baseUrl}/tokens/metadata`,
        {
          params: { 'api-key': this.apiKey },
          data: { mintAccounts: [tokenAddress] },
          timeout: 5000,
        }
      );

      return this.transformMetadata(response.data[0]);
    } catch (error) {
      throw new Error(`Helius API error: ${error.message}`);
    }
  }

  private transformMetadata(raw: any): TokenMetadata {
    return {
      address: raw.account,
      name: raw.onChainMetadata?.metadata?.data?.name,
      symbol: raw.onChainMetadata?.metadata?.data?.symbol,
      decimals: raw.onChainMetadata?.decimals,
      supply: raw.onChainMetadata?.supply,
      mintAuthority: raw.onChainMetadata?.mintAuthority,
      freezeAuthority: raw.onChainMetadata?.freezeAuthority,
    };
  }
}
```

**Similar providers**:
- `birdeye.provider.ts` - Liquidity data
- `rugcheck.provider.ts` - Honeypot detection

**See**: `docs/03-TECHNICAL/integrations/blockchain-api-integration.md` for complete provider implementations

#### 2.2 Risk Scoring Algorithm

**File**: `src/modules/scan/risk-scoring/risk-calculator.ts`

```typescript
import { RiskFactors, RiskScore } from '../scan.types';

export function calculateRiskScore(factors: RiskFactors): RiskScore {
  let score = 100;
  const breakdown: Record<string, { points: number; reason: string }> = {};

  // Metric 1: Total Liquidity (20% weight)
  if (factors.liquidity.usd < 5_000) {
    const deduction = 25;
    score -= deduction;
    breakdown.liquidity = {
      points: -deduction,
      reason: `Liquidity $${factors.liquidity.usd.toLocaleString()} is extremely low (< $5K)`,
    };
  } else if (factors.liquidity.usd < 20_000) {
    const deduction = 15;
    score -= deduction;
    breakdown.liquidity = {
      points: -deduction,
      reason: `Liquidity $${factors.liquidity.usd.toLocaleString()} is low (< $20K)`,
    };
  } else if (factors.liquidity.usd < 50_000) {
    const deduction = 5;
    score -= deduction;
    breakdown.liquidity = {
      points: -deduction,
      reason: `Liquidity $${factors.liquidity.usd.toLocaleString()} is moderate`,
    };
  }

  // Metric 2: LP Lock Status (15% weight)
  if (!factors.liquidity.locked) {
    const deduction = 20;
    score -= deduction;
    breakdown.lpLock = {
      points: -deduction,
      reason: 'Liquidity is NOT locked - instant rugpull risk',
    };
  } else if (factors.liquidity.lockDays < 30) {
    const deduction = 10;
    score -= deduction;
    breakdown.lpLock = {
      points: -deduction,
      reason: `LP locked for only ${factors.liquidity.lockDays} days (< 30)`,
    };
  }

  // ... Continue for all 12 metrics
  // See docs/03-TECHNICAL/integrations/telegram-bot-risk-algorithm.md
  // for complete implementation

  // Clamp score to 0-100
  score = Math.max(0, Math.min(100, score));

  const category = getRiskCategory(score);

  return {
    score,
    category,
    breakdown,
    timestamp: new Date(),
  };
}

function getRiskCategory(score: number): RiskCategory {
  if (score >= 80) return 'SAFE';
  if (score >= 60) return 'CAUTION';
  if (score >= 30) return 'HIGH_RISK';
  return 'LIKELY_SCAM';
}
```

**See**: `docs/03-TECHNICAL/integrations/telegram-bot-risk-algorithm.md` for complete 12-metric implementation

#### 2.3 Scan Processor

**File**: `src/modules/scan/scan.processor.ts`

```typescript
import { Job } from 'bullmq';
import { HeliusProvider } from './providers/helius.provider';
import { BirdeyeProvider } from './providers/birdeye.provider';
import { RugcheckProvider } from './providers/rugcheck.provider';
import { calculateRiskScore } from './risk-scoring/risk-calculator';
import { ScanRepository } from './scan.repository';
import { logger } from '../../config/logger';
import { metrics } from '../../config/monitoring';

export async function processScan(job: Job) {
  const startTime = Date.now();
  const { tokenAddress, userId } = job.data;

  try {
    logger.info({ jobId: job.id, tokenAddress, userId }, 'Processing scan');

    // Parallel API calls for performance
    const [metadata, liquidityData, honeypotCheck] = await Promise.all([
      new HeliusProvider(process.env.HELIUS_API_KEY!).getTokenMetadata(tokenAddress),
      new BirdeyeProvider(process.env.BIRDEYE_API_KEY!).getLiquidityData(tokenAddress),
      new RugcheckProvider(process.env.RUGCHECK_API_KEY!).checkHoneypot(tokenAddress),
    ]);

    // Calculate risk score
    const riskScore = calculateRiskScore({
      liquidity: liquidityData,
      holders: metadata.holderDistribution,
      honeypot: honeypotCheck,
      // ... other factors
    });

    // Save to database
    const scanRepository = new ScanRepository();
    const scan = await scanRepository.createScan({
      userId,
      tokenAddress,
      riskScore: riskScore.score,
      riskCategory: riskScore.category,
      breakdown: riskScore.breakdown,
      rawData: { metadata, liquidityData, honeypotCheck },
    });

    // Metrics
    const duration = Date.now() - startTime;
    metrics.timing('scan.duration', duration);
    metrics.increment('scan.success', 1, { category: riskScore.category });

    logger.info(
      { jobId: job.id, scanId: scan.id, duration, riskScore: riskScore.score },
      'Scan completed'
    );

    return { scanId: scan.id, riskScore };
  } catch (error) {
    const duration = Date.now() - startTime;
    metrics.increment('scan.failed', 1);

    logger.error(
      { jobId: job.id, tokenAddress, userId, error, duration },
      'Scan failed'
    );

    throw error; // BullMQ will retry
  }
}
```

**Verification**:
```bash
# Queue a real scan
curl -X POST http://localhost:3000/api/scan \
  -H "Content-Type: application/json" \
  -d '{"tokenAddress": "So11111111111111111111111111111111111111112"}'

# Check worker logs for processing
# Check database for scan record
npx prisma studio
```

---

### Phase 3: Telegram Bot (Week 5-6)

#### 3.1 Grammy.js Bot Setup

**File**: `src/modules/telegram/bot.ts`

```typescript
import { Bot, session } from 'grammy';
import { registerCommands } from './commands';
import { setupConversations } from './conversations';
import { logger } from '../../config/logger';

export function createBot() {
  const bot = new Bot(process.env.TELEGRAM_BOT_TOKEN!);

  // Session middleware
  bot.use(session({
    initial: () => ({
      scansToday: 0,
      tier: 'FREE',
    }),
  }));

  // Error handling
  bot.catch((err) => {
    logger.error({ error: err }, 'Bot error');
  });

  // Register commands
  registerCommands(bot);

  // Setup conversations
  setupConversations(bot);

  return bot;
}

// src/modules/telegram/server.ts
import { createBot } from './bot';
import { webhookCallback } from 'grammy';
import Fastify from 'fastify';

export async function startTelegramBot() {
  const bot = createBot();

  if (process.env.NODE_ENV === 'production') {
    // Webhook mode
    const app = Fastify();

    app.post(`/telegram-webhook/${process.env.TELEGRAM_BOT_TOKEN}`,
      webhookCallback(bot, 'fastify')
    );

    await app.listen({ port: 8443, host: '0.0.0.0' });

    await bot.api.setWebhook(
      `${process.env.WEBHOOK_URL}/telegram-webhook/${process.env.TELEGRAM_BOT_TOKEN}`
    );
  } else {
    // Polling mode (local dev)
    bot.start();
  }
}
```

#### 3.2 Command Handlers

**File**: `src/modules/telegram/commands/scan.command.ts`

```typescript
import { CommandContext } from 'grammy';
import { scanQueue } from '../../../config/queue';
import { UserRepository } from '../../auth/user.repository';
import { validateSolanaAddress } from '../../../utils/validators';

export async function scanCommand(ctx: CommandContext<Context>) {
  const userId = ctx.from!.id.toString();
  const tokenAddress = ctx.match; // Text after /scan

  // Validate address
  if (!validateSolanaAddress(tokenAddress)) {
    await ctx.reply(
      '❌ Invalid Solana token address.\n\n' +
      'Example: /scan So11111111111111111111111111111111111111112'
    );
    return;
  }

  // Check rate limits
  const userRepo = new UserRepository();
  const user = await userRepo.findByTelegramId(BigInt(userId));

  if (!user) {
    // First-time user, show consent
    await ctx.reply(
      '👋 Welcome to CryptoRugMunch!\n\n' +
      'Before we scan, we need your consent to store scan history.\n\n' +
      'Use /consent to accept our privacy policy.'
    );
    return;
  }

  const dailyScans = await userRepo.getDailyScans(user.id);
  const limit = user.tier === 'FREE' ? 10 : 50;

  if (dailyScans >= limit) {
    await ctx.reply(
      `❌ Daily scan limit reached (${dailyScans}/${limit}).\n\n` +
      (user.tier === 'FREE'
        ? 'Upgrade to Premium for 50 scans/day: /premium'
        : 'Limit resets at midnight UTC.')
    );
    return;
  }

  // Queue scan
  await ctx.reply(
    `🔍 Scanning ${tokenAddress.slice(0, 8)}...\n\n` +
    'This usually takes 2-3 seconds.'
  );

  const job = await scanQueue.add('scan', {
    userId: user.id,
    tokenAddress,
    telegramChatId: ctx.chat!.id,
  });

  // Job completion is handled by a separate listener
  // that sends results back to user
}
```

**File**: `src/modules/telegram/commands/index.ts`

```typescript
import { Bot } from 'grammy';
import { scanCommand } from './scan.command';
import { historyCommand } from './history.command';
import { premiumCommand } from './premium.command';
import { consentCommand } from './consent.command';

export function registerCommands(bot: Bot) {
  bot.command('start', startCommand);
  bot.command('scan', scanCommand);
  bot.command('history', historyCommand);
  bot.command('premium', premiumCommand);
  bot.command('consent', consentCommand);
  bot.command('help', helpCommand);

  // Default handler
  bot.on('message:text', async (ctx) => {
    await ctx.reply(
      'Use /help to see available commands.\n' +
      'To scan a token: /scan <address>'
    );
  });
}
```

**See**: `docs/03-TECHNICAL/integrations/telegram-bot-setup.md` for all command implementations

#### 3.3 Result Formatter

**File**: `src/modules/telegram/formatters/scan-result.formatter.ts`

```typescript
import { RiskScore } from '../../scan/scan.types';

export function formatScanResult(scan: {
  tokenAddress: string;
  riskScore: RiskScore;
  metadata: any;
}): string {
  const emoji = getRiskEmoji(scan.riskScore.category);
  const color = getRiskColor(scan.riskScore.category);

  let message = `${emoji} **Risk Assessment**\n\n`;
  message += `**Token**: ${scan.metadata.symbol || 'Unknown'}\n`;
  message += `**Address**: \`${scan.tokenAddress.slice(0, 8)}...${scan.tokenAddress.slice(-8)}\`\n\n`;
  message += `**Risk Score**: ${scan.riskScore.score}/100 (${scan.riskScore.category})\n\n`;

  message += `**Breakdown**:\n`;
  for (const [key, { points, reason }] of Object.entries(scan.riskScore.breakdown)) {
    message += `  ${points < 0 ? '⚠️' : '✅'} ${reason} (${points > 0 ? '+' : ''}${points})\n`;
  }

  message += `\n**Recommendation**:\n`;
  message += getRecommendation(scan.riskScore.category);

  return message;
}

function getRiskEmoji(category: RiskCategory): string {
  switch (category) {
    case 'SAFE': return '🟢';
    case 'CAUTION': return '🟡';
    case 'HIGH_RISK': return '🟠';
    case 'LIKELY_SCAM': return '🔴';
  }
}

function getRecommendation(category: RiskCategory): string {
  switch (category) {
    case 'SAFE':
      return '✅ Low risk. Token shows healthy fundamentals.';
    case 'CAUTION':
      return '⚠️ Proceed with caution. Some concerns detected.';
    case 'HIGH_RISK':
      return '🚨 High risk. Multiple red flags detected.';
    case 'LIKELY_SCAM':
      return '🛑 LIKELY SCAM. Do not invest. Strong indicators of rugpull.';
  }
}
```

**Verification**:
```bash
# Start bot
pnpm telegram:dev

# In Telegram app, message your bot:
# /start
# /consent
# /scan So11111111111111111111111111111111111111112

# Verify scan result message
```

---

### Phase 4: Testing & Monitoring (Week 7-8)

#### 4.1 Unit Tests

**File**: `tests/unit/risk-calculator.test.ts`

```typescript
import { describe, it, expect } from 'vitest';
import { calculateRiskScore } from '../../src/modules/scan/risk-scoring/risk-calculator';

describe('Risk Calculator', () => {
  it('should return SAFE for healthy token', () => {
    const result = calculateRiskScore({
      liquidity: {
        usd: 100_000,
        locked: true,
        lockDays: 365,
      },
      holders: {
        top10Percent: 25,
        whaleCount: 2,
      },
      honeypot: {
        isBuyable: true,
        isSellable: true,
        sellTax: 0,
      },
      // ... other factors
    });

    expect(result.score).toBeGreaterThanOrEqual(80);
    expect(result.category).toBe('SAFE');
  });

  it('should return LIKELY_SCAM for suspicious token', () => {
    const result = calculateRiskScore({
      liquidity: {
        usd: 1_000,
        locked: false,
        lockDays: 0,
      },
      holders: {
        top10Percent: 90,
        whaleCount: 1,
      },
      honeypot: {
        isBuyable: true,
        isSellable: false,
        sellTax: 99,
      },
      // ... other factors
    });

    expect(result.score).toBeLessThan(30);
    expect(result.category).toBe('LIKELY_SCAM');
  });
});
```

**Run tests**:
```bash
pnpm test:unit
pnpm test:unit --coverage  # With coverage report
```

#### 4.2 Integration Tests

**File**: `tests/integration/scan-api.test.ts`

```typescript
import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { createApp } from '../../src/app';
import type { FastifyInstance } from 'fastify';

describe('Scan API', () => {
  let app: FastifyInstance;

  beforeAll(async () => {
    app = await createApp();
    await app.ready();
  });

  afterAll(async () => {
    await app.close();
  });

  it('POST /api/scan should queue a scan job', async () => {
    const response = await app.inject({
      method: 'POST',
      url: '/api/scan',
      headers: {
        'x-api-key': 'test-api-key',
      },
      payload: {
        tokenAddress: 'So11111111111111111111111111111111111111112',
      },
    });

    expect(response.statusCode).toBe(202);
    expect(response.json()).toMatchObject({
      jobId: expect.any(String),
      status: 'queued',
    });
  });

  it('should return 429 when rate limit exceeded', async () => {
    // Make 101 requests (limit is 100/minute)
    const requests = Array(101).fill(null).map(() =>
      app.inject({
        method: 'POST',
        url: '/api/scan',
        headers: { 'x-api-key': 'test-api-key' },
        payload: { tokenAddress: 'So11111111111111111111111111111111111111112' },
      })
    );

    const responses = await Promise.all(requests);
    const rateLimited = responses.filter(r => r.statusCode === 429);

    expect(rateLimited.length).toBeGreaterThan(0);
  });
});
```

**See**: `docs/03-TECHNICAL/operations/testing-strategy.md` for complete testing guide

#### 4.3 DataDog Integration

**File**: `src/config/monitoring.ts`

```typescript
import StatsD from 'hot-shots';
import { FastifyInstance } from 'fastify';

export const metrics = new StatsD({
  host: process.env.DATADOG_AGENT_HOST || 'localhost',
  port: 8125,
  prefix: 'rugmunch.',
  globalTags: {
    env: process.env.NODE_ENV || 'development',
    service: 'api',
  },
});

export function setupMonitoring(app: FastifyInstance) {
  // Request metrics
  app.addHook('onRequest', (request, reply, done) => {
    request.startTime = Date.now();
    done();
  });

  app.addHook('onResponse', (request, reply, done) => {
    const duration = Date.now() - request.startTime;

    metrics.timing('http.request.duration', duration, {
      method: request.method,
      route: request.routeConfig?.url || 'unknown',
      status: reply.statusCode.toString(),
    });

    metrics.increment('http.request.count', 1, {
      method: request.method,
      status: reply.statusCode.toString(),
    });

    done();
  });

  // Health check endpoint
  app.get('/health', async () => {
    return { status: 'ok', timestamp: new Date().toISOString() };
  });
}
```

**See**: `docs/03-TECHNICAL/operations/monitoring-alerting-setup.md` for complete monitoring setup

---

## Testing Strategy

### Test Pyramid

```
        /\
       /  \      E2E Tests (5%)
      /----\     - User flows
     /      \    - Payment flows
    /--------\
   /          \  Integration Tests (20%)
  /------------\ - API endpoints
 /              \- Database operations
/________________\
   Unit Tests (75%)
   - Business logic
   - Utilities
```

### Running Tests

```bash
# Unit tests (fast, no external dependencies)
pnpm test:unit

# Integration tests (require DB/Redis)
pnpm test:integration

# E2E tests (full stack)
pnpm test:e2e

# Load tests
pnpm test:load

# All tests
pnpm test

# Watch mode
pnpm test:watch

# Coverage report
pnpm test:coverage
```

### Test Environment Setup

**File**: `.env.test`

```bash
NODE_ENV=test
DATABASE_URL="postgresql://user:password@localhost:5432/rugmunch_test"
REDIS_URL="redis://localhost:6379/1"  # Use DB 1 for tests
```

**Before tests**:
```bash
# Reset test database
npx prisma migrate reset --force
npx prisma db seed
```

---

## Deployment Process

### Local Development

```bash
# Start all services
docker-compose up -d

# Start API
pnpm dev

# Start worker
pnpm worker:dev

# Start Telegram bot
pnpm telegram:dev
```

### Railway (Staging)

**Step 1: Create Railway Project**

```bash
# Install Railway CLI
npm install -g @railway/cli

# Login
railway login

# Initialize project
railway init

# Link to GitHub repo
railway link
```

**Step 2: Configure Services**

Create `railway.toml`:

```toml
[build]
builder = "NIXPACKS"

[deploy]
startCommand = "npm run start"
healthcheckPath = "/health"
healthcheckTimeout = 100
restartPolicyType = "ON_FAILURE"
restartPolicyMaxRetries = 10
```

**Step 3: Set Environment Variables**

```bash
# Via Railway CLI
railway variables set TELEGRAM_BOT_TOKEN=<token>
railway variables set HELIUS_API_KEY=<key>
# ... all other variables

# Or via Railway dashboard
# https://railway.app/project/<project-id>/variables
```

**Step 4: Deploy**

```bash
# Deploy API
railway up --service api

# Deploy Worker
railway up --service worker

# Deploy Telegram Bot
railway up --service telegram-bot
```

**Step 5: Set up Webhook**

```bash
# Get Railway deployment URL
railway domain

# Set Telegram webhook
curl -X POST "https://api.telegram.org/bot<token>/setWebhook" \
  -d "url=https://<railway-domain>/telegram-webhook/<token>"
```

**See**: `docs/03-TECHNICAL/operations/worker-deployment.md` for AWS ECS production deployment

---

## Monitoring & Operations

### DataDog Dashboards

**Create dashboard** at https://app.datadoghq.com/dashboard/lists

**Key metrics to track**:

```
# Application Performance
- rugmunch.scan.duration (avg, p95, p99)
- rugmunch.scan.success (count)
- rugmunch.scan.failed (count)
- rugmunch.queue.depth (gauge)

# HTTP Metrics
- rugmunch.http.request.duration (avg, p95)
- rugmunch.http.request.count (count)

# Business Metrics
- rugmunch.user.signup (count)
- rugmunch.subscription.created (count)
- rugmunch.revenue.monthly (gauge)
```

### Alert Configuration

**File**: `datadog-monitors.json`

```json
{
  "monitors": [
    {
      "name": "High Scan Failure Rate",
      "type": "metric alert",
      "query": "avg(last_5m):sum:rugmunch.scan.failed{*}.as_count() / sum:rugmunch.scan.success{*}.as_count() > 0.1",
      "message": "@pagerduty-rugmunch Scan failure rate > 10% for 5 minutes",
      "tags": ["service:api", "severity:high"]
    },
    {
      "name": "Slow Scan Performance",
      "type": "metric alert",
      "query": "avg(last_5m):avg:rugmunch.scan.duration{*} > 5000",
      "message": "@slack-engineering Scan duration > 5s (SLA: 3s)",
      "tags": ["service:worker", "severity:medium"]
    }
  ]
}
```

**See**: `docs/03-TECHNICAL/operations/monitoring-alerting-setup.md` for complete monitoring setup

---

## Common Implementation Patterns

### Pattern 1: Repository Pattern

```typescript
// ✅ GOOD: Repository handles all database access
export class ScanRepository {
  async createScan(data: CreateScanInput): Promise<Scan> {
    return prisma.scan.create({ data });
  }

  async findByTokenAddress(address: string): Promise<Scan[]> {
    return prisma.scan.findMany({
      where: { tokenAddress: address },
      orderBy: { createdAt: 'desc' },
    });
  }
}

// Controller uses repository
export async function getScanHistory(req, reply) {
  const scanRepo = new ScanRepository();
  const scans = await scanRepo.findByTokenAddress(req.params.address);
  return scans;
}

// ❌ BAD: Direct Prisma in controller
export async function getScanHistory(req, reply) {
  const scans = await prisma.scan.findMany({
    where: { tokenAddress: req.params.address },
  });
  return scans;
}
```

### Pattern 2: Structured Logging

```typescript
// ✅ GOOD: Structured logging with context
logger.info(
  {
    userId,
    tokenAddress,
    duration,
    riskScore: score.score
  },
  'Scan completed'
);

// ❌ BAD: String concatenation
console.log(`Scan completed for user ${userId}`);
```

### Pattern 3: Error Handling

```typescript
// ✅ GOOD: Structured error with context
try {
  await riskyOperation();
} catch (error) {
  logger.error({ error, userId, tokenAddress }, 'Scan failed');

  Sentry.captureException(error, {
    tags: { operation: 'token_scan' },
    extra: { userId, tokenAddress },
  });

  metrics.increment('scan.failed', 1, { reason: error.code });

  throw new ScanError('Failed to scan token', { cause: error });
}

// ❌ BAD: Silent failure
try {
  await riskyOperation();
} catch (error) {
  console.log(error);
}
```

### Pattern 4: Caching

```typescript
// ✅ GOOD: Cache with TTL
async function getTokenMetadata(address: string) {
  const cached = await redis.get(`token:${address}`);
  if (cached) {
    metrics.increment('cache.hit', 1);
    return JSON.parse(cached);
  }

  metrics.increment('cache.miss', 1);

  const metadata = await helius.getTokenMetadata(address);

  await redis.setex(
    `token:${address}`,
    300, // 5 minutes
    JSON.stringify(metadata)
  );

  return metadata;
}
```

---

## Troubleshooting Common Issues

### Issue 1: Worker Not Processing Jobs

**Symptoms**: Jobs stuck in "waiting" state

**Debug**:
```bash
# Check Redis connection
redis-cli ping

# Check queue depth
redis-cli llen bull:token-scan:wait

# Check worker logs
pm2 logs worker
```

**Fix**:
```typescript
// Ensure worker is connected
scanWorker.on('error', (err) => {
  console.error('Worker error:', err);
});
```

### Issue 2: Slow Scan Performance

**Symptoms**: Scans taking > 5 seconds

**Debug**:
```bash
# Check DataDog metrics
# rugmunch.scan.duration p95

# Check individual API latencies
# rugmunch.provider.helius.duration
# rugmunch.provider.birdeye.duration
```

**Fix**:
```typescript
// Ensure parallel API calls
const [metadata, liquidity, honeypot] = await Promise.all([
  helius.getMetadata(address),
  birdeye.getLiquidity(address),
  rugcheck.checkHoneypot(address),
]);
```

### Issue 3: Rate Limit Exceeded

**Symptoms**: 429 errors from API providers

**Fix**:
```typescript
// Implement exponential backoff
async function fetchWithRetry(fn, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await fn();
    } catch (error) {
      if (error.status === 429 && i < maxRetries - 1) {
        const delay = Math.pow(2, i) * 1000;
        await sleep(delay);
        continue;
      }
      throw error;
    }
  }
}
```

---

## Related Documentation

- `docs/03-TECHNICAL/development/local-development-guide.md` - Complete local setup
- `docs/03-TECHNICAL/operations/worker-deployment.md` - Deployment guides
- `docs/03-TECHNICAL/operations/monitoring-alerting-setup.md` - Monitoring setup
- `docs/03-TECHNICAL/integrations/telegram-bot-risk-algorithm.md` - Risk scoring algorithm
- `docs/03-TECHNICAL/development/code-style-guide.md` - Coding standards
