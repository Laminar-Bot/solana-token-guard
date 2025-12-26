# Performance Rules

**Version:** 1.0.0
**Last Updated:** 07/10/2025
**Purpose:** Performance optimization guidelines and best practices for MDL Fan Dev

## Core Performance Principles

1. **Measure first, optimize second** - Use profiler and analytics
2. **Progressive enhancement** - Core functionality first, enhancements later
3. **Lazy load everything non-critical** - Reduce initial bundle size
4. **Cache aggressively** - React Query + browser caching
5. **Mobile-first performance** - Optimize for 3G networks

---

## Performance Budgets

### Core Web Vitals Targets

| Metric | Target | Maximum | Priority |
|--------|--------|---------|----------|
| **LCP** (Largest Contentful Paint) | <2.5s | 4.0s | Critical |
| **FID** (First Input Delay) | <100ms | 300ms | Critical |
| **CLS** (Cumulative Layout Shift) | <0.1 | 0.25 | High |
| **TTFB** (Time to First Byte) | <600ms | 1000ms | High |
| **FCP** (First Contentful Paint) | <1.8s | 3.0s | Medium |

### Bundle Size Budgets

```javascript
// Target bundle sizes (gzipped)
{
  "initial": "150KB",        // Critical path bundle
  "vendor-react": "50KB",    // React core
  "vendor-ui": "60KB",       // MUI + styled-components
  "vendor-player": "100KB",  // Bitmovin player
  "route-chunks": "30KB",    // Each route chunk
  "total": "500KB"           // Total JS downloaded
}
```

### Network Performance

- **3G Network:** Full page load <5s
- **4G Network:** Full page load <3s
- **WiFi:** Full page load <2s
- **API Response:** <500ms for data endpoints
- **Image Loading:** Progressive, lazy-loaded

**Reference:** [docs/06_PERFORMANCE_OPTIMIZATION.md](../docs/06_PERFORMANCE_OPTIMIZATION.md)

---

## Code Splitting & Lazy Loading

### Route-Based Code Splitting

**DO:**
```javascript
import { lazy, Suspense } from 'react';

// ✅ Lazy load route components
const Home = lazy(() => import('./pages/Home'));
const AccountManagement = lazy(() => import('./pages/AccountManagement'));
const Legal = lazy(() => import('./pages/Legal'));

function App() {
  return (
    <Suspense fallback={<SpinnerFullPage />}>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/account" element={<AccountManagement />} />
        <Route path="/legal" element={<Legal />} />
      </Routes>
    </Suspense>
  );
}
```

**DON'T:**
```javascript
// ❌ Import everything eagerly
import Home from './pages/Home';
import AccountManagement from './pages/AccountManagement';
import Legal from './pages/Legal';
```

### Component-Level Code Splitting

**When to lazy load components:**
- Heavy third-party components (video player, charts)
- Modal/dialog content
- Features behind feature flags
- Admin/low-usage features
- Components below the fold

**Example:**
```javascript
// ✅ Lazy load Bitmovin player
const BitmovinPlayer = lazy(() => import('./features/player/BitmovinPlayer'));

function StreamPage() {
  return (
    <Suspense fallback={<PlayerSkeleton />}>
      <BitmovinPlayer streamId={id} />
    </Suspense>
  );
}
```

### Vendor Code Splitting

**Configure in vite.config.js:**
```javascript
export default defineConfig({
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          // Core React libraries
          'vendor-react': [
            'react',
            'react-dom',
            'react-router-dom'
          ],

          // UI libraries
          'vendor-ui': [
            '@mui/material',
            'styled-components'
          ],

          // Utilities
          'vendor-utils': [
            '@tanstack/react-query',
            'date-fns'
          ],

          // Heavy libraries (load separately)
          'vendor-player': ['bitmovin-player'],
          'vendor-auth': ['amazon-cognito-identity-js'],
          'vendor-stripe': ['@stripe/react-stripe-js', '@stripe/stripe-js'],
        },
      },
    },
    chunkSizeWarningLimit: 500, // Warn at 500KB
  },
});
```

---

## React Performance Optimization

### Component Memoization

**Use React.memo for:**
- Pure presentational components
- Components that render often with same props
- Components with expensive render logic
- List items

**Example:**
```javascript
import { memo } from 'react';

// ✅ Memoize pure components
const FixtureCard = memo(function FixtureCard({ fixture }) {
  return (
    <Card>
      <h3>{fixture.title}</h3>
      <p>{fixture.date}</p>
    </Card>
  );
});
```

**DON'T memoize:**
```javascript
// ❌ Don't memoize if props always change
const UserGreeting = memo(({ currentTime }) => (
  <div>Hello! It's {currentTime}</div>
));

// ❌ Don't memoize simple components
const Divider = memo(() => <hr />);
```

### useMemo Hook

**Use useMemo for:**
- Expensive calculations
- Creating complex objects passed as props
- Filtering/sorting large arrays

**Example:**
```javascript
import { useMemo } from 'react';

function FixtureList({ fixtures, filter }) {
  // ✅ Memoize expensive filtering
  const filteredFixtures = useMemo(() => {
    return fixtures
      .filter(f => f.status === filter)
      .sort((a, b) => new Date(a.date) - new Date(b.date));
  }, [fixtures, filter]);

  return filteredFixtures.map(f => <FixtureCard key={f.id} fixture={f} />);
}
```

**DON'T use useMemo for:**
```javascript
// ❌ Simple calculations
const doubled = useMemo(() => count * 2, [count]);

// ❌ Creating arrays with few items
const items = useMemo(() => [1, 2, 3], []);
```

### useCallback Hook

**Use useCallback for:**
- Functions passed to memoized child components
- Functions used as dependencies in useEffect
- Event handlers in list items

**Example:**
```javascript
import { useCallback } from 'react';

function TodoList({ todos }) {
  // ✅ Memoize callback passed to children
  const handleToggle = useCallback((id) => {
    updateTodo(id, { completed: !completed });
  }, []);

  return todos.map(todo => (
    <TodoItem
      key={todo.id}
      todo={todo}
      onToggle={handleToggle}
    />
  ));
}
```

### Key Prop Optimization

**DO:**
```javascript
// ✅ Stable, unique keys
items.map(item => <Card key={item.id} {...item} />)

// ✅ Composite keys when needed
items.map((item, category) => (
  <Card key={`${category}-${item.id}`} {...item} />
))
```

**DON'T:**
```javascript
// ❌ Index as key (causes re-renders on reorder)
items.map((item, index) => <Card key={index} {...item} />)

// ❌ Random keys (causes full re-render)
items.map(item => <Card key={Math.random()} {...item} />)
```

---

## Image Optimization

### Responsive Images

**Pattern:**
```javascript
// ✅ Use srcset for responsive images
<img
  src="/images/hero-800.jpg"
  srcSet="/images/hero-400.jpg 400w,
          /images/hero-800.jpg 800w,
          /images/hero-1200.jpg 1200w"
  sizes="(max-width: 600px) 400px,
         (max-width: 1200px) 800px,
         1200px"
  alt="Hero banner"
  loading="lazy"
/>
```

### Lazy Loading Images

**DO:**
```javascript
// ✅ Native lazy loading
<img src="/image.jpg" loading="lazy" alt="Description" />

// ✅ Lazy load images below the fold
<img
  src={thumbnail}
  data-src={fullImage}
  className="lazy-load"
  loading="lazy"
/>
```

### Image Formats

**Priority order:**
1. WebP (smallest, modern browsers)
2. JPEG (photographs, gradients)
3. PNG (transparency, logos)
4. SVG (icons, simple graphics)

**Example:**
```javascript
<picture>
  <source srcset="/image.webp" type="image/webp" />
  <source srcset="/image.jpg" type="image/jpeg" />
  <img src="/image.jpg" alt="Fallback" />
</picture>
```

---

## Network Performance

### React Query Optimization

**Stale time strategy:**
```javascript
// From constants.js
export const QUERY_STALE_TIME_SEC = 10;

// Apply appropriate stale times
const queries = {
  // Frequently changing (live scores)
  liveData: { staleTime: 10 * 1000 },

  // Moderate (fixtures, playlists)
  fixtureData: { staleTime: 5 * 60 * 1000 },

  // Rarely changing (channel config)
  channelData: { staleTime: 15 * 60 * 1000 },

  // Static (terms, privacy)
  staticData: { staleTime: Infinity },
};
```

### Prefetching

**Prefetch on hover:**
```javascript
function FixtureCard({ fixture }) {
  const queryClient = useQueryClient();

  const handleMouseEnter = () => {
    // Prefetch fixture details on hover
    queryClient.prefetchQuery({
      queryKey: ['fixture', fixture.id],
      queryFn: () => fetchFixtureDetails(fixture.id),
    });
  };

  return (
    <Card onMouseEnter={handleMouseEnter}>
      <Link to={`/fixture/${fixture.id}`}>
        {fixture.title}
      </Link>
    </Card>
  );
}
```

### Request Deduplication

React Query automatically deduplicates requests. No additional code needed!

```javascript
// Three components mount simultaneously
<Component1 /> // useChannelData() - Makes request
<Component2 /> // useChannelData() - Reuses request
<Component3 /> // useChannelData() - Reuses request

// Only ONE network request is made ✅
```

---

## Build Optimization

### Vite Configuration

**Current configuration:**
```javascript
// vite.config.js
export default defineConfig({
  plugins: [react(), eslint()],

  build: {
    sourcemap: true, // For debugging (remove in production)
    rollupOptions: {
      output: {
        manualChunks: { /* vendor splitting */ }
      }
    }
  },

  optimizeDeps: {
    exclude: ['js-big-decimal'], // Exclude problematic deps
  },
});
```

### Production Build Optimization

**Recommended additions:**
```javascript
export default defineConfig({
  build: {
    target: 'es2015', // Support older browsers
    minify: 'terser',  // Better compression than esbuild
    terserOptions: {
      compress: {
        drop_console: true,     // Remove console.logs
        drop_debugger: true,    // Remove debugger statements
        pure_funcs: ['console.log'], // Remove specific functions
      },
    },
    rollupOptions: {
      output: {
        // Asset naming for cache busting
        assetFileNames: 'assets/[name]-[hash][extname]',
        chunkFileNames: 'chunks/[name]-[hash].js',
        entryFileNames: 'entries/[name]-[hash].js',
      },
    },
  },
});
```

### Analysing Bundle Size

```bash
# Install bundle analyser
npm install --save-dev rollup-plugin-visualizer

# Add to vite.config.js
import { visualizer } from 'rollup-plugin-visualizer';

plugins: [
  visualizer({
    open: true,
    gzipSize: true,
    brotliSize: true,
  }),
]

# Build and analyse
npm run build
```

---

## Runtime Performance

### Avoiding Re-renders

**Problem patterns:**
```javascript
// ❌ Creates new object every render
function Component() {
  const style = { color: 'red' };
  return <div style={style} />;
}

// ❌ Creates new function every render
function List({ items }) {
  return items.map(item => (
    <Item key={item.id} onClick={() => handle(item)} />
  ));
}

// ❌ Context value changes every render
function Provider({ children }) {
  const [state, setState] = useState();
  return (
    <Context.Provider value={{ state, setState }}>
      {children}
    </Context.Provider>
  );
}
```

**Solutions:**
```javascript
// ✅ Define outside component
const style = { color: 'red' };
function Component() {
  return <div style={style} />;
}

// ✅ Use useCallback
function List({ items }) {
  const handleClick = useCallback((item) => handle(item), []);
  return items.map(item => (
    <Item key={item.id} onClick={() => handleClick(item)} />
  ));
}

// ✅ Memoize context value
function Provider({ children }) {
  const [state, setState] = useState();
  const value = useMemo(() => ({ state, setState }), [state]);
  return (
    <Context.Provider value={value}>
      {children}
    </Context.Provider>
  );
}
```

### Virtualization

**For long lists:**
```javascript
import { FixedSizeList } from 'react-window';

function VirtualizedList({ items }) {
  return (
    <FixedSizeList
      height={600}
      itemCount={items.length}
      itemSize={50}
      width="100%"
    >
      {({ index, style }) => (
        <div style={style}>
          <FixtureCard fixture={items[index]} />
        </div>
      )}
    </FixedSizeList>
  );
}
```

**When to virtualize:**
- Lists with >50 items
- Infinite scroll lists
- Chat messages
- Data tables

---

## CSS Performance

### Styled Components Optimization

**DO:**
```javascript
// ✅ Define outside component (compiled once)
const Card = styled.div`
  padding: 16px;
  background: white;
`;

function Component() {
  return <Card>Content</Card>;
}
```

**DON'T:**
```javascript
// ❌ Define inside component (recompiled every render)
function Component() {
  const Card = styled.div`
    padding: 16px;
  `;
  return <Card>Content</Card>;
}
```

### CSS-in-JS Performance

**Minimize dynamic styles:**
```javascript
// ❌ Every prop change triggers style recalculation
const Box = styled.div`
  width: ${props => props.width}px;
  height: ${props => props.height}px;
  background: ${props => props.color};
`;

// ✅ Use CSS classes for variants
const Box = styled.div`
  &.large { width: 200px; height: 200px; }
  &.small { width: 100px; height: 100px; }
`;
```

### Avoiding Layout Thrashing

```javascript
// ❌ Causes layout thrashing (read-write-read-write)
element.style.width = element.offsetWidth + 10 + 'px';
element.style.height = element.offsetHeight + 10 + 'px';

// ✅ Batch reads, then batch writes
const width = element.offsetWidth;
const height = element.offsetHeight;
element.style.width = width + 10 + 'px';
element.style.height = height + 10 + 'px';
```

---

## Monitoring Performance

### Performance Measurement

**Add to production:**
```javascript
// Measure Core Web Vitals
import { getCLS, getFID, getFCP, getLCP, getTTFB } from 'web-vitals';

function sendToAnalytics(metric) {
  // Send to your analytics service
  console.log(metric);
}

getCLS(sendToAnalytics);
getFID(sendToAnalytics);
getFCP(sendToAnalytics);
getLCP(sendToAnalytics);
getTTFB(sendToAnalytics);
```

### React DevTools Profiler

```javascript
import { Profiler } from 'react';

function onRenderCallback(
  id,
  phase,
  actualDuration,
  baseDuration,
  startTime,
  commitTime,
) {
  // Log performance metrics
  if (actualDuration > 50) {
    console.warn(`Slow render: ${id} took ${actualDuration}ms`);
  }
}

<Profiler id="FixtureList" onRender={onRenderCallback}>
  <FixtureList />
</Profiler>
```

### Performance Budget Monitoring

```bash
# Check bundle sizes after build
npm run build

# Fail CI if bundles exceed limits
if [ $(stat -f%z dist/assets/*.js) -gt 524288 ]; then
  echo "Bundle size exceeds 500KB limit"
  exit 1
fi
```

---

## Performance Testing

### Lighthouse CI

```bash
# Install Lighthouse CI
npm install -g @lhci/cli

# Run Lighthouse
lhci autorun --config=lighthouserc.json
```

**lighthouserc.json:**
```json
{
  "ci": {
    "collect": {
      "numberOfRuns": 3,
      "url": ["http://localhost:5173"]
    },
    "assert": {
      "preset": "lighthouse:recommended",
      "assertions": {
        "categories:performance": ["error", {"minScore": 0.9}],
        "first-contentful-paint": ["error", {"maxNumericValue": 2000}],
        "largest-contentful-paint": ["error", {"maxNumericValue": 2500}]
      }
    }
  }
}
```

### Performance Regression Tests

```javascript
// Add to CI/CD
describe('Performance', () => {
  it('renders fixture list in under 100ms', () => {
    const start = performance.now();
    render(<FixtureList fixtures={mockFixtures} />);
    const end = performance.now();

    expect(end - start).toBeLessThan(100);
  });
});
```

---

## Performance Checklist

Before deploying to production:

- [ ] Run production build and check bundle sizes
- [ ] Lazy load routes and heavy components
- [ ] Memoize expensive calculations
- [ ] Optimize images (format, size, lazy loading)
- [ ] Configure appropriate React Query stale times
- [ ] Remove console.logs in production
- [ ] Test on 3G network (throttled)
- [ ] Run Lighthouse audit (score >90)
- [ ] Check Core Web Vitals
- [ ] Profile React components for slow renders
- [ ] Test with React DevTools Profiler
- [ ] Verify no memory leaks (event listeners, intervals)

---

## Common Performance Issues

### 1. Missing Keys in Lists

**Problem:**
```javascript
// ❌ Missing or incorrect keys
items.map((item, index) => <Card key={index} {...item} />)
```

**Impact:** React can't optimise list updates, causes unnecessary re-renders

**Fix:**
```javascript
// ✅ Stable unique keys
items.map(item => <Card key={item.id} {...item} />)
```

### 2. Inline Object/Function Props

**Problem:**
```javascript
// ❌ New object/function every render
<Component style={{ margin: 10 }} onClick={() => handle()} />
```

**Impact:** Child components re-render unnecessarily

**Fix:**
```javascript
const style = { margin: 10 };
const handleClick = useCallback(() => handle(), []);
<Component style={style} onClick={handleClick} />
```

### 3. Expensive Context Updates

**Problem:**
```javascript
// ❌ Context value changes every render
<Context.Provider value={{ data, setData }}>
```

**Impact:** All consumers re-render on every update

**Fix:**
```javascript
// ✅ Memoize context value
const value = useMemo(() => ({ data, setData }), [data]);
<Context.Provider value={value}>
```

### 4. No Code Splitting

**Problem:** Single large bundle (>500KB)

**Impact:** Slow initial load, poor FCP/LCP

**Fix:** Implement route-based and component-level code splitting

### 5. Aggressive Polling

**Problem:**
```javascript
// ❌ Polls every second
refetchInterval: 1000
```

**Impact:** Unnecessary network requests, battery drain

**Fix:**
```javascript
// ✅ Reasonable polling interval
refetchInterval: 30000 // 30 seconds
```

---

## Resources

### Internal Documentation
- [06_PERFORMANCE_OPTIMIZATION.md](../docs/06_PERFORMANCE_OPTIMIZATION.md)
- [31_CACHE_STRATEGY.md](../docs/31_CACHE_STRATEGY.md)
- [16_MOBILE_ARCHITECTURE.md](../docs/16_MOBILE_ARCHITECTURE.md)

### Tools
- [React DevTools Profiler](https://react.dev/learn/react-developer-tools)
- [Lighthouse](https://developers.google.com/web/tools/lighthouse)
- [Bundle Analyser](https://www.npmjs.com/package/rollup-plugin-visualizer)
- [web-vitals](https://github.com/GoogleChrome/web-vitals)

### External Resources
- [React Performance Optimization](https://react.dev/learn/render-and-commit)
- [Vite Performance Guide](https://vitejs.dev/guide/performance.html)
- [Core Web Vitals](https://web.dev/vitals/)