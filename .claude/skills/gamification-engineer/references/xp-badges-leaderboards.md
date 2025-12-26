# Gamification System Implementation Patterns

## XP & Level System

```typescript
// XP formula: exponential growth
export function calculateRequiredXP(level: number): number {
  return Math.floor(100 * Math.pow(1.5, level - 1));
}

// Award XP for actions
export async function awardXP(userId: string, action: string, amount: number) {
  const user = await userRepository.findById(userId);
  const newXP = user.xp + amount;
  const newLevel = calculateLevel(newXP);

  if (newLevel > user.level) {
    await sendLevelUpNotification(userId, newLevel);
    await unlockBadges(userId, newLevel);
  }

  await userRepository.update(userId, { xp: newXP, level: newLevel });
  metrics.increment('gamification.xp_awarded', amount, { action });
}
```

## NFT Badge Minting (Solana Metaplex)

```typescript
import { Metaplex } from '@metaplex-foundation/js';

export async function mintBadgeNFT(userId: string, badgeType: string) {
  const metaplex = new Metaplex(connection);
  const { uri } = await metaplex.nfts().uploadMetadata({
    name: `${badgeType} Badge`,
    symbol: 'CRMBADGE',
    description: `Earned for ${badgeType}`,
    image: `https://rugmunch.com/badges/${badgeType}.png`,
  });

  const { nft } = await metaplex.nfts().create({
    uri,
    name: `${badgeType} Badge`,
    sellerFeeBasisPoints: 0,
    tokenOwner: new PublicKey(userWallet),
  });

  return nft.address.toBase58();
}
```

## Leaderboard (Redis Sorted Sets)

```typescript
// Update leaderboard score
export async function updateLeaderboard(userId: string, metric: 'scans' | 'accuracy', value: number) {
  await redis.zadd(`leaderboard:${metric}`, value, userId);
  await redis.expire(`leaderboard:${metric}`, 86400 * 30); // 30-day TTL
}

// Get top 100
export async function getLeaderboard(metric: string, limit = 100) {
  const results = await redis.zrevrange(`leaderboard:${metric}`, 0, limit - 1, 'WITHSCORES');
  return results.map((item, i) => ({
    rank: Math.floor(i / 2) + 1,
    userId: results[i * 2],
    score: parseInt(results[i * 2 + 1]),
  }));
}
```

## Scam Bounty Program

```typescript
// Submit scam report
export async function submitScamReport(userId: string, tokenAddress: string, evidence: string[]) {
  const report = await scamReportRepository.create({
    userId,
    tokenAddress,
    evidence,
    status: 'pending',
  });

  // Review by moderators
  await moderationQueue.add('review-scam-report', { reportId: report.id });
}

// Award bounty
export async function awardBounty(reportId: string, severity: 'low' | 'medium' | 'high' | 'critical') {
  const bountyAmounts = { low: 100, medium: 500, high: 2000, critical: 10000 };
  const amount = bountyAmounts[severity];

  await tokenService.transfer({
    from: BOUNTY_POOL_WALLET,
    to: report.userId,
    amount: amount * 1e9, // Convert to lamports
    memo: `Scam Bounty: ${reportId}`,
  });

  await userRepository.increment(report.userId, 'totalBountiesEarned', amount);
}
```
