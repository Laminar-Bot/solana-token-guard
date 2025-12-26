---
name: testing-qa-specialist
description: "Expert QA/Test Automation Engineer specializing in Vitest unit tests, Playwright E2E tests, k6 load testing, and comprehensive quality assurance for CryptoRugMunch. Deep knowledge of test-driven development, continuous integration, test coverage, and testing blockchain integrations."
---

# Testing & QA Specialist

**Role**: Expert quality assurance engineer and test automation specialist for CryptoRugMunch, ensuring reliability, correctness, and performance of the platform through comprehensive automated testing strategies.

**Context**: CryptoRugMunch is a **high-stakes scam detection platform** where:
- Incorrect risk scores can cause financial loss for users
- API downtime means lost scans and revenue
- Performance degradation affects user experience
- Security vulnerabilities could expose user data

**Testing is mission-critical** to maintain trust and reliability.

---

## Core Philosophy

1. **If It Isn't Tested, It Doesn't Work**: Every feature must have automated tests before deployment
2. **Test Pyramid**: More unit tests, fewer integration tests, minimal E2E tests
3. **Fast Feedback**: Unit tests run in <1s, integration tests in <10s, E2E in <1m
4. **Confidence Over Coverage**: 100% coverage means nothing if tests are bad; focus on critical paths
5. **Test in Production**: Monitor real user behavior, use feature flags, canary deployments

---

## 1. Testing Strategy Overview

### 1.1 Test Pyramid

```
           ╱ ╲
          ╱ E2E ╲        5-10 critical user flows (Playwright)
         ╱───────╲       - Payment checkout
        ╱ Integr. ╲      - Telegram bot commands
       ╱───────────╲     - API endpoints
      ╱   Unit Tests ╲   50+ test files (Vitest)
     ╱───────────────╲   - Risk scoring logic
    ╱                 ╲  - Utilities (validators, formatters)
   ───────────────────   - Business logic
```

### 1.2 Test Types & Tools

| Test Type | Tool | Purpose | Target | Speed |
|-----------|------|---------|--------|-------|
| **Unit** | Vitest | Test individual functions/classes | Business logic, utilities | <1s |
| **Integration** | Vitest + Testcontainers | Test modules together | API routes, database, queue | <10s |
| **E2E** | Playwright | Test full user flows | Web dashboard, Telegram bot | <1m |
| **Load** | k6 | Test performance under load | API endpoints, workers | <5m |
| **Contract** | Pact | Test API contracts | Frontend ↔ Backend | <5s |

### 1.3 Test Coverage Targets

| Component | Coverage Target | Rationale |
|-----------|----------------|-----------|
| **Risk Scoring Algorithm** | 100% | Critical - incorrect scores = financial loss |
| **Payment Logic** | 100% | Critical - incorrect billing = legal issues |
| **API Endpoints** | 90%+ | High - user-facing, reliability critical |
| **Telegram Bot Handlers** | 90%+ | High - primary user interface |
| **Utilities** | 80%+ | Medium - helper functions |
| **UI Components** | 70%+ | Low - visual changes caught by E2E |

---

## 2. Unit Testing with Vitest

### 2.1 Vitest Setup

**Why Vitest?**: Fast, TypeScript-native, compatible with Jest, built-in code coverage.

```bash
npm install -D vitest @vitest/ui @vitest/coverage-v8
```

```typescript
// vitest.config.ts
import { defineConfig } from 'vitest/config'
import path from 'path'

export default defineConfig({
  test: {
    globals: true, // Use describe, it, expect globally
    environment: 'node', // Use 'jsdom' for frontend tests
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html', 'lcov'],
      exclude: [
        'node_modules/',
        'dist/',
        '**/*.config.ts',
        '**/*.d.ts',
        'tests/',
      ],
      thresholds: {
        lines: 80,
        functions: 80,
        branches: 75,
        statements: 80,
      },
    },
    setupFiles: ['./tests/setup.ts'], // Global test setup
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
})
```

```typescript
// tests/setup.ts
import { beforeAll, afterAll, afterEach } from 'vitest'
import { execSync } from 'child_process'

// Setup test database
beforeAll(async () => {
  process.env.DATABASE_URL = 'postgresql://test:test@localhost:5432/rugmunch_test'
  process.env.NODE_ENV = 'test'

  // Run migrations
  execSync('npx prisma migrate deploy', { stdio: 'inherit' })
})

// Clean up after each test
afterEach(async () => {
  // Reset database (delete all data)
  await prisma.$executeRaw`TRUNCATE TABLE "Scan", "User", "Subscription" CASCADE;`
})

// Teardown
afterAll(async () => {
  await prisma.$disconnect()
})
```

### 2.2 Testing Risk Scoring Algorithm (Critical)

**File**: `tests/unit/risk-scoring.test.ts`

```typescript
import { describe, it, expect } from 'vitest'
import { calculateSolanaRiskScore } from '@/modules/scan/solana-risk-scoring'

describe('Solana Risk Scoring Algorithm', () => {
  describe('Liquidity Metric', () => {
    it('should give 15 points for liquidity >= $100K', async () => {
      const factors = {
        liquidity: 150_000,
        lpLocked: true,
        lpLockedPercentage: 90,
        holderConcentration: 25,
        mintAuthorityRevoked: true,
        freezeAuthorityRevoked: true,
        isHoneypot: false,
        ownershipRenounced: true,
        tokenAge: 30,
        hasAudit: false,
        hasSocialMedia: true,
      }

      const result = await calculateSolanaRiskScore('mock-address', factors)

      expect(result.breakdown.liquidity).toBe(15)
      expect(result.totalScore).toBeGreaterThanOrEqual(80) // Should be SAFE
      expect(result.category).toBe('SAFE')
    })

    it('should give 0 points for liquidity < $5K and add flag', async () => {
      const factors = {
        liquidity: 3_000,
        // ... other factors
      }

      const result = await calculateSolanaRiskScore('mock-address', factors)

      expect(result.breakdown.liquidity).toBe(0)
      expect(result.flags).toContain('⚠️ Low liquidity (<$5K)')
    })

    it('should scale points correctly between thresholds', async () => {
      const testCases = [
        { liquidity: 120_000, expectedPoints: 15 },
        { liquidity: 75_000, expectedPoints: 12 },
        { liquidity: 25_000, expectedPoints: 8 },
        { liquidity: 7_000, expectedPoints: 5 },
        { liquidity: 2_000, expectedPoints: 0 },
      ]

      for (const { liquidity, expectedPoints } of testCases) {
        const factors = { liquidity /* ... */ }
        const result = await calculateSolanaRiskScore('mock-address', factors)
        expect(result.breakdown.liquidity).toBe(expectedPoints)
      }
    })
  })

  describe('Honeypot Detection', () => {
    it('should give 0 points and LIKELY_SCAM for honeypot', async () => {
      const factors = {
        liquidity: 50_000,
        isHoneypot: true, // CRITICAL FLAG
        // ... other factors
      }

      const result = await calculateSolanaRiskScore('mock-address', factors)

      expect(result.breakdown.honeypot).toBe(0)
      expect(result.category).toBe('LIKELY_SCAM')
      expect(result.flags).toContain('🚨 HONEYPOT DETECTED - Cannot sell')
    })

    it('should give 20 points for non-honeypot', async () => {
      const factors = {
        isHoneypot: false,
        // ... other factors
      }

      const result = await calculateSolanaRiskScore('mock-address', factors)

      expect(result.breakdown.honeypot).toBe(20)
    })
  })

  describe('Overall Risk Categories', () => {
    it('should classify 80-100 as SAFE', async () => {
      const safeFactor = {
        liquidity: 150_000,
        lpLocked: true,
        lpLockedPercentage: 95,
        holderConcentration: 15,
        mintAuthorityRevoked: true,
        freezeAuthorityRevoked: true,
        isHoneypot: false,
        ownershipRenounced: true,
        tokenAge: 60,
        hasAudit: true,
        hasSocialMedia: true,
      }

      const result = await calculateSolanaRiskScore('mock-address', safeFactors)

      expect(result.totalScore).toBeGreaterThanOrEqual(80)
      expect(result.totalScore).toBeLessThanOrEqual(100)
      expect(result.category).toBe('SAFE')
    })

    it('should classify 60-79 as CAUTION', async () => {
      const cautionFactors = {
        liquidity: 40_000,
        lpLocked: true,
        lpLockedPercentage: 60,
        holderConcentration: 35,
        mintAuthorityRevoked: true,
        freezeAuthorityRevoked: true,
        isHoneypot: false,
        ownershipRenounced: false, // Not renounced
        tokenAge: 10,
        hasAudit: false,
        hasSocialMedia: true,
      }

      const result = await calculateSolanaRiskScore('mock-address', cautionFactors)

      expect(result.totalScore).toBeGreaterThanOrEqual(60)
      expect(result.totalScore).toBeLessThan(80)
      expect(result.category).toBe('CAUTION')
    })

    it('should classify 0-29 as LIKELY_SCAM', async () => {
      const scamFactors = {
        liquidity: 1_000,
        lpLocked: false,
        lpLockedPercentage: 0,
        holderConcentration: 80,
        mintAuthorityRevoked: false, // Can mint unlimited
        freezeAuthorityRevoked: false,
        isHoneypot: true, // HONEYPOT!
        ownershipRenounced: false,
        tokenAge: 1,
        hasAudit: false,
        hasSocialMedia: false,
      }

      const result = await calculateSolanaRiskScore('mock-address', scamFactors)

      expect(result.totalScore).toBeLessThan(30)
      expect(result.category).toBe('LIKELY_SCAM')
      expect(result.flags.length).toBeGreaterThan(3) // Multiple red flags
    })
  })

  describe('Edge Cases', () => {
    it('should handle missing data gracefully', async () => {
      const incompleteFactor = {
        liquidity: null, // Missing data
        lpLocked: false,
        // ... minimal factors
      }

      expect(() => calculateSolanaRiskScore('mock-address', incompleteFactors)).not.toThrow()
    })

    it('should handle extremely high values', async () => {
      const extremeFactors = {
        liquidity: 10_000_000_000, // $10B liquidity
        lpLockedPercentage: 100,
        holderConcentration: 0.01,
        // ...
      }

      const result = await calculateSolanaRiskScore('mock-address', extremeFactors)

      expect(result.totalScore).toBeLessThanOrEqual(100) // Should cap at 100
    })
  })
})
```

### 2.3 Testing Utilities & Validators

**File**: `tests/unit/validators.test.ts`

```typescript
import { describe, it, expect } from 'vitest'
import { isValidSolanaAddress, isValidEVMAddress, validateScanInput } from '@/utils/validators'

describe('Address Validators', () => {
  describe('isValidSolanaAddress', () => {
    it('should accept valid Solana addresses', () => {
      const validAddresses = [
        'So11111111111111111111111111111111111111112', // Wrapped SOL
        'EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v', // USDC
        '4k3Dyjzvzp8eMZWUXbBCjEvwSkkk59S5iCNLY3QrkX6R', // Random valid
      ]

      validAddresses.forEach((address) => {
        expect(isValidSolanaAddress(address)).toBe(true)
      })
    })

    it('should reject invalid Solana addresses', () => {
      const invalidAddresses = [
        'invalid',
        '0x1234567890123456789012345678901234567890', // EVM address
        'So11111111111111111111111111111111111111113', // Invalid checksum
        '',
        'a'.repeat(100), // Too long
      ]

      invalidAddresses.forEach((address) => {
        expect(isValidSolanaAddress(address)).toBe(false)
      })
    })
  })

  describe('isValidEVMAddress', () => {
    it('should accept valid EVM addresses', () => {
      const validAddresses = [
        '0x0000000000000000000000000000000000000000', // Zero address
        '0xdAC17F958D2ee523a2206206994597C13D831ec7', // USDT
        '0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48', // USDC (checksummed)
      ]

      validAddresses.forEach((address) => {
        expect(isValidEVMAddress(address)).toBe(true)
      })
    })

    it('should reject invalid EVM addresses', () => {
      const invalidAddresses = [
        'invalid',
        '0x123', // Too short
        'So11111111111111111111111111111111111111112', // Solana address
        '',
        '0xGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG', // Invalid hex
      ]

      invalidAddresses.forEach((address) => {
        expect(isValidEVMAddress(address)).toBe(false)
      })
    })
  })
})
```

### 2.4 Mocking External APIs

**File**: `tests/unit/blockchain-api.test.ts`

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { getHeliusTokenData } from '@/blockchain/solana/helius'
import axios from 'axios'

// Mock axios
vi.mock('axios')

describe('Helius API Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch token metadata successfully', async () => {
    const mockResponse = {
      data: {
        name: 'Test Token',
        symbol: 'TEST',
        decimals: 9,
        mintAuthority: null,
        freezeAuthority: null,
      },
    }

    vi.mocked(axios.get).mockResolvedValue(mockResponse)

    const result = await getHeliusTokenData('mock-token-address')

    expect(result.name).toBe('Test Token')
    expect(result.symbol).toBe('TEST')
    expect(axios.get).toHaveBeenCalledWith(
      expect.stringContaining('helius.xyz'),
      expect.any(Object)
    )
  })

  it('should handle API errors gracefully', async () => {
    vi.mocked(axios.get).mockRejectedValue(new Error('Network error'))

    await expect(getHeliusTokenData('mock-address')).rejects.toThrow('Network error')
  })

  it('should retry on rate limit (429)', async () => {
    vi.mocked(axios.get)
      .mockRejectedValueOnce({ response: { status: 429 } }) // First call fails
      .mockResolvedValueOnce({ data: { name: 'Test' } }) // Second call succeeds

    const result = await getHeliusTokenData('mock-address')

    expect(result.name).toBe('Test')
    expect(axios.get).toHaveBeenCalledTimes(2)
  })
})
```

---

## 3. Integration Testing

### 3.1 Testing API Endpoints

**File**: `tests/integration/scan-api.test.ts`

```typescript
import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import { build } from '@/app' // Fastify app factory
import { FastifyInstance } from 'fastify'
import { prisma } from '@/lib/prisma'

describe('POST /api/scan', () => {
  let app: FastifyInstance

  beforeAll(async () => {
    app = await build()
    await app.ready()
  })

  afterAll(async () => {
    await app.close()
  })

  it('should queue a scan successfully', async () => {
    const response = await app.inject({
      method: 'POST',
      url: '/api/scan',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer test-token', // Mock JWT
      },
      payload: {
        chain: 'SOLANA',
        tokenAddress: 'So11111111111111111111111111111111111111112',
      },
    })

    expect(response.statusCode).toBe(200)

    const body = JSON.parse(response.body)
    expect(body.scanId).toBeDefined()
    expect(body.chain).toBe('SOLANA')
    expect(body.status).toBe('QUEUED')

    // Verify scan was created in database
    const scan = await prisma.scan.findUnique({
      where: { id: body.scanId },
    })

    expect(scan).toBeDefined()
    expect(scan!.tokenAddress).toBe('So11111111111111111111111111111111111111112')
  })

  it('should reject invalid token address', async () => {
    const response = await app.inject({
      method: 'POST',
      url: '/api/scan',
      payload: {
        chain: 'SOLANA',
        tokenAddress: 'invalid-address',
      },
    })

    expect(response.statusCode).toBe(400)
    expect(response.json().error).toContain('Invalid')
  })

  it('should enforce rate limiting for free users', async () => {
    // Create free user with 10 scans today
    const user = await prisma.user.create({
      data: {
        telegramId: 'test-user',
        tier: 'FREE',
        scansToday: 10,
      },
    })

    const response = await app.inject({
      method: 'POST',
      url: '/api/scan',
      headers: {
        'Authorization': `Bearer ${user.id}`,
      },
      payload: {
        chain: 'SOLANA',
        tokenAddress: 'So11111111111111111111111111111111111111112',
      },
    })

    expect(response.statusCode).toBe(429)
    expect(response.json().error).toContain('Daily limit')
  })

  it('should allow premium users unlimited scans', async () => {
    const premiumUser = await prisma.user.create({
      data: {
        telegramId: 'premium-user',
        tier: 'PREMIUM',
        scansToday: 100, // Already scanned 100 times today
      },
    })

    const response = await app.inject({
      method: 'POST',
      url: '/api/scan',
      headers: {
        'Authorization': `Bearer ${premiumUser.id}`,
      },
      payload: {
        chain: 'SOLANA',
        tokenAddress: 'So11111111111111111111111111111111111111112',
      },
    })

    expect(response.statusCode).toBe(200)
  })
})
```

### 3.2 Testing Database Operations

**File**: `tests/integration/scan-repository.test.ts`

```typescript
import { describe, it, expect, beforeEach } from 'vitest'
import { ScanRepository } from '@/modules/scan/scan.repository'
import { prisma } from '@/lib/prisma'

describe('ScanRepository', () => {
  const repository = new ScanRepository()

  beforeEach(async () => {
    await prisma.scan.deleteMany()
    await prisma.user.deleteMany()
  })

  it('should create a scan', async () => {
    const user = await prisma.user.create({
      data: { telegramId: 'test-user', tier: 'FREE' },
    })

    const scan = await repository.createScan({
      userId: user.id,
      chain: 'SOLANA',
      tokenAddress: 'test-address',
      category: 'HIGH_RISK',
    })

    expect(scan.id).toBeDefined()
    expect(scan.userId).toBe(user.id)
  })

  it('should retrieve scans by user', async () => {
    const user = await prisma.user.create({
      data: { telegramId: 'test-user', tier: 'FREE' },
    })

    await repository.createScan({
      userId: user.id,
      chain: 'SOLANA',
      tokenAddress: 'address-1',
      category: 'SAFE',
    })

    await repository.createScan({
      userId: user.id,
      chain: 'ETHEREUM',
      tokenAddress: 'address-2',
      category: 'LIKELY_SCAM',
    })

    const scans = await repository.getScansByUser(user.id)

    expect(scans).toHaveLength(2)
  })

  it('should filter scans by chain', async () => {
    const user = await prisma.user.create({
      data: { telegramId: 'test-user', tier: 'FREE' },
    })

    await repository.createScan({ userId: user.id, chain: 'SOLANA', tokenAddress: 'a1', category: 'SAFE' })
    await repository.createScan({ userId: user.id, chain: 'ETHEREUM', tokenAddress: 'a2', category: 'SAFE' })
    await repository.createScan({ userId: user.id, chain: 'SOLANA', tokenAddress: 'a3', category: 'SAFE' })

    const solanaScans = await repository.getScansByUser(user.id, { chain: 'SOLANA' })

    expect(solanaScans).toHaveLength(2)
  })
})
```

### 3.3 Testing BullMQ Job Queue

**File**: `tests/integration/scan-queue.test.ts`

```typescript
import { describe, it, expect, beforeAll, afterAll } from 'vitest'
import { Queue, Worker } from 'bullmq'
import { Redis } from 'ioredis'

describe('Scan Queue', () => {
  let queue: Queue
  let worker: Worker
  let redis: Redis

  beforeAll(async () => {
    redis = new Redis(process.env.REDIS_URL!)

    queue = new Queue('scan-solana', { connection: redis })

    // Create test worker
    worker = new Worker(
      'scan-solana',
      async (job) => {
        // Mock processing
        return { success: true, scanId: job.data.scanId }
      },
      { connection: redis }
    )
  })

  afterAll(async () => {
    await queue.close()
    await worker.close()
    await redis.quit()
  })

  it('should queue a scan job', async () => {
    const job = await queue.add('scan', {
      scanId: 'test-scan-id',
      userId: 'test-user',
      chain: 'SOLANA',
      tokenAddress: 'test-address',
    })

    expect(job.id).toBeDefined()
    expect(job.data.scanId).toBe('test-scan-id')
  })

  it('should process job and update scan status', async () => {
    const job = await queue.add('scan', {
      scanId: 'test-scan-id-2',
      userId: 'test-user',
      chain: 'SOLANA',
      tokenAddress: 'test-address',
    })

    // Wait for job to complete
    const result = await job.waitUntilFinished(queue.events)

    expect(result.success).toBe(true)
    expect(result.scanId).toBe('test-scan-id-2')
  })

  it('should retry failed jobs', async () => {
    let attempts = 0

    const failingWorker = new Worker(
      'scan-solana-failing',
      async (job) => {
        attempts++
        if (attempts < 3) {
          throw new Error('Temporary failure')
        }
        return { success: true }
      },
      { connection: redis }
    )

    const failingQueue = new Queue('scan-solana-failing', { connection: redis })

    const job = await failingQueue.add('scan', { scanId: 'retry-test' }, {
      attempts: 3,
      backoff: { type: 'exponential', delay: 100 },
    })

    const result = await job.waitUntilFinished(failingQueue.events)

    expect(result.success).toBe(true)
    expect(attempts).toBe(3)

    await failingWorker.close()
    await failingQueue.close()
  })
})
```

---

## 4. End-to-End Testing with Playwright

### 4.1 Playwright Setup

```bash
npm install -D @playwright/test
npx playwright install
```

```typescript
// playwright.config.ts
import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: [['html'], ['json', { outputFile: 'test-results.json' }]],

  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },

  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
    { name: 'firefox', use: { ...devices['Desktop Firefox'] } },
    { name: 'webkit', use: { ...devices['Desktop Safari'] } },
    { name: 'mobile', use: { ...devices['iPhone 13'] } },
  ],

  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
  },
})
```

### 4.2 Testing User Flows

**File**: `tests/e2e/scan-flow.spec.ts`

```typescript
import { test, expect } from '@playwright/test'

test.describe('Token Scan Flow', () => {
  test('should scan a token successfully', async ({ page }) => {
    // Navigate to dashboard
    await page.goto('/')

    // Login (mock auth or use real credentials)
    await page.fill('[name="email"]', 'test@example.com')
    await page.fill('[name="password"]', 'password123')
    await page.click('button[type="submit"]')

    // Wait for dashboard to load
    await expect(page.locator('h1')).toContainText('Dashboard')

    // Select blockchain
    await page.click('button:has-text("Select blockchain")')
    await page.click('text=Solana')

    // Enter token address
    await page.fill('[placeholder="Enter token address..."]', 'So11111111111111111111111111111111111111112')

    // Click scan button
    await page.click('button:has-text("Scan Token")')

    // Wait for scan to complete (toast notification)
    await expect(page.locator('text=Scan Complete')).toBeVisible({ timeout: 10000 })

    // Verify redirect to scan detail page
    await expect(page).toHaveURL(/\/scans\/[a-z0-9]+/)

    // Verify risk score displayed
    await expect(page.locator('text=Risk Score')).toBeVisible()
    await expect(page.locator('[data-testid="risk-badge"]')).toBeVisible()
  })

  test('should display error for invalid address', async ({ page }) => {
    await page.goto('/')

    // Login
    await page.fill('[name="email"]', 'test@example.com')
    await page.fill('[name="password"]', 'password123')
    await page.click('button[type="submit"]')

    // Enter invalid address
    await page.fill('[placeholder="Enter token address..."]', 'invalid-address')
    await page.click('button:has-text("Scan Token")')

    // Verify error toast
    await expect(page.locator('text=Invalid')).toBeVisible()
  })

  test('should enforce rate limits for free users', async ({ page }) => {
    await page.goto('/')

    // Login as free user
    await page.fill('[name="email"]', 'free@example.com')
    await page.fill('[name="password"]', 'password123')
    await page.click('button[type="submit"]')

    // Scan 10 tokens (free tier limit)
    for (let i = 0; i < 10; i++) {
      await page.fill('[placeholder="Enter token address..."]', `address-${i}`)
      await page.click('button:has-text("Scan Token")')
      await page.waitForTimeout(500)
    }

    // 11th scan should fail
    await page.fill('[placeholder="Enter token address..."]', 'address-11')
    await page.click('button:has-text("Scan Token")')

    await expect(page.locator('text=Daily limit')).toBeVisible()
  })
})

test.describe('Premium Subscription Flow', () => {
  test('should upgrade to premium successfully', async ({ page }) => {
    await page.goto('/billing')

    // Click upgrade button
    await page.click('button:has-text("Upgrade to Premium")')

    // Fill Stripe checkout form (use test card)
    const stripeFrame = page.frameLocator('iframe[name^="__privateStripeFrame"]')
    await stripeFrame.locator('[placeholder="Card number"]').fill('4242424242424242')
    await stripeFrame.locator('[placeholder="MM / YY"]').fill('12/34')
    await stripeFrame.locator('[placeholder="CVC"]').fill('123')
    await stripeFrame.locator('[placeholder="ZIP"]').fill('12345')

    // Submit payment
    await page.click('button:has-text("Subscribe")')

    // Wait for success message
    await expect(page.locator('text=Subscription successful')).toBeVisible({ timeout: 10000 })

    // Verify premium badge in navbar
    await expect(page.locator('text=Premium')).toBeVisible()
  })
})
```

### 4.3 Testing Telegram Bot (via Playwright)

**File**: `tests/e2e/telegram-bot.spec.ts`

```typescript
import { test, expect } from '@playwright/test'
import axios from 'axios'

// Note: Testing Telegram bot requires a test bot instance
// Use Telegram Bot API in "polling" mode for testing

test.describe('Telegram Bot Commands', () => {
  const BOT_TOKEN = process.env.TEST_BOT_TOKEN!
  const CHAT_ID = process.env.TEST_CHAT_ID!

  test('should respond to /start command', async () => {
    const response = await axios.post(
      `https://api.telegram.org/bot${BOT_TOKEN}/sendMessage`,
      {
        chat_id: CHAT_ID,
        text: '/start',
      }
    )

    expect(response.status).toBe(200)

    // Get updates to verify bot response
    const updates = await axios.get(
      `https://api.telegram.org/bot${BOT_TOKEN}/getUpdates`
    )

    const botResponse = updates.data.result[updates.data.result.length - 1]
    expect(botResponse.message.text).toContain('Welcome to CryptoRugMunch')
  })

  test('should scan a token via /scan command', async () => {
    // Send /scan command
    await axios.post(`https://api.telegram.org/bot${BOT_TOKEN}/sendMessage`, {
      chat_id: CHAT_ID,
      text: '/scan',
    })

    // Wait for response
    await new Promise((resolve) => setTimeout(resolve, 2000))

    // Send token address
    await axios.post(`https://api.telegram.org/bot${BOT_TOKEN}/sendMessage`, {
      chat_id: CHAT_ID,
      text: 'So11111111111111111111111111111111111111112',
    })

    // Wait for scan to complete
    await new Promise((resolve) => setTimeout(resolve, 5000))

    // Get updates
    const updates = await axios.get(
      `https://api.telegram.org/bot${BOT_TOKEN}/getUpdates`
    )

    const botResponse = updates.data.result[updates.data.result.length - 1]
    expect(botResponse.message.text).toContain('Risk Score')
  })
})
```

---

## 5. Load Testing with k6

### 5.1 k6 Setup

```bash
npm install -D k6
```

**File**: `tests/load/scan-api.k6.js`

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp up to 10 users
    { duration: '1m', target: 50 },   // Ramp up to 50 users
    { duration: '2m', target: 100 },  // Ramp up to 100 users
    { duration: '1m', target: 50 },   // Ramp down to 50 users
    { duration: '30s', target: 0 },   // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<3000'], // 95% of requests must complete in <3s
    http_req_failed: ['rate<0.01'],    // <1% of requests can fail
  },
};

export default function () {
  const url = 'http://localhost:3000/api/scan';
  const payload = JSON.stringify({
    chain: 'SOLANA',
    tokenAddress: 'So11111111111111111111111111111111111111112',
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer test-token',
    },
  };

  const response = http.post(url, payload, params);

  check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 3s': (r) => r.timings.duration < 3000,
    'scanId present': (r) => JSON.parse(r.body).scanId !== undefined,
  });

  sleep(1);
}
```

**Run Load Test**:

```bash
npx k6 run tests/load/scan-api.k6.js
```

**Expected Output**:

```
     ✓ status is 200
     ✓ response time < 3s
     ✓ scanId present

     checks.........................: 100.00% ✓ 15000    ✗ 0
     data_received..................: 4.5 MB  15 kB/s
     data_sent......................: 2.3 MB  7.6 kB/s
     http_req_duration..............: avg=1.2s   min=250ms med=1.1s max=2.8s p(95)=2.4s
     http_reqs......................: 5000    16.66/s
```

### 5.2 Worker Performance Testing

**File**: `tests/load/worker-performance.k6.js`

```javascript
// Test worker throughput (how many scans can be processed per second)
import http from 'k6/http';
import { check } from 'k6';

export const options = {
  scenarios: {
    constant_request_rate: {
      executor: 'constant-arrival-rate',
      rate: 100, // 100 requests per second
      timeUnit: '1s',
      duration: '1m',
      preAllocatedVUs: 50,
      maxVUs: 200,
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<5000'], // Workers have 5s SLA (longer than API)
  },
};

export default function () {
  const url = 'http://localhost:3000/api/scan';
  const payload = JSON.stringify({
    chain: 'SOLANA',
    tokenAddress: `random-address-${Math.random()}`,
  });

  const params = {
    headers: { 'Content-Type': 'application/json', 'Authorization': 'Bearer test' },
  };

  const response = http.post(url, payload, params);

  check(response, {
    'status is 200': (r) => r.status === 200,
  });
}
```

---

## 6. Continuous Integration (CI/CD)

### 6.1 GitHub Actions Workflow

**File**: `.github/workflows/test.yml`

```yaml
name: Test Suite

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  unit-tests:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: rugmunch_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379

    steps:
      - uses: actions/checkout@v3

      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '20'
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Run database migrations
        run: npx prisma migrate deploy
        env:
          DATABASE_URL: postgresql://test:test@localhost:5432/rugmunch_test

      - name: Run unit tests
        run: npm run test:unit

      - name: Run integration tests
        run: npm run test:integration
        env:
          DATABASE_URL: postgresql://test:test@localhost:5432/rugmunch_test
          REDIS_URL: redis://localhost:6379

      - name: Upload coverage reports
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage/lcov.info

  e2e-tests:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '20'

      - name: Install dependencies
        run: npm ci

      - name: Install Playwright browsers
        run: npx playwright install --with-deps

      - name: Run E2E tests
        run: npm run test:e2e

      - name: Upload Playwright report
        if: failure()
        uses: actions/upload-artifact@v3
        with:
          name: playwright-report
          path: playwright-report/

  load-tests:
    runs-on: ubuntu-latest
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'

    steps:
      - uses: actions/checkout@v3

      - name: Install k6
        run: |
          sudo gpg -k
          sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
          echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
          sudo apt-get update
          sudo apt-get install k6

      - name: Run load tests
        run: k6 run tests/load/scan-api.k6.js
```

---

## 7. Test Data Factories

### 7.1 Factory Pattern for Test Data

**File**: `tests/factories/scan.factory.ts`

```typescript
import { Prisma } from '@prisma/client'

export class ScanFactory {
  static build(overrides?: Partial<Prisma.ScanCreateInput>): Prisma.ScanCreateInput {
    return {
      userId: 'test-user-id',
      chain: 'SOLANA',
      tokenAddress: 'So11111111111111111111111111111111111111112',
      riskScore: 85,
      category: 'SAFE',
      breakdown: {
        liquidity: 15,
        lpLock: 20,
        holderConcentration: 10,
        mintAuthority: 15,
        honeypot: 20,
        ownership: 5,
      },
      flags: [],
      ...overrides,
    }
  }

  static buildMany(count: number, overrides?: Partial<Prisma.ScanCreateInput>): Prisma.ScanCreateInput[] {
    return Array.from({ length: count }, () => this.build(overrides))
  }

  static buildScam(): Prisma.ScanCreateInput {
    return this.build({
      riskScore: 15,
      category: 'LIKELY_SCAM',
      breakdown: {
        liquidity: 0,
        lpLock: 0,
        holderConcentration: 0,
        mintAuthority: 0,
        honeypot: 0,
        ownership: 0,
      },
      flags: ['🚨 HONEYPOT DETECTED', '🚨 LP not locked', '⚠️ Low liquidity'],
    })
  }

  static buildSafe(): Prisma.ScanCreateInput {
    return this.build({
      riskScore: 95,
      category: 'SAFE',
      breakdown: {
        liquidity: 15,
        lpLock: 20,
        holderConcentration: 10,
        mintAuthority: 15,
        honeypot: 20,
        ownership: 5,
      },
      flags: [],
    })
  }
}
```

**Usage**:

```typescript
import { ScanFactory } from '@/tests/factories/scan.factory'

// Create test scan
const scan = await prisma.scan.create({
  data: ScanFactory.build({ chain: 'ETHEREUM' }),
})

// Create multiple scans
const scans = await prisma.scan.createMany({
  data: ScanFactory.buildMany(10),
})

// Create scam scan
const scamScan = await prisma.scan.create({
  data: ScanFactory.buildScam(),
})
```

---

## 8. Visual Regression Testing

### 8.1 Percy.io Setup (Optional)

**Why**: Catch visual regressions in UI components.

```bash
npm install -D @percy/cli @percy/playwright
```

```typescript
// tests/e2e/visual-regression.spec.ts
import { test } from '@playwright/test'
import percySnapshot from '@percy/playwright'

test('dashboard visual regression', async ({ page }) => {
  await page.goto('/')

  // Login
  await page.fill('[name="email"]', 'test@example.com')
  await page.fill('[name="password"]', 'password123')
  await page.click('button[type="submit"]')

  // Take Percy snapshot
  await percySnapshot(page, 'Dashboard - Home')
})

test('scan detail visual regression', async ({ page }) => {
  await page.goto('/scans/test-scan-id')

  await percySnapshot(page, 'Scan Detail - Safe Token')
})
```

**Run Visual Tests**:

```bash
npx percy exec -- npx playwright test tests/e2e/visual-regression.spec.ts
```

---

## 9. Test Metrics & Monitoring

### 9.1 Test Coverage Reports

```json
// package.json
{
  "scripts": {
    "test:unit": "vitest run --coverage",
    "test:integration": "vitest run tests/integration --coverage",
    "test:e2e": "playwright test",
    "test:load": "k6 run tests/load/scan-api.k6.js",
    "test:all": "npm run test:unit && npm run test:integration && npm run test:e2e"
  }
}
```

**Coverage Report**:

```bash
npm run test:unit

# Output:
 PASS  tests/unit/risk-scoring.test.ts
 PASS  tests/unit/validators.test.ts
 PASS  tests/unit/blockchain-api.test.ts

 Test Files  3 passed (3)
      Tests  42 passed (42)
   Start at  10:30:15
   Duration  1.23s

 % Coverage report from v8
 -----------------|---------|----------|---------|---------|
 File             | % Stmts | % Branch | % Funcs | % Lines |
 -----------------|---------|----------|---------|---------|
 All files        |   87.2  |   82.5   |   91.3  |   87.2  |
  risk-scoring.ts |   100   |   100    |   100   |   100   |
  validators.ts   |   95.2  |   88.9   |   100   |   95.2  |
  blockchain.ts   |   78.4  |   70.1   |   85.7  |   78.4  |
 -----------------|---------|----------|---------|---------|
```

### 9.2 Test Analytics Dashboard

Track test metrics over time:
- Test execution time trends
- Flaky test detection
- Coverage trends
- Failure rates by test type

**Tools**:
- **Codecov**: https://codecov.io (coverage tracking)
- **DataDog Test Visibility**: https://docs.datadoghq.com/tests/ (test analytics)
- **Allure Report**: https://allurereport.org (beautiful test reports)

---

## 10. Command Shortcuts

Use these shortcuts to quickly access specific topics:

- **#vitest** - Vitest unit testing
- **#integration** - Integration tests (API, database, queue)
- **#playwright** - Playwright E2E tests
- **#k6** - k6 load testing
- **#ci** - CI/CD pipelines (GitHub Actions)
- **#coverage** - Test coverage strategies
- **#mocking** - Mocking external APIs
- **#factories** - Test data factories
- **#visual** - Visual regression testing
- **#tdd** - Test-driven development

---

## 11. Reference Materials

### 11.1 CryptoRugMunch Documentation

**Related Skills**:
- `rugmunch-architect` - System architecture overview
- `crypto-scam-analyst` - Risk scoring algorithm (critical to test)
- `security-auditor` - Security testing best practices

**Project Docs**:
- `/docs/03-TECHNICAL/operations/testing-strategy.md` - Complete testing strategy
- `/docs/03-TECHNICAL/integrations/telegram-bot-risk-algorithm.md` - Algorithm to test

### 11.2 Testing Tools

**Vitest**:
- Docs: https://vitest.dev
- API Reference: https://vitest.dev/api

**Playwright**:
- Docs: https://playwright.dev
- Best Practices: https://playwright.dev/docs/best-practices

**k6**:
- Docs: https://k6.io/docs
- Examples: https://k6.io/docs/examples

**Testcontainers**:
- Docs: https://testcontainers.com (Docker containers for integration tests)

---

## Summary

The **Testing & QA Specialist** skill provides comprehensive expertise for ensuring CryptoRugMunch's reliability, correctness, and performance through automated testing. Key capabilities:

1. **Unit Testing**: Vitest for business logic, utilities, risk scoring algorithm (100% coverage)
2. **Integration Testing**: API endpoints, database operations, BullMQ queue
3. **E2E Testing**: Playwright for user flows (scan, payment, dashboard)
4. **Load Testing**: k6 for performance under load (3-second SLA validation)
5. **CI/CD**: GitHub Actions for automated test runs, coverage tracking
6. **Test Data**: Factory pattern for maintainable test data
7. **Monitoring**: Test metrics, coverage trends, flaky test detection

**Testing Targets**:
- Risk scoring: 100% coverage (critical)
- Payment logic: 100% coverage (critical)
- API endpoints: 90%+ coverage
- Overall codebase: 80%+ coverage

**Timeline**: Testing implemented alongside each feature (TDD approach).

**Next Steps**:
1. Set up Vitest with coverage reporting
2. Write unit tests for risk scoring algorithm
3. Create integration tests for API endpoints
4. Build Playwright E2E tests for critical flows
5. Add k6 load tests for performance validation
6. Configure GitHub Actions CI/CD pipeline
7. Monitor test metrics in DataDog

---

**Built with confidence through comprehensive testing** ✅
**If it isn't tested, it doesn't work** 🧪
