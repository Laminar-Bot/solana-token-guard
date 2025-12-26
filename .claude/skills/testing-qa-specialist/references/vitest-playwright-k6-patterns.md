# Testing Patterns - Vitest, Playwright, k6

## Vitest Unit Tests

```typescript
// tests/unit/risk-scorer.test.ts
import { describe, it, expect } from 'vitest';
import { calculateRiskScore } from '@/modules/scan/risk-scorer';

describe('calculateRiskScore', () => {
  it('should return HIGH risk for active mint authority + high holder concentration', () => {
    const result = calculateRiskScore({
      mintAuthority: { isActive: true },
      holderConcentration: 85,
      liquidity: 5000,
      lpLockDays: 0,
    });

    expect(result.level).toBe('HIGH');
    expect(result.score).toBeGreaterThan(70);
  });

  it('should return LOW risk for revoked authorities + good liquidity', () => {
    const result = calculateRiskScore({
      mintAuthority: { isActive: false },
      holderConcentration: 25,
      liquidity: 100_000,
      lpLockDays: 90,
    });

    expect(result.level).toBe('LOW');
    expect(result.score).toBeLessThan(40);
  });
});
```

## Playwright E2E Tests

```typescript
// tests/e2e/scan-flow.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Token Scan Flow', () => {
  test('should scan token and display results', async ({ page }) => {
    await page.goto('https://app.rugmunch.com');

    // Enter token address
    await page.fill('[name="tokenAddress"]', 'So11111111111111111111111111111111111111112');
    await page.click('button[type="submit"]');

    // Wait for results
    await expect(page.locator('text=Risk Score')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('.risk-level')).toContainText(/LOW|MEDIUM|HIGH/);
  });
});
```

## k6 Load Tests

```javascript
// tests/load/scan-api.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 20 }, // Ramp up to 20 users
    { duration: '1m', target: 50 },  // Sustain 50 users
    { duration: '30s', target: 0 },  // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<3000'], // 95% of requests < 3s
    http_req_failed: ['rate<0.01'],    // Error rate < 1%
  },
};

export default function () {
  const res = http.post('https://api.rugmunch.com/v1/scan', JSON.stringify({
    tokenAddress: 'So11111111111111111111111111111111111111112',
  }), {
    headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${__ENV.API_KEY}` },
  });

  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 3s': (r) => r.timings.duration < 3000,
  });

  sleep(1);
}
```
