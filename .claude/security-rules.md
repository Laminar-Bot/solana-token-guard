# Security Rules

**Version:** 1.0.0
**Last Updated:** 07/10/2025
**Purpose:** Security standards and patterns for the MDL Fan Dev application

## Core Security Principles

1. **Never trust user input** - Validate and sanitise all data
2. **Defence in depth** - Multiple layers of security
3. **Least privilege** - Minimum necessary permissions
4. **Fail securely** - Errors should not expose sensitive information
5. **Keep secrets secret** - Never commit credentials or sensitive data

---

## Configuration & Secrets Management

### ⚠️ CRITICAL: Never Commit Secrets

**What NOT to commit:**
- API keys (except public ones like Stripe publishable key)
- Private tokens
- AWS credentials
- Database passwords
- Encryption keys
- User data or PII

**Current Pattern (constants.js):**
```javascript
// ❌ BAD - Sensitive keys in constants.js (legacy pattern)
export const SENTRY_DSN = "https://..."; // This is OK - DSN is public
export const BITMOVIN_KEY = "..."; // Player key - acceptable for client-side

// ✅ GOOD - Use environment variables for sensitive data
export const API_URL = import.meta.env.VITE_API_URL;
```

### Constants.js Security Checklist

When adding to `src/utils/constants.js`:

- [ ] Is this value safe to expose in client-side code?
- [ ] Could this value be used maliciously if discovered?
- [ ] Does this contain credentials or access tokens?
- [ ] Should this be environment-specific?

**Reference:** [docs/01_CONFIGURATION_MANAGEMENT.md](../docs/01_CONFIGURATION_MANAGEMENT.md)

---

## Authentication & Authorization

### Token Storage

**DO:**
```javascript
// ✅ Store tokens in localStorage (current pattern)
localStorage.setItem('accessToken', token);
localStorage.setItem('idToken', idToken);
localStorage.setItem('refreshToken', refreshToken);

// ✅ Always clear on logout
localStorage.removeItem('accessToken');
localStorage.removeItem('idToken');
localStorage.removeItem('refreshToken');
```

**DON'T:**
```javascript
// ❌ Never log tokens
console.log('Token:', token);

// ❌ Never send tokens in URL params
fetch(`/api/user?token=${token}`);

// ❌ Never store in global variables
window.userToken = token;
```

### Token Usage in API Calls

**Pattern from `apiFetch.js`:**
```javascript
// ✅ Correct pattern
let token = localStorage.getItem("accessToken");

if (token) {
  config.headers = {
    Authorization: `Bearer ${token}`,
    ...config.headers,
  };
}
```

### AWS Cognito Security

**MFA Implementation:**
- Always recommend MFA during registration
- Support TOTP authentication
- Never bypass MFA for "convenience"

**Password Requirements:**
- Minimum 8 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one number
- At least one special character

**Reference Implementation:** `src/features/authentication/PasswordValidation.jsx`

---

## Payment Security (Stripe)

### 🔒 CRITICAL: PCI Compliance

**NEVER log payment information:**
```javascript
// ❌ NEVER DO THIS
console.log('Card number:', cardNumber);
console.log('CVV:', cvv);
console.log('Payment intent:', paymentIntent);
console.log('Customer data:', stripeCustomer);

// ❌ NEVER store raw card data
localStorage.setItem('cardNumber', cardNumber);

// ❌ NEVER send card data directly
fetch('/api/payment', { body: { cardNumber, cvv } });
```

**DO use Stripe Elements:**
```javascript
// ✅ Use Stripe.js and Elements for all card input
import { CardElement, useStripe, useElements } from '@stripe/react-stripe-js';

const handleSubmit = async (event) => {
  event.preventDefault();
  const stripe = useStripe();
  const elements = useElements();

  // Stripe handles tokenization securely
  const {error, paymentMethod} = await stripe.createPaymentMethod({
    type: 'card',
    card: elements.getElement(CardElement),
  });
};
```

### Payment Logging Rules

**Safe to log:**
- Payment intent IDs (pi_xxx)
- Setup intent IDs (seti_xxx)
- Customer IDs (cus_xxx)
- Payment status/state

**NEVER log:**
- Card numbers
- CVV/CVC codes
- Full card details
- Bank account numbers
- Personal identification numbers

**Example:**
```javascript
// ✅ Safe logging
console.log('Payment intent created:', paymentIntent.id);
console.log('Payment status:', paymentIntent.status);

// ❌ Dangerous logging
console.log('Payment intent:', paymentIntent); // May contain sensitive data
```

**Reference:** [src/features/payments/__tests__/TESTING_GUIDE.md](../src/features/payments/__tests__/TESTING_GUIDE.md)

---

## Input Validation & Sanitisation

### Form Input Validation

**Always validate:**
1. **Client-side** - For UX and immediate feedback
2. **Server-side** - For security (never trust client)

**Email Validation:**
```javascript
// ✅ Use proper email regex
const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

function validateEmail(email) {
  return EMAIL_REGEX.test(email);
}
```

**Phone Number Validation:**
```javascript
// ✅ Based on authentication module pattern
function validatePhoneNumber(phone) {
  // Remove spaces and special characters
  const cleaned = phone.replace(/\D/g, '');
  return cleaned.length >= 10 && cleaned.length <= 15;
}
```

### XSS Prevention

**DO:**
```javascript
// ✅ React automatically escapes by default
<div>{userInput}</div>

// ✅ Use DOMPurify for HTML content
import DOMPurify from 'dompurify';
const cleanHTML = DOMPurify.sanitize(userHTML);
<div dangerouslySetInnerHTML={{ __html: cleanHTML }} />
```

**DON'T:**
```javascript
// ❌ Never use dangerouslySetInnerHTML without sanitisation
<div dangerouslySetInnerHTML={{ __html: userInput }} />

// ❌ Never use eval or Function constructor
eval(userInput);
new Function(userInput)();
```

### URL Validation

```javascript
// ✅ Validate external URLs
function isValidURL(url) {
  try {
    const parsed = new URL(url);
    // Only allow http/https
    return ['http:', 'https:'].includes(parsed.protocol);
  } catch {
    return false;
  }
}
```

---

## API Security

### CORS & Domain Validation

**Allowed Domains (from constants.js):**
```javascript
export const SENTRY_TARGETS = [
  'localhost',
  'matchday.live',
  'robinstv.bcfc.co.uk',
  // ... other verified domains
];
```

### Request Headers Security

**Always include:**
```javascript
headers: {
  'Accept': 'application/json',
  'Content-Type': 'application/json',
  'Authorization': `Bearer ${token}`, // Only if authenticated
}
```

**Never include:**
- Sensitive data in headers (use body instead)
- Tokens in custom headers (use Authorization header)

### Error Handling

**DO:**
```javascript
// ✅ Generic error messages to users
catch (error) {
  setError('An error occurred. Please try again.');
  // Log full error for debugging (server-side only)
}
```

**DON'T:**
```javascript
// ❌ Never expose internal errors
catch (error) {
  setError(error.message); // May contain sensitive info
  setError(error.stack); // Exposes code structure
}
```

---

## Session Management

### Session Storage Rules

**Store in localStorage:**
- Access tokens
- Refresh tokens
- User ID (not sensitive data)
- Preferences

**NEVER store:**
- Passwords (even hashed)
- Credit card numbers
- Social security numbers
- Full authentication responses

### Session Timeout

```javascript
// ✅ Implement session timeout
const SESSION_TIMEOUT = 30 * 60 * 1000; // 30 minutes

function checkSessionTimeout() {
  const lastActivity = localStorage.getItem('lastActivity');
  if (Date.now() - lastActivity > SESSION_TIMEOUT) {
    // Clear session and redirect to login
    logout();
  }
}
```

### Logout Security

```javascript
// ✅ Complete logout pattern
function logout() {
  // 1. Clear all tokens
  localStorage.removeItem('accessToken');
  localStorage.removeItem('idToken');
  localStorage.removeItem('refreshToken');

  // 2. Clear user data
  localStorage.removeItem('user');

  // 3. Invalidate session on server
  await revokeToken();

  // 4. Redirect to login
  navigate('/login');
}
```

**Reference:** `src/features/authentication/Logout.jsx`

---

## Security Testing

### Pre-Deployment Security Checklist

- [ ] All secrets removed from code
- [ ] Input validation on all forms
- [ ] Authentication required for protected routes
- [ ] Tokens stored securely
- [ ] No console.log of sensitive data
- [ ] Error messages don't expose internals
- [ ] HTTPS enforced in production
- [ ] CORS configured correctly
- [ ] CSP headers configured
- [ ] Dependencies updated (no known vulnerabilities)

### Security Testing Requirements

**Every PR must verify:**
1. No hardcoded secrets or credentials
2. Authentication flows tested
3. Input validation tested with:
   - SQL injection attempts
   - XSS payloads
   - Oversized inputs
   - Special characters
4. Error handling doesn't leak info

### npm audit

```bash
# Run before every commit
npm audit

# Fix vulnerabilities
npm audit fix

# Review and update dependencies monthly
npm outdated
```

---

## Common Security Vulnerabilities

### 1. Hardcoded Credentials

**❌ NEVER:**
```javascript
const API_KEY = 'sk_live_abc123...';
const DB_PASSWORD = 'password123';
```

**✅ ALWAYS:**
```javascript
const API_KEY = import.meta.env.VITE_API_KEY;
// Or use server-side secrets management
```

### 2. Console Logging Sensitive Data

**Current issue found:**
```javascript
// ❌ Found in CancelSubscriptionForm.jsx:59
console.log(err.message); // May contain sensitive data
```

**✅ Fix:**
```javascript
// Log sanitised error
console.error('Subscription cancellation failed:', err.code);
// Send full error to monitoring service (server-side)
Sentry.captureException(err);
```

### 3. localStorage XSS

**Risk:** If XSS vulnerability exists, attacker can read localStorage

**Mitigation:**
- Sanitise all user input
- Use Content Security Policy
- Consider httpOnly cookies for critical tokens (requires backend change)

### 4. CSRF Protection

**Current protection:** Cognito tokens + CORS

**Additional measures:**
- Verify Origin header
- Use custom request headers
- Implement CSRF tokens for state-changing operations

---

## Incident Response

### If Security Issue Discovered:

1. **Immediate Actions:**
   - Document the vulnerability
   - Assess severity and impact
   - Notify team lead immediately

2. **Critical Issues (passwords, payment data exposed):**
   - Rotate all affected credentials immediately
   - Notify affected users
   - Prepare incident report

3. **Post-Incident:**
   - Create tickets to prevent similar issues
   - Update security rules
   - Conduct security review

---

## Security Resources

### Internal Documentation
- [03_SECURITY.md](../docs/03_SECURITY.md) - Comprehensive security architecture
- [09_PAYMENT_SUBSCRIPTION_SYSTEM.md](../docs/09_PAYMENT_SUBSCRIPTION_SYSTEM.md) - Payment security
- [08_AUTHENTICATION_FLOWS.md](../docs/08_AUTHENTICATION_FLOWS.md) - Auth security

### External Resources
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Stripe Security Best Practices](https://stripe.com/docs/security/guide)
- [AWS Cognito Security](https://docs.aws.amazon.com/cognito/latest/developerguide/security.html)

### Reference Implementations
- **Gold Standard:** `src/features/authentication/` - Secure auth patterns
- **Payment Security:** `src/features/payments/` - PCI-compliant patterns
- **Error Handling:** `src/ui/ErrorFallback.jsx` - Secure error display

---

## Quick Security Checklist

Before committing code, verify:

- [ ] No secrets in code or config files
- [ ] Input validation implemented
- [ ] Authentication checked for protected features
- [ ] No sensitive data in logs
- [ ] Error messages are generic to users
- [ ] Tokens stored and used securely
- [ ] External URLs validated
- [ ] SQL/XSS injection prevented
- [ ] Dependencies have no known vulnerabilities
- [ ] Tests cover security scenarios