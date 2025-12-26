# Git Commit Rules

**Version:** 1.0.0
**Last Updated:** 07/10/2025
**Purpose:** Git workflow standards and commit message conventions for MDL Fan Dev

## Core Git Principles

1. **Meaningful commits** - Each commit should represent a logical unit of work
2. **Conventional commits** - Follow standardised commit message format
3. **Atomic commits** - One feature/fix per commit when possible
4. **Clear history** - Write commit messages for future developers
5. **Sync with CHANGELOG.md** - Major changes documented in both places

---

## Commit Message Format

### Structure

```
<type>(<scope>): <subject>

[optional body]

[optional footer]
```

### Type

**Required.** Must be one of:

- **feat**: New feature for the user
- **fix**: Bug fix for the user
- **docs**: Documentation changes
- **style**: Formatting, missing semicolons, etc. (no code change)
- **refactor**: Code change that neither fixes a bug nor adds a feature
- **perf**: Performance improvement
- **test**: Adding or updating tests
- **build**: Build system or external dependency changes
- **ci**: CI/CD configuration changes
- **chore**: Other changes that don't modify src or test files
- **revert**: Reverts a previous commit

### Scope

**Optional.** The section of the codebase affected:

**Feature scopes:**
- `auth` - Authentication module
- `account` - Account management
- `payments` - Payment/Stripe integration
- `player` - Video player
- `fixtures` - Sports fixtures
- `playlists` - Content playlists
- `collections` - Content collections
- `routing` - Navigation/routing
- `ui` - UI components

**Technical scopes:**
- `api` - API integration
- `testing` - Test infrastructure
- `performance` - Performance optimisations
- `security` - Security improvements
- `config` - Configuration changes

### Subject

**Required.** Short description:

- Use imperative mood ("add" not "added" or "adds")
- Don't capitalise first letter
- No period at the end
- Maximum 72 characters
- Be specific and descriptive

### Body

**Optional.** Provide additional context:

- Explain **why** the change was made (not what/how)
- Wrap at 72 characters
- Separate from subject with blank line
- Can include bullet points

### Footer

**Optional.** Reference issues, breaking changes:

```
BREAKING CHANGE: description of breaking change

Fixes #123
Closes #456
```

---

## Commit Message Examples

### Feature Commits

**Good examples:**
```bash
feat(auth): add multi-factor authentication support

feat(payments): integrate Stripe payment intents API

feat(player): add playback speed control

feat(fixtures): implement live score updates via WebSocket
```

**From recent history:**
```bash
feat(testing): Achieve 100% test coverage for authentication module
```

### Fix Commits

**Good examples:**
```bash
fix(auth): resolve session timeout not clearing tokens

fix(payments): handle failed payment method updates correctly

fix(player): prevent video buffering on network switch

fix(ui): correct mobile menu z-index issue
```

### Test Commits

**From recent history:**
```bash
test: achieve 99.6% test coverage across all active features

test: add comprehensive test coverage for 5 feature modules
```

### Refactor Commits

**Good examples:**
```bash
refactor(api): extract common API error handling logic

refactor(auth): simplify token refresh mechanism

refactor(ui): convert class components to functional components
```

### Documentation Commits

**Good examples:**
```bash
docs: add API integration patterns guide

docs: update authentication flow documentation

docs(readme): add development setup instructions
```

### Performance Commits

**Good examples:**
```bash
perf(player): implement video chunk preloading

perf(fixtures): add React Query caching for fixture list

perf: lazy load route components to reduce initial bundle
```

### Style Commits

**Good examples:**
```bash
style(auth): format login form with Prettier

style: update ESLint configuration and fix violations
```

### Build Commits

**Good examples:**
```bash
build: upgrade React to version 18.3.0

build(vite): add bundle size optimization

build: configure terser for production builds
```

---

## Multi-line Commit Examples

### With Body

```bash
feat(payments): add subscription cancellation flow

Implement complete subscription cancellation workflow including:
- Cancellation confirmation modal
- Immediate vs end-of-period cancellation options
- Stripe subscription update API integration
- Email notification trigger

The flow matches the design specifications in FFE_DESIGN_ACCOUNT_MANAGEMENT.md
```

### With Breaking Change

```bash
feat(auth)!: migrate to AWS Cognito v3 SDK

BREAKING CHANGE: AuthContext API has changed. The `signIn` method now
returns a Promise and requires async/await. Components using AuthContext
need to be updated.

Before:
  const { signIn } = useAuth();
  signIn(email, password);

After:
  const { signIn } = useAuth();
  await signIn(email, password);

Migration guide: docs/36_MIGRATION_GUIDE.md
```

### With Issue Reference

```bash
fix(player): resolve audio sync issue on Safari

Audio track was desyncing from video on Safari when seeking backwards.
Issue was caused by incorrect timestamp calculation in MediaElement.jsx.

Fixed by using the player's native currentTime property instead of
calculating from playback events.

Fixes #342
Closes #356
```

---

## Branch Naming Conventions

### Format

```
<type>/<ticket-id>-<short-description>
```

### Examples

```bash
feature/DEV-1336-playlist-naming
fix/DEV-1245-payment-failure
refactor/DEV-1189-api-error-handling
test/DEV-1432-fixture-coverage
docs/DEV-1501-api-documentation
chore/DEV-1612-dependency-updates
```

### Branch Types

- `feature/` - New features
- `fix/` - Bug fixes
- `refactor/` - Code refactoring
- `test/` - Test additions/improvements
- `docs/` - Documentation changes
- `chore/` - Maintenance tasks
- `hotfix/` - Urgent production fixes

**From recent history:**
```bash
DEV-1336-PLAYLIST-NAMING
terms-changes-20250910
```

---

## Commit Workflow

### 1. Make Atomic Commits

**DO:**
```bash
# ✅ Separate commits for different concerns
git add src/features/auth/LoginForm.jsx
git commit -m "feat(auth): add email validation to login form"

git add src/features/auth/__tests__/LoginForm.test.jsx
git commit -m "test(auth): add validation tests for login form"
```

**DON'T:**
```bash
# ❌ Mixed concerns in one commit
git add .
git commit -m "various updates"
```

### 2. Review Changes Before Committing

```bash
# Review staged changes
git diff --staged

# Review specific file
git diff --staged src/features/auth/LoginForm.jsx
```

### 3. Amend Recent Commits (Carefully)

```bash
# Add forgotten files to last commit
git add forgotten-file.js
git commit --amend --no-edit

# Fix commit message
git commit --amend -m "feat(auth): add email validation to login form"
```

**⚠️ WARNING:** Only amend commits that haven't been pushed!

### 4. Write Descriptive Commit Messages

**Take time to write good messages. Future you will thank present you!**

---

## Integration with CHANGELOG.md

### When to Update CHANGELOG

Update CHANGELOG.md for:
- **New features** (feat commits)
- **Bug fixes** (fix commits)
- **Breaking changes** (any commits with BREAKING CHANGE)
- **Deprecations** (when marking features as deprecated)
- **Security fixes** (critical security patches)
- **Performance improvements** (significant perf commits)

**Reference:** [.claude/changelog-rules.md](./changelog-rules.md)

### Commit Message → CHANGELOG Mapping

| Commit Type | CHANGELOG Section |
|-------------|------------------|
| `feat` | Added |
| `fix` | Fixed |
| `perf` | Changed (Performance) |
| `security` | Security |
| `BREAKING CHANGE` | Changed (Breaking) |
| `deprecate` | Deprecated |

### Example Workflow

```bash
# 1. Make changes
git add src/features/payments/CheckoutForm.jsx

# 2. Commit with conventional format
git commit -m "feat(payments): add Apple Pay support"

# 3. Update CHANGELOG.md
# Add entry under [Unreleased] → Added section:
# - Apple Pay payment method support (#789)

git add CHANGELOG.md
git commit -m "docs: update CHANGELOG for Apple Pay feature"
```

---

## Pull Request Guidelines

### PR Title Format

Use the same format as commit messages:

```
feat(payments): add Apple Pay support
fix(auth): resolve session timeout issue
test: achieve 100% coverage for payments module
```

### PR Description Template

```markdown
## Description
Brief description of what this PR does

## Type of Change
- [ ] New feature (feat)
- [ ] Bug fix (fix)
- [ ] Breaking change (BREAKING CHANGE)
- [ ] Documentation update (docs)
- [ ] Test coverage improvement (test)
- [ ] Performance improvement (perf)

## Testing
- [ ] Tests added/updated
- [ ] All tests passing
- [ ] Manual testing completed

## Documentation
- [ ] CHANGELOG.md updated (if needed)
- [ ] TESTING_GUIDE.md updated (if test changes)
- [ ] Code comments added/updated

## Screenshots (if applicable)
[Add screenshots for UI changes]

## Related Issues
Fixes #123
Closes #456
```

---

## Git Best Practices

### DO

- ✅ Write clear, descriptive commit messages
- ✅ Make small, focused commits
- ✅ Review changes before committing
- ✅ Pull latest changes before pushing
- ✅ Resolve merge conflicts carefully
- ✅ Use branches for all work (never commit to main)
- ✅ Update tests with code changes
- ✅ Run tests before pushing

### DON'T

- ❌ Commit commented-out code
- ❌ Commit console.logs (except in error handling)
- ❌ Commit secrets or credentials
- ❌ Use generic messages like "fixes", "updates"
- ❌ Mix multiple features in one commit
- ❌ Force push to main branch
- ❌ Amend commits after pushing
- ❌ Commit broken code

---

## Handling Sensitive Data

### Pre-Commit Checklist

Before committing:

- [ ] No API keys in code
- [ ] No passwords or tokens
- [ ] No user data or PII
- [ ] No `.env` files
- [ ] `constants.js` reviewed for secrets

**Reference:** [.claude/security-rules.md](./security-rules.md)

### If You Accidentally Commit Secrets

```bash
# 1. Immediately rotate the exposed credentials

# 2. Remove from git history (if not pushed)
git reset HEAD~1
# Edit files to remove secrets
git add .
git commit -m "fix(security): remove exposed credentials"

# 3. If already pushed, contact team lead immediately
```

---

## Commit Message Checklist

Before committing:

- [ ] Type is correct (feat, fix, etc.)
- [ ] Scope is accurate (or omitted if not needed)
- [ ] Subject is descriptive and concise (<72 chars)
- [ ] Subject uses imperative mood ("add" not "added")
- [ ] Subject doesn't end with period
- [ ] Body explains why (if needed)
- [ ] Breaking changes documented (if any)
- [ ] Issues referenced (if applicable)
- [ ] Tests updated
- [ ] No secrets in files
- [ ] Code linted and formatted

---

## Common Commit Mistakes

### 1. Vague Messages

**❌ Bad:**
```bash
git commit -m "fix stuff"
git commit -m "updates"
git commit -m "WIP"
```

**✅ Good:**
```bash
git commit -m "fix(auth): resolve token refresh race condition"
git commit -m "feat(fixtures): add filter by team functionality"
git commit -m "test(payments): add integration tests for checkout flow"
```

### 2. Too Large Commits

**❌ Bad:**
```bash
git add .
git commit -m "feat: complete authentication rewrite"
# 50 files changed, 2000 insertions, 1500 deletions
```

**✅ Good:**
Break into smaller commits:
```bash
git commit -m "refactor(auth): extract token management logic"
git commit -m "feat(auth): add MFA support"
git commit -m "test(auth): add comprehensive auth tests"
git commit -m "docs(auth): update authentication flow documentation"
```

### 3. Mixing Concerns

**❌ Bad:**
```bash
git commit -m "fix login bug and update readme"
```

**✅ Good:**
```bash
git commit -m "fix(auth): resolve login redirect issue"
git commit -m "docs: update authentication setup in readme"
```

---

## Squashing Commits

### When to Squash

**Before merging PR:**
- Multiple "WIP" commits
- Multiple "fix typo" commits
- Commits that fix previous commits in same PR

**Example:**
```bash
# Interactive rebase to squash last 3 commits
git rebase -i HEAD~3

# In editor, change 'pick' to 'squash' for commits to combine
pick abc123 feat(auth): add MFA support
squash def456 fix: typo in MFA form
squash ghi789 test: add MFA tests

# Result: Single commit with all changes
```

### When NOT to Squash

- Commits already pushed to shared branch
- Commits with different types (feat + fix)
- Commits from different authors

---

## Reverting Commits

### Revert Pattern

```bash
# Revert a single commit
git revert abc123

# This creates a new commit:
# "revert: feat(auth): add MFA support"
```

### Revert Commit Message

```bash
revert: feat(auth): add MFA support

This reverts commit abc123.

Reason: MFA implementation caused login issues on Safari.
Will be reimplemented with proper browser testing.
```

---

## Git Aliases (Optional)

Add to `~/.gitconfig`:

```bash
[alias]
  # Conventional commit helpers
  feat = "!f() { git commit -m \"feat: $*\"; }; f"
  fix = "!f() { git commit -m \"fix: $*\"; }; f"
  docs = "!f() { git commit -m \"docs: $*\"; }; f"
  test = "!f() { git commit -m \"test: $*\"; }; f"

  # Useful shortcuts
  st = status
  co = checkout
  br = branch
  last = log -1 HEAD
  unstage = reset HEAD --
```

**Usage:**
```bash
git feat "add Apple Pay support"
# Creates: feat: add Apple Pay support
```

---

## Resources

### Internal Documentation
- [.claude/changelog-rules.md](./changelog-rules.md) - CHANGELOG update guidelines
- [.claude/security-rules.md](./security-rules.md) - Security checklist

### External Resources
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Git Best Practices](https://git-scm.com/book/en/v2/Distributed-Git-Contributing-to-a-Project)
- [How to Write a Git Commit Message](https://chris.beams.io/posts/git-commit/)
- [Angular Commit Guidelines](https://github.com/angular/angular/blob/main/CONTRIBUTING.md#commit)

### Reference Examples
Check recent commits for patterns:
```bash
git log --oneline -20
```