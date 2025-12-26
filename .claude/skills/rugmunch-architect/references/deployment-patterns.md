# CryptoRugMunch Deployment Patterns

**Status**: ✅ Complete
**Last Updated**: 2025-01-19

This reference provides detailed deployment guides for all environments: Local, Railway (staging), and AWS ECS (production).

---

## Table of Contents

1. [Deployment Philosophy](#deployment-philosophy)
2. [Local Development Setup](#local-development-setup)
3. [Railway Staging Deployment](#railway-staging-deployment)
4. [AWS ECS Production Deployment](#aws-ecs-production-deployment)
5. [CI/CD Pipeline](#cicd-pipeline)
6. [Rollback Procedures](#rollback-procedures)
7. [Zero-Downtime Deployments](#zero-downtime-deployments)

---

## Deployment Philosophy

### Core Principles

1. **Environment Parity**: Dev, staging, production should be as similar as possible
2. **Infrastructure as Code**: All infrastructure defined in version control
3. **Automated Deployments**: Human pushes button, automation does the work
4. **Rollback-Ready**: Every deployment can be rolled back in < 5 minutes
5. **Monitored Deployments**: Watch metrics during and after deploys

### Three-Stage Strategy

```
┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│   Local     │      │   Railway   │      │   AWS ECS   │
│  (Docker)   │─────▶│  (Staging)  │─────▶│(Production) │
└─────────────┘      └─────────────┘      └─────────────┘
   Dev laptop         Auto-deploy PR        Manual deploy
   Polling mode       Webhook mode          Webhook mode
   SQLite option      PostgreSQL            RDS PostgreSQL
   Redis local        Railway Redis         ElastiCache
```

---

## Local Development Setup

### Prerequisites

```bash
# Required software
node --version    # v20+
docker --version  # 24+
psql --version    # 15+
redis-cli --version  # 7+

# Optional but recommended
npm install -g pnpm
brew install postgresql@15
brew install redis
```

### Option 1: Full Docker (Recommended)

**File**: `docker-compose.yml`

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: rugmunch
      POSTGRES_PASSWORD: dev_password
      POSTGRES_DB: rugmunch_dev
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U rugmunch"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5

  api:
    build: .
    command: npm run dev
    ports:
      - "3000:3000"
    environment:
      NODE_ENV: development
      DATABASE_URL: postgresql://rugmunch:dev_password@postgres:5432/rugmunch_dev
      REDIS_URL: redis://redis:6379
    volumes:
      - .:/app
      - /app/node_modules
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy

  worker:
    build: .
    command: npm run worker:dev
    environment:
      NODE_ENV: development
      DATABASE_URL: postgresql://rugmunch:dev_password@postgres:5432/rugmunch_dev
      REDIS_URL: redis://redis:6379
      WORKER_CONCURRENCY: 4
    volumes:
      - .:/app
      - /app/node_modules
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy

volumes:
  postgres_data:
  redis_data:
```

**Setup**:

```bash
# 1. Clone repo
git clone https://github.com/Laminar-Bot/rug-muncher.git
cd rug-muncher

# 2. Copy environment variables
cp .env.example .env

# 3. Start all services
docker-compose up -d

# 4. Run database migrations
docker-compose exec api npx prisma migrate dev

# 5. Seed database (optional)
docker-compose exec api npx prisma db seed

# 6. View logs
docker-compose logs -f api worker

# 7. Stop all services
docker-compose down
```

**Verify**:

```bash
# Check all services are running
docker-compose ps

# Expected output:
# NAME                COMMAND                  STATUS              PORTS
# rug-muncher-api     "npm run dev"            Up 2 minutes        0.0.0.0:3000->3000/tcp
# rug-muncher-worker  "npm run worker:dev"     Up 2 minutes
# rug-muncher-postgres "docker-entrypoint…"    Up 2 minutes        0.0.0.0:5432->5432/tcp
# rug-muncher-redis   "redis-server --app…"    Up 2 minutes        0.0.0.0:6379->6379/tcp

# Test API
curl http://localhost:3000/health
# Expected: {"status":"ok"}

# Test database
docker-compose exec postgres psql -U rugmunch -d rugmunch_dev -c "SELECT COUNT(*) FROM \"User\";"

# Test Redis
docker-compose exec redis redis-cli ping
# Expected: PONG
```

---

### Option 2: Native (No Docker)

**When to use**: If Docker causes performance issues on your machine

**Setup**:

```bash
# 1. Install PostgreSQL and Redis
brew install postgresql@15 redis

# 2. Start services
brew services start postgresql@15
brew services start redis

# 3. Create database
createdb rugmunch_dev

# 4. Clone and install
git clone https://github.com/Laminar-Bot/rug-muncher.git
cd rug-muncher
pnpm install

# 5. Configure environment
cp .env.example .env
# Edit .env:
# DATABASE_URL="postgresql://localhost:5432/rugmunch_dev"
# REDIS_URL="redis://localhost:6379"

# 6. Run migrations
npx prisma migrate dev

# 7. Start services in separate terminals
pnpm dev          # Terminal 1 (API)
pnpm worker:dev   # Terminal 2 (Worker)
pnpm telegram:dev # Terminal 3 (Telegram bot)
```

---

### Local Development Workflow

**Daily workflow**:

```bash
# Start work
git pull origin main
docker-compose up -d
pnpm dev

# Make changes
# ... edit code ...

# Run tests
pnpm test:unit
pnpm test:integration

# Commit
git add .
git commit -m "feat: add new feature"
git push origin feature-branch

# Create PR (triggers CI/CD)
gh pr create --title "Add new feature"

# End of day
docker-compose down
```

**Hot reload**: Code changes automatically reload API and worker

---

## Railway Staging Deployment

### Initial Setup

**Step 1: Create Railway Account**

```bash
# Install Railway CLI
npm install -g @railway/cli

# Login
railway login

# Link GitHub account
# https://railway.app/account/tokens
```

**Step 2: Create Project**

```bash
# From project root
railway init

# Create services
railway service add postgres
railway service add redis
railway service add api
railway service add worker
railway service add telegram-bot
```

**Step 3: Configure Each Service**

**API Service**:

```toml
# railway.toml (in repo root)
[build]
builder = "NIXPACKS"

[deploy]
startCommand = "npm run start"
healthcheckPath = "/health"
healthcheckTimeout = 100
restartPolicyType = "ON_FAILURE"
restartPolicyMaxRetries = 10
```

**Worker Service**:

```toml
# railway.worker.toml
[build]
builder = "NIXPACKS"

[deploy]
startCommand = "npm run worker:start"
restartPolicyType = "ON_FAILURE"
restartPolicyMaxRetries = 10
```

**Environment Variables** (set via Railway dashboard or CLI):

```bash
# Shared variables
railway variables set NODE_ENV=staging
railway variables set LOG_LEVEL=info

# Database (auto-set by Railway)
# DATABASE_URL (provided by Railway Postgres)
# REDIS_URL (provided by Railway Redis)

# Telegram
railway variables set TELEGRAM_BOT_TOKEN=<token>

# Blockchain APIs
railway variables set HELIUS_API_KEY=<key>
railway variables set BIRDEYE_API_KEY=<key>
railway variables set RUGCHECK_API_KEY=<key>

# Monitoring
railway variables set DATADOG_API_KEY=<key>
railway variables set SENTRY_DSN=<dsn>

# Stripe
railway variables set STRIPE_API_KEY=<key>
railway variables set STRIPE_WEBHOOK_SECRET=<secret>
```

**Step 4: Deploy**

```bash
# Deploy all services
railway up

# Or deploy specific service
railway up --service api

# View logs
railway logs --service api
railway logs --service worker

# View deployments
railway status
```

**Step 5: Set Telegram Webhook**

```bash
# Get Railway API URL
RAILWAY_URL=$(railway status --service api --json | jq -r '.url')

# Set webhook
curl -X POST "https://api.telegram.org/bot<BOT_TOKEN>/setWebhook" \
  -d "url=${RAILWAY_URL}/telegram-webhook/<BOT_TOKEN>"

# Verify webhook
curl "https://api.telegram.org/bot<BOT_TOKEN>/getWebhookInfo"
```

---

### Railway Auto-Deployments

**Setup GitHub integration**:

1. Go to Railway dashboard → Settings → Integrations
2. Connect GitHub repository
3. Configure auto-deploy:
   - **Branch**: `main`
   - **Deploy on push**: Enabled
   - **Deploy on PR**: Enabled (separate preview environment)

**Deploy workflow**:

```bash
# 1. Push to feature branch
git push origin feature-branch

# 2. Create PR
gh pr create

# 3. Railway automatically creates preview environment
# URL: https://rugmunch-pr-123.railway.app

# 4. Test preview environment
curl https://rugmunch-pr-123.railway.app/health

# 5. Merge PR
gh pr merge

# 6. Railway automatically deploys to staging
# URL: https://rugmunch-staging.railway.app
```

---

### Railway Scaling

**Manual scaling** (via dashboard):
- API: 1-2 instances
- Worker: 2 instances × 4 concurrency = 8 concurrent scans

**Resource limits**:
```yaml
# railway.json
{
  "deploy": {
    "numReplicas": 2,
    "sleepApplication": false,
    "restartPolicyType": "ON_FAILURE"
  }
}
```

---

## AWS ECS Production Deployment

### When to Migrate

**Triggers** (any one of):
- 100K scans/month sustained
- Railway costs > $500/month
- Need multi-region deployment
- Month 6 of roadmap

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                         AWS VPC                              │
│                                                              │
│  ┌────────────────┐                                         │
│  │  Route 53      │                                         │
│  │  api.rugmunch  │                                         │
│  └────────┬───────┘                                         │
│           │                                                  │
│  ┌────────▼────────────────────────────┐                   │
│  │     Application Load Balancer       │                   │
│  │  (ALB with SSL/TLS termination)     │                   │
│  └────────┬────────────────────────────┘                   │
│           │                                                  │
│  ┌────────▼───────────┐  ┌──────────────────┐             │
│  │   ECS Service      │  │  ECS Service     │             │
│  │   (API)            │  │  (Worker)        │             │
│  │                    │  │                  │             │
│  │  ┌──────────────┐  │  │  ┌────────────┐  │             │
│  │  │ Task (API)   │  │  │  │ Task (Wrkr)│  │             │
│  │  │ Fargate      │  │  │  │ Fargate    │  │             │
│  │  └──────────────┘  │  │  └────────────┘  │             │
│  └────────┬───────────┘  └──────────┬───────┘             │
│           │                          │                      │
│  ┌────────▼──────────────────────────▼───────┐             │
│  │           ElastiCache Redis               │             │
│  │         (Cluster mode enabled)            │             │
│  └───────────────────────────────────────────┘             │
│                                                              │
│  ┌──────────────────────────────────────────┐               │
│  │      RDS PostgreSQL Multi-AZ             │               │
│  │  (Primary + Read Replica)                │               │
│  └──────────────────────────────────────────┘               │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

### Infrastructure Setup (Terraform)

**File**: `terraform/main.tf`

```hcl
# VPC
module "vpc" {
  source = "terraform-aws-modules/vpc/aws"

  name = "rugmunch-vpc"
  cidr = "10.0.0.0/16"

  azs             = ["us-east-1a", "us-east-1b"]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24"]

  enable_nat_gateway = true
  single_nat_gateway = false
}

# RDS PostgreSQL
resource "aws_db_instance" "postgres" {
  identifier           = "rugmunch-postgres"
  engine              = "postgres"
  engine_version      = "15.4"
  instance_class      = "db.t3.medium"
  allocated_storage   = 100
  storage_encrypted   = true

  db_name  = "rugmunch"
  username = "rugmunch_admin"
  password = data.aws_secretsmanager_secret_version.db_password.secret_string

  multi_az               = true
  backup_retention_period = 7
  backup_window          = "03:00-04:00"
  maintenance_window     = "sun:04:00-sun:05:00"

  vpc_security_group_ids = [aws_security_group.rds.id]
  db_subnet_group_name   = aws_db_subnet_group.main.name

  skip_final_snapshot = false
  final_snapshot_identifier = "rugmunch-postgres-final-snapshot"
}

# Read Replica
resource "aws_db_instance" "postgres_replica" {
  identifier          = "rugmunch-postgres-replica"
  replicate_source_db = aws_db_instance.postgres.identifier
  instance_class      = "db.t3.medium"

  vpc_security_group_ids = [aws_security_group.rds.id]
}

# ElastiCache Redis
resource "aws_elasticache_replication_group" "redis" {
  replication_group_id       = "rugmunch-redis"
  replication_group_description = "Redis cluster for CryptoRugMunch"

  engine               = "redis"
  engine_version       = "7.0"
  node_type            = "cache.t3.medium"
  number_cache_clusters = 2

  automatic_failover_enabled = true
  multi_az_enabled          = true

  subnet_group_name = aws_elasticache_subnet_group.main.name
  security_group_ids = [aws_security_group.redis.id]
}

# ECS Cluster
resource "aws_ecs_cluster" "main" {
  name = "rugmunch-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

# ECS Task Definition (API)
resource "aws_ecs_task_definition" "api" {
  family                   = "rugmunch-api"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "512"
  memory                   = "1024"

  container_definitions = jsonencode([{
    name  = "api"
    image = "${aws_ecr_repository.api.repository_url}:latest"

    portMappings = [{
      containerPort = 3000
      protocol      = "tcp"
    }]

    environment = [
      { name = "NODE_ENV", value = "production" },
      { name = "PORT", value = "3000" }
    ]

    secrets = [
      {
        name      = "DATABASE_URL"
        valueFrom = "${aws_secretsmanager_secret.db_url.arn}"
      },
      {
        name      = "REDIS_URL"
        valueFrom = "${aws_secretsmanager_secret.redis_url.arn}"
      }
      # ... other secrets
    ]

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.api.name
        "awslogs-region"        = "us-east-1"
        "awslogs-stream-prefix" = "api"
      }
    }
  }])
}

# ECS Service (API)
resource "aws_ecs_service" "api" {
  name            = "rugmunch-api"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.api.arn
  desired_count   = 2
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = module.vpc.private_subnets
    security_groups  = [aws_security_group.api.id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.api.arn
    container_name   = "api"
    container_port   = 3000
  }

  depends_on = [aws_lb_listener.api]
}

# Auto Scaling (API)
resource "aws_appautoscaling_target" "api" {
  max_capacity       = 10
  min_capacity       = 2
  resource_id        = "service/${aws_ecs_cluster.main.name}/${aws_ecs_service.api.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "api_cpu" {
  name               = "api-cpu-scaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.api.resource_id
  scalable_dimension = aws_appautoscaling_target.api.scalable_dimension
  service_namespace  = aws_appautoscaling_target.api.service_namespace

  target_tracking_scaling_policy_configuration {
    target_value = 70.0

    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }

    scale_in_cooldown  = 300
    scale_out_cooldown = 60
  }
}

# Worker Task Definition & Service
# (Similar to API, but with queue-based scaling)

# Application Load Balancer
resource "aws_lb" "main" {
  name               = "rugmunch-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = module.vpc.public_subnets
}

resource "aws_lb_listener" "api" {
  load_balancer_arn = aws_lb.main.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS-1-2-2017-01"
  certificate_arn   = aws_acm_certificate.main.arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.api.arn
  }
}
```

---

### Deployment Process

**Initial deployment**:

```bash
# 1. Set up Terraform
cd terraform
terraform init

# 2. Plan infrastructure
terraform plan -out=plan.tfplan

# 3. Apply infrastructure
terraform apply plan.tfplan

# 4. Build and push Docker images
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-east-1.amazonaws.com

docker build -t rugmunch-api .
docker tag rugmunch-api:latest <account-id>.dkr.ecr.us-east-1.amazonaws.com/rugmunch-api:latest
docker push <account-id>.dkr.ecr.us-east-1.amazonaws.com/rugmunch-api:latest

# 5. Run database migrations
aws ecs run-task \
  --cluster rugmunch-cluster \
  --task-definition rugmunch-migration \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-xxx],securityGroups=[sg-xxx]}"

# 6. Deploy services
aws ecs update-service \
  --cluster rugmunch-cluster \
  --service rugmunch-api \
  --force-new-deployment

# 7. Verify deployment
aws ecs describe-services \
  --cluster rugmunch-cluster \
  --services rugmunch-api
```

**Subsequent deployments** (via CI/CD):

```yaml
# .github/workflows/deploy-production.yml
name: Deploy Production

on:
  push:
    tags:
      - 'v*'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v2
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-east-1

      - name: Login to ECR
        run: |
          aws ecr get-login-password | docker login --username AWS --password-stdin ${{ secrets.ECR_REGISTRY }}

      - name: Build and push
        run: |
          docker build -t ${{ secrets.ECR_REGISTRY }}/rugmunch-api:${{ github.sha }} .
          docker push ${{ secrets.ECR_REGISTRY }}/rugmunch-api:${{ github.sha }}

      - name: Update ECS service
        run: |
          aws ecs update-service \
            --cluster rugmunch-cluster \
            --service rugmunch-api \
            --force-new-deployment
```

---

### Worker Auto-Scaling

**Scale based on queue depth**:

```hcl
# CloudWatch metric for queue depth
resource "aws_cloudwatch_metric_alarm" "queue_depth_high" {
  alarm_name          = "rugmunch-queue-depth-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "ApproximateNumberOfMessagesVisible"
  namespace           = "AWS/SQS"
  period              = "60"
  statistic           = "Average"
  threshold           = "1000"
  alarm_description   = "Scale up workers when queue depth > 1000"
  alarm_actions       = [aws_appautoscaling_policy.worker_scale_up.arn]

  dimensions = {
    QueueName = "token-scan"
  }
}

resource "aws_appautoscaling_policy" "worker_scale_up" {
  name               = "worker-scale-up"
  service_namespace  = "ecs"
  resource_id        = "service/${aws_ecs_cluster.main.name}/${aws_ecs_service.worker.name}"
  scalable_dimension = "ecs:service:DesiredCount"

  step_scaling_policy_configuration {
    adjustment_type         = "ChangeInCapacity"
    cooldown               = 60
    metric_aggregation_type = "Average"

    step_adjustment {
      scaling_adjustment          = 2
      metric_interval_lower_bound = 0
    }
  }
}
```

---

## CI/CD Pipeline

### GitHub Actions Workflow

**File**: `.github/workflows/ci-cd.yml`

```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v3

      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '20'
          cache: 'pnpm'

      - name: Install dependencies
        run: pnpm install

      - name: Run database migrations
        env:
          DATABASE_URL: postgresql://postgres:test@localhost:5432/test
        run: npx prisma migrate deploy

      - name: Run unit tests
        run: pnpm test:unit

      - name: Run integration tests
        env:
          DATABASE_URL: postgresql://postgres:test@localhost:5432/test
          REDIS_URL: redis://localhost:6379
        run: pnpm test:integration

      - name: Upload coverage
        uses: codecov/codecov-action@v3

  deploy-staging:
    needs: test
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Deploy to Railway
        run: |
          curl -fsSL https://railway.app/install.sh | sh
          railway link ${{ secrets.RAILWAY_PROJECT_ID }}
          railway up --service api
          railway up --service worker

  deploy-production:
    needs: test
    if: startsWith(github.ref, 'refs/tags/v')
    runs-on: ubuntu-latest
    steps:
      # ... see AWS deployment steps above
```

---

## Rollback Procedures

### Railway Rollback

```bash
# Via CLI
railway rollback --service api

# Via dashboard
# Go to Deployments → Click previous deployment → "Redeploy"
```

### AWS ECS Rollback

```bash
# Option 1: Rollback to previous task definition
aws ecs update-service \
  --cluster rugmunch-cluster \
  --service rugmunch-api \
  --task-definition rugmunch-api:42  # Previous version

# Option 2: Rollback via tag
docker pull <ecr-registry>/rugmunch-api:v1.2.0
docker tag <ecr-registry>/rugmunch-api:v1.2.0 <ecr-registry>/rugmunch-api:latest
docker push <ecr-registry>/rugmunch-api:latest

aws ecs update-service \
  --cluster rugmunch-cluster \
  --service rugmunch-api \
  --force-new-deployment

# Option 3: Blue/Green rollback (instant)
aws deploy stop-deployment \
  --deployment-id d-XXXXXXXXX \
  --auto-rollback-enabled
```

---

## Zero-Downtime Deployments

### Strategy: Blue/Green with Health Checks

**ECS configuration**:

```hcl
resource "aws_ecs_service" "api" {
  # ... other config ...

  deployment_configuration {
    maximum_percent         = 200  # Allow double capacity during deploy
    minimum_healthy_percent = 100  # Keep all current tasks running

    deployment_circuit_breaker {
      enable   = true
      rollback = true
    }
  }

  health_check_grace_period_seconds = 60

  load_balancer {
    target_group_arn = aws_lb_target_group.api.arn
    container_name   = "api"
    container_port   = 3000
  }
}

# Target group health check
resource "aws_lb_target_group" "api" {
  # ... other config ...

  health_check {
    enabled             = true
    path                = "/health"
    protocol            = "HTTP"
    healthy_threshold   = 2
    unhealthy_threshold = 3
    timeout             = 5
    interval            = 30
    matcher             = "200"
  }

  deregistration_delay = 30  # Wait 30s before removing old tasks
}
```

**Deployment flow**:

```
1. Deploy new task definition (Green)
2. Start new tasks alongside old tasks (Blue)
3. Wait for health checks to pass (60s)
4. ALB routes traffic to Green tasks
5. Drain connections from Blue tasks (30s)
6. Terminate Blue tasks
```

---

## Deployment Checklist

### Pre-Deployment

- [ ] All tests passing (unit, integration, E2E)
- [ ] Changelog updated
- [ ] Database migrations tested
- [ ] Environment variables verified
- [ ] Monitoring dashboards ready
- [ ] Rollback plan documented
- [ ] Team notified (Slack #deployments)

### During Deployment

- [ ] Watch deployment progress (ECS console or CLI)
- [ ] Monitor error rates (Sentry)
- [ ] Monitor latency (DataDog)
- [ ] Check health endpoint returning 200
- [ ] Verify worker processing jobs
- [ ] Test critical user flows (scan, payment)

### Post-Deployment

- [ ] Verify all services healthy
- [ ] Check error rates baseline
- [ ] Review CloudWatch logs
- [ ] Test Telegram bot functionality
- [ ] Monitor for 30 minutes
- [ ] Mark deployment as successful (Slack)

### If Issues Detected

- [ ] Check metrics in DataDog
- [ ] Review error logs in Sentry
- [ ] Decide: Fix forward or rollback
- [ ] Execute rollback if needed (< 5 minutes)
- [ ] Post-mortem document issues

---

## Related Documentation

- `docs/03-TECHNICAL/operations/worker-deployment.md` - Detailed worker deployment
- `docs/03-TECHNICAL/operations/monitoring-alerting-setup.md` - Monitoring setup
- `docs/03-TECHNICAL/operations/environment-variables.md` - Environment configuration
- `docs/03-TECHNICAL/operations/ci-cd-pipeline.md` - CI/CD details
- `docs/03-TECHNICAL/operations/disaster-recovery.md` - DR procedures
