# EVM Multi-Chain Integration Patterns

## Ethers.js/Viem Provider Setup

```typescript
import { createPublicClient, http } from 'viem';
import { mainnet, base, bsc, polygon } from 'viem/chains';

export const clients = {
  ethereum: createPublicClient({ chain: mainnet, transport: http(process.env.ETH_RPC_URL) }),
  base: createPublicClient({ chain: base, transport: http(process.env.BASE_RPC_URL) }),
  bsc: createPublicClient({ chain: bsc, transport: http(process.env.BSC_RPC_URL) }),
  polygon: createPublicClient({ chain: polygon, transport: http(process.env.POLYGON_RPC_URL) }),
};

// Multi-chain token analysis
export async function analyzeERC20Token(tokenAddress: string, chainId: number) {
  const client = getClientForChain(chainId);

  const [name, symbol, decimals, totalSupply] = await Promise.all([
    client.readContract({ address: tokenAddress, abi: ERC20_ABI, functionName: 'name' }),
    client.readContract({ address: tokenAddress, abi: ERC20_ABI, functionName: 'symbol' }),
    client.readContract({ address: tokenAddress, abi: ERC20_ABI, functionName: 'decimals' }),
    client.readContract({ address: tokenAddress, abi: ERC20_ABI, functionName: 'totalSupply' }),
  ]);

  return { name, symbol, decimals, totalSupply: totalSupply.toString() };
}
```

## Uniswap V2/V3 Liquidity Analysis

```typescript
export async function getUniswapLiquidity(pairAddress: string, chainId: number) {
  const client = getClientForChain(chainId);

  const [reserves, token0, token1] = await Promise.all([
    client.readContract({ address: pairAddress, abi: UNISWAP_PAIR_ABI, functionName: 'getReserves' }),
    client.readContract({ address: pairAddress, abi: UNISWAP_PAIR_ABI, functionName: 'token0' }),
    client.readContract({ address: pairAddress, abi: UNISWAP_PAIR_ABI, functionName: 'token1' }),
  ]);

  return {
    reserve0: reserves[0].toString(),
    reserve1: reserves[1].toString(),
    token0,
    token1,
  };
}
```

## Honeypot Detection (EVM-specific)

```typescript
// Detect buy/sell tax difference
export async function detectHoneypot(tokenAddress: string, chainId: number) {
  const client = getClientForChain(chainId);

  // Simulate buy
  const buyTax = await simulateTrade(tokenAddress, 'buy', client);

  // Simulate sell
  const sellTax = await simulateTrade(tokenAddress, 'sell', client);

  const taxDifference = Math.abs(sellTax - buyTax);

  return {
    buyTax,
    sellTax,
    taxDifference,
    isHoneypot: taxDifference > 10, // >10% difference
  };
}
```
