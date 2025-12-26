# ML, A/B Testing & Analytics Patterns

## XGBoost Scam Detection Model

```python
# train_model.py
import xgboost as xgb
from sklearn.model_selection import train_test_split

# Load labeled dataset (10,000 tokens: 50% scam, 50% legit)
X, y = load_labeled_data()
X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2)

# Train XGBoost classifier
model = xgb.XGBClassifier(
    max_depth=6,
    learning_rate=0.1,
    n_estimators=100,
    objective='binary:logistic',
)

model.fit(X_train, y_train)

# Evaluate
y_pred = model.predict(X_test)
accuracy = (y_pred == y_test).mean()
print(f"Accuracy: {accuracy:.3f}")  # Target: >95%

# Export to ONNX for production
import onnxmltools
onnx_model = onnxmltools.convert_xgboost(model)
onnxmltools.utils.save_model(onnx_model, 'models/scam-detector-v1.onnx')
```

## Feature Engineering (50+ Features)

```typescript
export function engineerFeatures(tokenData: any): TokenFeatures {
  return {
    // On-chain
    tokenAge: (Date.now() - tokenData.createdAt) / (1000 * 60 * 60 * 24),
    holderCount: tokenData.holders.length,
    top10HolderPercent: calculateTop10Percent(tokenData.holders),
    liquidityUSD: tokenData.liquidity,
    volumeToLiquidityRatio: tokenData.volume24h / tokenData.liquidity,

    // Authorities
    mintAuthorityActive: tokenData.mintAuthority ? 1 : 0,
    freezeAuthorityActive: tokenData.freezeAuthority ? 1 : 0,

    // Trading
    txCount24h: tokenData.transactions24h,
    uniqueWallets24h: tokenData.uniqueWallets,
    avgTradeSize: tokenData.volume24h / tokenData.transactions24h,

    // Derived
    holderToVolumeRatio: tokenData.holders.length / (tokenData.volume24h || 1),
    walletConcentration: tokenData.uniqueWallets / tokenData.transactions24h,
  };
}
```

## A/B Testing (GrowthBook)

```typescript
import { GrowthBook } from '@growthbook/growthbook';

const growthbook = new GrowthBook({
  apiHost: process.env.GROWTHBOOK_API_HOST,
  clientKey: process.env.GROWTHBOOK_CLIENT_KEY,
  enableDevMode: process.env.NODE_ENV === 'development',
});

// Test: New risk scoring algorithm
export async function getR iskAlgorithm(userId: string): 'v1' | 'v2' {
  const variant = growthbook.getFeatureValue('risk-algorithm-test', 'v1');
  return variant;
}

// Track conversion (user upgrades after scan)
growthbook.trackExperiment('risk-algorithm-test', userId, {
  event: 'upgrade',
  revenue: 19.99,
});
```

## Business Analytics (DAU, MRR, Churn)

```typescript
// Daily Active Users
export async function calculateDAU(date: Date): Promise<number> {
  const startOfDay = new Date(date).setHours(0, 0, 0, 0);
  const endOfDay = new Date(date).setHours(23, 59, 59, 999);

  const activeUsers = await userRepository.count({
    where: {
      lastActiveAt: { gte: new Date(startOfDay), lte: new Date(endOfDay) },
    },
  });

  return activeUsers;
}

// Monthly Recurring Revenue
export async function calculateMRR(): Promise<number> {
  const activeSubs = await subscriptionRepository.findMany({
    where: { status: 'active' },
  });

  return activeSubs.reduce((sum, sub) => sum + sub.monthlyPrice, 0);
}

// Churn Rate
export async function calculateChurnRate(month: number, year: number): Promise<number> {
  const startOfMonth = new Date(year, month - 1, 1);
  const endOfMonth = new Date(year, month, 0);

  const totalStart = await subscriptionRepository.count({
    where: { createdAt: { lt: startOfMonth }, status: 'active' },
  });

  const churned = await subscriptionRepository.count({
    where: {
      canceledAt: { gte: startOfMonth, lte: endOfMonth },
    },
  });

  return (churned / totalStart) * 100;
}
```

## Metabase Dashboard Queries

```sql
-- Cohort Retention Analysis
SELECT
  DATE_TRUNC('month', u.created_at) AS cohort_month,
  DATE_TRUNC('month', s.created_at) AS activity_month,
  COUNT(DISTINCT u.id) AS active_users
FROM users u
LEFT JOIN scans s ON u.id = s.user_id
GROUP BY cohort_month, activity_month
ORDER BY cohort_month, activity_month;
```
