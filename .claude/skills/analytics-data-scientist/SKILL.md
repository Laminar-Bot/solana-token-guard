---
name: analytics-data-scientist
description: "Expert data scientist specializing in analytics, machine learning for scam detection, anomaly detection, user behavior analysis, A/B testing, and data-driven product decisions for CryptoRugMunch. Deep knowledge of Python ML stack, PostgreSQL analytics, and business intelligence."
---

# Analytics & Data Science Specialist

**Role**: Expert data scientist and analytics engineer for CryptoRugMunch, providing insights from user behavior, improving scam detection through machine learning, and enabling data-driven product decisions.

**Context**: CryptoRugMunch generates **massive amounts of data**:
- **Scan data**: Millions of token scans with risk metrics
- **User behavior**: Telegram interactions, dashboard usage, conversion funnels
- **Blockchain data**: Token metadata, liquidity pools, holder distributions
- **Business metrics**: MRR, churn, CAC, LTV

**Goal**: Extract actionable insights to:
1. Improve scam detection accuracy (ML models)
2. Optimize user retention and conversion
3. Identify new scam patterns
4. Inform product roadmap decisions

---

## Core Philosophy

1. **Data-Driven Decisions**: Every product change backed by metrics and experiments
2. **Predictive Analytics**: Move from reactive (manual risk scoring) to predictive (ML models)
3. **Anomaly Detection**: Automatically flag new scam patterns before they spread
4. **Experimentation Culture**: A/B test features, pricing, messaging
5. **Ethical AI**: Ensure ML models are fair, explainable, and don't discriminate

---

## 1. Analytics Stack

### 1.1 Tech Stack

| Component | Tool | Purpose |
|-----------|------|---------|
| **Data Warehouse** | PostgreSQL (OLAP) | Historical scan data, aggregations |
| **ETL Pipelines** | dbt (Data Build Tool) | Transform raw data into analytics tables |
| **BI Dashboards** | Metabase / Looker | Self-serve analytics for non-technical team |
| **ML Platform** | Python (scikit-learn, XGBoost) | Train scam detection models |
| **Feature Store** | Feast (optional) | Store ML features for reuse |
| **Experiment Tracking** | MLflow | Track ML experiments, model versions |
| **A/B Testing** | GrowthBook / Statsig | Feature flags, A/B tests |
| **Observability** | DataDog | Monitor model performance in production |

### 1.2 Data Warehouse Schema

**Goal**: Separate OLTP (transactions) from OLAP (analytics) for performance.

```sql
-- Analytics schema (read-optimized)
CREATE SCHEMA analytics;

-- Fact table: Scans
CREATE TABLE analytics.fact_scans (
  scan_id VARCHAR PRIMARY KEY,
  user_id VARCHAR NOT NULL,
  token_address VARCHAR NOT NULL,
  chain VARCHAR NOT NULL,
  risk_score INT NOT NULL,
  category VARCHAR NOT NULL,
  scanned_at TIMESTAMP NOT NULL,
  completed_at TIMESTAMP,
  duration_ms INT,

  -- Denormalized dimensions (for query performance)
  user_tier VARCHAR,
  user_level INT,
  user_xp INT,

  -- Risk breakdown (JSON)
  breakdown JSONB,
  flags TEXT[]
);

-- Dimension table: Tokens
CREATE TABLE analytics.dim_tokens (
  token_address VARCHAR PRIMARY KEY,
  chain VARCHAR NOT NULL,
  name VARCHAR,
  symbol VARCHAR,
  first_scanned_at TIMESTAMP,
  last_scanned_at TIMESTAMP,
  total_scans INT,
  avg_risk_score FLOAT,
  is_known_scam BOOLEAN DEFAULT FALSE
);

-- Dimension table: Users
CREATE TABLE analytics.dim_users (
  user_id VARCHAR PRIMARY KEY,
  telegram_id VARCHAR,
  tier VARCHAR,
  level INT,
  xp INT,
  created_at TIMESTAMP,
  last_active_at TIMESTAMP,
  total_scans INT,
  scams_detected INT,
  subscription_mrr FLOAT
);

-- Aggregated table: Daily scan volume
CREATE TABLE analytics.agg_daily_scans (
  date DATE PRIMARY KEY,
  total_scans INT,
  scans_safe INT,
  scans_caution INT,
  scans_high_risk INT,
  scans_scam INT,
  unique_users INT,
  new_users INT,
  avg_scan_duration_ms FLOAT
);

-- Indexes for common queries
CREATE INDEX idx_fact_scans_scanned_at ON analytics.fact_scans (scanned_at DESC);
CREATE INDEX idx_fact_scans_user_id ON analytics.fact_scans (user_id);
CREATE INDEX idx_fact_scans_category ON analytics.fact_scans (category);
CREATE INDEX idx_dim_tokens_avg_risk ON analytics.dim_tokens (avg_risk_score ASC);
```

---

## 2. Business Analytics

### 2.1 Key Metrics (KPIs)

#### Growth Metrics

```sql
-- Daily Active Users (DAU)
SELECT
  DATE(scanned_at) AS date,
  COUNT(DISTINCT user_id) AS dau
FROM analytics.fact_scans
WHERE scanned_at >= NOW() - INTERVAL '30 days'
GROUP BY DATE(scanned_at)
ORDER BY date DESC;

-- Monthly Active Users (MAU)
SELECT
  DATE_TRUNC('month', scanned_at) AS month,
  COUNT(DISTINCT user_id) AS mau
FROM analytics.fact_scans
GROUP BY DATE_TRUNC('month', scanned_at)
ORDER BY month DESC;

-- DAU/MAU Ratio (stickiness)
WITH dau AS (
  SELECT DATE(scanned_at) AS date, COUNT(DISTINCT user_id) AS users
  FROM analytics.fact_scans
  WHERE scanned_at >= NOW() - INTERVAL '30 days'
  GROUP BY DATE(scanned_at)
),
mau AS (
  SELECT
    DATE_TRUNC('month', scanned_at) AS month,
    COUNT(DISTINCT user_id) AS users
  FROM analytics.fact_scans
  GROUP BY DATE_TRUNC('month', scanned_at)
)
SELECT
  dau.date,
  dau.users AS dau,
  mau.users AS mau,
  ROUND((dau.users::FLOAT / mau.users) * 100, 2) AS stickiness_pct
FROM dau
JOIN mau ON DATE_TRUNC('month', dau.date) = mau.month;
```

#### Revenue Metrics

```sql
-- Monthly Recurring Revenue (MRR)
SELECT
  DATE_TRUNC('month', created_at) AS month,
  SUM(CASE WHEN tier = 'PREMIUM' THEN 9.99 ELSE 0 END) AS mrr
FROM analytics.dim_users
WHERE tier = 'PREMIUM'
GROUP BY DATE_TRUNC('month', created_at)
ORDER BY month DESC;

-- Churn Rate (monthly)
WITH subscribers AS (
  SELECT
    DATE_TRUNC('month', created_at) AS month,
    COUNT(*) AS new_subscribers
  FROM "Subscription"
  WHERE status = 'ACTIVE'
  GROUP BY DATE_TRUNC('month', created_at)
),
churned AS (
  SELECT
    DATE_TRUNC('month', cancelled_at) AS month,
    COUNT(*) AS churned_subscribers
  FROM "Subscription"
  WHERE status = 'CANCELLED'
  GROUP BY DATE_TRUNC('month', cancelled_at)
)
SELECT
  subscribers.month,
  subscribers.new_subscribers,
  COALESCE(churned.churned_subscribers, 0) AS churned,
  ROUND((COALESCE(churned.churned_subscribers, 0)::FLOAT / subscribers.new_subscribers) * 100, 2) AS churn_rate_pct
FROM subscribers
LEFT JOIN churned ON subscribers.month = churned.month
ORDER BY subscribers.month DESC;

-- Customer Lifetime Value (LTV)
WITH user_lifetimes AS (
  SELECT
    user_id,
    EXTRACT(EPOCH FROM (cancelled_at - created_at)) / (30.44 * 86400) AS lifetime_months,
    9.99 AS monthly_revenue
  FROM "Subscription"
  WHERE status = 'CANCELLED' OR status = 'ACTIVE'
)
SELECT
  AVG(lifetime_months * monthly_revenue) AS avg_ltv,
  PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY lifetime_months * monthly_revenue) AS median_ltv
FROM user_lifetimes;
```

#### Product Metrics

```sql
-- Top Scammed Tokens (most detected scams)
SELECT
  token_address,
  COUNT(*) AS times_scanned,
  AVG(risk_score) AS avg_risk_score,
  COUNT(CASE WHEN category = 'LIKELY_SCAM' THEN 1 END) AS scam_votes
FROM analytics.fact_scans
GROUP BY token_address
HAVING AVG(risk_score) < 30
ORDER BY scam_votes DESC
LIMIT 20;

-- Scan Volume by Chain
SELECT
  chain,
  COUNT(*) AS total_scans,
  AVG(risk_score) AS avg_risk_score
FROM analytics.fact_scans
WHERE scanned_at >= NOW() - INTERVAL '30 days'
GROUP BY chain
ORDER BY total_scans DESC;

-- Risk Distribution (SAFE vs SCAM)
SELECT
  category,
  COUNT(*) AS scans,
  ROUND((COUNT(*)::FLOAT / SUM(COUNT(*)) OVER ()) * 100, 2) AS percentage
FROM analytics.fact_scans
GROUP BY category
ORDER BY scans DESC;
```

### 2.2 User Behavior Analysis

#### Cohort Retention Analysis

```sql
-- Cohort retention (weekly)
WITH user_cohorts AS (
  SELECT
    user_id,
    DATE_TRUNC('week', MIN(scanned_at)) AS cohort_week
  FROM analytics.fact_scans
  GROUP BY user_id
),
user_activity AS (
  SELECT
    user_id,
    DATE_TRUNC('week', scanned_at) AS activity_week
  FROM analytics.fact_scans
),
cohort_activity AS (
  SELECT
    c.cohort_week,
    a.activity_week,
    COUNT(DISTINCT a.user_id) AS active_users
  FROM user_cohorts c
  JOIN user_activity a ON c.user_id = a.user_id
  GROUP BY c.cohort_week, a.activity_week
),
cohort_sizes AS (
  SELECT
    cohort_week,
    COUNT(*) AS cohort_size
  FROM user_cohorts
  GROUP BY cohort_week
)
SELECT
  ca.cohort_week,
  ca.activity_week,
  EXTRACT(WEEK FROM ca.activity_week - ca.cohort_week) AS week_number,
  ca.active_users,
  cs.cohort_size,
  ROUND((ca.active_users::FLOAT / cs.cohort_size) * 100, 2) AS retention_pct
FROM cohort_activity ca
JOIN cohort_sizes cs ON ca.cohort_week = cs.cohort_week
ORDER BY ca.cohort_week, ca.activity_week;
```

#### Conversion Funnel

```sql
-- Free to Premium conversion funnel
WITH funnel AS (
  SELECT
    user_id,
    COUNT(*) AS total_scans,
    MAX(CASE WHEN scans_today >= 10 THEN 1 ELSE 0 END) AS hit_free_limit,
    MAX(CASE WHEN tier = 'PREMIUM' THEN 1 ELSE 0 END) AS converted_to_premium
  FROM analytics.fact_scans
  JOIN analytics.dim_users USING (user_id)
  GROUP BY user_id
)
SELECT
  'Total Users' AS stage,
  COUNT(*) AS users,
  100.0 AS conversion_pct
FROM funnel
UNION ALL
SELECT
  'Hit Free Limit (10 scans/day)' AS stage,
  SUM(hit_free_limit) AS users,
  ROUND((SUM(hit_free_limit)::FLOAT / COUNT(*)) * 100, 2) AS conversion_pct
FROM funnel
UNION ALL
SELECT
  'Converted to Premium' AS stage,
  SUM(converted_to_premium) AS users,
  ROUND((SUM(converted_to_premium)::FLOAT / SUM(hit_free_limit)) * 100, 2) AS conversion_pct
FROM funnel;
```

---

## 3. Machine Learning for Scam Detection

### 3.1 Current vs ML Approach

| Aspect | Current (Rule-Based) | ML Approach |
|--------|---------------------|-------------|
| **Algorithm** | Weighted scoring (12 metrics) | Gradient Boosting (XGBoost, LightGBM) |
| **Accuracy** | ~85% (manual tuning) | ~95%+ (trained on data) |
| **Adaptability** | Manual updates | Auto-adapts to new scams |
| **Explainability** | Fully transparent | SHAP values for interpretability |
| **Performance** | Fast (<1s) | Slightly slower (~2s) |
| **Maintenance** | High (manual threshold tuning) | Low (retraining pipeline) |

### 3.2 ML Feature Engineering

**Goal**: Extract 50+ features from raw blockchain data for ML model.

```python
# ml/features.py
import pandas as pd
import numpy as np
from datetime import datetime, timedelta

def extract_features(token_data: dict) -> dict:
    """
    Extract ML features from raw token data.

    Returns 50+ features for scam detection model.
    """
    features = {}

    # === Liquidity Features ===
    features['liquidity_usd'] = token_data['liquidity']
    features['liquidity_log'] = np.log1p(token_data['liquidity'])  # Log-scale
    features['liquidity_24h_change_pct'] = token_data.get('liquidity_24h_change', 0)

    # === LP Lock Features ===
    features['lp_locked'] = 1 if token_data['lpLocked'] else 0
    features['lp_locked_percentage'] = token_data['lpLockedPercentage']
    features['lp_unlock_days'] = (token_data.get('unlockDate', datetime.now()) - datetime.now()).days
    features['lp_unlock_days_log'] = np.log1p(max(features['lp_unlock_days'], 0))

    # === Holder Distribution Features ===
    features['holder_concentration_top10'] = token_data['holderConcentration']
    features['holder_concentration_top50'] = token_data.get('top50Percentage', 0)
    features['holder_count'] = token_data.get('holderCount', 0)
    features['holder_count_log'] = np.log1p(features['holder_count'])
    features['holder_gini_coefficient'] = calculate_gini(token_data.get('holders', []))

    # === Authority Features ===
    features['mint_authority_revoked'] = 1 if token_data['mintAuthorityRevoked'] else 0
    features['freeze_authority_revoked'] = 1 if token_data['freezeAuthorityRevoked'] else 0
    features['ownership_renounced'] = 1 if token_data['ownershipRenounced'] else 0

    # === Honeypot Features ===
    features['is_honeypot'] = 1 if token_data['isHoneypot'] else 0
    features['buy_tax_pct'] = token_data.get('buyTax', 0)
    features['sell_tax_pct'] = token_data.get('sellTax', 0)
    features['tax_difference'] = features['sell_tax_pct'] - features['buy_tax_pct']

    # === Token Age Features ===
    features['token_age_days'] = token_data['tokenAge']
    features['token_age_log'] = np.log1p(features['token_age_days'])
    features['is_very_new'] = 1 if features['token_age_days'] < 7 else 0
    features['is_established'] = 1 if features['token_age_days'] > 90 else 0

    # === Social Media Features ===
    features['has_twitter'] = 1 if token_data.get('twitterUrl') else 0
    features['has_telegram'] = 1 if token_data.get('telegramUrl') else 0
    features['has_website'] = 1 if token_data.get('websiteUrl') else 0
    features['social_media_score'] = features['has_twitter'] + features['has_telegram'] + features['has_website']

    # === Audit Features ===
    features['has_audit'] = 1 if token_data['hasAudit'] else 0
    features['audit_score'] = token_data.get('auditScore', 0)

    # === Price Features ===
    features['price_usd'] = token_data.get('priceUSD', 0)
    features['price_24h_change_pct'] = token_data.get('price24hChange', 0)
    features['market_cap_usd'] = token_data.get('marketCap', 0)
    features['market_cap_log'] = np.log1p(features['market_cap_usd'])
    features['volume_24h_usd'] = token_data.get('volume24h', 0)
    features['volume_24h_log'] = np.log1p(features['volume_24h_usd'])

    # === Trading Activity Features ===
    features['trade_count_24h'] = token_data.get('tradeCount24h', 0)
    features['trade_count_log'] = np.log1p(features['trade_count_24h'])
    features['buy_sell_ratio'] = token_data.get('buySellRatio', 1.0)

    # === Derived Features (Ratios) ===
    features['liquidity_to_mcap_ratio'] = (
        features['liquidity_usd'] / features['market_cap_usd']
        if features['market_cap_usd'] > 0 else 0
    )
    features['volume_to_liquidity_ratio'] = (
        features['volume_24h_usd'] / features['liquidity_usd']
        if features['liquidity_usd'] > 0 else 0
    )

    # === Categorical Features (One-Hot Encoded) ===
    features['chain_solana'] = 1 if token_data['chain'] == 'SOLANA' else 0
    features['chain_ethereum'] = 1 if token_data['chain'] == 'ETHEREUM' else 0
    features['chain_bsc'] = 1 if token_data['chain'] == 'BSC' else 0

    return features

def calculate_gini(holders: list) -> float:
    """Calculate Gini coefficient for holder distribution (inequality measure)."""
    if not holders:
        return 0.0

    balances = np.array([h['balance'] for h in holders])
    balances = np.sort(balances)
    n = len(balances)
    index = np.arange(1, n + 1)
    return (2 * np.sum(index * balances)) / (n * np.sum(balances)) - (n + 1) / n
```

### 3.3 Training XGBoost Model

```python
# ml/train.py
import xgboost as xgb
import pandas as pd
import numpy as np
from sklearn.model_selection import train_test_split
from sklearn.metrics import accuracy_score, precision_recall_fscore_support, roc_auc_score
import mlflow
import mlflow.xgboost

def train_scam_detection_model():
    """
    Train XGBoost model for scam detection.

    Binary classification: 0 = SAFE (score >= 60), 1 = SCAM (score < 60)
    """
    # Load historical scan data
    df = pd.read_sql("""
        SELECT
            scan_id,
            breakdown,
            category,
            risk_score
        FROM analytics.fact_scans
        WHERE scanned_at >= NOW() - INTERVAL '90 days'
    """, con=db_connection)

    # Extract features
    features_list = []
    labels = []

    for _, row in df.iterrows():
        features = extract_features(row['breakdown'])
        features_list.append(features)

        # Label: 1 = SCAM, 0 = SAFE
        label = 1 if row['risk_score'] < 60 else 0
        labels.append(label)

    X = pd.DataFrame(features_list)
    y = np.array(labels)

    # Train/test split
    X_train, X_test, y_train, y_test = train_test_split(
        X, y, test_size=0.2, random_state=42, stratify=y
    )

    # Train XGBoost model
    model = xgb.XGBClassifier(
        n_estimators=500,
        max_depth=6,
        learning_rate=0.05,
        subsample=0.8,
        colsample_bytree=0.8,
        objective='binary:logistic',
        eval_metric='auc',
        use_label_encoder=False,
        random_state=42
    )

    # Start MLflow run
    with mlflow.start_run():
        # Train
        model.fit(
            X_train, y_train,
            eval_set=[(X_test, y_test)],
            early_stopping_rounds=50,
            verbose=False
        )

        # Evaluate
        y_pred = model.predict(X_test)
        y_pred_proba = model.predict_proba(X_test)[:, 1]

        accuracy = accuracy_score(y_test, y_pred)
        precision, recall, f1, _ = precision_recall_fscore_support(y_test, y_pred, average='binary')
        roc_auc = roc_auc_score(y_test, y_pred_proba)

        # Log metrics
        mlflow.log_metric('accuracy', accuracy)
        mlflow.log_metric('precision', precision)
        mlflow.log_metric('recall', recall)
        mlflow.log_metric('f1', f1)
        mlflow.log_metric('roc_auc', roc_auc)

        # Log model
        mlflow.xgboost.log_model(model, 'scam-detection-model')

        print(f"Model Performance:")
        print(f"  Accuracy: {accuracy:.4f}")
        print(f"  Precision: {precision:.4f}")
        print(f"  Recall: {recall:.4f}")
        print(f"  F1 Score: {f1:.4f}")
        print(f"  ROC AUC: {roc_auc:.4f}")

        # Feature importance
        feature_importance = pd.DataFrame({
            'feature': X.columns,
            'importance': model.feature_importances_
        }).sort_values('importance', ascending=False)

        print("\nTop 10 Most Important Features:")
        print(feature_importance.head(10))

        return model
```

### 3.4 Model Deployment & Serving

```python
# ml/serve.py
import mlflow.xgboost
import xgboost as xgb

# Load trained model
model_uri = "runs:/<run-id>/scam-detection-model"
model = mlflow.xgboost.load_model(model_uri)

def predict_scam_probability(token_data: dict) -> float:
    """
    Predict scam probability for a token.

    Returns:
        float: Probability of being a scam (0.0 = SAFE, 1.0 = SCAM)
    """
    features = extract_features(token_data)
    X = pd.DataFrame([features])

    scam_probability = model.predict_proba(X)[0][1]

    return scam_probability

def get_ml_risk_score(token_data: dict) -> int:
    """
    Get ML-based risk score (0-100).

    Converts scam probability to CryptoRugMunch's 0-100 scale.
    """
    scam_probability = predict_scam_probability(token_data)

    # Invert: 0% scam = 100 risk score, 100% scam = 0 risk score
    risk_score = int((1 - scam_probability) * 100)

    return risk_score
```

### 3.5 Model Explainability (SHAP)

```python
# ml/explainability.py
import shap
import matplotlib.pyplot as plt

def explain_prediction(model, token_data: dict):
    """
    Explain ML model prediction using SHAP values.

    Shows which features contributed most to the scam classification.
    """
    features = extract_features(token_data)
    X = pd.DataFrame([features])

    # Create SHAP explainer
    explainer = shap.TreeExplainer(model)
    shap_values = explainer.shap_values(X)

    # Plot feature importance for this prediction
    shap.force_plot(
        explainer.expected_value,
        shap_values[0],
        X.iloc[0],
        matplotlib=True
    )
    plt.savefig('shap_explanation.png')

    # Return top contributing features
    feature_contributions = pd.DataFrame({
        'feature': X.columns,
        'shap_value': shap_values[0]
    }).sort_values('shap_value', key=abs, ascending=False)

    return feature_contributions.head(10)
```

---

## 4. Anomaly Detection for New Scams

### 4.1 Isolation Forest for Outlier Detection

**Goal**: Detect tokens with unusual patterns (potential new scam types).

```python
# ml/anomaly_detection.py
from sklearn.ensemble import IsolationForest
import pandas as pd

def train_anomaly_detector():
    """
    Train Isolation Forest to detect unusual tokens.

    Flags tokens with anomalous feature combinations.
    """
    # Load historical SAFE tokens (to learn "normal" patterns)
    df = pd.read_sql("""
        SELECT breakdown
        FROM analytics.fact_scans
        WHERE category = 'SAFE'
        AND scanned_at >= NOW() - INTERVAL '90 days'
    """, con=db_connection)

    # Extract features
    features_list = [extract_features(row['breakdown']) for _, row in df.iterrows()]
    X = pd.DataFrame(features_list)

    # Train Isolation Forest
    model = IsolationForest(
        n_estimators=200,
        contamination=0.05,  # Expect 5% of data to be anomalies
        random_state=42
    )
    model.fit(X)

    return model

def detect_anomaly(model, token_data: dict) -> dict:
    """
    Detect if token is anomalous (potential new scam pattern).

    Returns:
        dict: {
            'is_anomaly': bool,
            'anomaly_score': float,
            'explanation': str
        }
    """
    features = extract_features(token_data)
    X = pd.DataFrame([features])

    # Predict (-1 = anomaly, 1 = normal)
    prediction = model.predict(X)[0]
    anomaly_score = model.score_samples(X)[0]

    is_anomaly = prediction == -1

    explanation = ""
    if is_anomaly:
        # Find which features are most unusual
        feature_means = X.mean(axis=0)
        feature_stds = X.std(axis=0)
        z_scores = ((X.iloc[0] - feature_means) / feature_stds).abs().sort_values(ascending=False)

        top_unusual = z_scores.head(3)
        explanation = f"Unusual features: {', '.join(top_unusual.index.tolist())}"

    return {
        'is_anomaly': is_anomaly,
        'anomaly_score': float(anomaly_score),
        'explanation': explanation
    }
```

---

## 5. A/B Testing & Experimentation

### 5.1 A/B Testing Framework

**Tool**: GrowthBook (open-source A/B testing platform)

```typescript
// src/lib/experiments.ts
import { GrowthBook } from '@growthbook/growthbook'

const gb = new GrowthBook({
  apiHost: process.env.GROWTHBOOK_API_HOST!,
  clientKey: process.env.GROWTHBOOK_CLIENT_KEY!,
  enableDevMode: process.env.NODE_ENV === 'development',
  trackingCallback: (experiment, result) => {
    // Log experiment exposure to DataDog
    metrics.increment('experiment.exposure', 1, {
      experiment_key: experiment.key,
      variant: result.value,
    })
  },
})

// Load features from GrowthBook
await gb.loadFeatures()

export { gb }
```

```typescript
// Example: A/B test for pricing page CTA
import { gb } from '@/lib/experiments'

export function PricingPage() {
  // Get variant (control vs treatment)
  const ctaText = gb.feature('pricing-cta-text').value || 'Upgrade to Premium'

  return (
    <div>
      <h1>Pricing</h1>
      <button onClick={handleUpgrade}>
        {ctaText}
      </button>
    </div>
  )
}
```

### 5.2 Experiment Analysis

**Goal**: Determine if experiment had statistically significant impact.

```python
# analytics/experiment_analysis.py
from scipy import stats
import pandas as pd

def analyze_ab_test(experiment_key: str):
    """
    Analyze A/B test results for statistical significance.

    Uses Welch's t-test for continuous metrics (e.g., revenue)
    and Chi-squared test for categorical metrics (e.g., conversion).
    """
    # Load experiment data
    df = pd.read_sql(f"""
        SELECT
            variant,
            converted,
            revenue
        FROM experiment_events
        WHERE experiment_key = '{experiment_key}'
    """, con=db_connection)

    control = df[df['variant'] == 'control']
    treatment = df[df['variant'] == 'treatment']

    # === Conversion Rate Test (Chi-squared) ===
    control_conversions = control['converted'].sum()
    control_total = len(control)
    treatment_conversions = treatment['converted'].sum()
    treatment_total = len(treatment)

    contingency_table = [
        [control_conversions, control_total - control_conversions],
        [treatment_conversions, treatment_total - treatment_conversions]
    ]

    chi2, p_value_conversion, _, _ = stats.chi2_contingency(contingency_table)

    control_cvr = control_conversions / control_total
    treatment_cvr = treatment_conversions / treatment_total
    cvr_lift = ((treatment_cvr - control_cvr) / control_cvr) * 100

    # === Revenue Test (Welch's t-test) ===
    t_stat, p_value_revenue = stats.ttest_ind(
        control['revenue'],
        treatment['revenue'],
        equal_var=False  # Welch's t-test
    )

    control_arpu = control['revenue'].mean()
    treatment_arpu = treatment['revenue'].mean()
    arpu_lift = ((treatment_arpu - control_arpu) / control_arpu) * 100

    # === Results ===
    print(f"Experiment: {experiment_key}")
    print(f"\nConversion Rate:")
    print(f"  Control: {control_cvr:.2%} ({control_conversions}/{control_total})")
    print(f"  Treatment: {treatment_cvr:.2%} ({treatment_conversions}/{treatment_total})")
    print(f"  Lift: {cvr_lift:+.2f}%")
    print(f"  P-value: {p_value_conversion:.4f}")
    print(f"  Significant: {'✅ YES' if p_value_conversion < 0.05 else '❌ NO'}")

    print(f"\nRevenue (ARPU):")
    print(f"  Control: ${control_arpu:.2f}")
    print(f"  Treatment: ${treatment_arpu:.2f}")
    print(f"  Lift: {arpu_lift:+.2f}%")
    print(f"  P-value: {p_value_revenue:.4f}")
    print(f"  Significant: {'✅ YES' if p_value_revenue < 0.05 else '❌ NO'}")

    return {
        'conversion_rate': {
            'control': control_cvr,
            'treatment': treatment_cvr,
            'lift_pct': cvr_lift,
            'p_value': p_value_conversion,
            'significant': p_value_conversion < 0.05
        },
        'revenue': {
            'control_arpu': control_arpu,
            'treatment_arpu': treatment_arpu,
            'lift_pct': arpu_lift,
            'p_value': p_value_revenue,
            'significant': p_value_revenue < 0.05
        }
    }
```

---

## 6. Real-Time Analytics

### 6.1 Stream Processing with Kafka (Optional)

**Goal**: Real-time dashboards, live metrics, instant alerts.

```python
# analytics/stream_processor.py
from kafka import KafkaConsumer
import json

consumer = KafkaConsumer(
    'scan-completed',
    bootstrap_servers=os.getenv('KAFKA_BROKERS'),
    value_deserializer=lambda m: json.loads(m.decode('utf-8'))
)

def process_scan_stream():
    """
    Process scan events in real-time.

    - Update real-time dashboard metrics
    - Trigger alerts for anomalies
    - Update recommendation engine
    """
    for message in consumer:
        scan = message.value

        # Update real-time metrics in Redis
        redis.incr(f"scans:today:{scan['chain']}")
        redis.incr(f"scans:category:{scan['category']}")

        # Check for anomalies
        if scan['category'] == 'LIKELY_SCAM':
            # Alert if unusual scam pattern
            anomaly_result = detect_anomaly(anomaly_model, scan)
            if anomaly_result['is_anomaly']:
                send_alert(f"🚨 New scam pattern detected: {scan['tokenAddress']}")

        # Update token reputation
        update_token_reputation(scan['tokenAddress'], scan['riskScore'])
```

---

## 7. Predictive Analytics

### 7.1 Churn Prediction

**Goal**: Predict which premium users are likely to cancel.

```python
# ml/churn_prediction.py
from sklearn.ensemble import RandomForestClassifier
import pandas as pd

def train_churn_model():
    """
    Train model to predict premium subscriber churn.

    Features:
    - Days since subscription
    - Scan frequency (last 7/30 days)
    - Engagement (Telegram bot usage)
    - Value received (scams detected)
    """
    # Load subscriber data
    df = pd.read_sql("""
        SELECT
            s.user_id,
            s.created_at AS subscription_start,
            s.cancelled_at,
            CASE WHEN s.cancelled_at IS NOT NULL THEN 1 ELSE 0 END AS churned,
            COUNT(sc.id) AS scans_total,
            COUNT(sc.id) FILTER (WHERE sc.scanned_at >= NOW() - INTERVAL '7 days') AS scans_7d,
            COUNT(sc.id) FILTER (WHERE sc.scanned_at >= NOW() - INTERVAL '30 days') AS scans_30d,
            COUNT(sc.id) FILTER (WHERE sc.category = 'LIKELY_SCAM') AS scams_detected,
            u.xp,
            u.level
        FROM "Subscription" s
        LEFT JOIN "Scan" sc ON s.user_id = sc.user_id
        LEFT JOIN "User" u ON s.user_id = u.id
        WHERE s.tier = 'PREMIUM'
        GROUP BY s.user_id, s.created_at, s.cancelled_at, u.xp, u.level
    """, con=db_connection)

    # Feature engineering
    df['subscription_days'] = (pd.Timestamp.now() - df['subscription_start']).dt.days
    df['scans_per_week'] = df['scans_7d']
    df['scans_per_month'] = df['scans_30d']
    df['scams_detected_per_month'] = df['scams_detected']

    # Select features
    feature_cols = [
        'subscription_days', 'scans_total', 'scans_per_week', 'scans_per_month',
        'scams_detected_per_month', 'xp', 'level'
    ]
    X = df[feature_cols]
    y = df['churned']

    # Train model
    model = RandomForestClassifier(n_estimators=200, random_state=42)
    model.fit(X, y)

    return model

def predict_churn_risk(user_id: str) -> float:
    """
    Predict churn risk for a premium user.

    Returns:
        float: Churn probability (0.0 = unlikely to churn, 1.0 = likely to churn)
    """
    # Load user data
    user_data = get_user_features(user_id)
    X = pd.DataFrame([user_data])

    churn_probability = churn_model.predict_proba(X)[0][1]

    # Send alert if high churn risk
    if churn_probability > 0.7:
        send_retention_campaign(user_id, churn_probability)

    return churn_probability
```

---

## 8. BI Dashboards

### 8.1 Metabase Setup

**Why Metabase?**: Free, open-source, self-hosted BI tool.

```bash
# Run Metabase with Docker
docker run -d -p 3000:3000 \
  -e MB_DB_TYPE=postgres \
  -e MB_DB_DBNAME=metabase \
  -e MB_DB_PORT=5432 \
  -e MB_DB_USER=metabase \
  -e MB_DB_PASS=password \
  -e MB_DB_HOST=postgres \
  --name metabase \
  metabase/metabase
```

**Pre-built Dashboards**:

1. **Executive Dashboard**:
   - DAU/MAU/WAU
   - MRR, churn rate
   - Top scammed tokens
   - Risk distribution

2. **Product Metrics**:
   - Scan volume by chain
   - Average scan duration
   - Free vs premium usage
   - Conversion funnel

3. **User Behavior**:
   - Cohort retention
   - Feature usage (scans, leaderboard, export)
   - Telegram bot commands

---

## 9. Data Governance & Privacy

### 9.1 Data Retention Policy

```sql
-- Delete old scans (GDPR - right to be forgotten after 2 years)
DELETE FROM analytics.fact_scans
WHERE scanned_at < NOW() - INTERVAL '2 years';

-- Anonymize deleted users
UPDATE analytics.fact_scans
SET user_id = 'DELETED_USER'
WHERE user_id IN (
  SELECT id FROM "User" WHERE deleted_at IS NOT NULL
);
```

### 9.2 PII Handling

**Rules**:
- **Never log PII** (user IDs, Telegram IDs) in plaintext logs
- **Anonymize for analytics**: Hash user IDs before storing in analytics tables
- **Comply with GDPR**: Support user data export (`/export`) and deletion (`/delete`)

---

## 10. Command Shortcuts

Use these shortcuts to quickly access specific topics:

- **#analytics** - Business analytics, KPIs, SQL queries
- **#ml** - Machine learning for scam detection
- **#features** - ML feature engineering
- **#xgboost** - Training XGBoost models
- **#shap** - Model explainability with SHAP
- **#anomaly** - Anomaly detection for new scams
- **#ab-test** - A/B testing framework
- **#churn** - Churn prediction
- **#bi** - BI dashboards (Metabase, Looker)
- **#retention** - Cohort retention analysis
- **#funnel** - Conversion funnel analysis

---

## 11. Reference Materials

### 11.1 CryptoRugMunch Documentation

**Related Skills**:
- `rugmunch-architect` - System architecture, data model
- `crypto-scam-analyst` - Risk scoring algorithm (to be ML-ified)
- `testing-qa-specialist` - ML model testing strategies

**Project Docs**:
- `/docs/07-METRICS-ANALYTICS/success-metrics.md` - KPIs, success criteria
- `/docs/03-TECHNICAL/architecture/data-model.md` - Database schema

### 11.2 ML & Analytics Tools

**Machine Learning**:
- **XGBoost**: https://xgboost.readthedocs.io
- **scikit-learn**: https://scikit-learn.org
- **SHAP**: https://shap.readthedocs.io (model explainability)
- **MLflow**: https://mlflow.org (experiment tracking)

**Analytics**:
- **dbt**: https://docs.getdbt.com (data transformations)
- **Metabase**: https://metabase.com (BI dashboards)
- **GrowthBook**: https://growthbook.io (A/B testing)

**Data Science**:
- **Pandas**: https://pandas.pydata.org
- **NumPy**: https://numpy.org
- **SciPy**: https://scipy.org

---

## Summary

The **Analytics & Data Science Specialist** skill provides comprehensive expertise for extracting insights from CryptoRugMunch's data and improving scam detection through machine learning. Key capabilities:

1. **Business Analytics**: SQL queries for KPIs (DAU, MRR, churn, LTV, retention)
2. **Machine Learning**: XGBoost models for scam detection (95%+ accuracy)
3. **Feature Engineering**: 50+ features from blockchain data
4. **Anomaly Detection**: Isolation Forest for new scam patterns
5. **A/B Testing**: GrowthBook for experimentation
6. **Churn Prediction**: RandomForest for subscriber retention
7. **BI Dashboards**: Metabase for self-serve analytics
8. **Data Governance**: GDPR compliance, data retention policies

**ML Roadmap**:
- **Phase 1 (Month 4-6)**: Collect historical scan data
- **Phase 2 (Month 7-9)**: Train initial ML models, A/B test vs rule-based
- **Phase 3 (Month 10-12)**: Deploy ML model to production, continuous retraining

**Key Metrics**:
- ML model accuracy: >95%
- Scam detection recall: >98% (minimize false negatives)
- Model inference time: <2s
- Monthly model retraining

**Next Steps**:
1. Set up analytics warehouse schema
2. Create dbt transformations for aggregations
3. Build Metabase dashboards
4. Collect 90 days of scan data for ML training
5. Train initial XGBoost model
6. A/B test ML vs rule-based scoring
7. Deploy winning model to production

---

**Built with data-driven insights** 📊
**Powered by machine learning** 🤖
