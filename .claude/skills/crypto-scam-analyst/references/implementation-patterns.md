# Risk Scoring Implementation Patterns

> Reference material for crypto-scam-analyst skill
> Links to: `.claude/agents/risk-scoring-agent.md`

## Pattern 1: Analyzer Plugin Architecture

Each of the 12 metrics is implemented as a self-contained analyzer class:

```typescript
interface RiskAnalyzer {
  name: string;
  weight: number;
  analyze(tokenAddress: string): Promise<MetricResult>;
}

interface MetricResult {
  name: string;
  score: number; // 0-100
  risk: 'LOW' | 'MEDIUM' | 'HIGH';
  weight: number;
  data: Record<string, any>;
  explanation: string;
}

// Registry of all analyzers
const ANALYZERS: RiskAnalyzer[] = [
  new LiquidityAnalyzer(),      // 25% weight
  new LPLockAnalyzer(),          // 20% weight
  new HolderConcentrationAnalyzer(), // 15% weight
  new MintAuthorityAnalyzer(),   // 15% weight
  new FreezeAuthorityAnalyzer(), // 10% weight
  new HoneypotAnalyzer(),        // 10% weight
  new VolumeRatioAnalyzer(),     // 5% weight
  new TokenAgeAnalyzer(),        // 5% weight
  new CreatorHistoryAnalyzer(),  // 5% weight
  new SocialVerificationAnalyzer(), // 5% weight
  new MetadataQualityAnalyzer(), // 3% weight
  new UpdateAuthorityAnalyzer(), // 2% weight
];

// Aggregate all metrics
const results = await Promise.all(
  ANALYZERS.map(a => a.analyze(tokenAddress))
);

// Calculate weighted score
const finalScore = results.reduce(
  (sum, result) => sum + (result.score * result.weight),
  0
);
```

**Benefits:**
- Easy to add new metrics
- Each metric is independently testable
- Weights can be tuned without changing logic
- Metrics can be disabled/enabled per tier

## Pattern 2: Multi-Provider Fallback

Never depend on a single blockchain API provider:

```typescript
async function fetchLiquidity(address: string): Promise<number> {
  const providers = [
    { name: 'birdeye', fn: () => birdeyeApi.getLiquidity(address) },
    { name: 'helius', fn: () => heliusApi.getLiquidity(address) },
    { name: 'rugcheck', fn: () => rugcheckApi.getLiquidity(address) },
  ];

  for (const provider of providers) {
    try {
      const result = await provider.fn();
      logger.info({ provider: provider.name, result }, 'Provider succeeded');
      return result;
    } catch (error) {
      logger.warn(
        { provider: provider.name, error },
        'Provider failed, trying next'
      );
    }
  }

  throw new Error('All providers failed');
}
```

**Benefits:**
- Resilient to single provider outages
- Can switch providers based on rate limits
- Logs provider reliability metrics

## Pattern 3: Red Flag Override System

Some combinations of metrics are instant HIGH_RISK:

```typescript
function checkRedFlags(metrics: MetricResult[]): RedFlag[] {
  const flags: RedFlag[] = [];

  const mintAuth = metrics.find(m => m.name === 'mint_authority');
  const holderConc = metrics.find(m => m.name === 'holder_concentration');
  const honeypot = metrics.find(m => m.name === 'honeypot');

  // Critical: Mint authority + high concentration = rug pull risk
  if (mintAuth?.data.isActive && holderConc?.data.top10Percent > 80) {
    flags.push({
      severity: 'CRITICAL',
      message: 'Creator can mint unlimited tokens with 80%+ holder concentration',
      override: true, // Force HIGH_RISK regardless of weighted score
    });
  }

  // Critical: Honeypot detected
  if (honeypot?.data.taxDifference > 10) {
    flags.push({
      severity: 'CRITICAL',
      message: `Transfer tax asymmetry detected: ${honeypot.data.taxDifference}%`,
      override: true,
    });
  }

  // Warning: No social verification + new token
  const social = metrics.find(m => m.name === 'social_verification');
  const age = metrics.find(m => m.name === 'token_age');
  if (!social?.data.verified && age?.data.ageHours < 24) {
    flags.push({
      severity: 'WARNING',
      message: 'New token with no social verification',
      override: false,
    });
  }

  return flags;
}

// Apply overrides
let finalLevel: RiskLevel;
const hasOverride = redFlags.some(f => f.override);

if (hasOverride) {
  finalLevel = 'HIGH_RISK';
} else if (finalScore >= 70) {
  finalLevel = 'HIGH_RISK';
} else if (finalScore >= 40) {
  finalLevel = 'MEDIUM_RISK';
} else {
  finalLevel = 'LOW_RISK';
}
```

**Benefits:**
- Catches critical scam patterns immediately
- Can't be fooled by good metrics in other areas
- Clear severity levels for user communication

## Pattern 4: Tier-Based Feature Gating

Premium users get more detailed analysis:

```typescript
async function scanToken(
  address: string,
  tier: 'free' | 'premium'
): Promise<ScanResult> {
  // Basic metrics (all tiers)
  const basicMetrics = await Promise.all([
    liquidityAnalyzer.analyze(address),
    lpLockAnalyzer.analyze(address),
    holderConcentrationAnalyzer.analyze(address),
  ]);

  let detailedMetrics = [];
  let socialAnalysis = undefined;
  let historicalData = undefined;

  // Premium-only features
  if (tier === 'premium') {
    detailedMetrics = await Promise.all([
      creatorHistoryAnalyzer.analyze(address),
      socialVerificationAnalyzer.analyze(address),
      metadataQualityAnalyzer.analyze(address),
    ]);

    socialAnalysis = await fetchSocialMetrics(address);
    historicalData = await fetchHistoricalScans(address);
  }

  return {
    riskScore,
    riskLevel,
    metrics: [...basicMetrics, ...detailedMetrics],
    detailedAnalysis: tier === 'premium' ? socialAnalysis : undefined,
    historicalData: tier === 'premium' ? historicalData : undefined,
    tier,
  };
}
```

**Benefits:**
- Clear value proposition for premium
- Easy to add new premium features
- Can't be bypassed by client

## Pattern 5: Caching Strategy

Reduce API calls and improve performance:

```typescript
import { Redis } from 'ioredis';

const redis = new Redis(process.env.REDIS_URL);

async function getCachedTokenMetrics(address: string) {
  const cacheKey = `token:metrics:${address}`;

  // Try cache first
  const cached = await redis.get(cacheKey);
  if (cached) {
    metrics.increment('cache.hit', 1, { key: 'token_metrics' });
    return JSON.parse(cached);
  }

  // Fetch from blockchain APIs
  metrics.increment('cache.miss', 1, { key: 'token_metrics' });
  const metrics = await fetchTokenMetrics(address);

  // Cache for 5 minutes (token metrics change frequently)
  await redis.setex(cacheKey, 300, JSON.stringify(metrics));

  return metrics;
}
```

**Cache TTLs:**
- Token metadata: 1 hour (rarely changes)
- Liquidity: 5 minutes (changes frequently)
- Holder distribution: 10 minutes
- Creator history: 1 day (static)

## Pattern 6: Metric Confidence Scoring

Not all metrics are equally reliable:

```typescript
interface MetricResult {
  // ... existing fields
  confidence: number; // 0-100 (how reliable is this metric?)
  dataSource: string; // Which API provided the data
}

// Adjust weight based on confidence
const adjustedScore = results.reduce((sum, result) => {
  const confidenceFactor = result.confidence / 100;
  const adjustedWeight = result.weight * confidenceFactor;
  return sum + (result.score * adjustedWeight);
}, 0);

// Show confidence to user
if (avgConfidence < 70) {
  warnings.push(
    'Some metrics have low confidence due to limited data availability'
  );
}
```

**Benefits:**
- Handles missing/unreliable data gracefully
- Transparent to users about data quality
- Can weight more reliable sources higher

## Related Files

- **Agent**: `.claude/agents/risk-scoring-agent.md` - Expert guidance on risk metrics
- **Docs**: `docs/03-TECHNICAL/integrations/telegram-bot-risk-algorithm.md` - Full specification
- **Implementation**: `src/modules/scan/` - Actual code (when implemented)
