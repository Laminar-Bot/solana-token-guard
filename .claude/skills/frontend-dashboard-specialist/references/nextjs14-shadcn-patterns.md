# Next.js 14 + shadcn/ui Dashboard Patterns

## Server Component Data Fetching

```tsx
// app/dashboard/page.tsx (Server Component)
export default async function DashboardPage() {
  const scans = await prisma.scan.findMany({
    where: { userId: session.user.id },
    orderBy: { createdAt: 'desc' },
    take: 10,
  });

  return (
    <div className="container mx-auto py-8">
      <h1 className="text-3xl font-bold mb-6">Recent Scans</h1>
      <ScanTable scans={scans} />
    </div>
  );
}
```

## Client Component with React Query

```tsx
'use client';

import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

export function UsageStats() {
  const { data, isLoading } = useQuery({
    queryKey: ['usage'],
    queryFn: () => fetch('/api/usage').then(res => res.json()),
  });

  if (isLoading) return <Skeleton className="h-32" />;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Usage This Month</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{data.scansUsed} / {data.scansLimit}</div>
        <Progress value={(data.scansUsed / data.scansLimit) * 100} className="mt-2" />
      </CardContent>
    </Card>
  );
}
```

## Form Validation (React Hook Form + Zod)

```tsx
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';

const schema = z.object({
  tokenAddress: z.string().regex(/^[1-9A-HJ-NP-Za-km-z]{32,44}$/),
});

export function ScanForm() {
  const form = useForm({ resolver: zodResolver(schema) });

  const onSubmit = async (data: z.infer<typeof schema>) => {
    const result = await fetch('/api/scan', {
      method: 'POST',
      body: JSON.stringify(data),
    }).then(res => res.json());

    toast.success('Scan completed!');
  };

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)}>
        <FormField
          control={form.control}
          name="tokenAddress"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Token Address</FormLabel>
              <FormControl>
                <Input placeholder="Enter Solana address..." {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button type="submit">Scan Token</Button>
      </form>
    </Form>
  );
}
```

## Real-Time Updates (WebSocket)

```tsx
'use client';

import { useEffect, useState } from 'react';
import { io } from 'socket.io-client';

export function LiveScanResults() {
  const [scans, setScans] = useState([]);

  useEffect(() => {
    const socket = io(process.env.NEXT_PUBLIC_WS_URL!);

    socket.on('scan:complete', (scan) => {
      setScans(prev => [scan, ...prev]);
      toast.success(`Scan completed for ${scan.tokenSymbol}`);
    });

    return () => socket.disconnect();
  }, []);

  return <div>{scans.map(scan => <ScanCard key={scan.id} scan={scan} />)}</div>;
}
```
