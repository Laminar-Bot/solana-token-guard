# Security Audit Checklist & OWASP Top 10

## API Security Checklist

```typescript
// 1. Input Validation (Prevent Injection)
import { z } from 'zod';

const ScanRequestSchema = z.object({
  tokenAddress: z.string().regex(/^[1-9A-HJ-NP-Za-km-z]{32,44}$/),
  tier: z.enum(['free', 'premium', 'enterprise']),
});

// 2. Authentication (JWT with short expiry)
export function generateJWT(userId: string): string {
  return jwt.sign({ userId }, process.env.JWT_SECRET!, {
    expiresIn: '15m',
    issuer: 'cryptorugmunch',
  });
}

// 3. Rate Limiting
import rateLimit from '@fastify/rate-limit';

app.register(rateLimit, {
  max: 100,
  timeWindow: '15 minutes',
  redis, // Distributed rate limiting
});

// 4. CORS (Restrict origins)
app.register(cors, {
  origin: ['https://rugmunch.com', 'https://app.rugmunch.com'],
  credentials: true,
});

// 5. Helmet (Security headers)
app.register(helmet, {
  contentSecurityPolicy: {
    directives: {
      defaultSrc: ["'self'"],
      scriptSrc: ["'self'", "'unsafe-inline'"],
    },
  },
});

// 6. Secrets Management (Never commit secrets!)
// Use AWS Secrets Manager or environment variables ONLY
const secrets = await getSecretValue('rugmunch/production');
```

## GDPR Compliance Patterns

```typescript
// Right to Access
export async function exportUserData(userId: string) {
  const [user, scans, subscriptions] = await Promise.all([
    userRepository.findById(userId),
    scanRepository.findByUser(userId),
    subscriptionRepository.findByUser(userId),
  ]);

  return {
    user: { id: user.id, email: user.email, createdAt: user.createdAt },
    scans: scans.map(s => ({ tokenAddress: s.tokenAddress, riskScore: s.riskScore, scannedAt: s.createdAt })),
    subscriptions: subscriptions.map(s => ({ tier: s.tier, startedAt: s.createdAt })),
  };
}

// Right to Deletion
export async function deleteUserData(userId: string) {
  await Promise.all([
    userRepository.delete(userId),
    scanRepository.deleteByUser(userId),
    subscriptionRepository.cancelByUser(userId),
  ]);

  logger.info({ userId }, 'GDPR deletion completed');
}
```

## SQL Injection Prevention

```typescript
// ✅ SAFE: Parameterized queries (Prisma)
const scans = await prisma.scan.findMany({
  where: { userId, tokenAddress },
});

// ❌ UNSAFE: String concatenation
const query = `SELECT * FROM scans WHERE userId = '${userId}'`; // NEVER DO THIS!
```

## XSS Prevention

```typescript
// Escape user input before rendering
import DOMPurify from 'isomorphic-dompurify';

export function sanitizeHTML(dirty: string): string {
  return DOMPurify.sanitize(dirty, {
    ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'a'],
    ALLOWED_ATTR: ['href'],
  });
}
```
