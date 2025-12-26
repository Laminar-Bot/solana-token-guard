# CryptoRugMunch Tech Stack Decisions

**Status**: ✅ Complete
**Last Updated**: 2025-01-19

This reference consolidates all Architecture Decision Records (ADRs) and technology choices made for CryptoRugMunch.

---

## Table of Contents

1. [Core Technology Choices](#core-technology-choices)
2. [Architecture Decision Records (ADRs)](#architecture-decision-records-adrs)
3. [Trade-offs Analysis](#trade-offs-analysis)
4. [Alternative Approaches Considered](#alternative-approaches-considered)
5. [Future Evolution Path](#future-evolution-path)

---

## Core Technology Choices

### Backend Framework: Fastify

**Decision**: Use Fastify as the HTTP server framework

**Rationale**:
- **Performance**: 30-40% faster than Express (~40k req/s vs ~30k req/s)
- **TypeScript-first**: Native TypeScript support with excellent type inference
- **Schema validation**: Built-in JSON Schema validation (faster than Joi/Zod)
- **Plugin system**: Excellent ecosystem with encapsulation
- **Low overhead**: Minimal memory footprint critical for worker performance

**Trade-offs**:
- ✅ Pro: Best-in-class performance for Node.js
- ✅ Pro: Growing ecosystem with strong community
- ⚠️ Con: Smaller ecosystem than Express
- ⚠️ Con: More opinionated (good for consistency)

**Alternatives considered**:
- Express: More mature ecosystem, but slower and less TypeScript-friendly
- NestJS: Too heavy for our needs, unnecessary abstractions
- Hono: Very fast, but immature ecosystem

**Related**: `docs/03-TECHNICAL/architecture/tech-stack-rationale.md`

---

### Job Queue: BullMQ

**Decision**: Use BullMQ for asynchronous job processing

**Rationale**:
- **Reliability**: Redis-backed with automatic retries and exponential backoff
- **Performance**: Handles 10,000+ jobs/second easily
- **Features**: Priority queues, delayed jobs, job events, metrics
- **Observability**: Built-in metrics, event listeners for monitoring
- **Worker scaling**: Easy horizontal scaling (add more worker processes)

**Trade-offs**:
- ✅ Pro: Battle-tested in production (used by Vercel, Linear, etc.)
- ✅ Pro: Excellent TypeScript support
- ✅ Pro: Redis as single dependency (we already need Redis for caching)
- ⚠️ Con: Redis is a single point of failure (mitigated with Redis Cluster)

**Alternatives considered**:
- AWS SQS: Higher latency, requires AWS, more expensive
- RabbitMQ: More complex, requires separate infrastructure
- Kafka: Overkill for our use case, complex operational overhead

**Configuration**:
```typescript
{
  defaultJobOptions: {
    attempts: 3,
    backoff: { type: 'exponential', delay: 2000 },
    removeOnComplete: { count: 1000, age: 86400 },
    removeOnFail: { count: 5000, age: 604800 },
  },
}
```

**Related**: `docs/03-TECHNICAL/integrations/telegram-bot-risk-algorithm.md`

---

### Database: PostgreSQL with Prisma ORM

**Decision**: Use PostgreSQL as primary database with Prisma ORM

**Rationale**:
- **PostgreSQL advantages**:
  - ACID compliance for financial transactions (subscriptions)
  - JSONB support for flexible scan data storage
  - Excellent performance with proper indexing
  - Advanced features: CTEs, window functions, full-text search
  - Proven reliability at scale

- **Prisma advantages**:
  - Type-safe database client (no runtime errors from typos)
  - Excellent TypeScript integration
  - Migration system with rollback support
  - Prisma Studio for database exploration
  - Query builder prevents SQL injection

**Trade-offs**:
- ✅ Pro: Strong consistency guarantees
- ✅ Pro: Rich query capabilities
- ✅ Pro: Excellent tooling (pgAdmin, Prisma Studio)
- ⚠️ Con: More expensive than DynamoDB at scale (mitigated with read replicas)
- ⚠️ Con: Vertical scaling limits (acceptable for our scale)

**Alternatives considered**:
- MongoDB: Flexible schema, but no transactions, harder to maintain consistency
- DynamoDB: Cheaper at scale, but complex query model, vendor lock-in
- MySQL: Good, but PostgreSQL has better JSON support and features

**Related**: `docs/03-TECHNICAL/architecture/data-model.md`

---

### Caching: Redis

**Decision**: Use Redis for caching and session storage

**Rationale**:
- **Performance**: Sub-millisecond latency, 100k+ ops/second
- **Features**: TTL, pub/sub, sorted sets for leaderboards
- **Use cases**:
  - Token metadata caching (5-minute TTL)
  - Rate limiting (sliding window algorithm)
  - Session storage
  - BullMQ backing store
  - Leaderboard (sorted sets)

**Trade-offs**:
- ✅ Pro: Extremely fast, simple to use
- ✅ Pro: Multiple use cases with single instance
- ⚠️ Con: In-memory, limited by RAM (acceptable with eviction policies)
- ⚠️ Con: Single point of failure (mitigated with Redis Cluster in production)

**Configuration**:
```bash
# Eviction policy
maxmemory-policy allkeys-lru

# Persistence
save 900 1      # Save after 900s if 1 key changed
save 300 10     # Save after 300s if 10 keys changed
appendonly yes  # AOF for durability
```

**Related**: `docs/03-TECHNICAL/architecture/scalability-performance.md`

---

### Telegram Bot: Grammy.js

**Decision**: Use Grammy.js as the Telegram Bot framework

**Rationale**:
- **TypeScript-first**: Excellent type safety and IDE support
- **Modern API**: Async/await, promises (no callbacks)
- **Plugin system**: Conversations, sessions, rate limiting
- **Performance**: Handles 100k+ messages/day easily
- **Active development**: Regular updates, responsive maintainer

**Trade-offs**:
- ✅ Pro: Best TypeScript experience in Telegram bot ecosystem
- ✅ Pro: Clean, modern API design
- ✅ Pro: Excellent documentation with examples
- ⚠️ Con: Smaller community than node-telegram-bot-api
- ⚠️ Con: Less mature (v1.0 released in 2022)

**Alternatives considered**:
- node-telegram-bot-api: More mature, but callback-based, poor TypeScript support
- Telegraf: Good, but Grammy has better TypeScript experience
- python-telegram-bot: Python ecosystem, requires separate service

**Related**: `docs/03-TECHNICAL/integrations/telegram-bot-setup.md`

---

### Monitoring: DataDog + Sentry

**Decision**: Use DataDog for metrics/logs, Sentry for error tracking

**Rationale**:

**DataDog**:
- **Unified platform**: Metrics, logs, APM, dashboards in one place
- **StatsD integration**: Simple metric collection with hot-shots
- **Log aggregation**: Structured logging with Pino → DataDog
- **Alerting**: Flexible alert rules with PagerDuty integration
- **Cost**: Reasonable at our scale (~$200/month for 1M metrics/day)

**Sentry**:
- **Error tracking**: Automatic error grouping, stack traces
- **Performance monitoring**: Transaction traces, bottleneck detection
- **Release tracking**: Correlate errors with deployments
- **Breadcrumbs**: User actions leading to errors
- **Cost**: Free tier sufficient for MVP (5k events/month)

**Trade-offs**:
- ✅ Pro: Best-in-class observability
- ✅ Pro: Excellent integrations (Slack, PagerDuty)
- ⚠️ Con: Cost scales with usage (acceptable, ~$300/month at 100K scans/month)
- ⚠️ Con: Two separate platforms (acceptable, they serve different purposes)

**Alternatives considered**:
- New Relic: More expensive, similar features
- Prometheus + Grafana: Open-source, but requires self-hosting, more ops overhead
- CloudWatch: AWS-specific, poor querying experience

**Related**: `docs/03-TECHNICAL/operations/monitoring-alerting-setup.md`

---

### Deployment: Railway (Staging) → AWS ECS (Production)

**Decision**: Start with Railway, migrate to AWS ECS at scale

**Rationale**:

**Railway (Staging)**:
- **Speed**: Deploy in < 5 minutes
- **Simplicity**: Managed PostgreSQL, Redis included
- **Cost**: ~$20/month for staging environment
- **CI/CD**: Automatic deploys from GitHub
- **Developer experience**: Excellent for early development

**AWS ECS (Production)**:
- **Scale**: Auto-scaling workers based on queue depth
- **Cost efficiency**: ~$200/month for production (vs ~$500 on Railway)
- **Control**: Full infrastructure control, VPC, IAM roles
- **Reliability**: Multi-AZ deployments, load balancers
- **Migration path**: Clear upgrade path when needed

**Trade-offs**:
- ✅ Pro: Fast MVP launch on Railway
- ✅ Pro: Cost savings at scale with AWS
- ⚠️ Con: Migration effort when moving to AWS (acceptable, planned for Month 6)
- ⚠️ Con: AWS complexity (mitigated with Terraform, good docs)

**Migration trigger**: 100K scans/month or $500/month Railway costs

**Related**: `docs/03-TECHNICAL/operations/worker-deployment.md`

---

## Architecture Decision Records (ADRs)

### ADR-001: Event-Driven Architecture with BullMQ

**Context**: Need reliable, scalable job processing for token scans

**Decision**: Use event-driven architecture with BullMQ queue

**Consequences**:
- ✅ Workers can scale independently from API servers
- ✅ Automatic retries for failed scans
- ✅ Graceful degradation under load
- ⚠️ Increased system complexity (queue management)
- ⚠️ Redis as critical dependency

**Status**: Accepted

**Related**: `docs/03-TECHNICAL/architecture/system-architecture.md`

---

### ADR-002: Modular Monolith (Not Microservices)

**Context**: Need to balance code organization with operational simplicity

**Decision**: Use modular monolith with clear module boundaries

**Rationale**:
- Simpler deployment (single application)
- Easier debugging (no distributed tracing needed at MVP)
- Lower infrastructure costs (1 app vs 5+ services)
- Easy to extract modules into microservices later if needed

**Module boundaries**:
```
src/modules/
├── scan/       # Risk scoring, providers, queue
├── telegram/   # Bot handlers, commands
├── payment/    # Stripe integration
└── auth/       # User management
```

**Rules**:
- Modules communicate via well-defined interfaces
- No circular dependencies
- Shared types in `src/types/`
- Shared utilities in `src/utils/`

**Consequences**:
- ✅ Faster development velocity
- ✅ Lower operational overhead
- ✅ Easier to maintain consistency
- ⚠️ Harder to scale individual components (acceptable at our scale)
- ⚠️ Deployment is all-or-nothing (mitigated with blue/green deployments)

**Status**: Accepted

**Migration path**: Extract `scan` module to microservice when queue depth > 10K consistently

**Related**: `docs/03-TECHNICAL/architecture/system-architecture.md`

---

### ADR-003: Parallel API Calls for Performance

**Context**: 3-second SLA for scans requires fast external API calls

**Decision**: Make all blockchain API calls in parallel with Promise.all

**Implementation**:
```typescript
// ✅ GOOD: Parallel (fast)
const [metadata, liquidity, honeypot] = await Promise.all([
  helius.getMetadata(address),      // ~500ms
  birdeye.getLiquidity(address),    // ~700ms
  rugcheck.checkHoneypot(address),  // ~600ms
]);
// Total: ~700ms (slowest API)

// ❌ BAD: Sequential (slow)
const metadata = await helius.getMetadata(address);    // 500ms
const liquidity = await birdeye.getLiquidity(address); // 700ms
const honeypot = await rugcheck.checkHoneypot(address); // 600ms
// Total: 1,800ms
```

**Consequences**:
- ✅ 60-70% latency reduction (1.8s → 0.7s for API calls)
- ✅ Consistent performance under load
- ⚠️ All requests fail if one provider is down (mitigated with fallbacks)

**Status**: Accepted

**Related**: `docs/03-TECHNICAL/integrations/blockchain-api-integration.md`

---

### ADR-004: Aggressive Caching Strategy

**Context**: Blockchain data changes slowly, API costs add up quickly

**Decision**: Cache token metadata for 5 minutes, liquidity data for 1 minute

**Rationale**:
- Token metadata (name, symbol, supply) rarely changes
- Liquidity data changes more frequently, but 1-minute staleness acceptable
- 70%+ cache hit rate = 70% cost savings on API calls

**Cache invalidation**:
```typescript
// Metadata: 5 minutes
await redis.setex(`token:metadata:${address}`, 300, data);

// Liquidity: 1 minute
await redis.setex(`token:liquidity:${address}`, 60, data);

// Honeypot check: 10 minutes (rarely changes)
await redis.setex(`token:honeypot:${address}`, 600, data);
```

**Consequences**:
- ✅ 70%+ cost savings on API calls ($4.50/month vs $15/month)
- ✅ Faster response times (sub-millisecond cache reads)
- ⚠️ Slight staleness (acceptable for risk assessment)

**Status**: Accepted

**Monitoring**: Track cache hit rate in DataDog (target: 70%+)

**Related**: `docs/03-TECHNICAL/architecture/scalability-performance.md`

---

### ADR-005: JSONB for Scan Data Storage

**Context**: Need to store flexible, varying scan data efficiently

**Decision**: Use PostgreSQL JSONB for `rawData` and `breakdown` fields

**Rationale**:
- API responses vary by provider
- Future-proof for adding new metrics without schema migrations
- JSONB supports indexing and querying (unlike JSON)
- Easier to store full API responses for debugging

**Schema**:
```prisma
model Scan {
  id           String   @id @default(cuid())
  riskScore    Int      // 0-100
  riskCategory RiskCategory
  breakdown    Json     // Risk breakdown by metric
  rawData      Json     // Full API responses for debugging
}
```

**Consequences**:
- ✅ Flexible schema for varying data
- ✅ No migrations needed for new metrics
- ✅ Full data retention for debugging
- ⚠️ Larger database size (mitigated with data retention policy)

**Data retention**: Delete `rawData` after 30 days to save space

**Status**: Accepted

**Related**: `docs/03-TECHNICAL/architecture/data-model.md`

---

### ADR-006: Webhook Mode (Production) vs Polling (Development)

**Context**: Telegram supports both webhook and long-polling for bots

**Decision**: Use webhooks in production, polling in development

**Rationale**:

**Production (Webhook)**:
- Lower latency (instant message delivery)
- Lower server load (no constant polling)
- Required for high-volume bots (10k+ messages/day)

**Development (Polling)**:
- No HTTPS required (easier local dev)
- No webhook setup needed
- Easier to debug (logs visible immediately)

**Implementation**:
```typescript
if (process.env.NODE_ENV === 'production') {
  // Webhook mode
  await bot.api.setWebhook(`${WEBHOOK_URL}/telegram-webhook/${BOT_TOKEN}`);
} else {
  // Polling mode
  bot.start();
}
```

**Consequences**:
- ✅ Optimal mode for each environment
- ✅ Easier local development
- ⚠️ Need to remember to switch modes (automated in deployment scripts)

**Status**: Accepted

**Related**: `docs/03-TECHNICAL/integrations/telegram-bot-setup.md`

---

### ADR-007: StatsD + DataDog for Metrics (Not Custom Metrics)

**Context**: Need to track application metrics efficiently

**Decision**: Use StatsD protocol with `hot-shots` library to send metrics to DataDog

**Rationale**:
- **StatsD advantages**:
  - Fire-and-forget UDP (no blocking)
  - Minimal overhead (< 1ms per metric)
  - Buffer batching (efficient network usage)
  - Industry standard

- **Implementation**:
```typescript
import StatsD from 'hot-shots';

const metrics = new StatsD({
  host: 'localhost',
  port: 8125,
  prefix: 'rugmunch.',
  globalTags: { env: 'production', service: 'api' },
});

// Usage
metrics.increment('scan.success', 1, { tier: 'premium' });
metrics.timing('scan.duration', durationMs);
metrics.gauge('queue.depth', queueDepth);
```

**Consequences**:
- ✅ Zero blocking (UDP)
- ✅ Minimal overhead
- ✅ Automatic batching
- ⚠️ UDP can lose packets (acceptable for metrics)

**Status**: Accepted

**Related**: `docs/03-TECHNICAL/operations/monitoring-alerting-setup.md`

---

### ADR-008: Structured Logging with Pino

**Context**: Need efficient, queryable logging

**Decision**: Use Pino for structured JSON logging

**Rationale**:
- **Performance**: 5x faster than Winston, 10x faster than Bunyan
- **Structured**: JSON logs for easy DataDog querying
- **Context**: Automatic request ID, user ID, etc.
- **Ecosystem**: Excellent Fastify integration

**Implementation**:
```typescript
import pino from 'pino';

const logger = pino({
  level: process.env.LOG_LEVEL || 'info',
  transport: process.env.NODE_ENV === 'development'
    ? { target: 'pino-pretty' }
    : undefined,
});

// Usage
logger.info({ userId, tokenAddress, duration }, 'Scan completed');
```

**Log levels**:
- `error`: Critical issues requiring immediate attention
- `warn`: Degraded performance, fallback used
- `info`: Normal operations (scan completed, user signup)
- `debug`: Detailed debugging (local dev only)

**Consequences**:
- ✅ Extremely fast (minimal overhead)
- ✅ Easy DataDog querying (`@userId:123`)
- ✅ Automatic context propagation
- ⚠️ JSON logs harder to read locally (mitigated with pino-pretty)

**Status**: Accepted

**Related**: `docs/03-TECHNICAL/development/code-style-guide.md`

---

## Trade-offs Analysis

### Performance vs Developer Experience

**Decision**: Prioritize developer experience at MVP, optimize later

**Examples**:
- ✅ Use Prisma (slower than raw SQL, but much better DX)
- ✅ Use TypeScript (slower compile times, but type safety)
- ✅ Use Railway (more expensive, but faster deployment)

**Rationale**: At MVP scale (< 10K scans/day), performance isn't bottleneck. Developer velocity matters more.

**Optimization trigger**: When performance impacts user experience (p95 latency > 3s)

---

### Cost vs Simplicity

**Decision**: Pay for managed services early, self-host later if needed

**Examples**:
- ✅ Railway managed PostgreSQL ($20/month vs AWS RDS $50/month vs self-hosted $0)
- ✅ Helius API ($49/month vs running own Solana RPC node $500+/month)
- ✅ DataDog ($200/month vs self-hosted Prometheus/Grafana + ops time)

**Rationale**: At early stage, developer time is more valuable than infrastructure savings.

**Cost optimization trigger**: When infrastructure costs > $500/month, evaluate self-hosting

---

### Flexibility vs Constraints

**Decision**: Use opinionated tools to reduce decision fatigue

**Examples**:
- ✅ Fastify (opinionated plugin system) vs Express (unopinionated)
- ✅ Prisma (opinionated ORM) vs Knex (query builder)
- ✅ Pino (opinionated logging) vs Winston (configurable)

**Rationale**: Consistency more valuable than flexibility at our scale.

---

## Alternative Approaches Considered

### Alternative 1: Serverless (AWS Lambda)

**Why we didn't choose it**:
- ❌ Cold starts unacceptable for 3-second SLA
- ❌ Limited execution time (15 minutes max)
- ❌ More complex to debug
- ❌ Higher cost at our expected scale

**When it makes sense**: If traffic is extremely spiky (10x variance)

---

### Alternative 2: GraphQL API

**Why we didn't choose it**:
- ❌ Overkill for simple Telegram bot queries
- ❌ REST is simpler for external integrations
- ❌ Additional complexity (schema, resolvers, n+1 queries)

**When it makes sense**: If we build complex web dashboard with many related queries

---

### Alternative 3: Microservices

**Why we didn't choose it**:
- ❌ Operational complexity too high for team of 1-2
- ❌ Network overhead between services
- ❌ Distributed tracing required
- ❌ Harder to debug

**When it makes sense**: When team grows to 10+, or specific modules need independent scaling

---

### Alternative 4: NoSQL (MongoDB)

**Why we didn't choose it**:
- ❌ No transactions for subscription management
- ❌ Harder to maintain data consistency
- ❌ Less mature TypeScript tooling than Prisma

**When it makes sense**: If we need to store highly variable document structures

---

## Future Evolution Path

### Phase 1: MVP (Current)
- Modular monolith
- Railway hosting
- Manual scaling

### Phase 2: Growth (Month 6)
- Migrate to AWS ECS
- Implement auto-scaling
- Add read replicas for database
- Redis Cluster for high availability

### Phase 3: Scale (Month 12)
- Consider extracting `scan` module to microservice
- Add message queue for events (Kafka/SNS)
- Multi-region deployment
- CDN for web dashboard

### Phase 4: Optimization (Month 18+)
- Evaluate self-hosting Solana RPC node
- Consider edge functions for scan history
- Implement custom analytics pipeline
- Add machine learning for scam detection

---

## Decision Criteria Framework

When evaluating new technology choices, use this framework:

### 1. Performance
- Does it meet our SLA (3-second p95 latency)?
- Can it handle 10x our current scale?

### 2. Developer Experience
- Does it have good TypeScript support?
- Is the documentation excellent?
- Does it reduce boilerplate?

### 3. Cost
- What's the total cost of ownership (infrastructure + ops time)?
- Is there a managed option?
- Does it scale linearly with usage?

### 4. Reliability
- Is it battle-tested in production?
- What's the failure mode?
- Can we monitor it effectively?

### 5. Team Fit
- Can our team maintain it?
- Is there community support?
- Are there good learning resources?

**Priority**: Reliability > Performance > Developer Experience > Cost

---

## Related Documentation

- `docs/03-TECHNICAL/architecture/tech-stack-rationale.md` - Detailed tech stack rationale
- `docs/03-TECHNICAL/architecture/system-architecture.md` - Overall architecture
- `docs/03-TECHNICAL/architecture/scalability-performance.md` - Performance strategies
- `docs/03-TECHNICAL/operations/deployment-strategy.md` - Deployment evolution
