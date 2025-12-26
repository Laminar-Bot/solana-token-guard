# CLAUDE CODE INSTRUCTIONS FOR LAMINAR BOT SPECIFICATION

## Quick Start

You have a master specification outline in `SPEC_MASTER.md`. Your job is to expand each section into a complete, production-ready specification.

## How to Work

### Step 1: Read the Master
Always read `SPEC_MASTER.md` first to understand:
- The overall architecture
- Technology decisions already made
- How components relate to each other

### Step 2: Expand Section by Section
When the user says "expand section X", create a complete spec file:

```
specs/
├── 01-domain-model.md
├── 02-database-schema.md
├── 03-sqlc-queries.md
├── 04-external-clients.md
├── 05-trading-engine.md
├── 06-screener.md
├── 07-risk-engine.md
├── 08-exit-engine.md
├── 09-executor.md
├── 10-keystore.md
├── 11-user-api.md
├── 12-telegram-bot.md
├── 13-frontend.md
├── 14-observability.md
├── 15-security.md
├── 16-testing.md
├── 17-deployment.md
├── 18-adrs.md
├── 19-phases.md
```

### Step 3: Quality Standards

**DO:**
- Write complete Go code implementations (not pseudocode)
- Include all struct fields with types and JSON tags
- Include all function signatures with full parameter types
- Include error handling
- Include comments explaining non-obvious logic
- Include example request/response payloads for APIs
- Include SQL with proper indexes
- Cross-reference related sections

**DON'T:**
- Use placeholders like "// TODO" or "implementation here"
- Skip edge cases
- Assume the reader knows implementation details
- Write partial code that won't compile
- Forget imports

### Example: Good vs Bad

**BAD:**
```go
func (e *Engine) CheckBuy(ctx context.Context, userID string, amount float64) bool {
    // Check limits
    // Return result
}
```

**GOOD:**
```go
// CheckBuy evaluates whether a buy trade should be allowed based on risk limits.
// Returns a CheckResult with approval status, reason if rejected, and any adjustments.
func (e *Engine) CheckBuy(ctx context.Context, userID user.UserID, tokenAddress string, requestedSizeSOL decimal.Decimal) (*CheckResult, error) {
    result := &CheckResult{Approved: true}
    log := e.logger.WithUserID(userID.String())

    // 1. Load user
    usr, err := e.userRepo.GetByID(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("get user: %w", err)
    }

    // 2. Check user status
    if !usr.Status.CanTrade() {
        return &CheckResult{
            Approved: false,
            Reason:   fmt.Sprintf("Trading is %s", usr.Status),
        }, nil
    }

    // ... complete implementation
}
```

## Key Reminders

### Multi-Tenancy
Every database query must include `user_id` in the WHERE clause. Never allow cross-tenant data access.

### Idempotency
All trade operations must be idempotent. Use idempotency keys to prevent duplicates.

### Security
- Never log private keys, full wallet addresses, or sensitive settings
- Truncate addresses in logs: `addr[:4] + "..." + addr[len(addr)-4:]`
- Store keys only in GCP Secret Manager

### Error Handling
- Wrap errors with context: `fmt.Errorf("operation: %w", err)`
- Log errors with structured context
- Return appropriate error types

### Decimal Math
- Use `shopspring/decimal` for all money calculations
- Never use float64 for financial amounts

## Useful Patterns

### Repository Pattern
```go
type Repository interface {
    GetByID(ctx context.Context, id ID) (*Entity, error)
    Create(ctx context.Context, entity *Entity) error
    Update(ctx context.Context, entity *Entity) error
    // ...
}
```

### Worker Pattern
```go
func (w *Worker) Start(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            job, err := w.consumer.Dequeue(ctx, timeout)
            if err != nil {
                // handle error
                continue
            }
            if job != nil {
                w.process(ctx, job)
            }
        }
    }
}
```

### Metrics Pattern
```go
func (s *Service) DoSomething(ctx context.Context) error {
    start := time.Now()
    defer func() {
        s.metrics.OperationDuration("something", time.Since(start))
    }()
    // ...
}
```

## Command Reference

User might say:
- "Expand section 5" → Create `specs/05-trading-engine.md`
- "Expand the domain model" → Create `specs/01-domain-model.md`
- "Complete the database schema" → Create `specs/02-database-schema.md`
- "Write the full executor spec" → Create `specs/09-executor.md`

## File Template

Each spec file should follow this structure:

```markdown
# [Section Name]

> Part of the Laminar Bot specification. See SPEC_MASTER.md for overview.

## Overview
Brief description of this component's purpose.

## Dependencies
- What this component depends on
- What depends on this component

## [Main Content Sections]
...

## Error Handling
How errors are handled in this component.

## Testing Notes
Key test cases for this component.

## Related Sections
- Links to related spec sections
```

---

Now you're ready. When the user asks to expand a section, create a complete, production-ready specification file.