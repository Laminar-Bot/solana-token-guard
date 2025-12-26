---
name: frontend-dashboard-specialist
description: "Expert Next.js 14 frontend developer specializing in responsive dashboards, real-time data visualization, shadcn/ui component library, TailwindCSS styling, and accessible user interfaces. Deep knowledge of CryptoRugMunch's web dashboard for scan history, analytics, and premium features."
---

# Frontend & Dashboard Specialist

**Role**: Expert frontend developer for Next.js 14 with deep knowledge of modern React patterns, TailwindCSS, shadcn/ui, real-time updates, data visualization, and accessible user interfaces for CryptoRugMunch's web dashboard.

**Context**: CryptoRugMunch's **primary interface is the Telegram bot**, but users need a **web dashboard** for:
- Viewing scan history with detailed breakdowns
- Analyzing trends (most scammed tokens, risk distribution)
- Managing subscription and billing
- Accessing premium features (export, advanced filters)
- Gamification (leaderboards, NFT badges, XP progress)

---

## Core Philosophy

1. **Mobile-First Design**: 80% of users access via mobile, dashboard must be responsive
2. **Performance-First**: Fast page loads (<1s), optimistic UI updates, aggressive caching
3. **Accessibility (WCAG 2.1 AA)**: Semantic HTML, keyboard navigation, screen reader support
4. **Component Reusability**: shadcn/ui base, custom variants, consistent design system
5. **Real-Time Updates**: WebSocket for live scan results, optimistic UI for instant feedback

---

## 1. Tech Stack Overview

### 1.1 Frontend Framework

**Next.js 14** (App Router, React Server Components)

```json
// package.json (frontend dependencies)
{
  "dependencies": {
    "next": "^14.2.0",
    "react": "^18.3.0",
    "react-dom": "^18.3.0",

    // UI Components
    "tailwindcss": "^3.4.0",
    "@radix-ui/react-alert-dialog": "^1.0.5",
    "@radix-ui/react-dropdown-menu": "^2.0.6",
    "@radix-ui/react-toast": "^1.1.5",
    "lucide-react": "^0.344.0", // Icons
    "class-variance-authority": "^0.7.0", // Component variants
    "clsx": "^2.1.0",
    "tailwind-merge": "^2.2.0",

    // Data Fetching & State
    "@tanstack/react-query": "^5.24.0",
    "swr": "^2.2.5", // Alternative to React Query
    "zustand": "^4.5.0", // Global state

    // Charts & Visualization
    "recharts": "^2.12.0",
    "framer-motion": "^11.0.5", // Animations

    // Forms & Validation
    "react-hook-form": "^7.50.0",
    "zod": "^3.22.4",
    "@hookform/resolvers": "^3.3.4",

    // Date & Time
    "date-fns": "^3.3.0",

    // Web3 (for wallet connect)
    "@solana/wallet-adapter-react": "^0.15.35",
    "@solana/wallet-adapter-react-ui": "^0.9.35"
  },
  "devDependencies": {
    "@types/node": "^20",
    "@types/react": "^18",
    "@types/react-dom": "^18",
    "typescript": "^5",
    "autoprefixer": "^10.4.17",
    "postcss": "^8.4.35",
    "eslint": "^8",
    "eslint-config-next": "14.2.0"
  }
}
```

### 1.2 Project Structure

```
apps/web/                          # Next.js frontend
├── app/                           # App Router pages
│   ├── (auth)/                    # Auth layout group
│   │   ├── login/page.tsx         # Login page
│   │   └── signup/page.tsx        # Signup page
│   ├── (dashboard)/               # Dashboard layout group
│   │   ├── layout.tsx             # Shared dashboard layout
│   │   ├── page.tsx               # Dashboard home
│   │   ├── scans/                 # Scan history
│   │   │   ├── page.tsx           # Scan list
│   │   │   └── [id]/page.tsx      # Scan detail
│   │   ├── analytics/page.tsx     # Analytics dashboard
│   │   ├── leaderboard/page.tsx   # Gamification leaderboard
│   │   ├── settings/page.tsx      # User settings
│   │   └── billing/page.tsx       # Subscription management
│   ├── api/                       # API routes (proxy to backend)
│   │   ├── auth/[...nextauth]/route.ts
│   │   └── webhook/stripe/route.ts
│   └── layout.tsx                 # Root layout
├── components/                    # React components
│   ├── ui/                        # shadcn/ui components
│   │   ├── button.tsx
│   │   ├── card.tsx
│   │   ├── dialog.tsx
│   │   ├── table.tsx
│   │   └── ... (50+ components)
│   ├── scan/                      # Scan-related components
│   │   ├── ScanCard.tsx
│   │   ├── RiskBadge.tsx
│   │   ├── ScanDetailModal.tsx
│   │   └── ScanTimeline.tsx
│   ├── analytics/                 # Analytics components
│   │   ├── RiskDistributionChart.tsx
│   │   ├── ScanVolumeChart.tsx
│   │   └── TopScammedTokens.tsx
│   └── layout/                    # Layout components
│       ├── Navbar.tsx
│       ├── Sidebar.tsx
│       └── Footer.tsx
├── lib/                           # Utilities
│   ├── api.ts                     # API client
│   ├── utils.ts                   # Helpers (cn, formatters)
│   └── constants.ts               # Constants
├── hooks/                         # Custom React hooks
│   ├── useScans.ts
│   ├── useAuth.ts
│   └── useWebSocket.ts
├── styles/
│   └── globals.css                # Global styles + Tailwind
├── public/                        # Static assets
│   ├── images/
│   └── fonts/
└── tailwind.config.ts             # Tailwind configuration
```

---

## 2. shadcn/ui Component Library

### 2.1 Setup & Configuration

**shadcn/ui** is a collection of **copy-paste components** built on Radix UI + Tailwind.

```bash
# Install shadcn/ui CLI
npx shadcn-ui@latest init

# Add components (examples)
npx shadcn-ui@latest add button
npx shadcn-ui@latest add card
npx shadcn-ui@latest add dialog
npx shadcn-ui@latest add table
npx shadcn-ui@latest add toast
npx shadcn-ui@latest add dropdown-menu
```

```typescript
// components/ui/button.tsx (example shadcn component)
import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"

const buttonVariants = cva(
  "inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground hover:bg-primary/90",
        destructive: "bg-destructive text-destructive-foreground hover:bg-destructive/90",
        outline: "border border-input bg-background hover:bg-accent hover:text-accent-foreground",
        secondary: "bg-secondary text-secondary-foreground hover:bg-secondary/80",
        ghost: "hover:bg-accent hover:text-accent-foreground",
        link: "text-primary underline-offset-4 hover:underline",
      },
      size: {
        default: "h-10 px-4 py-2",
        sm: "h-9 rounded-md px-3",
        lg: "h-11 rounded-md px-8",
        icon: "h-10 w-10",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : "button"
    return (
      <Comp
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        {...props}
      />
    )
  }
)
Button.displayName = "Button"

export { Button, buttonVariants }
```

**Usage**:

```tsx
import { Button } from "@/components/ui/button"

<Button variant="default">Scan Token</Button>
<Button variant="destructive">Delete</Button>
<Button variant="outline" size="sm">Cancel</Button>
```

### 2.2 Custom Component Variants

**Goal**: Extend shadcn components for CryptoRugMunch-specific use cases.

```tsx
// components/ui/risk-badge.tsx (custom component)
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"

const riskBadgeVariants = cva(
  "inline-flex items-center rounded-full px-3 py-1 text-xs font-semibold ring-1 ring-inset",
  {
    variants: {
      risk: {
        SAFE: "bg-green-50 text-green-700 ring-green-600/20",
        CAUTION: "bg-yellow-50 text-yellow-800 ring-yellow-600/20",
        HIGH_RISK: "bg-orange-50 text-orange-700 ring-orange-600/20",
        LIKELY_SCAM: "bg-red-50 text-red-700 ring-red-600/20",
      },
    },
  }
)

export interface RiskBadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof riskBadgeVariants> {
  score: number // 0-100
}

export function RiskBadge({ risk, score, className, ...props }: RiskBadgeProps) {
  const emoji = {
    SAFE: "🟢",
    CAUTION: "🟡",
    HIGH_RISK: "🟠",
    LIKELY_SCAM: "🔴",
  }

  return (
    <div className={cn(riskBadgeVariants({ risk }), className)} {...props}>
      <span className="mr-1">{emoji[risk!]}</span>
      {risk} ({score}/100)
    </div>
  )
}
```

**Usage**:

```tsx
<RiskBadge risk="SAFE" score={92} />
<RiskBadge risk="LIKELY_SCAM" score={18} />
```

---

## 3. Dashboard Pages

### 3.1 Dashboard Home (Scan Overview)

**Route**: `/` (authenticated users)

**Features**:
- Recent scans (last 10)
- Quick scan input
- Stats: Total scans, scams detected, XP earned
- Premium upsell banner (free users)

```tsx
// app/(dashboard)/page.tsx
import { Suspense } from 'react'
import { auth } from '@/lib/auth'
import { redirect } from 'next/navigation'
import { ScanInput } from '@/components/scan/ScanInput'
import { RecentScans } from '@/components/scan/RecentScans'
import { StatsCards } from '@/components/analytics/StatsCards'
import { PremiumBanner } from '@/components/billing/PremiumBanner'

export default async function DashboardPage() {
  const session = await auth()

  if (!session) {
    redirect('/login')
  }

  return (
    <div className="container mx-auto py-10 space-y-8">
      {/* Stats */}
      <Suspense fallback={<StatsCardsSkeleton />}>
        <StatsCards userId={session.user.id} />
      </Suspense>

      {/* Premium Banner (free users only) */}
      {session.user.tier === 'FREE' && <PremiumBanner />}

      {/* Quick Scan */}
      <div className="max-w-2xl mx-auto">
        <h2 className="text-2xl font-bold mb-4">Scan a Token</h2>
        <ScanInput userId={session.user.id} />
      </div>

      {/* Recent Scans */}
      <Suspense fallback={<RecentScansSkeleton />}>
        <RecentScans userId={session.user.id} />
      </Suspense>
    </div>
  )
}
```

```tsx
// components/scan/ScanInput.tsx
'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { toast } from '@/components/ui/use-toast'
import { scanToken } from '@/lib/api'

export function ScanInput({ userId }: { userId: string }) {
  const [chain, setChain] = useState<'SOLANA' | 'ETHEREUM'>('SOLANA')
  const [address, setAddress] = useState('')
  const [loading, setLoading] = useState(false)

  const handleScan = async () => {
    if (!address) {
      toast({ title: 'Error', description: 'Please enter a token address', variant: 'destructive' })
      return
    }

    setLoading(true)

    try {
      const result = await scanToken({ chain, tokenAddress: address })

      toast({
        title: '✅ Scan Complete',
        description: `Risk Score: ${result.riskScore}/100 (${result.category})`,
      })

      // Redirect to scan detail
      window.location.href = `/scans/${result.scanId}`
    } catch (error: any) {
      toast({
        title: '❌ Scan Failed',
        description: error.message || 'Unknown error',
        variant: 'destructive',
      })
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <Select value={chain} onValueChange={(val) => setChain(val as 'SOLANA' | 'ETHEREUM')}>
        <SelectTrigger>
          <SelectValue placeholder="Select blockchain" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="SOLANA">🟣 Solana</SelectItem>
          <SelectItem value="ETHEREUM">🔷 Ethereum</SelectItem>
          <SelectItem value="BASE">🔵 Base</SelectItem>
          <SelectItem value="BSC">🟡 BSC</SelectItem>
        </SelectContent>
      </Select>

      <Input
        type="text"
        placeholder="Enter token address..."
        value={address}
        onChange={(e) => setAddress(e.target.value)}
      />

      <Button onClick={handleScan} disabled={loading} className="w-full">
        {loading ? '🔍 Scanning...' : '🔍 Scan Token'}
      </Button>
    </div>
  )
}
```

### 3.2 Scan History Page

**Route**: `/scans`

**Features**:
- Table of all scans (paginated)
- Filters: Chain, risk category, date range
- Export to CSV (premium)

```tsx
// app/(dashboard)/scans/page.tsx
import { Suspense } from 'react'
import { auth } from '@/lib/auth'
import { redirect } from 'next/navigation'
import { ScanTable } from '@/components/scan/ScanTable'
import { ScanFilters } from '@/components/scan/ScanFilters'

export default async function ScansPage() {
  const session = await auth()

  if (!session) {
    redirect('/login')
  }

  return (
    <div className="container mx-auto py-10">
      <h1 className="text-3xl font-bold mb-8">Scan History</h1>

      <ScanFilters />

      <Suspense fallback={<div>Loading scans...</div>}>
        <ScanTable userId={session.user.id} />
      </Suspense>
    </div>
  )
}
```

```tsx
// components/scan/ScanTable.tsx
'use client'

import { useScans } from '@/hooks/useScans'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { RiskBadge } from '@/components/ui/risk-badge'
import { formatDistanceToNow } from 'date-fns'
import { Button } from '@/components/ui/button'
import { ExternalLink } from 'lucide-react'

export function ScanTable({ userId }: { userId: string }) {
  const { data: scans, isLoading, error } = useScans(userId)

  if (isLoading) return <div>Loading...</div>
  if (error) return <div>Error: {error.message}</div>

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Token</TableHead>
            <TableHead>Chain</TableHead>
            <TableHead>Risk Score</TableHead>
            <TableHead>Scanned</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {scans.map((scan) => (
            <TableRow key={scan.id}>
              <TableCell className="font-mono text-sm">
                {scan.tokenAddress.slice(0, 8)}...{scan.tokenAddress.slice(-6)}
              </TableCell>
              <TableCell>{scan.chain}</TableCell>
              <TableCell>
                <RiskBadge risk={scan.category} score={scan.riskScore} />
              </TableCell>
              <TableCell className="text-muted-foreground">
                {formatDistanceToNow(new Date(scan.scannedAt), { addSuffix: true })}
              </TableCell>
              <TableCell className="text-right">
                <Button variant="ghost" size="sm" asChild>
                  <a href={`/scans/${scan.id}`}>
                    View <ExternalLink className="ml-2 h-4 w-4" />
                  </a>
                </Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
```

### 3.3 Scan Detail Page

**Route**: `/scans/[id]`

**Features**:
- Full risk breakdown (12 metrics)
- Token metadata (name, symbol, supply)
- Timeline (when scanned, result)
- Share button (copy link)
- Rescan button (premium)

```tsx
// app/(dashboard)/scans/[id]/page.tsx
import { auth } from '@/lib/auth'
import { redirect } from 'next/navigation'
import { getScanById } from '@/lib/api'
import { RiskBadge } from '@/components/ui/risk-badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Share2, RefreshCw } from 'lucide-react'

export default async function ScanDetailPage({ params }: { params: { id: string } }) {
  const session = await auth()

  if (!session) {
    redirect('/login')
  }

  const scan = await getScanById(params.id)

  if (!scan || scan.userId !== session.user.id) {
    return <div>Scan not found</div>
  }

  return (
    <div className="container mx-auto py-10 max-w-4xl">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold mb-2">Scan Details</h1>
          <p className="text-muted-foreground font-mono text-sm">{scan.tokenAddress}</p>
        </div>

        <div className="flex gap-2">
          <Button variant="outline" size="sm">
            <Share2 className="mr-2 h-4 w-4" /> Share
          </Button>
          {session.user.tier === 'PREMIUM' && (
            <Button size="sm">
              <RefreshCw className="mr-2 h-4 w-4" /> Rescan
            </Button>
          )}
        </div>
      </div>

      {/* Risk Score */}
      <Card className="mb-8">
        <CardHeader>
          <CardTitle>Risk Assessment</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-muted-foreground mb-2">Overall Risk Score</p>
              <RiskBadge risk={scan.category} score={scan.riskScore} className="text-lg px-4 py-2" />
            </div>

            <div className="text-6xl font-bold text-muted-foreground">
              {scan.riskScore}
              <span className="text-2xl">/100</span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Breakdown */}
      <Card className="mb-8">
        <CardHeader>
          <CardTitle>Risk Breakdown</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {Object.entries(scan.breakdown as Record<string, number>).map(([metric, score]) => (
              <div key={metric} className="flex items-center justify-between">
                <span className="text-sm capitalize">{metric.replace(/([A-Z])/g, ' $1')}</span>
                <div className="flex items-center gap-4">
                  <div className="w-48 bg-secondary rounded-full h-2">
                    <div
                      className="bg-primary h-2 rounded-full transition-all"
                      style={{ width: `${(score / 20) * 100}%` }}
                    />
                  </div>
                  <span className="text-sm font-semibold w-12 text-right">{score}/20</span>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Flags */}
      {scan.flags && scan.flags.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>⚠️ Warning Flags</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="space-y-2">
              {scan.flags.map((flag, index) => (
                <li key={index} className="text-sm text-muted-foreground">
                  • {flag}
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
```

### 3.4 Analytics Dashboard

**Route**: `/analytics`

**Features**:
- Risk distribution chart (pie chart)
- Scan volume over time (line chart)
- Top scammed tokens (table)
- Chain breakdown (bar chart)

```tsx
// app/(dashboard)/analytics/page.tsx
import { auth } from '@/lib/auth'
import { redirect } from 'next/navigation'
import { RiskDistributionChart } from '@/components/analytics/RiskDistributionChart'
import { ScanVolumeChart } from '@/components/analytics/ScanVolumeChart'
import { TopScammedTokens } from '@/components/analytics/TopScammedTokens'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export default async function AnalyticsPage() {
  const session = await auth()

  if (!session) {
    redirect('/login')
  }

  return (
    <div className="container mx-auto py-10">
      <h1 className="text-3xl font-bold mb-8">Analytics</h1>

      <div className="grid gap-6 md:grid-cols-2 mb-8">
        <Card>
          <CardHeader>
            <CardTitle>Risk Distribution</CardTitle>
          </CardHeader>
          <CardContent>
            <RiskDistributionChart userId={session.user.id} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Scan Volume (Last 30 Days)</CardTitle>
          </CardHeader>
          <CardContent>
            <ScanVolumeChart userId={session.user.id} />
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Top Scammed Tokens</CardTitle>
        </CardHeader>
        <CardContent>
          <TopScammedTokens />
        </CardContent>
      </Card>
    </div>
  )
}
```

```tsx
// components/analytics/RiskDistributionChart.tsx
'use client'

import { useAnalytics } from '@/hooks/useAnalytics'
import { PieChart, Pie, Cell, ResponsiveContainer, Legend, Tooltip } from 'recharts'

const COLORS = {
  SAFE: '#22c55e',
  CAUTION: '#eab308',
  HIGH_RISK: '#f97316',
  LIKELY_SCAM: '#ef4444',
}

export function RiskDistributionChart({ userId }: { userId: string }) {
  const { data: analytics, isLoading } = useAnalytics(userId)

  if (isLoading) return <div>Loading...</div>

  const data = [
    { name: 'Safe', value: analytics.riskDistribution.SAFE, color: COLORS.SAFE },
    { name: 'Caution', value: analytics.riskDistribution.CAUTION, color: COLORS.CAUTION },
    { name: 'High Risk', value: analytics.riskDistribution.HIGH_RISK, color: COLORS.HIGH_RISK },
    { name: 'Likely Scam', value: analytics.riskDistribution.LIKELY_SCAM, color: COLORS.LIKELY_SCAM },
  ]

  return (
    <ResponsiveContainer width="100%" height={300}>
      <PieChart>
        <Pie
          data={data}
          cx="50%"
          cy="50%"
          labelLine={false}
          label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
          outerRadius={80}
          fill="#8884d8"
          dataKey="value"
        >
          {data.map((entry, index) => (
            <Cell key={`cell-${index}`} fill={entry.color} />
          ))}
        </Pie>
        <Tooltip />
        <Legend />
      </PieChart>
    </ResponsiveContainer>
  )
}
```

### 3.5 Leaderboard Page

**Route**: `/leaderboard`

**Features**:
- Top users by XP (global)
- User's rank
- Badges earned
- Weekly/monthly/all-time filters

```tsx
// app/(dashboard)/leaderboard/page.tsx
import { auth } from '@/lib/auth'
import { redirect } from 'next/navigation'
import { getLeaderboard } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Trophy, Award } from 'lucide-react'

export default async function LeaderboardPage() {
  const session = await auth()

  if (!session) {
    redirect('/login')
  }

  const leaderboard = await getLeaderboard()

  return (
    <div className="container mx-auto py-10 max-w-4xl">
      <h1 className="text-3xl font-bold mb-8 flex items-center gap-2">
        <Trophy className="h-8 w-8 text-yellow-500" />
        Leaderboard
      </h1>

      <Card>
        <CardHeader>
          <CardTitle>Top Scam Hunters</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {leaderboard.map((user, index) => (
              <div
                key={user.id}
                className="flex items-center justify-between p-4 rounded-lg bg-secondary/50"
              >
                <div className="flex items-center gap-4">
                  <div className="text-2xl font-bold text-muted-foreground w-8">
                    #{index + 1}
                  </div>

                  <Avatar>
                    <AvatarImage src={user.avatarUrl} />
                    <AvatarFallback>{user.username[0].toUpperCase()}</AvatarFallback>
                  </Avatar>

                  <div>
                    <p className="font-semibold">{user.username}</p>
                    <p className="text-sm text-muted-foreground">
                      Level {user.level} • {user.title}
                    </p>
                  </div>
                </div>

                <div className="text-right">
                  <p className="text-2xl font-bold">{user.xp.toLocaleString()} XP</p>
                  <div className="flex items-center gap-1 text-sm text-muted-foreground">
                    <Award className="h-4 w-4" />
                    {user.badgeCount} badges
                  </div>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
```

---

## 4. Real-Time Updates (WebSocket)

### 4.1 WebSocket Setup

**Goal**: Push scan results to dashboard in real-time (avoid polling).

```typescript
// lib/websocket.ts
import { useEffect, useState } from 'react'
import { io, Socket } from 'socket.io-client'

let socket: Socket | null = null

export function useWebSocket(userId: string) {
  const [connected, setConnected] = useState(false)

  useEffect(() => {
    if (!socket) {
      socket = io(process.env.NEXT_PUBLIC_WS_URL!, {
        auth: { userId },
      })

      socket.on('connect', () => {
        console.log('WebSocket connected')
        setConnected(true)
      })

      socket.on('disconnect', () => {
        console.log('WebSocket disconnected')
        setConnected(false)
      })
    }

    return () => {
      // Don't disconnect on unmount (keep persistent connection)
    }
  }, [userId])

  return { socket, connected }
}

export function useScanUpdates(userId: string, onUpdate: (scan: any) => void) {
  const { socket, connected } = useWebSocket(userId)

  useEffect(() => {
    if (!socket) return

    socket.on('scan:completed', (scan) => {
      console.log('Scan completed:', scan)
      onUpdate(scan)
    })

    return () => {
      socket.off('scan:completed')
    }
  }, [socket, onUpdate])

  return { connected }
}
```

**Usage in Component**:

```tsx
// components/scan/ScanInput.tsx (updated with WebSocket)
'use client'

import { useScanUpdates } from '@/lib/websocket'
import { toast } from '@/components/ui/use-toast'

export function ScanInput({ userId }: { userId: string }) {
  // ... (existing state)

  useScanUpdates(userId, (scan) => {
    toast({
      title: '✅ Scan Complete',
      description: `${scan.tokenAddress.slice(0, 8)}... scored ${scan.riskScore}/100`,
    })

    // Optionally navigate to scan detail
    // window.location.href = `/scans/${scan.id}`
  })

  // ... (existing render)
}
```

---

## 5. Data Fetching Patterns

### 5.1 React Query (TanStack Query)

**Recommended for**: Client-side data fetching, caching, revalidation.

```tsx
// hooks/useScans.ts
import { useQuery } from '@tanstack/react-query'
import { getScans } from '@/lib/api'

export function useScans(userId: string) {
  return useQuery({
    queryKey: ['scans', userId],
    queryFn: () => getScans(userId),
    staleTime: 5 * 60 * 1000, // 5 minutes
    refetchInterval: 60 * 1000, // Refetch every 1 minute (polling)
  })
}
```

```tsx
// app/(dashboard)/layout.tsx (setup QueryClientProvider)
'use client'

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useState } from 'react'

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(() => new QueryClient())

  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  )
}
```

### 5.2 Server Components (Next.js 14 App Router)

**Recommended for**: Initial page load data (SEO, fast initial render).

```tsx
// app/(dashboard)/scans/page.tsx
import { getScans } from '@/lib/api'

export default async function ScansPage() {
  const scans = await getScans() // Fetch on server

  return (
    <div>
      {scans.map((scan) => (
        <ScanCard key={scan.id} scan={scan} />
      ))}
    </div>
  )
}
```

### 5.3 API Client

```typescript
// lib/api.ts
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL!

async function fetchAPI(endpoint: string, options?: RequestInit) {
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    credentials: 'include', // Send cookies
  })

  if (!response.ok) {
    const error = await response.json()
    throw new Error(error.message || 'API request failed')
  }

  return response.json()
}

export async function getScans(userId: string) {
  return fetchAPI(`/api/scans?userId=${userId}`)
}

export async function getScanById(scanId: string) {
  return fetchAPI(`/api/scans/${scanId}`)
}

export async function scanToken(data: { chain: string; tokenAddress: string }) {
  return fetchAPI('/api/scan', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function getAnalytics(userId: string) {
  return fetchAPI(`/api/analytics?userId=${userId}`)
}

export async function getLeaderboard() {
  return fetchAPI('/api/leaderboard')
}
```

---

## 6. Forms & Validation

### 6.1 React Hook Form + Zod

```tsx
// components/forms/ScanForm.tsx
'use client'

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'

const scanFormSchema = z.object({
  chain: z.enum(['SOLANA', 'ETHEREUM', 'BASE', 'BSC', 'POLYGON']),
  tokenAddress: z.string().min(32, 'Invalid address').max(66, 'Invalid address'),
})

type ScanFormValues = z.infer<typeof scanFormSchema>

export function ScanForm({ onSubmit }: { onSubmit: (values: ScanFormValues) => void }) {
  const form = useForm<ScanFormValues>({
    resolver: zodResolver(scanFormSchema),
    defaultValues: {
      chain: 'SOLANA',
      tokenAddress: '',
    },
  })

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="chain"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Blockchain</FormLabel>
              <FormControl>
                <Select onValueChange={field.onChange} defaultValue={field.value}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select chain" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="SOLANA">Solana</SelectItem>
                    <SelectItem value="ETHEREUM">Ethereum</SelectItem>
                  </SelectContent>
                </Select>
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="tokenAddress"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Token Address</FormLabel>
              <FormControl>
                <Input placeholder="Enter token address..." {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <Button type="submit" className="w-full">
          Scan Token
        </Button>
      </form>
    </Form>
  )
}
```

---

## 7. Responsive Design & Mobile-First

### 7.1 TailwindCSS Breakpoints

```tsx
// Example: Responsive grid
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
  {/* Stacks on mobile, 2 cols on tablet, 3 cols on desktop */}
</div>

// Example: Hide sidebar on mobile
<aside className="hidden lg:block w-64">
  <Sidebar />
</aside>

// Example: Responsive text size
<h1 className="text-2xl md:text-3xl lg:text-4xl font-bold">
  Dashboard
</h1>
```

### 7.2 Mobile Navigation

```tsx
// components/layout/MobileNav.tsx
'use client'

import { useState } from 'react'
import { Menu } from 'lucide-react'
import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet'
import { Button } from '@/components/ui/button'

export function MobileNav() {
  const [open, setOpen] = useState(false)

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button variant="ghost" size="icon" className="lg:hidden">
          <Menu className="h-6 w-6" />
        </Button>
      </SheetTrigger>
      <SheetContent side="left">
        <nav className="flex flex-col gap-4">
          <a href="/" onClick={() => setOpen(false)}>Dashboard</a>
          <a href="/scans" onClick={() => setOpen(false)}>Scan History</a>
          <a href="/analytics" onClick={() => setOpen(false)}>Analytics</a>
          <a href="/leaderboard" onClick={() => setOpen(false)}>Leaderboard</a>
          <a href="/settings" onClick={() => setOpen(false)}>Settings</a>
        </nav>
      </SheetContent>
    </Sheet>
  )
}
```

---

## 8. Accessibility (WCAG 2.1 AA)

### 8.1 Semantic HTML

```tsx
// ✅ GOOD: Semantic HTML
<header>
  <nav>
    <ul>
      <li><a href="/">Home</a></li>
    </ul>
  </nav>
</header>

<main>
  <article>
    <h1>Scan Results</h1>
    <section>
      <h2>Risk Breakdown</h2>
    </section>
  </article>
</main>

<footer>
  <p>&copy; 2025 CryptoRugMunch</p>
</footer>

// ❌ BAD: Non-semantic divs
<div className="header">
  <div className="nav">
    <div><a href="/">Home</a></div>
  </div>
</div>
```

### 8.2 Keyboard Navigation

```tsx
// Ensure all interactive elements are keyboard accessible
<Button
  onClick={handleClick}
  onKeyDown={(e) => {
    if (e.key === 'Enter' || e.key === ' ') {
      handleClick()
    }
  }}
  tabIndex={0}
>
  Scan Token
</Button>
```

### 8.3 ARIA Labels

```tsx
// components/scan/ScanCard.tsx
<div role="article" aria-labelledby="scan-title">
  <h2 id="scan-title">Scan #{scan.id}</h2>
  <p aria-label={`Risk score: ${scan.riskScore} out of 100`}>
    Score: {scan.riskScore}/100
  </p>
</div>

// Screen reader announcements
<div role="status" aria-live="polite">
  {loading ? 'Scanning token...' : 'Scan complete'}
</div>
```

### 8.4 Color Contrast

```css
/* Ensure all text meets WCAG AA contrast ratio (4.5:1 for normal text) */
.text-muted-foreground {
  color: hsl(240 3.8% 46.1%); /* #6b7280 - passes AA */
}

.bg-primary {
  background-color: hsl(221.2 83.2% 53.3%); /* #3b82f6 - passes AA with white text */
}
```

---

## 9. Performance Optimization

### 9.1 Image Optimization

```tsx
import Image from 'next/image'

// ✅ GOOD: Next.js Image component (auto-optimization)
<Image
  src="/logo.png"
  alt="CryptoRugMunch Logo"
  width={200}
  height={50}
  priority // Load immediately (above fold)
/>

// ❌ BAD: Regular img tag (no optimization)
<img src="/logo.png" alt="Logo" />
```

### 9.2 Code Splitting

```tsx
// Dynamic imports for large components
import dynamic from 'next/dynamic'

const HeavyChart = dynamic(() => import('@/components/analytics/HeavyChart'), {
  loading: () => <div>Loading chart...</div>,
  ssr: false, // Don't render on server (client-only)
})

export function AnalyticsPage() {
  return (
    <div>
      <h1>Analytics</h1>
      <HeavyChart />
    </div>
  )
}
```

### 9.3 Memoization

```tsx
'use client'

import { useMemo } from 'react'

export function ExpensiveComponent({ data }: { data: any[] }) {
  const processedData = useMemo(() => {
    // Expensive calculation
    return data.map((item) => /* ... */)
  }, [data]) // Only recalculate when data changes

  return <div>{/* Render processedData */}</div>
}
```

---

## 10. Command Shortcuts

Use these shortcuts to quickly access specific topics:

- **#nextjs** - Next.js 14 App Router
- **#shadcn** - shadcn/ui component library
- **#tailwind** - TailwindCSS styling
- **#forms** - React Hook Form + Zod validation
- **#charts** - Recharts data visualization
- **#websocket** - Real-time WebSocket updates
- **#query** - React Query data fetching
- **#mobile** - Responsive design, mobile-first
- **#a11y** - Accessibility (WCAG 2.1 AA)
- **#perf** - Performance optimization

---

## 11. Reference Materials

### 11.1 CryptoRugMunch Documentation

**Related Skills**:
- `rugmunch-architect` - System architecture overview
- `telegram-bot-developer` - Telegram bot UI patterns
- `gamification-engineer` - Leaderboard, XP, badges

**Project Docs**:
- `/docs/02-PRODUCT/ux-design-principles.md` - Design system, UX flows
- `/docs/03-TECHNICAL/architecture/api-specification.md` - API endpoints for dashboard
- `/docs/03-TECHNICAL/integrations/telegram-bot-risk-algorithm.md` - Risk scoring breakdown display

### 11.2 External Libraries

**Next.js 14**:
- Docs: https://nextjs.org/docs
- App Router: https://nextjs.org/docs/app

**shadcn/ui**:
- Docs: https://ui.shadcn.com
- Components: https://ui.shadcn.com/docs/components

**TailwindCSS**:
- Docs: https://tailwindcss.com/docs
- Cheat Sheet: https://nerdcave.com/tailwind-cheat-sheet

**React Query**:
- Docs: https://tanstack.com/query/latest

**Recharts**:
- Docs: https://recharts.org/en-US

**React Hook Form**:
- Docs: https://react-hook-form.com

---

## Summary

The **Frontend & Dashboard Specialist** skill provides comprehensive expertise for building CryptoRugMunch's web dashboard with Next.js 14, shadcn/ui, and TailwindCSS. Key capabilities:

1. **Modern Stack**: Next.js 14 App Router, React Server Components, shadcn/ui, TailwindCSS
2. **Dashboard Pages**: Scan history, analytics charts, leaderboard, billing
3. **Real-Time Updates**: WebSocket for live scan results, optimistic UI
4. **Data Fetching**: React Query for client-side, Server Components for SSR
5. **Forms & Validation**: React Hook Form + Zod for type-safe forms
6. **Responsive Design**: Mobile-first, TailwindCSS breakpoints
7. **Accessibility**: WCAG 2.1 AA compliance, semantic HTML, keyboard navigation
8. **Performance**: Image optimization, code splitting, memoization

**Timeline**: Dashboard developed in parallel with backend (Week 5-8).

**Next Steps**:
1. Set up Next.js 14 project with TypeScript
2. Install shadcn/ui components (button, card, table, dialog, toast)
3. Build dashboard layout (navbar, sidebar, footer)
4. Implement scan history page with pagination
5. Create scan detail page with risk breakdown
6. Add analytics charts (Recharts)
7. Integrate WebSocket for real-time updates
8. Deploy to Vercel

---

**Built with modern React patterns** ⚛️
**Accessible, performant, beautiful** ✨
