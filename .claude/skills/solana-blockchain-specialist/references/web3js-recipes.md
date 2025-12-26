# Solana Web3.js Recipes & Patterns

> Reference material for solana-blockchain-specialist skill
> Links to: `docs/03-TECHNICAL/integrations/blockchain-api-integration.md`

## Pattern 1: Connection Management & RPC Configuration

### Helius RPC Setup (Production)

```typescript
// src/shared/blockchain/connection.ts
import { Connection, ConnectionConfig } from '@solana/web3.js';
import { logger } from '@/shared/logger';

const RPC_CONFIG: ConnectionConfig = {
  commitment: 'confirmed', // Balance between speed and finality
  confirmTransactionInitialTimeout: 60000, // 60 seconds
  wsEndpoint: process.env.HELIUS_WS_URL, // WebSocket for real-time updates
};

export const connection = new Connection(
  process.env.HELIUS_RPC_URL!,
  RPC_CONFIG
);

// Health check
export async function checkRpcHealth(): Promise<boolean> {
  try {
    const version = await connection.getVersion();
    logger.info({ version }, 'RPC connection healthy');
    return true;
  } catch (error) {
    logger.error({ error }, 'RPC connection failed');
    return false;
  }
}

// Get slot with retry
export async function getCurrentSlot(retries = 3): Promise<number> {
  for (let i = 0; i < retries; i++) {
    try {
      return await connection.getSlot();
    } catch (error) {
      if (i === retries - 1) throw error;
      logger.warn({ attempt: i + 1, error }, 'Slot fetch failed, retrying...');
      await new Promise(resolve => setTimeout(resolve, 1000 * (i + 1)));
    }
  }
  throw new Error('Failed to get current slot after retries');
}
```

### Connection Pooling for High Throughput

```typescript
// src/shared/blockchain/connection-pool.ts
import { Connection } from '@solana/web3.js';

class ConnectionPool {
  private connections: Connection[] = [];
  private currentIndex = 0;

  constructor(rpcUrls: string[], config: ConnectionConfig) {
    this.connections = rpcUrls.map(url => new Connection(url, config));
  }

  // Round-robin load balancing
  getConnection(): Connection {
    const conn = this.connections[this.currentIndex];
    this.currentIndex = (this.currentIndex + 1) % this.connections.length;
    return conn;
  }

  // Execute with automatic failover
  async execute<T>(
    fn: (connection: Connection) => Promise<T>,
    retries = this.connections.length
  ): Promise<T> {
    let lastError: Error | undefined;

    for (let i = 0; i < retries; i++) {
      const conn = this.getConnection();
      try {
        return await fn(conn);
      } catch (error) {
        lastError = error as Error;
        logger.warn({ error, attempt: i + 1 }, 'RPC call failed, trying next connection');
      }
    }

    throw new Error(`All RPC connections failed: ${lastError?.message}`);
  }
}

// Export singleton pool
export const connectionPool = new ConnectionPool(
  [
    process.env.HELIUS_RPC_URL!,
    process.env.HELIUS_RPC_URL_BACKUP!,
    process.env.QUICKNODE_RPC_URL!, // Fallback provider
  ],
  { commitment: 'confirmed' }
);
```

---

## Pattern 2: Token Metadata Fetching

### Basic Token Info (Mint, Supply, Decimals)

```typescript
// src/modules/blockchain-api/token-metadata.service.ts
import { PublicKey } from '@solana/web3.js';
import { getMint, getAccount } from '@solana/spl-token';
import { connection } from '@/shared/blockchain/connection';
import { logger } from '@/shared/logger';
import { metrics } from '@/shared/metrics';

export interface TokenMetadata {
  mint: string;
  decimals: number;
  supply: string;
  mintAuthority: string | null;
  freezeAuthority: string | null;
  isInitialized: boolean;
}

export async function getTokenMetadata(
  mintAddress: string
): Promise<TokenMetadata> {
  const startTime = Date.now();

  try {
    const mintPubkey = new PublicKey(mintAddress);
    const mintInfo = await getMint(connection, mintPubkey);

    const metadata: TokenMetadata = {
      mint: mintAddress,
      decimals: mintInfo.decimals,
      supply: mintInfo.supply.toString(),
      mintAuthority: mintInfo.mintAuthority?.toBase58() || null,
      freezeAuthority: mintInfo.freezeAuthority?.toBase58() || null,
      isInitialized: mintInfo.isInitialized,
    };

    const duration = Date.now() - startTime;
    metrics.timing('blockchain.get_token_metadata.duration', duration);
    logger.info({ mintAddress, duration }, 'Token metadata fetched');

    return metadata;
  } catch (error) {
    metrics.increment('blockchain.get_token_metadata.error', 1);
    logger.error({ error, mintAddress }, 'Failed to fetch token metadata');
    throw new Error(`Failed to fetch token metadata: ${error.message}`);
  }
}
```

### Metaplex Metadata (Name, Symbol, URI)

```typescript
// src/modules/blockchain-api/metaplex-metadata.service.ts
import { Metaplex } from '@metaplex-foundation/js';
import { connection } from '@/shared/blockchain/connection';
import { PublicKey } from '@solana/web3.js';

const metaplex = Metaplex.make(connection);

export interface MetaplexMetadata {
  name: string;
  symbol: string;
  uri: string;
  isMutable: boolean;
  updateAuthority: string;
  creators: Array<{ address: string; verified: boolean; share: number }>;
}

export async function getMetaplexMetadata(
  mintAddress: string
): Promise<MetaplexMetadata | null> {
  try {
    const mintPubkey = new PublicKey(mintAddress);
    const nft = await metaplex.nfts().findByMint({ mintAddress: mintPubkey });

    if (!nft || !nft.json) {
      logger.warn({ mintAddress }, 'No Metaplex metadata found');
      return null;
    }

    return {
      name: nft.name,
      symbol: nft.symbol,
      uri: nft.uri,
      isMutable: nft.isMutable,
      updateAuthority: nft.updateAuthorityAddress.toBase58(),
      creators: nft.creators.map(c => ({
        address: c.address.toBase58(),
        verified: c.verified,
        share: c.share,
      })),
    };
  } catch (error) {
    logger.error({ error, mintAddress }, 'Failed to fetch Metaplex metadata');
    return null;
  }
}

// Fetch off-chain JSON metadata
export async function fetchOffChainMetadata(uri: string): Promise<any> {
  try {
    const response = await fetch(uri);
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }
    return await response.json();
  } catch (error) {
    logger.warn({ error, uri }, 'Failed to fetch off-chain metadata');
    return null;
  }
}
```

---

## Pattern 3: Authority Checking (Mint, Freeze, Update)

### Check Mint & Freeze Authorities

```typescript
// src/modules/blockchain-api/authority-checker.service.ts
import { PublicKey } from '@solana/web3.js';
import { getMint } from '@solana/spl-token';
import { connection } from '@/shared/blockchain/connection';

export interface AuthorityStatus {
  mintAuthority: {
    isActive: boolean;
    address: string | null;
  };
  freezeAuthority: {
    isActive: boolean;
    address: string | null;
  };
  updateAuthority: {
    isActive: boolean;
    address: string | null;
  };
}

export async function checkAuthorities(
  mintAddress: string
): Promise<AuthorityStatus> {
  const mintPubkey = new PublicKey(mintAddress);
  const mintInfo = await getMint(connection, mintPubkey);

  // Mint authority allows creating more tokens (inflation risk)
  const mintAuthority = mintInfo.mintAuthority
    ? {
        isActive: true,
        address: mintInfo.mintAuthority.toBase58(),
      }
    : {
        isActive: false,
        address: null,
      };

  // Freeze authority allows freezing token accounts (honeypot risk)
  const freezeAuthority = mintInfo.freezeAuthority
    ? {
        isActive: true,
        address: mintInfo.freezeAuthority.toBase58(),
      }
    : {
        isActive: false,
        address: null,
      };

  // Update authority (from Metaplex metadata)
  const metaplexMetadata = await getMetaplexMetadata(mintAddress);
  const updateAuthority = metaplexMetadata
    ? {
        isActive: true,
        address: metaplexMetadata.updateAuthority,
      }
    : {
        isActive: false,
        address: null,
      };

  return { mintAuthority, freezeAuthority, updateAuthority };
}

// Calculate risk score based on authorities
export function calculateAuthorityRisk(status: AuthorityStatus): {
  score: number;
  redFlags: string[];
} {
  const redFlags: string[] = [];
  let score = 0;

  if (status.mintAuthority.isActive) {
    score += 30; // 30 points (HIGH RISK)
    redFlags.push('Mint authority is active - creator can mint unlimited tokens');
  }

  if (status.freezeAuthority.isActive) {
    score += 20; // 20 points (MEDIUM-HIGH RISK)
    redFlags.push('Freeze authority is active - creator can freeze user wallets');
  }

  if (status.updateAuthority.isActive) {
    score += 5; // 5 points (LOW RISK)
    redFlags.push('Update authority is active - metadata can be changed');
  }

  return { score, redFlags };
}
```

---

## Pattern 4: Holder Distribution Analysis

### Get Token Accounts (Top Holders)

```typescript
// src/modules/blockchain-api/holder-distribution.service.ts
import { PublicKey } from '@solana/web3.js';
import { getAssociatedTokenAddress, getAccount } from '@solana/spl-token';
import { connection } from '@/shared/blockchain/connection';

export interface HolderDistribution {
  totalHolders: number;
  top10HoldersPercent: number;
  top10Holders: Array<{
    address: string;
    balance: string;
    percent: number;
  }>;
  concentration: 'LOW' | 'MEDIUM' | 'HIGH';
}

export async function getHolderDistribution(
  mintAddress: string
): Promise<HolderDistribution> {
  const mintPubkey = new PublicKey(mintAddress);

  // Get all token accounts for this mint
  const accounts = await connection.getParsedProgramAccounts(
    new PublicKey('TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA'), // SPL Token program
    {
      filters: [
        {
          dataSize: 165, // Token account size
        },
        {
          memcmp: {
            offset: 0,
            bytes: mintPubkey.toBase58(), // Filter by mint address
          },
        },
      ],
    }
  );

  // Parse balances
  const holders = accounts
    .map(acc => {
      const parsed = acc.account.data.parsed.info;
      return {
        address: parsed.owner,
        balance: parsed.tokenAmount.uiAmount,
      };
    })
    .filter(h => h.balance > 0) // Exclude empty accounts
    .sort((a, b) => b.balance - a.balance); // Sort by balance desc

  const totalSupply = holders.reduce((sum, h) => sum + h.balance, 0);

  // Calculate top 10 holders percentage
  const top10Holders = holders.slice(0, 10);
  const top10Balance = top10Holders.reduce((sum, h) => sum + h.balance, 0);
  const top10Percent = (top10Balance / totalSupply) * 100;

  // Determine concentration level
  let concentration: 'LOW' | 'MEDIUM' | 'HIGH';
  if (top10Percent > 50) {
    concentration = 'HIGH'; // HIGH RISK
  } else if (top10Percent > 30) {
    concentration = 'MEDIUM'; // MEDIUM RISK
  } else {
    concentration = 'LOW'; // LOW RISK
  }

  return {
    totalHolders: holders.length,
    top10HoldersPercent: top10Percent,
    top10Holders: top10Holders.map(h => ({
      address: h.address,
      balance: h.balance.toString(),
      percent: (h.balance / totalSupply) * 100,
    })),
    concentration,
  };
}
```

---

## Pattern 5: Liquidity Pool Analysis (Raydium, Orca)

### Get Raydium Pool Info

```typescript
// src/modules/blockchain-api/raydium-pools.service.ts
import { PublicKey } from '@solana/web3.js';
import { connection } from '@/shared/blockchain/connection';
import { getAccount } from '@solana/spl-token';

export interface RaydiumPool {
  poolAddress: string;
  tokenAMint: string;
  tokenBMint: string;
  tokenAReserve: string;
  tokenBReserve: string;
  liquidityUSD: number;
}

// Raydium AMM program ID
const RAYDIUM_AMM_PROGRAM_ID = new PublicKey(
  '675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8'
);

export async function getRaydiumPools(
  tokenMint: string
): Promise<RaydiumPool[]> {
  const mintPubkey = new PublicKey(tokenMint);

  // Get all Raydium pools containing this token
  const accounts = await connection.getProgramAccounts(RAYDIUM_AMM_PROGRAM_ID, {
    filters: [
      {
        dataSize: 752, // Raydium AMM account size
      },
      {
        memcmp: {
          offset: 400, // Offset for token A mint
          bytes: mintPubkey.toBase58(),
        },
      },
    ],
  });

  const pools: RaydiumPool[] = [];

  for (const account of accounts) {
    // Parse Raydium pool data (simplified - real implementation needs borsh deserialization)
    const poolData = parseRaydiumPoolData(account.account.data);

    // Get reserves
    const tokenAAccount = await getAccount(
      connection,
      new PublicKey(poolData.tokenAVault)
    );
    const tokenBAccount = await getAccount(
      connection,
      new PublicKey(poolData.tokenBVault)
    );

    // Calculate USD liquidity (requires price oracle)
    const liquidityUSD = await calculatePoolLiquidityUSD(
      poolData.tokenAMint,
      poolData.tokenBMint,
      tokenAAccount.amount.toString(),
      tokenBAccount.amount.toString()
    );

    pools.push({
      poolAddress: account.pubkey.toBase58(),
      tokenAMint: poolData.tokenAMint,
      tokenBMint: poolData.tokenBMint,
      tokenAReserve: tokenAAccount.amount.toString(),
      tokenBReserve: tokenBAccount.amount.toString(),
      liquidityUSD,
    });
  }

  return pools;
}

// Helper to parse Raydium pool data (borsh deserialization)
function parseRaydiumPoolData(data: Buffer): any {
  // Simplified - real implementation uses @project-serum/borsh
  // See: https://github.com/raydium-io/raydium-sdk
  return {
    tokenAMint: '', // Parse from buffer
    tokenBMint: '',
    tokenAVault: '',
    tokenBVault: '',
  };
}
```

### Aggregate Liquidity Across DEXs

```typescript
// src/modules/blockchain-api/liquidity-aggregator.service.ts
export interface TotalLiquidity {
  totalUSD: number;
  sources: Array<{
    dex: 'raydium' | 'orca' | 'serum';
    pools: number;
    liquidityUSD: number;
  }>;
  riskLevel: 'LOW' | 'MEDIUM' | 'HIGH';
}

export async function getTotalLiquidity(
  tokenMint: string
): Promise<TotalLiquidity> {
  const [raydiumPools, orcaPools] = await Promise.all([
    getRaydiumPools(tokenMint),
    getOrcaPools(tokenMint), // Similar implementation
  ]);

  const raydiumLiquidity = raydiumPools.reduce((sum, p) => sum + p.liquidityUSD, 0);
  const orcaLiquidity = orcaPools.reduce((sum, p) => sum + p.liquidityUSD, 0);
  const totalUSD = raydiumLiquidity + orcaLiquidity;

  // Determine risk level based on total liquidity
  let riskLevel: 'LOW' | 'MEDIUM' | 'HIGH';
  if (totalUSD < 10_000) {
    riskLevel = 'HIGH'; // <$10K = HIGH RISK (easy to rug)
  } else if (totalUSD < 50_000) {
    riskLevel = 'MEDIUM'; // $10K-$50K = MEDIUM RISK
  } else {
    riskLevel = 'LOW'; // >$50K = LOW RISK
  }

  return {
    totalUSD,
    sources: [
      { dex: 'raydium', pools: raydiumPools.length, liquidityUSD: raydiumLiquidity },
      { dex: 'orca', pools: orcaPools.length, liquidityUSD: orcaLiquidity },
    ],
    riskLevel,
  };
}
```

---

## Pattern 6: Transaction History Analysis

### Get Recent Token Transfers

```typescript
// src/modules/blockchain-api/transaction-history.service.ts
import { PublicKey, ParsedTransactionWithMeta } from '@solana/web3.js';
import { connection } from '@/shared/blockchain/connection';

export interface TokenTransfer {
  signature: string;
  blockTime: number;
  from: string;
  to: string;
  amount: string;
  type: 'transfer' | 'mint' | 'burn';
}

export async function getRecentTransfers(
  tokenMint: string,
  limit = 100
): Promise<TokenTransfer[]> {
  const mintPubkey = new PublicKey(tokenMint);

  // Get recent signatures for this token
  const signatures = await connection.getSignaturesForAddress(mintPubkey, {
    limit,
  });

  // Fetch full transactions
  const transactions = await connection.getParsedTransactions(
    signatures.map(s => s.signature),
    { maxSupportedTransactionVersion: 0 }
  );

  const transfers: TokenTransfer[] = [];

  for (const tx of transactions) {
    if (!tx || !tx.meta) continue;

    // Parse SPL token transfers from transaction
    const tokenTransfers = parseTokenTransfers(tx, tokenMint);
    transfers.push(...tokenTransfers);
  }

  return transfers;
}

function parseTokenTransfers(
  tx: ParsedTransactionWithMeta,
  tokenMint: string
): TokenTransfer[] {
  const transfers: TokenTransfer[] = [];

  for (const instruction of tx.transaction.message.instructions) {
    if (
      'parsed' in instruction &&
      instruction.program === 'spl-token' &&
      instruction.parsed.type === 'transfer'
    ) {
      const info = instruction.parsed.info;

      if (info.mint === tokenMint) {
        transfers.push({
          signature: tx.transaction.signatures[0],
          blockTime: tx.blockTime || 0,
          from: info.source,
          to: info.destination,
          amount: info.amount,
          type: 'transfer',
        });
      }
    }
  }

  return transfers;
}

// Analyze transfer patterns for suspicious activity
export function analyzeTransferPatterns(transfers: TokenTransfer[]): {
  suspiciousPatterns: string[];
  score: number;
} {
  const suspiciousPatterns: string[] = [];
  let score = 0;

  // Check for rapid consecutive transfers (potential bot activity)
  const rapidTransfers = transfers.filter((t, i) => {
    if (i === 0) return false;
    const prevTime = transfers[i - 1].blockTime;
    return t.blockTime - prevTime < 5; // <5 seconds between transfers
  });

  if (rapidTransfers.length > 10) {
    suspiciousPatterns.push('High frequency bot-like transfer activity detected');
    score += 15;
  }

  // Check for large single-holder dumps
  const totalVolume = transfers.reduce((sum, t) => sum + Number(t.amount), 0);
  const largeTransfers = transfers.filter(t => Number(t.amount) > totalVolume * 0.1);

  if (largeTransfers.length > 0) {
    suspiciousPatterns.push(
      `Large transfers detected (>10% of volume): ${largeTransfers.length}`
    );
    score += 10;
  }

  return { suspiciousPatterns, score };
}
```

---

## Pattern 7: Parallel RPC Calls for Performance

### Batch Multiple Calls

```typescript
// src/shared/blockchain/batch-fetcher.ts
import { PublicKey } from '@solana/web3.js';
import { connection } from './connection';
import { getMint } from '@solana/spl-token';

export async function batchFetchTokenData(
  tokenMints: string[]
): Promise<Map<string, TokenMetadata>> {
  const results = new Map<string, TokenMetadata>();

  // Execute all calls in parallel
  const promises = tokenMints.map(async mint => {
    try {
      const mintPubkey = new PublicKey(mint);
      const mintInfo = await getMint(connection, mintPubkey);

      return {
        mint,
        data: {
          mint,
          decimals: mintInfo.decimals,
          supply: mintInfo.supply.toString(),
          mintAuthority: mintInfo.mintAuthority?.toBase58() || null,
          freezeAuthority: mintInfo.freezeAuthority?.toBase58() || null,
          isInitialized: mintInfo.isInitialized,
        },
      };
    } catch (error) {
      logger.warn({ error, mint }, 'Failed to fetch token data');
      return { mint, data: null };
    }
  });

  const settled = await Promise.allSettled(promises);

  for (const result of settled) {
    if (result.status === 'fulfilled' && result.value.data) {
      results.set(result.value.mint, result.value.data);
    }
  }

  return results;
}

// Batch get multiple accounts
export async function batchGetAccounts(
  publicKeys: PublicKey[]
): Promise<(AccountInfo<Buffer> | null)[]> {
  // Web3.js supports batching up to 100 accounts
  const BATCH_SIZE = 100;
  const batches: PublicKey[][] = [];

  for (let i = 0; i < publicKeys.length; i += BATCH_SIZE) {
    batches.push(publicKeys.slice(i, i + BATCH_SIZE));
  }

  const results = await Promise.all(
    batches.map(batch => connection.getMultipleAccountsInfo(batch))
  );

  return results.flat();
}
```

---

## Pattern 8: Caching Blockchain Data

```typescript
// src/shared/blockchain/cache.ts
import { Redis } from 'ioredis';
import { logger } from '@/shared/logger';

const redis = new Redis(process.env.REDIS_URL!);

export async function getCachedTokenMetadata(
  mintAddress: string
): Promise<TokenMetadata | null> {
  const cacheKey = `token:metadata:${mintAddress}`;

  try {
    const cached = await redis.get(cacheKey);
    if (cached) {
      metrics.increment('blockchain.cache.hit', 1, { type: 'token_metadata' });
      return JSON.parse(cached);
    }

    metrics.increment('blockchain.cache.miss', 1, { type: 'token_metadata' });

    // Fetch from blockchain
    const metadata = await getTokenMetadata(mintAddress);

    // Cache for 1 hour (token metadata rarely changes)
    await redis.setex(cacheKey, 3600, JSON.stringify(metadata));

    return metadata;
  } catch (error) {
    logger.error({ error, mintAddress }, 'Cache error');
    // Fallback to direct fetch
    return await getTokenMetadata(mintAddress);
  }
}

// Different TTLs for different data types
export const CACHE_TTLS = {
  TOKEN_METADATA: 3600, // 1 hour
  LIQUIDITY: 300, // 5 minutes (changes frequently)
  HOLDER_DISTRIBUTION: 600, // 10 minutes
  CREATOR_HISTORY: 86400, // 24 hours (static)
  AUTHORITIES: 3600, // 1 hour
};
```

---

## Related Files

- **Docs**: `docs/03-TECHNICAL/integrations/blockchain-api-integration.md` - Full API integration guide
- **Skill**: `.claude/skills/solana-blockchain-specialist/SKILL.md` - Main skill definition
- **Agent**: `.claude/agents/risk-scoring-agent.md` - Uses these patterns for risk analysis
