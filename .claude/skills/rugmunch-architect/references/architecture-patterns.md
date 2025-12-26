# Modular Monolith Architecture Patterns

> Reference material for rugmunch-architect skill
> Links to: `.claude/agents/modular-monolith-architect.md`

## Pattern 1: Module Structure & Public API

Every module follows this strict structure:

```
src/modules/{module-name}/
├── {module}.service.ts       # Business logic (pure, testable)
├── {module}.repository.ts    # Database access (Prisma only)
├── {module}.controller.ts    # API routes (Fastify)
├── {module}.types.ts         # TypeScript types/interfaces
├── {module}.errors.ts        # Module-specific errors
├── index.ts                  # PUBLIC API (ONLY this can be imported!)
└── __tests__/
    ├── {module}.service.test.ts
    ├── {module}.repository.test.ts
    └── {module}.integration.test.ts
```

### Public API Pattern (index.ts)

**The index.ts file is the ONLY entry point to a module:**

```typescript
// src/modules/scan/index.ts
// This is the PUBLIC API - external modules can ONLY import from here

export { ScanService } from './scan.service';
export { ScanRepository } from './scan.repository';
export { registerScanRoutes } from './scan.controller';

// Export types
export type {
  Scan,
  CreateScanInput,
  ScanResult,
  RiskMetric,
  RiskLevel,
} from './scan.types';

// Export errors
export { ScanError, ScanNotFoundError, InvalidTokenAddressError } from './scan.errors';

// DO NOT export internal helpers, utilities, or private functions
// Those should remain encapsulated within the module
```

### Import Rules (ENFORCED)

✅ **ALLOWED** (Public API imports):
```typescript
// Other modules import from index.ts ONLY
import { ScanService } from '@/modules/scan';
import { UserRepository } from '@/modules/user';
import { registerPaymentRoutes } from '@/modules/payment';
```

❌ **FORBIDDEN** (Direct internal imports - will be BLOCKED by hook):
```typescript
// Direct imports to module internals
import { ScanService } from '@/modules/scan/scan.service'; // ❌ NO!
import { calculateScore } from '@/modules/scan/helpers'; // ❌ NO!
import { scanRepository } from '@/modules/scan/scan.repository'; // ❌ NO!
```

**Why this matters:**
- Encapsulation: Internal refactoring doesn't break external code
- Clear contracts: index.ts defines the module's interface
- Testability: Easy to mock the entire module
- Maintainability: Obvious what's public vs private

---

## Pattern 2: Service Layer (Business Logic)

Services contain pure business logic with NO side effects:

```typescript
// src/modules/scan/scan.service.ts
import { ScanRepository } from './scan.repository';
import { BlockchainApiService } from '@/modules/blockchain-api';
import type { CreateScanInput, Scan, RiskScore } from './scan.types';
import { ScanError, InvalidTokenAddressError } from './scan.errors';
import { logger } from '@/shared/logger';
import { metrics } from '@/shared/metrics';

export class ScanService {
  constructor(
    private repository: ScanRepository,
    private blockchainApi: BlockchainApiService
  ) {}

  async scanToken(
    tokenAddress: string,
    userId: string,
    tier: 'free' | 'premium'
  ): Promise<Scan> {
    const startTime = Date.now();

    try {
      // 1. Validation
      if (!this.isValidSolanaAddress(tokenAddress)) {
        throw new InvalidTokenAddressError(tokenAddress);
      }

      // 2. Fetch blockchain data
      logger.info({ tokenAddress, userId, tier }, 'Starting token scan');
      const metrics = await this.blockchainApi.getTokenMetrics(tokenAddress);

      // 3. Calculate risk score
      const riskScore = await this.calculateRiskScore(metrics, tier);

      // 4. Persist to database
      const scan = await this.repository.create({
        tokenAddress,
        userId,
        tier,
        riskScore: riskScore.score,
        riskLevel: riskScore.level,
        metrics: riskScore.metrics,
        redFlags: riskScore.redFlags,
      });

      // 5. Metrics and logging
      const duration = Date.now() - startTime;
      metrics.timing('scan.duration', duration, { tier, risk: scan.riskLevel });
      metrics.increment('scan.success', 1, { tier, risk: scan.riskLevel });

      logger.info(
        { scanId: scan.id, duration, riskLevel: scan.riskLevel },
        'Token scan completed'
      );

      return scan;
    } catch (error) {
      const duration = Date.now() - startTime;
      metrics.increment('scan.error', 1, { tier, error_type: error.name });
      logger.error({ error, tokenAddress, userId, tier }, 'Token scan failed');
      throw new ScanError('Failed to scan token', { cause: error });
    }
  }

  async calculateRiskScore(
    metrics: TokenMetrics,
    tier: 'free' | 'premium'
  ): Promise<RiskScore> {
    // Pure calculation logic (no I/O, fully testable)
    const analyzers = this.getAnalyzersForTier(tier);
    const results = await Promise.all(
      analyzers.map(analyzer => analyzer.analyze(metrics))
    );

    const finalScore = results.reduce(
      (sum, result) => sum + result.score * result.weight,
      0
    );

    const redFlags = this.checkRedFlags(results);

    return {
      score: finalScore,
      level: this.determineRiskLevel(finalScore, redFlags),
      metrics: results,
      redFlags,
    };
  }

  private isValidSolanaAddress(address: string): boolean {
    return /^[1-9A-HJ-NP-Za-km-z]{32,44}$/.test(address);
  }

  // ... more private helpers
}
```

**Service Layer Rules:**
- ✅ Pure business logic, testable without mocks
- ✅ Dependency injection (constructor parameters)
- ✅ Comprehensive logging and metrics
- ✅ Structured error handling
- ❌ NO direct database calls (use repository)
- ❌ NO direct HTTP/external API calls (use dedicated service)
- ❌ NO console.log (use logger)

---

## Pattern 3: Repository Layer (Database Access)

Repositories handle ALL database operations:

```typescript
// src/modules/scan/scan.repository.ts
import { prisma } from '@/shared/database';
import type { Scan, CreateScanInput, UpdateScanInput } from './scan.types';
import { ScanNotFoundError } from './scan.errors';
import { logger } from '@/shared/logger';

export class ScanRepository {
  async create(data: CreateScanInput): Promise<Scan> {
    try {
      return await prisma.scan.create({
        data: {
          tokenAddress: data.tokenAddress,
          userId: data.userId,
          tier: data.tier,
          riskScore: data.riskScore,
          riskLevel: data.riskLevel,
          metrics: data.metrics, // JSON field
          redFlags: data.redFlags, // JSON field
        },
        include: {
          user: true, // Include related user if needed
        },
      });
    } catch (error) {
      logger.error({ error, data }, 'Failed to create scan');
      throw error;
    }
  }

  async findById(id: string): Promise<Scan | null> {
    return await prisma.scan.findUnique({
      where: { id },
      include: { user: true },
    });
  }

  async findByTokenAddress(tokenAddress: string): Promise<Scan[]> {
    return await prisma.scan.findMany({
      where: { tokenAddress },
      orderBy: { createdAt: 'desc' },
      take: 10, // Limit to recent scans
    });
  }

  async findByUser(userId: string, limit = 50): Promise<Scan[]> {
    return await prisma.scan.findMany({
      where: { userId },
      orderBy: { createdAt: 'desc' },
      take: limit,
    });
  }

  async update(id: string, data: UpdateScanInput): Promise<Scan> {
    const existing = await this.findById(id);
    if (!existing) {
      throw new ScanNotFoundError(id);
    }

    return await prisma.scan.update({
      where: { id },
      data,
    });
  }

  async delete(id: string): Promise<void> {
    await prisma.scan.delete({ where: { id } });
  }

  async countByUser(userId: string, since?: Date): Promise<number> {
    return await prisma.scan.count({
      where: {
        userId,
        createdAt: since ? { gte: since } : undefined,
      },
    });
  }
}
```

**Repository Layer Rules:**
- ✅ ALL database access goes through repositories
- ✅ Simple CRUD operations (create, read, update, delete)
- ✅ Domain-specific queries (findByTokenAddress, countByUser)
- ✅ Include related entities when needed
- ❌ NO business logic in repositories
- ❌ NO external API calls
- ❌ Keep queries simple (complex analytics belong in dedicated analytics module)

---

## Pattern 4: Controller Layer (API Routes)

Controllers handle HTTP requests and responses:

```typescript
// src/modules/scan/scan.controller.ts
import type { FastifyInstance, FastifyRequest, FastifyReply } from 'fastify';
import { ScanService } from './scan.service';
import { ScanRepository } from './scan.repository';
import { BlockchainApiService } from '@/modules/blockchain-api';
import { authenticate } from '@/shared/auth';
import { rateLimit } from '@/shared/rate-limit';
import { ScanRequestSchema, ScanResponseSchema } from './scan.schemas';
import { InvalidTokenAddressError, ScanError } from './scan.errors';
import { logger } from '@/shared/logger';

export async function registerScanRoutes(app: FastifyInstance) {
  const scanService = new ScanService(
    new ScanRepository(),
    new BlockchainApiService()
  );

  // POST /api/scan - Scan a token
  app.post(
    '/api/scan',
    {
      schema: {
        body: ScanRequestSchema,
        response: {
          200: ScanResponseSchema,
        },
        tags: ['Scan'],
        description: 'Analyze a Solana token for scam risk',
      },
      preHandler: [authenticate, rateLimit({ tier: 'user' })],
    },
    async (
      request: FastifyRequest<{ Body: { tokenAddress: string } }>,
      reply: FastifyReply
    ) => {
      const { tokenAddress } = request.body;
      const userId = request.user!.id;
      const tier = request.user!.tier;

      try {
        const scan = await scanService.scanToken(tokenAddress, userId, tier);

        return reply.status(200).send({
          success: true,
          data: {
            scanId: scan.id,
            tokenAddress: scan.tokenAddress,
            riskScore: scan.riskScore,
            riskLevel: scan.riskLevel,
            metrics: scan.metrics,
            redFlags: scan.redFlags,
            scannedAt: scan.createdAt,
          },
        });
      } catch (error) {
        if (error instanceof InvalidTokenAddressError) {
          return reply.status(400).send({
            success: false,
            error: {
              code: 'INVALID_TOKEN_ADDRESS',
              message: error.message,
            },
          });
        }

        logger.error({ error, tokenAddress, userId }, 'Scan request failed');

        return reply.status(500).send({
          success: false,
          error: {
            code: 'SCAN_FAILED',
            message: 'Failed to scan token. Please try again.',
          },
        });
      }
    }
  );

  // GET /api/scan/:scanId - Get scan by ID
  app.get(
    '/api/scan/:scanId',
    {
      schema: {
        params: { type: 'object', properties: { scanId: { type: 'string' } } },
        response: { 200: ScanResponseSchema },
        tags: ['Scan'],
      },
      preHandler: [authenticate],
    },
    async (
      request: FastifyRequest<{ Params: { scanId: string } }>,
      reply: FastifyReply
    ) => {
      const { scanId } = request.params;
      const userId = request.user!.id;

      const repository = new ScanRepository();
      const scan = await repository.findById(scanId);

      if (!scan) {
        return reply.status(404).send({
          success: false,
          error: { code: 'SCAN_NOT_FOUND', message: 'Scan not found' },
        });
      }

      // Authorization check
      if (scan.userId !== userId) {
        return reply.status(403).send({
          success: false,
          error: { code: 'FORBIDDEN', message: 'Access denied' },
        });
      }

      return reply.status(200).send({ success: true, data: scan });
    }
  );

  // GET /api/scan/history - Get user's scan history
  app.get(
    '/api/scan/history',
    {
      schema: {
        querystring: {
          type: 'object',
          properties: { limit: { type: 'number', default: 50 } },
        },
        tags: ['Scan'],
      },
      preHandler: [authenticate],
    },
    async (
      request: FastifyRequest<{ Querystring: { limit?: number } }>,
      reply: FastifyReply
    ) => {
      const userId = request.user!.id;
      const limit = request.query.limit || 50;

      const repository = new ScanRepository();
      const scans = await repository.findByUser(userId, limit);

      return reply.status(200).send({ success: true, data: scans });
    }
  );
}
```

**Controller Layer Rules:**
- ✅ Handle HTTP request/response only
- ✅ Schema validation (Fastify JSON Schema)
- ✅ Authentication and authorization
- ✅ Rate limiting
- ✅ Error handling with proper HTTP status codes
- ✅ OpenAPI/Swagger documentation (schema tags)
- ❌ NO business logic in controllers
- ❌ NO direct database access (use service/repository)

---

## Pattern 5: Dependency Graph & Module Boundaries

### Allowed Dependencies (NO CYCLES)

```
Dependency Flow (Bottom-Up):

shared/           # Utilities, logger, metrics, database client
  ↑
blockchain-api/   # External API integrations (Helius, Birdeye)
  ↑
scan/             # Risk scoring, analysis
  ↑
telegram/         # Bot commands, message handling
payment/          # Stripe integration
  ↑
api/              # Main API server, route registration
```

### Example Dependency Rules

✅ **ALLOWED**:
```typescript
// scan module can depend on blockchain-api
import { BlockchainApiService } from '@/modules/blockchain-api';

// telegram module can depend on scan
import { ScanService } from '@/modules/scan';

// payment module can depend on user
import { UserService } from '@/modules/user';
```

❌ **FORBIDDEN** (creates cycles):
```typescript
// blockchain-api CANNOT depend on scan (creates cycle)
import { ScanService } from '@/modules/scan'; // ❌ NO!

// user CANNOT depend on payment (creates cycle)
import { PaymentService } from '@/modules/payment'; // ❌ NO!
```

### Detecting Circular Dependencies

Use `madge` to check for cycles:

```bash
# Check for circular dependencies
npx madge --circular --extensions ts src/modules/

# Visualize dependency graph
npx madge --image graph.svg src/modules/
```

---

## Pattern 6: Testing Modular Monoliths

### Unit Tests (Service Layer)

```typescript
// src/modules/scan/__tests__/scan.service.test.ts
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ScanService } from '../scan.service';
import { ScanRepository } from '../scan.repository';
import { BlockchainApiService } from '@/modules/blockchain-api';

// Mock dependencies
vi.mock('../scan.repository');
vi.mock('@/modules/blockchain-api');

describe('ScanService', () => {
  let service: ScanService;
  let mockRepository: ScanRepository;
  let mockBlockchainApi: BlockchainApiService;

  beforeEach(() => {
    mockRepository = new ScanRepository() as any;
    mockBlockchainApi = new BlockchainApiService() as any;
    service = new ScanService(mockRepository, mockBlockchainApi);
  });

  describe('scanToken', () => {
    it('should return LOW_RISK for token with good metrics', async () => {
      const tokenAddress = 'So11111111111111111111111111111111111111112';
      const userId = 'user123';
      const tier = 'premium';

      mockBlockchainApi.getTokenMetrics = vi.fn().mockResolvedValue({
        liquidity: 150000,
        holderConcentration: 25,
        mintAuthority: null,
        freezeAuthority: null,
      });

      mockRepository.create = vi.fn().mockResolvedValue({
        id: 'scan123',
        tokenAddress,
        userId,
        tier,
        riskScore: 15,
        riskLevel: 'LOW_RISK',
        metrics: [],
        redFlags: [],
      });

      const result = await service.scanToken(tokenAddress, userId, tier);

      expect(result.riskLevel).toBe('LOW_RISK');
      expect(result.riskScore).toBeLessThan(40);
      expect(mockBlockchainApi.getTokenMetrics).toHaveBeenCalledWith(tokenAddress);
      expect(mockRepository.create).toHaveBeenCalled();
    });

    it('should throw InvalidTokenAddressError for malformed address', async () => {
      await expect(
        service.scanToken('invalid', 'user123', 'free')
      ).rejects.toThrow('Invalid Solana address');
    });
  });
});
```

### Integration Tests (API + Database)

```typescript
// src/modules/scan/__tests__/scan.integration.test.ts
import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { buildApp } from '@/app';
import type { FastifyInstance } from 'fastify';
import { prisma } from '@/shared/database';

describe('Scan API Integration', () => {
  let app: FastifyInstance;
  let authToken: string;
  let userId: string;

  beforeAll(async () => {
    app = await buildApp();
    await app.ready();

    // Create test user and get auth token
    const user = await prisma.user.create({
      data: { telegramId: 'test123', tier: 'premium' },
    });
    userId = user.id;
    authToken = generateTestToken(userId);
  });

  afterAll(async () => {
    await prisma.scan.deleteMany({ where: { userId } });
    await prisma.user.delete({ where: { id: userId } });
    await app.close();
  });

  it('POST /api/scan - should scan token and return risk score', async () => {
    const response = await app.inject({
      method: 'POST',
      url: '/api/scan',
      headers: { Authorization: `Bearer ${authToken}` },
      payload: {
        tokenAddress: 'So11111111111111111111111111111111111111112',
      },
    });

    expect(response.statusCode).toBe(200);
    const body = JSON.parse(response.body);
    expect(body.success).toBe(true);
    expect(body.data.riskScore).toBeGreaterThanOrEqual(0);
    expect(body.data.riskScore).toBeLessThanOrEqual(100);
    expect(['LOW_RISK', 'MEDIUM_RISK', 'HIGH_RISK']).toContain(body.data.riskLevel);
  });

  it('GET /api/scan/history - should return user scan history', async () => {
    const response = await app.inject({
      method: 'GET',
      url: '/api/scan/history',
      headers: { Authorization: `Bearer ${authToken}` },
    });

    expect(response.statusCode).toBe(200);
    const body = JSON.parse(response.body);
    expect(body.success).toBe(true);
    expect(Array.isArray(body.data)).toBe(true);
  });
});
```

---

## Pattern 7: Error Handling

### Module-Specific Errors

```typescript
// src/modules/scan/scan.errors.ts
export class ScanError extends Error {
  constructor(message: string, options?: ErrorOptions) {
    super(message, options);
    this.name = 'ScanError';
  }
}

export class InvalidTokenAddressError extends ScanError {
  constructor(address: string) {
    super(`Invalid Solana address: ${address}`);
    this.name = 'InvalidTokenAddressError';
  }
}

export class ScanNotFoundError extends ScanError {
  constructor(scanId: string) {
    super(`Scan not found: ${scanId}`);
    this.name = 'ScanNotFoundError';
  }
}

export class RateLimitExceededError extends ScanError {
  constructor(tier: string, limit: number) {
    super(`Rate limit exceeded for ${tier} tier (${limit} scans/day)`);
    this.name = 'RateLimitExceededError';
  }
}
```

### Error Handling in Controllers

```typescript
// Map errors to HTTP status codes
try {
  const result = await service.scanToken(tokenAddress, userId, tier);
  return reply.status(200).send({ success: true, data: result });
} catch (error) {
  if (error instanceof InvalidTokenAddressError) {
    return reply.status(400).send({
      success: false,
      error: { code: 'INVALID_TOKEN_ADDRESS', message: error.message },
    });
  }

  if (error instanceof RateLimitExceededError) {
    return reply.status(429).send({
      success: false,
      error: { code: 'RATE_LIMIT_EXCEEDED', message: error.message },
    });
  }

  if (error instanceof ScanNotFoundError) {
    return reply.status(404).send({
      success: false,
      error: { code: 'SCAN_NOT_FOUND', message: error.message },
    });
  }

  // Unknown error
  logger.error({ error, tokenAddress, userId }, 'Unexpected scan error');
  Sentry.captureException(error, { extra: { tokenAddress, userId } });

  return reply.status(500).send({
    success: false,
    error: { code: 'INTERNAL_ERROR', message: 'An unexpected error occurred' },
  });
}
```

---

## Related Files

- **Agent**: `.claude/agents/modular-monolith-architect.md` - Architecture enforcement
- **Docs**: `docs/03-TECHNICAL/architecture/modular-monolith-structure.md` - Full specification
- **Hook**: `.claude/hooks/enforce-module-boundaries.sh` - Automated boundary checking
