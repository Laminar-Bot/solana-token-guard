# API Integration Rules

**Version:** 1.0.0
**Last Updated:** 07/10/2025
**Purpose:** API integration patterns and best practices for MDL Fan Dev

## Core API Principles

1. **Single source of truth** - All API calls through service layer
2. **React Query for server state** - Never useState for API data
3. **Consistent error handling** - Unified error patterns
4. **Token management** - Automatic token injection
5. **Request deduplication** - React Query handles this automatically

---

## Architecture Overview

```
Component → Custom Hook → React Query → Service Layer → API
                                           (apiFetch.js)
```

**Key Files:**
- `src/services/apiFetch.js` - Base API client
- `src/services/apiAuth.js` - Authentication endpoints
- `src/services/apiStripe.js` - Payment endpoints
- `src/services/apiChannel.js` - Channel/content endpoints

**Reference:** [docs/12_API_INTEGRATION_PATTERNS.md](../docs/12_API_INTEGRATION_PATTERNS.md)

---

## Service Layer Pattern

### Base API Client (apiFetch.js)

**DO use the central API client:**
```javascript
import { getData } from '../services/apiFetch';

// ✅ Correct pattern
const data = await getData({
  method: 'get',
  path: '/channels',
  headers: {},
});
```

**DON'T bypass the service layer:**
```javascript
// ❌ Never make direct fetch calls
fetch('https://api.example.com/data')
  .then(res => res.json());

// ❌ Never create separate axios instances
const customAxios = axios.create({...});
```

### API Client Features

**Automatic token injection:**
```javascript
// Token automatically added from localStorage
let token = localStorage.getItem("accessToken");

if (token) {
  config.headers = {
    Authorization: `Bearer ${token}`,
    ...config.headers,
  };
}
```

**Standard headers:**
```javascript
config.headers = {
  ...headers,
  Accept: "application/json",
  "Content-type": "application/json",
};
```

---

## React Query Integration

### When to Use React Query

**✅ USE React Query for:**
- GET requests (data fetching)
- POST/PUT/DELETE with server state
- Data that needs caching
- Data that multiple components need
- Data that needs automatic refetching
- Background data updates

**❌ DON'T use React Query for:**
- Form submission without refetch needs
- One-off actions (downloads, exports)
- Local-only state
- Data that never changes

### Query Patterns

**Basic Query:**
```javascript
import { useQuery } from '@tanstack/react-query';
import { getData } from '../services/apiFetch';

function useChannelData() {
  return useQuery({
    queryKey: ['channel'],
    queryFn: () => getData({ method: 'get', path: '/channel' }),
    staleTime: 5 * 60 * 1000, // 5 minutes (from constants.js)
  });
}
```

**Query with Parameters:**
```javascript
function useFixture(fixtureId) {
  return useQuery({
    queryKey: ['fixture', fixtureId],
    queryFn: () => getData({
      method: 'get',
      path: `/fixtures/${fixtureId}`
    }),
    enabled: !!fixtureId, // Only run if fixtureId exists
  });
}
```

**Query with Dependencies:**
```javascript
function useUserSubscription(userId) {
  return useQuery({
    queryKey: ['subscription', userId],
    queryFn: () => getData({
      method: 'get',
      path: `/subscriptions/${userId}`
    }),
    enabled: !!userId,
    staleTime: 5 * 60 * 1000,
    refetchOnWindowFocus: false, // Don't refetch on window focus
  });
}
```

### Mutation Patterns

**Basic Mutation:**
```javascript
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { getData } from '../services/apiFetch';

function useUpdateProfile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (profileData) => getData({
      method: 'put',
      path: '/profile',
      body: profileData,
    }),
    onSuccess: () => {
      // Invalidate and refetch
      queryClient.invalidateQueries({ queryKey: ['profile'] });
    },
  });
}
```

**Mutation with Optimistic Updates:**
```javascript
function useToggleFavourite() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (streamId) => getData({
      method: 'post',
      path: `/streams/${streamId}/favourite`,
    }),
    onMutate: async (streamId) => {
      // Cancel outgoing queries
      await queryClient.cancelQueries({ queryKey: ['stream', streamId] });

      // Snapshot previous value
      const previousStream = queryClient.getQueryData(['stream', streamId]);

      // Optimistically update
      queryClient.setQueryData(['stream', streamId], (old) => ({
        ...old,
        isFavourite: !old.isFavourite,
      }));

      return { previousStream };
    },
    onError: (err, streamId, context) => {
      // Rollback on error
      queryClient.setQueryData(
        ['stream', streamId],
        context.previousStream
      );
    },
    onSettled: (data, error, streamId) => {
      // Always refetch after error or success
      queryClient.invalidateQueries({ queryKey: ['stream', streamId] });
    },
  });
}
```

### Query Key Conventions

**Pattern:** `[resource, ...identifiers, ...filters]`

**Examples:**
```javascript
// List queries
['channels']
['fixtures']
['playlists']

// Detail queries
['channel', channelId]
['fixture', fixtureId]
['stream', streamId]

// Filtered queries
['fixtures', { status: 'live' }]
['playlists', { channelId, type: 'upcoming' }]

// User-specific queries
['subscription', userId]
['billing-history', userId]
['payment-methods', customerId]
```

### Stale Time Configuration

**From constants.js:**
```javascript
export const QUERY_STALE_TIME_SEC = 10; // 10 seconds
```

**Usage:**
```javascript
// Frequently changing data (live scores, etc.)
staleTime: 10 * 1000, // 10 seconds

// Moderate changes (fixtures, playlists)
staleTime: 5 * 60 * 1000, // 5 minutes

// Rarely changing (channel config, plans)
staleTime: 15 * 60 * 1000, // 15 minutes

// Static data (terms, privacy policy)
staleTime: Infinity,
```

---

## Error Handling

### Service Layer Error Pattern

**Current pattern (apiFetch.js):**
```javascript
try {
  res = await axios({ ...config });
  return res.data;
} catch (err) {
  throw new Error(err.response.data);
}
```

### React Query Error Handling

**Component level:**
```javascript
function MyComponent() {
  const { data, error, isError, isLoading } = useQuery({
    queryKey: ['data'],
    queryFn: fetchData,
  });

  if (isLoading) return <Spinner />;

  if (isError) {
    return <ErrorMessage error={error.message} />;
  }

  return <DataDisplay data={data} />;
}
```

**Global error handling:**
```javascript
// In App.jsx or query client setup
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      onError: (error) => {
        console.error('Query error:', error);
        // Show toast notification
        toast.error('Failed to load data');
      },
    },
    mutations: {
      onError: (error) => {
        console.error('Mutation error:', error);
        toast.error('Action failed. Please try again.');
      },
    },
  },
});
```

### Authentication Errors

**401 Handling:**
```javascript
// In apiFetch.js or interceptor
if (error.response?.status === 401) {
  // Clear tokens
  localStorage.removeItem('accessToken');
  localStorage.removeItem('idToken');

  // Redirect to login
  window.location.href = '/login';

  throw new Error('Session expired. Please log in again.');
}
```

**403 Handling:**
```javascript
if (error.response?.status === 403) {
  throw new Error('You do not have permission to perform this action.');
}
```

---

## Caching Strategy

### Cache Invalidation

**Invalidate after mutations:**
```javascript
const mutation = useMutation({
  mutationFn: updateData,
  onSuccess: () => {
    // Invalidate specific query
    queryClient.invalidateQueries({ queryKey: ['data', id] });

    // Invalidate all queries starting with 'data'
    queryClient.invalidateQueries({ queryKey: ['data'] });
  },
});
```

**Invalidate on navigation:**
```javascript
// In useEffect or on component mount
useEffect(() => {
  queryClient.invalidateQueries({ queryKey: ['fresh-data'] });
}, [location.pathname]);
```

### Cache Pre-population

```javascript
// Pre-populate cache with list data
const { data: list } = useQuery({
  queryKey: ['items'],
  queryFn: fetchItems,
  onSuccess: (items) => {
    // Pre-populate individual item caches
    items.forEach(item => {
      queryClient.setQueryData(['item', item.id], item);
    });
  },
});
```

### Persistent Cache

```javascript
// For data that should persist across sessions
import { createSyncStoragePersister } from '@tanstack/query-sync-storage-persister';
import { persistQueryClient } from '@tanstack/react-query-persist-client';

const persister = createSyncStoragePersister({
  storage: window.localStorage,
});

persistQueryClient({
  queryClient,
  persister,
  maxAge: 24 * 60 * 60 * 1000, // 24 hours
});
```

---

## Loading States

### Skeleton Loading Pattern

**DO:**
```javascript
function FixtureList() {
  const { data, isLoading } = useQuery({
    queryKey: ['fixtures'],
    queryFn: fetchFixtures,
  });

  // ✅ Show skeleton loader
  if (isLoading) {
    return <FixtureSkeleton count={3} />;
  }

  return <FixtureCards data={data} />;
}
```

**DON'T:**
```javascript
// ❌ Generic spinner everywhere
if (isLoading) return <Spinner />;

// ❌ No loading state
return data?.map(...) // Breaks on initial load
```

### Suspense Integration

```javascript
import { Suspense } from 'react';

// Wrap components that use suspense-enabled queries
<Suspense fallback={<LoadingSkeleton />}>
  <DataComponent />
</Suspense>
```

---

## Request Deduplication

React Query automatically deduplicates requests with the same query key.

**Example:**
```javascript
// These three components mount simultaneously
function Component1() {
  const { data } = useChannelData(); // Makes request
}

function Component2() {
  const { data } = useChannelData(); // Reuses first request
}

function Component3() {
  const { data } = useChannelData(); // Reuses first request
}

// Only ONE network request is made
```

---

## Polling & Real-time Updates

### Polling Pattern

```javascript
function useliveScroes(matchId) {
  return useQuery({
    queryKey: ['live-scores', matchId],
    queryFn: () => fetchScores(matchId),
    refetchInterval: 10000, // Poll every 10 seconds
    refetchIntervalInBackground: true, // Poll when tab is not visible
    enabled: !!matchId,
  });
}
```

### WebSocket Integration

```javascript
// For real-time updates, use WebSocket + React Query
function useStreamStatus(streamId) {
  const queryClient = useQueryClient();

  useEffect(() => {
    const ws = new WebSocket(WEB_SOCKET_URL);

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);

      // Update React Query cache
      queryClient.setQueryData(['stream', streamId], (old) => ({
        ...old,
        status: data.status,
      }));
    };

    return () => ws.close();
  }, [streamId]);

  return useQuery({
    queryKey: ['stream', streamId],
    queryFn: () => fetchStream(streamId),
  });
}
```

**Reference:** [docs/21_WEBSOCKET_IMPLEMENTATION.md](../docs/21_WEBSOCKET_IMPLEMENTATION.md)

---

## Stripe API Integration

### Payment Intent Pattern

```javascript
function useCreatePaymentIntent() {
  return useMutation({
    mutationFn: async (productId) => {
      const response = await getData({
        method: 'post',
        path: '/stripe/create-payment-intent',
        body: { productId },
      });
      return response;
    },
    onSuccess: (data) => {
      // Navigate to checkout with client secret
      navigate(`/checkout?payment_intent=${data.clientSecret}`);
    },
  });
}
```

### Setup Intent Pattern (Payment Methods)

```javascript
function useCreateSetupIntent() {
  return useMutation({
    mutationFn: async () => {
      const response = await getData({
        method: 'post',
        path: '/stripe/create-setup-intent',
      });
      return response;
    },
  });
}
```

**Security Note:** Never log Stripe responses. See [security-rules.md](./security-rules.md)

**Reference:** [src/features/payments/__tests__/TESTING_GUIDE.md](../src/features/payments/__tests__/TESTING_GUIDE.md)

---

## Testing API Integration

### Mocking API Calls

**Pattern from authentication tests:**
```javascript
import { getData } from '../../services/apiFetch';

jest.mock('../../services/apiFetch', () => ({
  getData: jest.fn(),
}));

describe('useUserData', () => {
  it('fetches user data successfully', async () => {
    const mockData = { id: '1', name: 'Test User' };
    getData.mockResolvedValueOnce(mockData);

    const { result } = renderHook(() => useUserData(), {
      wrapper: TestQueryWrapper,
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockData);
  });

  it('handles errors', async () => {
    getData.mockRejectedValueOnce(new Error('API Error'));

    const { result } = renderHook(() => useUserData(), {
      wrapper: TestQueryWrapper,
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});
```

### React Query Test Wrapper

```javascript
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

export function TestQueryWrapper({ children }) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false, // Don't retry in tests
        cacheTime: 0, // No cache in tests
      },
    },
  });

  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
}
```

**Reference:** [.claude/testing-rules.md](./.claude/testing-rules.md)

---

## Common Patterns

### Dependent Queries

```javascript
function useStreamWithChannel(streamId) {
  // First query
  const { data: stream } = useQuery({
    queryKey: ['stream', streamId],
    queryFn: () => fetchStream(streamId),
  });

  // Second query depends on first
  const { data: channel } = useQuery({
    queryKey: ['channel', stream?.channelId],
    queryFn: () => fetchChannel(stream.channelId),
    enabled: !!stream?.channelId, // Only run if stream exists
  });

  return { stream, channel };
}
```

### Infinite Queries (Pagination)

```javascript
function useInfiniteFixtures() {
  return useInfiniteQuery({
    queryKey: ['fixtures', 'infinite'],
    queryFn: ({ pageParam = 0 }) =>
      getData({
        method: 'get',
        path: `/fixtures?page=${pageParam}`,
      }),
    getNextPageParam: (lastPage, pages) => {
      return lastPage.hasMore ? pages.length : undefined;
    },
  });
}
```

### Parallel Queries

```javascript
function useDashboardData() {
  const channels = useQuery({ queryKey: ['channels'], queryFn: fetchChannels });
  const fixtures = useQuery({ queryKey: ['fixtures'], queryFn: fetchFixtures });
  const playlists = useQuery({ queryKey: ['playlists'], queryFn: fetchPlaylists });

  // All queries run in parallel
  const isLoading = channels.isLoading || fixtures.isLoading || playlists.isLoading;

  return { channels: channels.data, fixtures: fixtures.data, playlists: playlists.data, isLoading };
}
```

---

## API Integration Checklist

Before implementing a new API integration:

- [ ] Use service layer (apiFetch.js) - never direct fetch
- [ ] Use React Query for server state - never useState
- [ ] Define proper query key structure
- [ ] Set appropriate staleTime
- [ ] Implement error handling
- [ ] Add loading states (skeleton preferred)
- [ ] Handle authentication errors (401/403)
- [ ] Invalidate cache after mutations
- [ ] Write tests with mocked API calls
- [ ] Document any new API patterns

---

## Performance Considerations

### Query Performance

**DO:**
```javascript
// ✅ Specific query keys for targeted invalidation
queryKey: ['fixture', fixtureId]

// ✅ Disable unnecessary refetches
refetchOnWindowFocus: false
refetchOnReconnect: false

// ✅ Appropriate stale time
staleTime: 5 * 60 * 1000 // 5 minutes
```

**DON'T:**
```javascript
// ❌ Generic query keys cause over-fetching
queryKey: ['data']

// ❌ Aggressive refetching
refetchInterval: 1000 // Every second

// ❌ No stale time (refetches constantly)
staleTime: 0
```

### Bundle Size

React Query adds ~13KB gzipped. This is acceptable for the features it provides.

**Reference:** [.claude/performance-rules.md](./.claude/performance-rules.md)

---

## Resources

### Internal Documentation
- [12_API_INTEGRATION_PATTERNS.md](../docs/12_API_INTEGRATION_PATTERNS.md)
- [11_STATE_MANAGEMENT_DATA_FLOW.md](../docs/11_STATE_MANAGEMENT_DATA_FLOW.md)
- [05_ERROR_HANDLING_MONITORING.md](../docs/05_ERROR_HANDLING_MONITORING.md)

### Reference Implementations
- **Authentication API:** `src/services/apiAuth.js`
- **Payment API:** `src/services/apiStripe.js`
- **Channel API:** `src/services/apiChannel.js`
- **Custom Hooks:** `src/features/*/use*.js`

### External Resources
- [TanStack Query Docs](https://tanstack.com/query/latest)
- [React Query Best Practices](https://tkdodo.eu/blog/practical-react-query)