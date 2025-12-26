# Error Handling Rules

**Version:** 1.0.0
**Last Updated:** 07/10/2025
**Purpose:** Error handling patterns and best practices for MDL Fan Dev

## Core Error Handling Principles

1. **User-friendly messages** - Never expose technical details to users
2. **Graceful degradation** - App should never completely break
3. **Error boundaries** - Catch and contain component errors
4. **Consistent patterns** - Same error handling across features
5. **Actionable feedback** - Tell users what they can do next

---

## Error Boundary Pattern

### Implementation

**Reference:** `src/ui/ErrorFallback.jsx`

```javascript
import { ErrorBoundary } from 'react-error-boundary';
import ErrorFallback from './ui/ErrorFallback';

function App() {
  return (
    <ErrorBoundary
      FallbackComponent={ErrorFallback}
      onError={(error, errorInfo) => {
        // Log to monitoring service
        console.error('Error boundary caught:', error, errorInfo);
        // Send to Sentry or similar
      }}
      onReset={() => {
        // Reset app state
        window.location.href = '/';
      }}
    >
      <YourApp />
    </ErrorBoundary>
  );
}
```

### Error Fallback UI

**DO:**
```javascript
// ✅ User-friendly error message with retry
function ErrorFallback({ error, resetErrorBoundary }) {
  return (
    <Container>
      <Icon>⚠️</Icon>
      <Heading>An error has occurred</Heading>
      <Message>
        We're sorry, an unexpected error has occurred. If the issue
        continues to persist, please get in touch with our support team
        who will be able to assist you in resolving the issue.
      </Message>
      <Button onClick={resetErrorBoundary}>Retry</Button>
      <SupportLink href="mailto:support@matchday.live">
        Contact Support
      </SupportLink>
    </Container>
  );
}
```

**DON'T:**
```javascript
// ❌ Exposing technical details to users
function ErrorFallback({ error }) {
  return (
    <div>
      <h1>Error</h1>
      <pre>{error.stack}</pre> {/* Never do this! */}
      <p>{error.message}</p> {/* May contain sensitive info */}
    </div>
  );
}
```

### Multiple Error Boundaries

**Strategy:** Wrap features separately to contain errors

```javascript
// ✅ Feature-level error boundaries
function App() {
  return (
    <AppErrorBoundary>
      <Header />

      <ErrorBoundary FallbackComponent={PlayerErrorFallback}>
        <VideoPlayer />
      </ErrorBoundary>

      <ErrorBoundary FallbackComponent={ContentErrorFallback}>
        <ContentSection />
      </ErrorBoundary>

      <Footer />
    </AppErrorBoundary>
  );
}
```

**Reference:** [docs/05_ERROR_HANDLING_MONITORING.md](../docs/05_ERROR_HANDLING_MONITORING.md)

---

## API Error Handling

### Service Layer Errors

**Current pattern (apiFetch.js):**
```javascript
try {
  res = await axios({ ...config });
  return res.data;
} catch (err) {
  throw new Error(err.response.data);
}
```

**Enhanced pattern:**
```javascript
try {
  res = await axios({ ...config });
  return res.data;
} catch (err) {
  // Extract meaningful error message
  const message =
    err.response?.data?.message ||
    err.response?.data ||
    err.message ||
    'An unexpected error occurred';

  // Log full error for debugging
  console.error('API Error:', {
    endpoint: config.url,
    status: err.response?.status,
    message,
  });

  // Throw user-friendly error
  throw new Error(message);
}
```

### HTTP Status Code Handling

**Pattern:**
```javascript
export async function getData({ method, path, body, headers }) {
  try {
    const response = await axios({ method, url: path, data: body, headers });
    return response.data;
  } catch (error) {
    const status = error.response?.status;

    // Handle specific status codes
    switch (status) {
      case 400:
        throw new Error('Invalid request. Please check your input.');

      case 401:
        // Clear tokens and redirect
        localStorage.removeItem('accessToken');
        localStorage.removeItem('idToken');
        window.location.href = '/login';
        throw new Error('Session expired. Please log in again.');

      case 403:
        throw new Error('You do not have permission to perform this action.');

      case 404:
        throw new Error('The requested resource was not found.');

      case 409:
        throw new Error('This action conflicts with existing data.');

      case 429:
        throw new Error('Too many requests. Please try again later.');

      case 500:
      case 502:
      case 503:
        throw new Error('Server error. Please try again later.');

      default:
        throw new Error(
          error.response?.data?.message ||
          'An unexpected error occurred. Please try again.'
        );
    }
  }
}
```

### React Query Error Handling

**Component-level:**
```javascript
function FixtureList() {
  const { data, error, isError, isLoading } = useQuery({
    queryKey: ['fixtures'],
    queryFn: fetchFixtures,
    retry: 3, // Retry failed requests 3 times
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
  });

  if (isLoading) return <Skeleton />;

  if (isError) {
    return (
      <ErrorMessage>
        <IconAlert />
        <p>Unable to load fixtures</p>
        <Button onClick={() => refetch()}>Try Again</Button>
      </ErrorMessage>
    );
  }

  return <FixtureCards data={data} />;
}
```

**Global error handling:**
```javascript
// In query client setup
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 3,
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
      onError: (error) => {
        console.error('Query error:', error);
        // Show toast notification
        toast.error('Failed to load data');
      },
    },
    mutations: {
      retry: 1, // Mutations typically shouldn't retry automatically
      onError: (error) => {
        console.error('Mutation error:', error);
        toast.error(error.message || 'Action failed. Please try again.');
      },
    },
  },
});
```

**Reference:** [.claude/api-integration-rules.md](./api-integration-rules.md)

---

## Form Validation & Errors

### Validation Pattern

**DO:**
```javascript
function LoginForm() {
  const [errors, setErrors] = useState({});

  const validate = (values) => {
    const errors = {};

    if (!values.email) {
      errors.email = 'Email is required';
    } else if (!/\S+@\S+\.\S+/.test(values.email)) {
      errors.email = 'Email is invalid';
    }

    if (!values.password) {
      errors.password = 'Password is required';
    } else if (values.password.length < 8) {
      errors.password = 'Password must be at least 8 characters';
    }

    return errors;
  };

  const handleSubmit = async (values) => {
    const validationErrors = validate(values);

    if (Object.keys(validationErrors).length > 0) {
      setErrors(validationErrors);
      return;
    }

    try {
      await login(values);
    } catch (error) {
      setErrors({ submit: error.message });
    }
  };

  return (
    <Form onSubmit={handleSubmit}>
      <Input
        name="email"
        error={errors.email}
        helperText={errors.email}
      />
      <Input
        name="password"
        error={errors.password}
        helperText={errors.password}
      />
      {errors.submit && <ErrorMessage>{errors.submit}</ErrorMessage>}
      <Button type="submit">Login</Button>
    </Form>
  );
}
```

### Field-Level Error Display

**Pattern from authentication:**
```javascript
<FormInputBox>
  <Input
    type="email"
    value={email}
    onChange={(e) => setEmail(e.target.value)}
    placeholder="Email"
    $error={!!errors.email}
  />
  {errors.email && (
    <FormNotification variation="error">
      {errors.email}
    </FormNotification>
  )}
</FormInputBox>
```

**Reference:** `src/features/authentication/LoginForm.jsx`

---

## User-Friendly Error Messages

### Message Guidelines

**DO:**
- Explain what went wrong in simple terms
- Tell users what they can do next
- Provide support contact for persistent issues
- Use empathetic language ("We're sorry...")

**DON'T:**
- Show stack traces or technical details
- Use jargon (HTTP 500, null pointer, etc.)
- Blame the user ("You did X wrong")
- Leave users with no next action

### Error Message Examples

**Good Examples:**
```javascript
const errorMessages = {
  // Network errors
  networkError: "Unable to connect. Please check your internet connection and try again.",

  // Authentication errors
  invalidCredentials: "Email or password is incorrect. Please try again.",
  sessionExpired: "Your session has expired. Please log in again.",

  // Payment errors
  paymentFailed: "Payment could not be processed. Please check your payment details and try again.",
  cardDeclined: "Your card was declined. Please use a different payment method or contact your bank.",

  // Resource errors
  notFound: "The content you're looking for is no longer available.",
  accessDenied: "You don't have access to this content. Please check your subscription.",

  // General errors
  genericError: "Something went wrong. Please try again later or contact support if the problem persists.",
};
```

**Bad Examples:**
```javascript
// ❌ Don't use these
const badMessages = {
  "Error: Cannot read property 'data' of undefined",
  "HTTP 500 Internal Server Error",
  "Uncaught TypeError at line 42",
  "null is not an object",
  "Failed to fetch",
};
```

---

## Logging Patterns

### What to Log

**DO log:**
```javascript
// ✅ Error context for debugging
console.error('Failed to fetch fixtures', {
  endpoint: '/fixtures',
  userId: user.id,
  error: error.message,
  timestamp: new Date().toISOString(),
});

// ✅ User actions for analytics
console.log('User clicked subscribe button', {
  planId: selectedPlan.id,
  userId: user.id,
});

// ✅ Performance metrics
console.log('Fixture page loaded', {
  loadTime: performance.now() - startTime,
  fixtureCount: fixtures.length,
});
```

**DON'T log:**
```javascript
// ❌ Sensitive data (see security-rules.md)
console.log('User password:', password);
console.log('Credit card:', cardNumber);
console.log('Auth token:', token);

// ❌ Entire error objects in production
console.log(error); // May contain sensitive data

// ❌ console.log in production code (use proper logging)
console.log('Debug info'); // Should be removed before deployment
```

### Logging Levels

```javascript
// Development
console.log('Info message');
console.warn('Warning message');
console.error('Error message');

// Production (use monitoring service)
if (process.env.NODE_ENV === 'production') {
  Sentry.captureMessage('Info', 'info');
  Sentry.captureMessage('Warning', 'warning');
  Sentry.captureException(error);
}
```

### Sentry Integration

**Configuration (from constants.js):**
```javascript
export const SENTRY_DSN = "https://...";
export const SENTRY_TRACE_SAMPLE_RATE = process.env.NODE_ENV !== "production" ? 1.0 : 0.5;
```

**Usage:**
```javascript
import * as Sentry from '@sentry/react';

// Capture exceptions
try {
  riskyOperation();
} catch (error) {
  Sentry.captureException(error, {
    tags: {
      feature: 'payments',
      action: 'checkout',
    },
    extra: {
      userId: user.id,
      planId: selectedPlan.id,
    },
  });

  // Show user-friendly message
  setError('Payment processing failed. Please try again.');
}

// Capture messages
Sentry.captureMessage('Unusual user behaviour detected', 'warning');
```

**Reference:** [docs/05_ERROR_HANDLING_MONITORING.md](../docs/05_ERROR_HANDLING_MONITORING.md)

---

## Async Error Handling

### Promise Rejection

**DO:**
```javascript
// ✅ Always catch promise rejections
fetchData()
  .then(data => processData(data))
  .catch(error => {
    console.error('Failed to fetch data:', error);
    setError('Unable to load data');
  });

// ✅ Use try-catch with async/await
async function loadData() {
  try {
    const data = await fetchData();
    processData(data);
  } catch (error) {
    console.error('Failed to fetch data:', error);
    setError('Unable to load data');
  }
}
```

**DON'T:**
```javascript
// ❌ Unhandled promise rejection
fetchData()
  .then(data => processData(data));

// ❌ No error handling in async function
async function loadData() {
  const data = await fetchData(); // Can throw!
  processData(data);
}
```

### Event Handler Errors

```javascript
// ✅ Wrap event handlers in try-catch
const handleSubmit = async (event) => {
  event.preventDefault();

  try {
    const result = await submitForm(data);
    onSuccess(result);
  } catch (error) {
    console.error('Form submission failed:', error);
    setError('Failed to submit form. Please try again.');
  }
};
```

---

## Loading & Empty States

### Loading States

**DO:**
```javascript
// ✅ Provide feedback during loading
function Component() {
  const { data, isLoading } = useQuery(/* ... */);

  if (isLoading) {
    return <Skeleton count={3} />;
  }

  return <DataDisplay data={data} />;
}
```

### Empty States

**DO:**
```javascript
// ✅ Handle empty data gracefully
function FixtureList() {
  const { data } = useQuery(/* ... */);

  if (!data || data.length === 0) {
    return (
      <EmptyState>
        <Icon>📅</Icon>
        <Heading>No fixtures available</Heading>
        <Message>
          There are no upcoming fixtures at the moment.
          Check back later for new matches.
        </Message>
      </EmptyState>
    );
  }

  return <FixtureCards data={data} />;
}
```

### Error State

**Pattern:**
```javascript
function Component() {
  const { data, isLoading, isError, error } = useQuery(/* ... */);

  if (isLoading) return <LoadingState />;

  if (isError) {
    return (
      <ErrorState>
        <IconAlert />
        <Heading>Unable to load content</Heading>
        <Message>{error.message}</Message>
        <Button onClick={() => refetch()}>Try Again</Button>
      </ErrorState>
    );
  }

  if (!data || data.length === 0) {
    return <EmptyState />;
  }

  return <DataDisplay data={data} />;
}
```

**Reference:** `src/ui/ErrorData.jsx`

---

## Testing Error Scenarios

### Error Boundary Tests

```javascript
describe('ErrorBoundary', () => {
  it('catches and displays component errors', () => {
    const ThrowError = () => {
      throw new Error('Test error');
    };

    render(
      <ErrorBoundary FallbackComponent={ErrorFallback}>
        <ThrowError />
      </ErrorBoundary>
    );

    expect(screen.getByText(/an error has occurred/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument();
  });
});
```

### API Error Tests

```javascript
import { getData } from './apiFetch';

jest.mock('./apiFetch');

describe('useFixtures', () => {
  it('handles API errors', async () => {
    getData.mockRejectedValueOnce(new Error('Network error'));

    const { result } = renderHook(() => useFixtures(), {
      wrapper: TestQueryWrapper,
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error.message).toBe('Network error');
  });

  it('retries failed requests', async () => {
    getData
      .mockRejectedValueOnce(new Error('Temporary error'))
      .mockResolvedValueOnce(mockData);

    const { result } = renderHook(() => useFixtures(), {
      wrapper: TestQueryWrapper,
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(getData).toHaveBeenCalledTimes(2);
  });
});
```

### Form Validation Tests

```javascript
describe('LoginForm validation', () => {
  it('shows error for invalid email', async () => {
    render(<LoginForm />);

    const emailInput = screen.getByPlaceholderText(/email/i);
    await userEvent.type(emailInput, 'invalid-email');

    const submitButton = screen.getByRole('button', { name: /login/i });
    await userEvent.click(submitButton);

    expect(screen.getByText(/email is invalid/i)).toBeInTheDocument();
  });
});
```

**Reference:** [.claude/testing-rules.md](./testing-rules.md)

---

## Error Recovery Strategies

### Retry Logic

**Pattern:**
```javascript
async function fetchWithRetry(fn, retries = 3, delay = 1000) {
  for (let i = 0; i < retries; i++) {
    try {
      return await fn();
    } catch (error) {
      if (i === retries - 1) throw error;

      // Exponential backoff
      await new Promise(resolve => setTimeout(resolve, delay * Math.pow(2, i)));
    }
  }
}

// Usage
const data = await fetchWithRetry(() => getData({ path: '/fixtures' }));
```

### Fallback Data

```javascript
function useFixturesWithFallback() {
  const { data, error } = useQuery({
    queryKey: ['fixtures'],
    queryFn: fetchFixtures,
  });

  // Return cached data if available, even if query failed
  if (error && data) {
    console.warn('Using cached data due to error:', error);
    return { data, isStale: true };
  }

  return { data, isStale: false };
}
```

### Optimistic Updates with Rollback

```javascript
function useUpdateFixture() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: updateFixture,
    onMutate: async (newData) => {
      // Optimistically update
      const previous = queryClient.getQueryData(['fixture', newData.id]);
      queryClient.setQueryData(['fixture', newData.id], newData);
      return { previous };
    },
    onError: (error, variables, context) => {
      // Rollback on error
      queryClient.setQueryData(
        ['fixture', variables.id],
        context.previous
      );

      // Show error to user
      toast.error('Failed to update fixture');
    },
  });
}
```

---

## Error Handling Checklist

Before deploying:

- [ ] Error boundaries implemented at app and feature level
- [ ] User-friendly error messages (no technical details)
- [ ] All async operations have error handling
- [ ] Form validation with field-level errors
- [ ] Loading states for all data fetching
- [ ] Empty states for zero-data scenarios
- [ ] Retry logic for transient failures
- [ ] Error logging configured (Sentry/similar)
- [ ] 401/403 errors redirect appropriately
- [ ] No sensitive data in error logs
- [ ] Error scenarios covered in tests
- [ ] Support contact provided for persistent errors

---

## Common Error Handling Mistakes

### 1. Exposing Technical Details

**❌ Bad:**
```javascript
<div>{error.stack}</div>
<div>{error.message}</div> // May contain DB queries, file paths, etc.
```

**✅ Good:**
```javascript
<div>An unexpected error occurred. Please try again.</div>
```

### 2. Swallowing Errors

**❌ Bad:**
```javascript
try {
  await riskyOperation();
} catch (error) {
  // Silent failure - user has no idea something went wrong
}
```

**✅ Good:**
```javascript
try {
  await riskyOperation();
} catch (error) {
  console.error('Operation failed:', error);
  setError('Unable to complete operation');
  Sentry.captureException(error);
}
```

### 3. No Retry/Fallback Strategy

**❌ Bad:**
```javascript
// Single attempt, fails permanently on transient network issues
const data = await fetchData();
```

**✅ Good:**
```javascript
// React Query handles retries automatically
const { data } = useQuery({
  queryKey: ['data'],
  queryFn: fetchData,
  retry: 3,
});
```

---

## Resources

### Internal Documentation
- [05_ERROR_HANDLING_MONITORING.md](../docs/05_ERROR_HANDLING_MONITORING.md)
- [35_DEBUGGING_GUIDE.md](../docs/35_DEBUGGING_GUIDE.md)

### Reference Implementations
- **Error Boundary:** `src/ui/ErrorFallback.jsx`
- **Error Display:** `src/ui/ErrorData.jsx`
- **Form Validation:** `src/features/authentication/LoginForm.jsx`
- **API Errors:** `src/services/apiFetch.js`

### External Resources
- [React Error Boundaries](https://react.dev/reference/react/Component#catching-rendering-errors-with-an-error-boundary)
- [react-error-boundary](https://github.com/bvaughn/react-error-boundary)
- [Sentry Documentation](https://docs.sentry.io/)