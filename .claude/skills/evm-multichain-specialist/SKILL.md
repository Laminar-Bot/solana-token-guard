---
name: evm-multichain-specialist
description: "Expert EVM blockchain developer specializing in multi-chain token analysis across Ethereum, Base, BSC, and Polygon. Deep knowledge of Ethers.js/Viem, ERC-20 tokens, Uniswap/SushiSwap liquidity pools, EVM-specific scam patterns, and cross-chain architecture for CryptoRugMunch's expansion beyond Solana."
---

# EVM & Multi-Chain Specialist

**Role**: Expert blockchain developer for EVM-compatible chains (Ethereum, Base, BSC, Polygon) with deep knowledge of smart contract analysis, ERC-20 token standards, DEX liquidity pools, and multi-chain architecture patterns.

**Context**: CryptoRugMunch launches on **Solana (MVP)** but plans to expand to **EVM chains** (Ethereum, Base, BSC, Polygon) in **Month 4-6** post-launch. This skill provides expertise for implementing multi-chain token analysis with chain-specific scam detection.

---

## Core Philosophy

1. **Chain-Agnostic Architecture**: Design systems that work across multiple blockchains with minimal code duplication
2. **EVM-Specific Patterns**: Understand differences between Solana and EVM (accounts vs UTXOs, gas vs compute units, program derived addresses vs contract addresses)
3. **Liquidity Pool Analysis**: Master Uniswap V2/V3, SushiSwap, PancakeSwap pool mechanics for scam detection
4. **Gas Optimization**: Write efficient smart contracts and minimize RPC calls for cost-effective multi-chain operations
5. **Security-First**: Validate all on-chain data, handle reorgs, protect against sandwich attacks and MEV exploitation

---

## 1. EVM Blockchain Fundamentals

### 1.1 EVM vs Solana Comparison

| Aspect | Solana | EVM (Ethereum/Base/BSC) |
|--------|--------|-------------------------|
| **Account Model** | Account-based (rent-exempt) | Account-based (state stored in contracts) |
| **Transaction Cost** | ~0.00001 SOL (compute units) | Variable (gas price × gas used) |
| **Finality** | ~400ms (PoH) | 12-15s (Ethereum), 2s (BSC/Base) |
| **RPC Calls** | Free (Helius generous limits) | Expensive (Infura/Alchemy quotas) |
| **Token Standard** | SPL Token (program-based) | ERC-20 (smart contract) |
| **Liquidity Pools** | Raydium, Orca (AMM programs) | Uniswap, SushiSwap (smart contracts) |
| **Authorities** | Mint/Freeze authority (revocable) | Owner role in contract (immutable after renounce) |
| **Upgradability** | Program upgrade authority | Proxy patterns (UUPS, Transparent, Beacon) |
| **Scam Vectors** | Mint authority dilution, freeze | Hidden minting, transfer fees, honeypots |

### 1.2 Target EVM Chains for CryptoRugMunch

#### **Ethereum Mainnet**
- **Why**: Largest liquidity, most established DeFi ecosystem
- **Challenges**: High gas costs ($5-50 per transaction), slower finality (12-15s)
- **RPC Providers**: Infura, Alchemy, QuickNode
- **DEXs**: Uniswap V2/V3, SushiSwap, Curve
- **Use Case**: High-value tokens, established projects

#### **Base (Coinbase L2)**
- **Why**: Low fees (~$0.01), fast finality (~2s), growing ecosystem
- **Challenges**: Newer chain, fewer scam databases
- **RPC Providers**: Base RPC (free), Alchemy, QuickNode
- **DEXs**: Uniswap V3 (official), BaseSwap, Aerodrome
- **Use Case**: Emerging memecoins, L2 DeFi

#### **BSC (Binance Smart Chain)**
- **Why**: Most scam activity, low fees (~$0.10), fast finality (~3s)
- **Challenges**: Centralized (21 validators), high scam volume
- **RPC Providers**: BSC RPC (free), Ankr, NodeReal
- **DEXs**: PancakeSwap V2/V3, BiSwap
- **Use Case**: High scam detection demand, memecoins

#### **Polygon (PoS)**
- **Why**: Low fees (~$0.01), fast finality (~2s), Ethereum-compatible
- **Challenges**: Occasional congestion, bridge risks
- **RPC Providers**: Polygon RPC (free), Alchemy, Infura
- **DEXs**: QuickSwap, SushiSwap
- **Use Case**: Gaming tokens, NFT projects

---

## 2. Ethers.js & Viem for EVM Interactions

### 2.1 Ethers.js (v6) Setup

CryptoRugMunch uses **Ethers.js v6** for EVM chains (familiar API, TypeScript support).

```typescript
// src/blockchain/evm/providers.ts
import { ethers, JsonRpcProvider, FallbackProvider } from 'ethers';

// Multi-provider setup with fallback (critical for uptime)
export function getEthereumProvider(): FallbackProvider {
  const providers = [
    new JsonRpcProvider(process.env.INFURA_ETHEREUM_RPC!, 1, {
      staticNetwork: ethers.Network.from('mainnet')
    }),
    new JsonRpcProvider(process.env.ALCHEMY_ETHEREUM_RPC!, 1, {
      staticNetwork: ethers.Network.from('mainnet')
    }),
    new JsonRpcProvider(process.env.QUICKNODE_ETHEREUM_RPC!, 1, {
      staticNetwork: ethers.Network.from('mainnet')
    }),
  ];

  return new FallbackProvider(providers, 1, {
    cacheTimeout: 5000, // 5s cache for static data
  });
}

export function getBaseProvider(): JsonRpcProvider {
  return new JsonRpcProvider(process.env.BASE_RPC!, 8453, {
    staticNetwork: ethers.Network.from('base'),
  });
}

export function getBscProvider(): JsonRpcProvider {
  return new JsonRpcProvider(process.env.BSC_RPC!, 56, {
    staticNetwork: ethers.Network.from('bsc'),
  });
}

export function getPolygonProvider(): JsonRpcProvider {
  return new JsonRpcProvider(process.env.POLYGON_RPC!, 137, {
    staticNetwork: ethers.Network.from('matic'),
  });
}

// Factory function for chain-agnostic access
export function getProviderForChain(chain: 'ethereum' | 'base' | 'bsc' | 'polygon'): JsonRpcProvider | FallbackProvider {
  switch (chain) {
    case 'ethereum': return getEthereumProvider();
    case 'base': return getBaseProvider();
    case 'bsc': return getBscProvider();
    case 'polygon': return getPolygonProvider();
    default: throw new Error(`Unsupported chain: ${chain}`);
  }
}
```

### 2.2 ERC-20 Token Contract Interface

```typescript
// src/blockchain/evm/erc20.ts
import { ethers, Contract } from 'ethers';

// Minimal ERC-20 ABI (only what we need for scam detection)
const ERC20_ABI = [
  'function name() view returns (string)',
  'function symbol() view returns (string)',
  'function decimals() view returns (uint8)',
  'function totalSupply() view returns (uint256)',
  'function balanceOf(address owner) view returns (uint256)',
  'function owner() view returns (address)', // Not standard, but common
  'function renounceOwnership() external', // Ownable pattern
  'function transfer(address to, uint256 amount) returns (bool)',
  'function allowance(address owner, address spender) view returns (uint256)',
  'event Transfer(address indexed from, address indexed to, uint256 value)',
  'event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)',
];

export interface ERC20Metadata {
  address: string;
  name: string;
  symbol: string;
  decimals: number;
  totalSupply: bigint;
  owner: string | null; // null if ownership renounced or no owner() function
}

export async function getERC20Metadata(
  tokenAddress: string,
  provider: JsonRpcProvider
): Promise<ERC20Metadata> {
  const contract = new Contract(tokenAddress, ERC20_ABI, provider);

  const [name, symbol, decimals, totalSupply] = await Promise.all([
    contract.name(),
    contract.symbol(),
    contract.decimals(),
    contract.totalSupply(),
  ]);

  // Try to get owner (may not exist or may throw)
  let owner: string | null = null;
  try {
    owner = await contract.owner();
    // Check if ownership renounced (owner = 0x0000...0000)
    if (owner === ethers.ZeroAddress) {
      owner = null;
    }
  } catch (error) {
    // owner() function doesn't exist (not Ownable)
    owner = null;
  }

  return {
    address: tokenAddress,
    name,
    symbol,
    decimals: Number(decimals),
    totalSupply,
    owner,
  };
}
```

### 2.3 Viem Alternative (Optional)

**Viem** is a modern, TypeScript-first alternative to Ethers.js with better tree-shaking and type safety.

```typescript
// Alternative: Using Viem instead of Ethers.js
import { createPublicClient, http, parseAbi } from 'viem';
import { mainnet, base, bsc, polygon } from 'viem/chains';

export const ethereumClient = createPublicClient({
  chain: mainnet,
  transport: http(process.env.INFURA_ETHEREUM_RPC),
});

export const baseClient = createPublicClient({
  chain: base,
  transport: http(process.env.BASE_RPC),
});

// Read ERC-20 data with Viem
const erc20Abi = parseAbi([
  'function name() view returns (string)',
  'function symbol() view returns (string)',
  'function decimals() view returns (uint8)',
  'function totalSupply() view returns (uint256)',
]);

const [name, symbol, decimals, totalSupply] = await ethereumClient.multicall({
  contracts: [
    { address: tokenAddress, abi: erc20Abi, functionName: 'name' },
    { address: tokenAddress, abi: erc20Abi, functionName: 'symbol' },
    { address: tokenAddress, abi: erc20Abi, functionName: 'decimals' },
    { address: tokenAddress, abi: erc20Abi, functionName: 'totalSupply' },
  ],
});
```

**Decision**: CryptoRugMunch uses **Ethers.js v6** initially for familiarity, but Viem can be adopted later for performance.

---

## 3. ERC-20 Token Analysis

### 3.1 Smart Contract Source Code Verification

**Critical Difference from Solana**: EVM contracts can be verified on block explorers (Etherscan, BscScan) to inspect source code.

```typescript
// src/blockchain/evm/contract-verification.ts
import axios from 'axios';

export interface VerificationResult {
  isVerified: boolean;
  sourceCode: string | null;
  contractName: string | null;
  compilerVersion: string | null;
  optimizationEnabled: boolean;
  runs: number;
  constructorArguments: string | null;
  evmVersion: string | null;
  license: string | null;
}

export async function getContractSource(
  tokenAddress: string,
  chain: 'ethereum' | 'bsc' | 'polygon' | 'base'
): Promise<VerificationResult> {
  const apiKeys = {
    ethereum: process.env.ETHERSCAN_API_KEY!,
    bsc: process.env.BSCSCAN_API_KEY!,
    polygon: process.env.POLYGONSCAN_API_KEY!,
    base: process.env.BASESCAN_API_KEY!,
  };

  const apiUrls = {
    ethereum: 'https://api.etherscan.io/api',
    bsc: 'https://api.bscscan.com/api',
    polygon: 'https://api.polygonscan.com/api',
    base: 'https://api.basescan.org/api',
  };

  const response = await axios.get(apiUrls[chain], {
    params: {
      module: 'contract',
      action: 'getsourcecode',
      address: tokenAddress,
      apikey: apiKeys[chain],
    },
  });

  const result = response.data.result[0];

  if (result.SourceCode === '') {
    return {
      isVerified: false,
      sourceCode: null,
      contractName: null,
      compilerVersion: null,
      optimizationEnabled: false,
      runs: 0,
      constructorArguments: null,
      evmVersion: null,
      license: null,
    };
  }

  return {
    isVerified: true,
    sourceCode: result.SourceCode,
    contractName: result.ContractName,
    compilerVersion: result.CompilerVersion,
    optimizationEnabled: result.OptimizationUsed === '1',
    runs: parseInt(result.Runs, 10),
    constructorArguments: result.ConstructorArguments,
    evmVersion: result.EVMVersion,
    license: result.LicenseType,
  };
}
```

### 3.2 Detecting EVM-Specific Scam Patterns

EVM contracts have **different scam vectors** than Solana:

| Scam Pattern | Solana | EVM |
|--------------|--------|-----|
| **Mint Authority Dilution** | ✅ Common (unlimited minting) | ✅ Common (hidden mint function) |
| **Freeze Authority** | ✅ Can freeze wallets | ❌ Rare (requires custom code) |
| **Honeypot (can't sell)** | ⚠️ Less common | ✅ Very common (transfer restrictions) |
| **Transfer Tax/Fee** | ❌ Not possible | ✅ Common (10-99% sell tax) |
| **Ownership Renounce** | ✅ Via authority revocation | ✅ Via renounceOwnership() |
| **Proxy/Upgradable** | ✅ Via upgrade authority | ✅ Via proxy patterns (UUPS, Transparent) |
| **Liquidity Lock** | ✅ Time-locked LP tokens | ✅ Time-locked LP tokens (Unicrypt, Team.Finance) |

#### 3.2.1 Honeypot Detection (Transfer Restrictions)

**Common Pattern**: Contract allows buying but blocks selling via `require()` checks.

```solidity
// Example honeypot contract (malicious)
contract HoneypotToken is ERC20 {
    address public owner;
    bool public tradingEnabled = false;

    modifier onlyOwner() {
        require(msg.sender == owner, "Not owner");
        _;
    }

    function transfer(address to, uint256 amount) public override returns (bool) {
        // Only owner can sell before trading enabled
        require(tradingEnabled || msg.sender == owner, "Trading not enabled");
        return super.transfer(to, amount);
    }

    function enableTrading() external onlyOwner {
        tradingEnabled = true;
    }
}
```

**Detection Strategy**: Simulate a sell transaction and check if it reverts.

```typescript
// src/blockchain/evm/honeypot-detection.ts
import { ethers, Contract } from 'ethers';

const ROUTER_ABI = [
  'function swapExactTokensForETH(uint amountIn, uint amountOutMin, address[] calldata path, address to, uint deadline) external returns (uint[] memory amounts)',
];

export interface HoneypotResult {
  isHoneypot: boolean;
  canBuy: boolean;
  canSell: boolean;
  buyTax: number; // 0-100%
  sellTax: number; // 0-100%
  error: string | null;
}

export async function detectHoneypot(
  tokenAddress: string,
  chain: 'ethereum' | 'base' | 'bsc' | 'polygon',
  provider: JsonRpcProvider
): Promise<HoneypotResult> {
  // Use honeypot.is API (faster than manual simulation)
  const honeypotAPIs = {
    ethereum: 'https://api.honeypot.is/v2/IsHoneypot',
    bsc: 'https://api.honeypot.is/v2/IsHoneypot',
    polygon: 'https://api.honeypot.is/v2/IsHoneypot',
    base: null, // Not supported yet, fallback to manual
  };

  if (honeypotAPIs[chain]) {
    try {
      const response = await axios.get(honeypotAPIs[chain]!, {
        params: {
          address: tokenAddress,
          chainID: chain === 'ethereum' ? 1 : chain === 'bsc' ? 56 : 137,
        },
      });

      const data = response.data;

      return {
        isHoneypot: data.isHoneypot || data.honeypotResult?.isHoneypot || false,
        canBuy: !data.isHoneypot,
        canSell: !(data.isHoneypot || data.honeypotResult?.isHoneypot),
        buyTax: data.simulationResult?.buyTax || 0,
        sellTax: data.simulationResult?.sellTax || 0,
        error: data.honeypotResult?.honeypotReason || null,
      };
    } catch (error) {
      // Fallback to manual detection
      return await manualHoneypotDetection(tokenAddress, chain, provider);
    }
  }

  // Manual detection for chains without API support
  return await manualHoneypotDetection(tokenAddress, chain, provider);
}

async function manualHoneypotDetection(
  tokenAddress: string,
  chain: 'ethereum' | 'base' | 'bsc' | 'polygon',
  provider: JsonRpcProvider
): Promise<HoneypotResult> {
  // Simulate buy and sell using callStatic (no actual transaction)
  const routers = {
    ethereum: '0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D', // Uniswap V2
    base: '0x4752ba5DBc23f44D87826276BF6Fd6b1C372aD24', // BaseSwap
    bsc: '0x10ED43C718714eb63d5aA57B78B54704E256024E', // PancakeSwap V2
    polygon: '0xa5E0829CaCEd8fFDD4De3c43696c57F7D7A678ff', // QuickSwap
  };

  const router = new Contract(routers[chain], ROUTER_ABI, provider);
  const weth = await router.WETH();

  // Try to simulate sell (swap tokens for ETH)
  try {
    await router.swapExactTokensForETH.staticCall(
      ethers.parseEther('0.01'), // 0.01 tokens
      0, // Min amount out
      [tokenAddress, weth],
      ethers.ZeroAddress, // Dummy recipient
      Math.floor(Date.now() / 1000) + 300 // 5 min deadline
    );

    // If no revert, not a honeypot
    return {
      isHoneypot: false,
      canBuy: true,
      canSell: true,
      buyTax: 0,
      sellTax: 0,
      error: null,
    };
  } catch (error: any) {
    // Check error message
    const errorMessage = error.message || error.toString();

    if (errorMessage.includes('Trading not enabled') ||
        errorMessage.includes('Transfer failed') ||
        errorMessage.includes('Blacklisted')) {
      return {
        isHoneypot: true,
        canBuy: true, // Usually can buy
        canSell: false,
        buyTax: 0,
        sellTax: 100, // Can't sell = 100% tax
        error: errorMessage,
      };
    }

    // Unknown error (could be liquidity issue, not honeypot)
    return {
      isHoneypot: false,
      canBuy: true,
      canSell: true,
      buyTax: 0,
      sellTax: 0,
      error: null,
    };
  }
}
```

#### 3.2.2 Transfer Tax Detection

**Common Pattern**: Contract charges a fee on transfers (buy/sell tax).

```solidity
// Example transfer tax contract
contract TaxToken is ERC20 {
    uint256 public buyTax = 5; // 5%
    uint256 public sellTax = 10; // 10%
    address public taxWallet;

    function _transfer(address from, address to, uint256 amount) internal override {
        uint256 tax = 0;

        // Detect buy (from = DEX pair)
        if (from == uniswapPair && to != owner) {
            tax = (amount * buyTax) / 100;
        }

        // Detect sell (to = DEX pair)
        if (to == uniswapPair && from != owner) {
            tax = (amount * sellTax) / 100;
        }

        if (tax > 0) {
            super._transfer(from, taxWallet, tax);
            amount -= tax;
        }

        super._transfer(from, to, amount);
    }
}
```

**Detection**: Simulate buy/sell and compare balance changes.

```typescript
// src/blockchain/evm/transfer-tax-detection.ts
export interface TransferTaxResult {
  hasBuyTax: boolean;
  hasSellTax: boolean;
  buyTax: number; // 0-100%
  sellTax: number; // 0-100%
}

export async function detectTransferTax(
  tokenAddress: string,
  chain: 'ethereum' | 'base' | 'bsc' | 'polygon',
  provider: JsonRpcProvider
): Promise<TransferTaxResult> {
  // Use GoPlus Security API for tax detection
  const response = await axios.get('https://api.gopluslabs.io/api/v1/token_security/' +
    (chain === 'ethereum' ? '1' : chain === 'bsc' ? '56' : chain === 'polygon' ? '137' : '8453'), {
    params: {
      contract_addresses: tokenAddress,
    },
  });

  const data = response.data.result[tokenAddress.toLowerCase()];

  if (!data) {
    return { hasBuyTax: false, hasSellTax: false, buyTax: 0, sellTax: 0 };
  }

  return {
    hasBuyTax: parseFloat(data.buy_tax || '0') > 0,
    hasSellTax: parseFloat(data.sell_tax || '0') > 0,
    buyTax: parseFloat(data.buy_tax || '0') * 100,
    sellTax: parseFloat(data.sell_tax || '0') * 100,
  };
}
```

#### 3.2.3 Hidden Mint Function Detection

**Common Pattern**: Contract has a `mint()` function callable by owner (not visible in standard ERC-20 ABI).

```solidity
// Example contract with hidden mint
contract MintableToken is ERC20, Ownable {
    function mint(address to, uint256 amount) external onlyOwner {
        _mint(to, amount);
    }
}
```

**Detection Strategy**:
1. Check verified source code for `_mint()` or `mint()` functions
2. Check if `totalSupply()` increases over time (historical data)

```typescript
// src/blockchain/evm/mint-detection.ts
export interface MintFunctionResult {
  hasMintFunction: boolean;
  isMintable: boolean;
  mintAuthority: string | null;
  totalSupplyIncreased: boolean;
}

export async function detectMintFunction(
  tokenAddress: string,
  chain: 'ethereum' | 'base' | 'bsc' | 'polygon',
  provider: JsonRpcProvider
): Promise<MintFunctionResult> {
  // Step 1: Check verified source code
  const verification = await getContractSource(tokenAddress, chain);

  let hasMintFunction = false;
  let mintAuthority: string | null = null;

  if (verification.isVerified && verification.sourceCode) {
    const sourceCode = verification.sourceCode;

    // Check for mint function
    hasMintFunction = sourceCode.includes('function mint(') ||
                      sourceCode.includes('function _mint(');

    // Check if Ownable (mint authority = owner)
    if (sourceCode.includes('Ownable') || sourceCode.includes('onlyOwner')) {
      const contract = new Contract(tokenAddress, ['function owner() view returns (address)'], provider);
      try {
        mintAuthority = await contract.owner();
        if (mintAuthority === ethers.ZeroAddress) {
          mintAuthority = null; // Ownership renounced
        }
      } catch (error) {
        // No owner function
      }
    }
  }

  // Step 2: Check if totalSupply increased (historical)
  const contract = new Contract(tokenAddress, ['function totalSupply() view returns (uint256)'], provider);
  const currentSupply = await contract.totalSupply();

  // Check supply 1000 blocks ago
  const currentBlock = await provider.getBlockNumber();
  const oldSupply = await contract.totalSupply({ blockTag: currentBlock - 1000 });

  const totalSupplyIncreased = currentSupply > oldSupply;

  return {
    hasMintFunction,
    isMintable: hasMintFunction && mintAuthority !== null,
    mintAuthority,
    totalSupplyIncreased,
  };
}
```

---

## 4. Uniswap & DEX Liquidity Pool Analysis

### 4.1 Uniswap V2 Pool Detection

**Uniswap V2** is the most common DEX for EVM tokens. Pools are created via factory contract.

```typescript
// src/blockchain/evm/uniswap-v2.ts
import { ethers, Contract } from 'ethers';

const UNISWAP_V2_FACTORY_ABI = [
  'function getPair(address tokenA, address tokenB) external view returns (address pair)',
];

const UNISWAP_V2_PAIR_ABI = [
  'function token0() external view returns (address)',
  'function token1() external view returns (address)',
  'function getReserves() external view returns (uint112 reserve0, uint112 reserve1, uint32 blockTimestampLast)',
  'function totalSupply() external view returns (uint256)',
  'function balanceOf(address owner) external view returns (uint256)',
];

export interface UniswapV2Pool {
  pairAddress: string;
  token0: string;
  token1: string;
  reserve0: bigint;
  reserve1: bigint;
  totalLPSupply: bigint;
  liquidityUSD: number;
}

export async function getUniswapV2Pool(
  tokenAddress: string,
  chain: 'ethereum' | 'base' | 'bsc' | 'polygon',
  provider: JsonRpcProvider
): Promise<UniswapV2Pool | null> {
  const factories = {
    ethereum: '0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f', // Uniswap V2
    base: '0x8909Dc15e40173Ff4699343b6eB8132c65e18eC6', // BaseSwap
    bsc: '0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73', // PancakeSwap V2
    polygon: '0x5757371414417b8C6CAad45bAeF941aBc7d3Ab32', // QuickSwap
  };

  const weth = {
    ethereum: '0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2', // WETH
    base: '0x4200000000000000000000000000000000000006', // WETH on Base
    bsc: '0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c', // WBNB
    polygon: '0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270', // WMATIC
  };

  const factory = new Contract(factories[chain], UNISWAP_V2_FACTORY_ABI, provider);

  // Get pair address
  const pairAddress = await factory.getPair(tokenAddress, weth[chain]);

  if (pairAddress === ethers.ZeroAddress) {
    return null; // No pool exists
  }

  const pair = new Contract(pairAddress, UNISWAP_V2_PAIR_ABI, provider);

  const [token0, token1, reserves, totalLPSupply] = await Promise.all([
    pair.token0(),
    pair.token1(),
    pair.getReserves(),
    pair.totalSupply(),
  ]);

  // Calculate liquidity USD (simplified, assumes WETH = $3000)
  const wethReserve = token0.toLowerCase() === weth[chain].toLowerCase()
    ? reserves.reserve0
    : reserves.reserve1;

  const liquidityUSD = Number(ethers.formatEther(wethReserve)) * 3000 * 2; // *2 for both sides

  return {
    pairAddress,
    token0,
    token1,
    reserve0: reserves.reserve0,
    reserve1: reserves.reserve1,
    totalLPSupply,
    liquidityUSD,
  };
}
```

### 4.2 Liquidity Lock Detection

**Critical for Scam Detection**: If LP tokens are not locked, dev can rug pull by removing liquidity.

**Common Lock Platforms**:
- **Ethereum**: Unicrypt, Team.Finance, PinkSale
- **BSC**: PancakeSwap Lock, Mudra, PinkLock
- **Base**: BaseSwap Lock (new)
- **Polygon**: QuickSwap Lock

```typescript
// src/blockchain/evm/liquidity-lock.ts
export interface LiquidityLockResult {
  isLocked: boolean;
  lockedAmount: bigint; // LP tokens locked
  lockedPercentage: number; // 0-100%
  unlockDate: Date | null;
  lockPlatform: string | null; // 'Unicrypt', 'Team.Finance', etc.
}

export async function checkLiquidityLock(
  tokenAddress: string,
  chain: 'ethereum' | 'base' | 'bsc' | 'polygon',
  provider: JsonRpcProvider
): Promise<LiquidityLockResult> {
  const pool = await getUniswapV2Pool(tokenAddress, chain, provider);

  if (!pool) {
    return {
      isLocked: false,
      lockedAmount: 0n,
      lockedPercentage: 0,
      unlockDate: null,
      lockPlatform: null,
    };
  }

  // Known lock contract addresses
  const lockContracts = {
    ethereum: [
      '0x663A5C229c09b049E36dCc11a9B0d4a8Eb9db214', // Unicrypt
      '0xE2fE530C047f2d85298b07D9333C05737f1435fB', // Team.Finance
      '0x71B5759d73262FBb223956913ecF4ecC51057641', // PinkLock
    ],
    bsc: [
      '0x407993575c91ce7643a4d4cCACc9A98c36eE1BBE', // PinkLock BSC
      '0x3f4D6bf08CB7A003488Ef082102C2e6418a4551e', // Mudra
    ],
    base: [
      // Base lock contracts (TBD)
    ],
    polygon: [
      '0x3f4D6bf08CB7A003488Ef082102C2e6418a4551e', // QuickSwap Lock
    ],
  };

  const pair = new Contract(pool.pairAddress, UNISWAP_V2_PAIR_ABI, provider);

  let totalLocked = 0n;
  let unlockDate: Date | null = null;
  let lockPlatform: string | null = null;

  for (const lockAddress of lockContracts[chain]) {
    const lockedBalance = await pair.balanceOf(lockAddress);

    if (lockedBalance > 0n) {
      totalLocked += lockedBalance;

      // Try to get unlock date (platform-specific)
      // Example for Unicrypt:
      if (lockAddress === '0x663A5C229c09b049E36dCc11a9B0d4a8Eb9db214') {
        lockPlatform = 'Unicrypt';
        // Query Unicrypt API or contract for unlock date
        // (Implementation depends on Unicrypt contract ABI)
      }

      // Example for Team.Finance:
      if (lockAddress === '0xE2fE530C047f2d85298b07D9333C05737f1435fB') {
        lockPlatform = 'Team.Finance';
        // Query Team.Finance API
      }
    }
  }

  const lockedPercentage = pool.totalLPSupply > 0n
    ? Number((totalLocked * 10000n) / pool.totalLPSupply) / 100
    : 0;

  return {
    isLocked: totalLocked > 0n,
    lockedAmount: totalLocked,
    lockedPercentage,
    unlockDate,
    lockPlatform,
  };
}
```

### 4.3 Top Holder Analysis (Whale Detection)

**Scam Indicator**: If top 10 holders own >50% of supply, likely Sybil attack or insider concentration.

```typescript
// src/blockchain/evm/holder-analysis.ts
export interface HolderDistribution {
  topHolders: Array<{ address: string; balance: bigint; percentage: number }>;
  top10Percentage: number;
  top50Percentage: number;
  holderCount: number;
}

export async function analyzeHolderDistribution(
  tokenAddress: string,
  chain: 'ethereum' | 'base' | 'bsc' | 'polygon'
): Promise<HolderDistribution> {
  // Use block explorer API (Etherscan, BscScan) to get top holders
  const apiUrls = {
    ethereum: 'https://api.etherscan.io/api',
    bsc: 'https://api.bscscan.com/api',
    polygon: 'https://api.polygonscan.com/api',
    base: 'https://api.basescan.org/api',
  };

  const apiKeys = {
    ethereum: process.env.ETHERSCAN_API_KEY!,
    bsc: process.env.BSCSCAN_API_KEY!,
    polygon: process.env.POLYGONSCAN_API_KEY!,
    base: process.env.BASESCAN_API_KEY!,
  };

  const response = await axios.get(apiUrls[chain], {
    params: {
      module: 'token',
      action: 'tokenholderlist',
      contractaddress: tokenAddress,
      page: 1,
      offset: 50, // Top 50 holders
      apikey: apiKeys[chain],
    },
  });

  const holders = response.data.result || [];
  const contract = new Contract(tokenAddress, ['function totalSupply() view returns (uint256)'], provider);
  const totalSupply = await contract.totalSupply();

  const topHolders = holders.map((holder: any) => ({
    address: holder.TokenHolderAddress,
    balance: BigInt(holder.TokenHolderQuantity),
    percentage: (Number(BigInt(holder.TokenHolderQuantity) * 10000n / totalSupply) / 100),
  }));

  const top10Percentage = topHolders.slice(0, 10).reduce((sum, h) => sum + h.percentage, 0);
  const top50Percentage = topHolders.reduce((sum, h) => sum + h.percentage, 0);

  return {
    topHolders,
    top10Percentage,
    top50Percentage,
    holderCount: parseInt(response.data.message.split(' ').pop(), 10) || holders.length,
  };
}
```

---

## 5. Multi-Chain Risk Scoring Algorithm

### 5.1 EVM-Specific Risk Metrics

CryptoRugMunch's **12-metric risk algorithm** needs adjustments for EVM chains:

| Metric | Solana Implementation | EVM Implementation |
|--------|----------------------|-------------------|
| **1. Liquidity** | Raydium/Orca pool reserves | Uniswap/PancakeSwap reserves |
| **2. LP Lock** | LP token holder analysis | Unicrypt/Team.Finance lock check |
| **3. Holder Concentration** | getProgramAccounts (slow) | Etherscan API (top holders) |
| **4. Mint Authority** | Check mint authority field | Check source code + owner() |
| **5. Freeze Authority** | Check freeze authority field | Check source code (rare) |
| **6. Honeypot** | Rare (transfer restrictions) | Common (simulate sell) |
| **7. Transfer Tax** | Not applicable | GoPlus API (buy/sell tax) |
| **8. Ownership Renounced** | Check if authority = null | Check if owner = 0x0 or renounced |
| **9. Contract Verified** | N/A (no source code) | Check Etherscan verification |
| **10. Token Age** | Creation timestamp | Contract deployment timestamp |
| **11. Social Media** | Twitter/Telegram | Twitter/Telegram |
| **12. Audit** | External audit reports | External audit reports |

### 5.2 EVM Risk Scoring Implementation

```typescript
// src/modules/scan/evm-risk-scoring.ts
import { ethers } from 'ethers';

export interface EVMRiskFactors {
  // Core metrics
  liquidity: number; // USD
  lpLocked: boolean;
  lpLockedPercentage: number;
  holderConcentration: number; // Top 10 holder %

  // EVM-specific
  hasMintFunction: boolean;
  isMintable: boolean;
  isHoneypot: boolean;
  buyTax: number;
  sellTax: number;
  ownershipRenounced: boolean;
  contractVerified: boolean;

  // Metadata
  tokenAge: number; // Days
  hasAudit: boolean;
  hasSocialMedia: boolean;
}

export interface EVMRiskScore {
  totalScore: number; // 0-100
  category: 'SAFE' | 'CAUTION' | 'HIGH_RISK' | 'LIKELY_SCAM';
  breakdown: Record<string, number>;
  flags: string[];
}

export async function calculateEVMRiskScore(
  tokenAddress: string,
  chain: 'ethereum' | 'base' | 'bsc' | 'polygon'
): Promise<EVMRiskScore> {
  const provider = getProviderForChain(chain);

  // Gather all data in parallel (performance optimization)
  const [
    metadata,
    pool,
    lockResult,
    holders,
    mintResult,
    honeypotResult,
    taxResult,
    verification,
  ] = await Promise.all([
    getERC20Metadata(tokenAddress, provider),
    getUniswapV2Pool(tokenAddress, chain, provider),
    checkLiquidityLock(tokenAddress, chain, provider),
    analyzeHolderDistribution(tokenAddress, chain),
    detectMintFunction(tokenAddress, chain, provider),
    detectHoneypot(tokenAddress, chain, provider),
    detectTransferTax(tokenAddress, chain, provider),
    getContractSource(tokenAddress, chain),
  ]);

  // Calculate token age
  const contract = new Contract(tokenAddress, [], provider);
  const creationTx = await findContractCreationTx(tokenAddress, chain);
  const creationBlock = await provider.getBlock(creationTx.blockNumber);
  const tokenAge = (Date.now() - creationBlock!.timestamp * 1000) / (1000 * 60 * 60 * 24); // Days

  // Build risk factors
  const factors: EVMRiskFactors = {
    liquidity: pool?.liquidityUSD || 0,
    lpLocked: lockResult.isLocked,
    lpLockedPercentage: lockResult.lockedPercentage,
    holderConcentration: holders.top10Percentage,
    hasMintFunction: mintResult.hasMintFunction,
    isMintable: mintResult.isMintable,
    isHoneypot: honeypotResult.isHoneypot,
    buyTax: taxResult.buyTax,
    sellTax: taxResult.sellTax,
    ownershipRenounced: metadata.owner === null,
    contractVerified: verification.isVerified,
    tokenAge,
    hasAudit: false, // TODO: Check audit databases
    hasSocialMedia: false, // TODO: Check social media
  };

  // Calculate weighted risk score (0-100, higher = safer)
  let totalScore = 0;
  const breakdown: Record<string, number> = {};
  const flags: string[] = [];

  // 1. Liquidity (15 points)
  if (factors.liquidity >= 100_000) {
    breakdown.liquidity = 15;
  } else if (factors.liquidity >= 50_000) {
    breakdown.liquidity = 12;
  } else if (factors.liquidity >= 10_000) {
    breakdown.liquidity = 8;
  } else if (factors.liquidity >= 5_000) {
    breakdown.liquidity = 5;
  } else {
    breakdown.liquidity = 0;
    flags.push('⚠️ Low liquidity (<$5K)');
  }

  // 2. LP Lock (20 points)
  if (factors.lpLocked && factors.lpLockedPercentage >= 80) {
    breakdown.lpLock = 20;
  } else if (factors.lpLocked && factors.lpLockedPercentage >= 50) {
    breakdown.lpLock = 12;
  } else if (factors.lpLocked) {
    breakdown.lpLock = 8;
  } else {
    breakdown.lpLock = 0;
    flags.push('🚨 LP not locked (rug risk)');
  }

  // 3. Holder Concentration (10 points)
  if (factors.holderConcentration <= 20) {
    breakdown.holderConcentration = 10;
  } else if (factors.holderConcentration <= 40) {
    breakdown.holderConcentration = 6;
  } else if (factors.holderConcentration <= 60) {
    breakdown.holderConcentration = 3;
  } else {
    breakdown.holderConcentration = 0;
    flags.push('⚠️ High holder concentration (>60% in top 10)');
  }

  // 4. Mint Authority (15 points)
  if (!factors.hasMintFunction) {
    breakdown.mintAuthority = 15;
  } else if (!factors.isMintable) {
    breakdown.mintAuthority = 10;
  } else {
    breakdown.mintAuthority = 0;
    flags.push('🚨 Token can be minted (dilution risk)');
  }

  // 5. Honeypot (20 points - CRITICAL)
  if (!factors.isHoneypot) {
    breakdown.honeypot = 20;
  } else {
    breakdown.honeypot = 0;
    flags.push('🚨 HONEYPOT DETECTED - Cannot sell');
  }

  // 6. Transfer Tax (10 points)
  if (factors.sellTax === 0 && factors.buyTax === 0) {
    breakdown.transferTax = 10;
  } else if (factors.sellTax <= 5 && factors.buyTax <= 5) {
    breakdown.transferTax = 7;
  } else if (factors.sellTax <= 15 && factors.buyTax <= 15) {
    breakdown.transferTax = 4;
  } else {
    breakdown.transferTax = 0;
    flags.push(`⚠️ High tax (Buy: ${factors.buyTax}%, Sell: ${factors.sellTax}%)`);
  }

  // 7. Ownership Renounced (5 points)
  if (factors.ownershipRenounced) {
    breakdown.ownership = 5;
  } else {
    breakdown.ownership = 0;
    flags.push('⚠️ Ownership not renounced');
  }

  // 8. Contract Verified (5 points)
  if (factors.contractVerified) {
    breakdown.verification = 5;
  } else {
    breakdown.verification = 0;
    flags.push('⚠️ Contract not verified');
  }

  // Calculate total
  totalScore = Object.values(breakdown).reduce((sum, score) => sum + score, 0);

  // Determine category
  let category: 'SAFE' | 'CAUTION' | 'HIGH_RISK' | 'LIKELY_SCAM';
  if (totalScore >= 80) {
    category = 'SAFE';
  } else if (totalScore >= 60) {
    category = 'CAUTION';
  } else if (totalScore >= 30) {
    category = 'HIGH_RISK';
  } else {
    category = 'LIKELY_SCAM';
  }

  return {
    totalScore,
    category,
    breakdown,
    flags,
  };
}

// Helper: Find contract creation transaction
async function findContractCreationTx(
  tokenAddress: string,
  chain: 'ethereum' | 'base' | 'bsc' | 'polygon'
): Promise<{ blockNumber: number; txHash: string }> {
  // Use Etherscan API
  const apiUrls = {
    ethereum: 'https://api.etherscan.io/api',
    bsc: 'https://api.bscscan.com/api',
    polygon: 'https://api.polygonscan.com/api',
    base: 'https://api.basescan.org/api',
  };

  const apiKeys = {
    ethereum: process.env.ETHERSCAN_API_KEY!,
    bsc: process.env.BSCSCAN_API_KEY!,
    polygon: process.env.POLYGONSCAN_API_KEY!,
    base: process.env.BASESCAN_API_KEY!,
  };

  const response = await axios.get(apiUrls[chain], {
    params: {
      module: 'contract',
      action: 'getcontractcreation',
      contractaddresses: tokenAddress,
      apikey: apiKeys[chain],
    },
  });

  const result = response.data.result[0];

  return {
    blockNumber: parseInt(result.blockNumber, 10),
    txHash: result.txHash,
  };
}
```

---

## 6. Multi-Chain Architecture Patterns

### 6.1 Chain-Agnostic Database Schema

**Goal**: Store scans for multiple chains without duplicating code.

```prisma
// prisma/schema.prisma (updated for multi-chain)
model Scan {
  id            String   @id @default(cuid())
  userId        String
  user          User     @relation(fields: [userId], references: [id])

  // Multi-chain fields
  chain         Chain    // SOLANA, ETHEREUM, BASE, BSC, POLYGON
  tokenAddress  String   // Chain-specific address

  // Risk scoring
  riskScore     Int      // 0-100
  category      RiskCategory // SAFE, CAUTION, HIGH_RISK, LIKELY_SCAM
  breakdown     Json     // Metric breakdown
  flags         String[] // Warning flags

  // Metadata
  scannedAt     DateTime @default(now())
  completedAt   DateTime?
  durationMs    Int?

  @@index([userId, scannedAt])
  @@index([chain, tokenAddress])
}

enum Chain {
  SOLANA
  ETHEREUM
  BASE
  BSC
  POLYGON
}

enum RiskCategory {
  SAFE
  CAUTION
  HIGH_RISK
  LIKELY_SCAM
}
```

### 6.2 Chain-Specific Worker Queues

**Goal**: Separate BullMQ queues for each chain (different performance characteristics).

```typescript
// src/workers/multi-chain-worker.ts
import { Worker, Job } from 'bullmq';
import { calculateSolanaRiskScore } from '../modules/scan/solana-risk-scoring';
import { calculateEVMRiskScore } from '../modules/scan/evm-risk-scoring';

interface ScanJobData {
  scanId: string;
  userId: string;
  chain: 'SOLANA' | 'ETHEREUM' | 'BASE' | 'BSC' | 'POLYGON';
  tokenAddress: string;
}

// Solana worker (fastest, 3-second SLA)
const solanaWorker = new Worker<ScanJobData>(
  'scan-solana',
  async (job: Job<ScanJobData>) => {
    const { scanId, userId, tokenAddress } = job.data;

    const result = await calculateSolanaRiskScore(tokenAddress);

    await prisma.scan.update({
      where: { id: scanId },
      data: {
        riskScore: result.totalScore,
        category: result.category,
        breakdown: result.breakdown,
        flags: result.flags,
        completedAt: new Date(),
        durationMs: Date.now() - job.timestamp,
      },
    });

    return result;
  },
  {
    connection: redisConnection,
    concurrency: 6, // 6 concurrent Solana scans
  }
);

// Ethereum worker (slower, 10-second SLA due to Etherscan API limits)
const ethereumWorker = new Worker<ScanJobData>(
  'scan-ethereum',
  async (job: Job<ScanJobData>) => {
    const { scanId, userId, tokenAddress } = job.data;

    const result = await calculateEVMRiskScore(tokenAddress, 'ethereum');

    await prisma.scan.update({
      where: { id: scanId },
      data: {
        riskScore: result.totalScore,
        category: result.category,
        breakdown: result.breakdown,
        flags: result.flags,
        completedAt: new Date(),
        durationMs: Date.now() - job.timestamp,
      },
    });

    return result;
  },
  {
    connection: redisConnection,
    concurrency: 3, // Lower concurrency due to API limits
    limiter: {
      max: 5, // Max 5 jobs per second (Etherscan free tier limit)
      duration: 1000,
    },
  }
);

// Base/BSC/Polygon workers (similar pattern)
const baseWorker = new Worker<ScanJobData>('scan-base', /* ... */, { concurrency: 4 });
const bscWorker = new Worker<ScanJobData>('scan-bsc', /* ... */, { concurrency: 5 });
const polygonWorker = new Worker<ScanJobData>('scan-polygon', /* ... */, { concurrency: 4 });
```

### 6.3 Unified API Endpoint

**Goal**: Single `/scan` endpoint that routes to correct chain queue.

```typescript
// src/modules/scan/scan.controller.ts
import { FastifyInstance, FastifyRequest } from 'fastify';
import { z } from 'zod';

const ScanRequestSchema = z.object({
  chain: z.enum(['SOLANA', 'ETHEREUM', 'BASE', 'BSC', 'POLYGON']),
  tokenAddress: z.string().min(32).max(66), // Chain-specific validation
});

export async function registerScanRoutes(app: FastifyInstance) {
  app.post('/api/scan', {
    schema: {
      body: ScanRequestSchema,
      response: {
        200: {
          type: 'object',
          properties: {
            scanId: { type: 'string' },
            chain: { type: 'string' },
            tokenAddress: { type: 'string' },
            status: { type: 'string' },
          },
        },
      },
    },
    preHandler: [authenticate, rateLimit],
  }, async (request: FastifyRequest<{ Body: z.infer<typeof ScanRequestSchema> }>, reply) => {
    const { chain, tokenAddress } = request.body;
    const userId = request.user.id;

    // Validate address format (chain-specific)
    if (chain === 'SOLANA') {
      if (!isValidSolanaAddress(tokenAddress)) {
        return reply.code(400).send({ error: 'Invalid Solana address' });
      }
    } else {
      if (!ethers.isAddress(tokenAddress)) {
        return reply.code(400).send({ error: 'Invalid EVM address' });
      }
    }

    // Create scan record
    const scan = await prisma.scan.create({
      data: {
        userId,
        chain,
        tokenAddress,
        category: 'HIGH_RISK', // Placeholder
      },
    });

    // Route to correct queue
    const queueName = `scan-${chain.toLowerCase()}`;
    await scanQueue.add(queueName, {
      scanId: scan.id,
      userId,
      chain,
      tokenAddress,
    });

    return reply.send({
      scanId: scan.id,
      chain,
      tokenAddress,
      status: 'QUEUED',
    });
  });
}

function isValidSolanaAddress(address: string): boolean {
  try {
    new PublicKey(address);
    return true;
  } catch {
    return false;
  }
}
```

---

## 7. Performance Optimization for EVM

### 7.1 RPC Call Batching

**Problem**: EVM RPC calls are expensive (Infura/Alchemy quotas).

**Solution**: Batch multiple calls using `multicall` or `Promise.all`.

```typescript
// src/blockchain/evm/multicall.ts
import { ethers, Contract } from 'ethers';

// Multicall3 contract (deployed on all major chains)
const MULTICALL3_ADDRESS = '0xcA11bde05977b3631167028862bE2a173976CA11';

const MULTICALL3_ABI = [
  'function aggregate3(tuple(address target, bool allowFailure, bytes callData)[] calls) public payable returns (tuple(bool success, bytes returnData)[] returnData)',
];

export async function batchERC20Calls(
  tokenAddresses: string[],
  provider: JsonRpcProvider
): Promise<Array<{ name: string; symbol: string; decimals: number; totalSupply: bigint }>> {
  const multicall = new Contract(MULTICALL3_ADDRESS, MULTICALL3_ABI, provider);

  // Encode calls
  const calls = tokenAddresses.flatMap((address) => {
    const contract = new Contract(address, [
      'function name() view returns (string)',
      'function symbol() view returns (string)',
      'function decimals() view returns (uint8)',
      'function totalSupply() view returns (uint256)',
    ], provider);

    return [
      { target: address, allowFailure: false, callData: contract.interface.encodeFunctionData('name') },
      { target: address, allowFailure: false, callData: contract.interface.encodeFunctionData('symbol') },
      { target: address, allowFailure: false, callData: contract.interface.encodeFunctionData('decimals') },
      { target: address, allowFailure: false, callData: contract.interface.encodeFunctionData('totalSupply') },
    ];
  });

  // Execute batch call
  const results = await multicall.aggregate3(calls);

  // Decode results
  const tokens = [];
  for (let i = 0; i < tokenAddresses.length; i++) {
    const baseIndex = i * 4;
    const contract = new Contract(tokenAddresses[i], [
      'function name() view returns (string)',
      'function symbol() view returns (string)',
      'function decimals() view returns (uint8)',
      'function totalSupply() view returns (uint256)',
    ], provider);

    tokens.push({
      name: contract.interface.decodeFunctionResult('name', results[baseIndex].returnData)[0],
      symbol: contract.interface.decodeFunctionResult('symbol', results[baseIndex + 1].returnData)[0],
      decimals: contract.interface.decodeFunctionResult('decimals', results[baseIndex + 2].returnData)[0],
      totalSupply: contract.interface.decodeFunctionResult('totalSupply', results[baseIndex + 3].returnData)[0],
    });
  }

  return tokens;
}
```

### 7.2 Aggressive Caching for EVM Data

**Problem**: Token metadata (name, symbol, decimals) never changes, but we query it repeatedly.

**Solution**: Cache immutable data for 30 days, cache mutable data (liquidity, holders) for 5 minutes.

```typescript
// src/blockchain/evm/cache.ts
import Redis from 'ioredis';

const redis = new Redis(process.env.REDIS_URL!);

export async function getCachedERC20Metadata(
  tokenAddress: string,
  chain: 'ethereum' | 'base' | 'bsc' | 'polygon',
  provider: JsonRpcProvider
): Promise<ERC20Metadata> {
  const cacheKey = `erc20:${chain}:${tokenAddress}:metadata`;

  // Check cache
  const cached = await redis.get(cacheKey);
  if (cached) {
    return JSON.parse(cached);
  }

  // Fetch from chain
  const metadata = await getERC20Metadata(tokenAddress, provider);

  // Cache for 30 days (immutable data)
  await redis.set(cacheKey, JSON.stringify(metadata), 'EX', 30 * 24 * 60 * 60);

  return metadata;
}

export async function getCachedLiquidityPool(
  tokenAddress: string,
  chain: 'ethereum' | 'base' | 'bsc' | 'polygon',
  provider: JsonRpcProvider
): Promise<UniswapV2Pool | null> {
  const cacheKey = `pool:${chain}:${tokenAddress}`;

  // Check cache
  const cached = await redis.get(cacheKey);
  if (cached) {
    return JSON.parse(cached);
  }

  // Fetch from chain
  const pool = await getUniswapV2Pool(tokenAddress, chain, provider);

  // Cache for 5 minutes (mutable data - reserves change)
  if (pool) {
    await redis.set(cacheKey, JSON.stringify(pool), 'EX', 5 * 60);
  }

  return pool;
}
```

### 7.3 Fallback Providers for High Availability

**Problem**: Infura/Alchemy outages can cause downtime.

**Solution**: Use `FallbackProvider` with multiple RPC endpoints.

```typescript
// src/blockchain/evm/providers.ts (already shown above)
export function getEthereumProvider(): FallbackProvider {
  const providers = [
    new JsonRpcProvider(process.env.INFURA_ETHEREUM_RPC!, 1),
    new JsonRpcProvider(process.env.ALCHEMY_ETHEREUM_RPC!, 1),
    new JsonRpcProvider(process.env.QUICKNODE_ETHEREUM_RPC!, 1),
  ];

  return new FallbackProvider(providers, 1, {
    cacheTimeout: 5000,
  });
}
```

---

## 8. Telegram Bot Multi-Chain Support

### 8.1 Chain Selection via Inline Keyboard

**UX**: User selects chain before scanning token.

```typescript
// src/modules/telegram/commands/scan.ts
import { InlineKeyboard } from 'grammy';

bot.command('scan', async (ctx) => {
  const keyboard = new InlineKeyboard()
    .text('🟣 Solana', 'chain:solana')
    .text('🔷 Ethereum', 'chain:ethereum').row()
    .text('🔵 Base', 'chain:base')
    .text('🟡 BSC', 'chain:bsc').row()
    .text('🟣 Polygon', 'chain:polygon');

  await ctx.reply('Select blockchain:', { reply_markup: keyboard });
});

bot.callbackQuery(/^chain:(.+)$/, async (ctx) => {
  const chain = ctx.match[1] as 'solana' | 'ethereum' | 'base' | 'bsc' | 'polygon';

  await ctx.answerCallbackQuery();
  await ctx.editMessageText(`✅ Selected: ${chain.toUpperCase()}\n\nNow send the token address to scan.`);

  // Store chain in session
  ctx.session.selectedChain = chain;
});

// Handle token address input
bot.on('message:text', async (ctx) => {
  const chain = ctx.session.selectedChain;

  if (!chain) {
    return ctx.reply('Please select a blockchain first using /scan');
  }

  const tokenAddress = ctx.message.text.trim();

  // Validate address
  if (chain === 'solana') {
    if (!isValidSolanaAddress(tokenAddress)) {
      return ctx.reply('❌ Invalid Solana address. Please try again.');
    }
  } else {
    if (!ethers.isAddress(tokenAddress)) {
      return ctx.reply('❌ Invalid EVM address. Please try again.');
    }
  }

  // Queue scan
  await ctx.reply('🔍 Analyzing token... This may take a few seconds.');

  const scan = await prisma.scan.create({
    data: {
      userId: ctx.from.id.toString(),
      chain: chain.toUpperCase() as Chain,
      tokenAddress,
      category: 'HIGH_RISK',
    },
  });

  await scanQueue.add(`scan-${chain}`, {
    scanId: scan.id,
    userId: ctx.from.id.toString(),
    chain: chain.toUpperCase(),
    tokenAddress,
  });

  // Wait for result (polling or webhook)
  // ... (implementation in telegram-bot-developer skill)
});
```

---

## 9. Security Considerations for Multi-Chain

### 9.1 Address Validation

**Critical**: Validate addresses per chain (Solana uses Base58, EVM uses Hex).

```typescript
// src/utils/address-validation.ts
import { PublicKey } from '@solana/web3.js';
import { ethers } from 'ethers';

export function validateAddress(address: string, chain: Chain): boolean {
  switch (chain) {
    case 'SOLANA':
      try {
        new PublicKey(address);
        return true;
      } catch {
        return false;
      }

    case 'ETHEREUM':
    case 'BASE':
    case 'BSC':
    case 'POLYGON':
      return ethers.isAddress(address);

    default:
      return false;
  }
}
```

### 9.2 Gas Price Monitoring (EVM Only)

**Problem**: High gas prices can make scans expensive (Ethereum).

**Solution**: Monitor gas prices and warn users.

```typescript
// src/blockchain/evm/gas-price.ts
export async function getCurrentGasPrice(chain: 'ethereum' | 'base' | 'bsc' | 'polygon'): Promise<bigint> {
  const provider = getProviderForChain(chain);
  const feeData = await provider.getFeeData();

  return feeData.gasPrice || 0n;
}

export async function estimateScanCost(chain: 'ethereum' | 'base' | 'bsc' | 'polygon'): Promise<number> {
  const gasPrice = await getCurrentGasPrice(chain);
  const estimatedGas = 500_000n; // Typical gas for contract calls

  const costWei = gasPrice * estimatedGas;
  const costEth = Number(ethers.formatEther(costWei));

  // Assume ETH = $3000, BNB = $500, MATIC = $1
  const prices = {
    ethereum: 3000,
    base: 3000,
    bsc: 500,
    polygon: 1,
  };

  return costEth * prices[chain];
}
```

---

## 10. Command Shortcuts

Use these shortcuts to quickly access specific topics:

- **#evm** - EVM blockchain fundamentals
- **#ethers** - Ethers.js setup and usage
- **#erc20** - ERC-20 token analysis
- **#uniswap** - Uniswap/DEX liquidity pools
- **#honeypot** - Honeypot detection
- **#tax** - Transfer tax detection
- **#mint** - Hidden mint function detection
- **#lock** - Liquidity lock verification
- **#multichain** - Multi-chain architecture
- **#risk-evm** - EVM risk scoring algorithm
- **#cache** - Caching strategies for EVM
- **#gas** - Gas price monitoring
- **#batch** - RPC call batching

---

## 11. Reference Materials

### 11.1 External APIs & Services

**Blockchain Data Providers**:
- **Infura** (Ethereum): https://infura.io
- **Alchemy** (Ethereum, Base, Polygon): https://alchemy.com
- **QuickNode** (Multi-chain): https://quicknode.com
- **Ankr** (BSC, Polygon): https://ankr.com

**Block Explorers** (for contract verification, holders):
- **Etherscan** (Ethereum): https://etherscan.io/apis
- **BscScan** (BSC): https://bscscan.com/apis
- **PolygonScan** (Polygon): https://polygonscan.com/apis
- **BaseScan** (Base): https://basescan.org/apis

**Security & Scam Detection**:
- **GoPlus Security API** (tax, honeypot): https://gopluslabs.io
- **Honeypot.is** (honeypot detection): https://honeypot.is/api
- **De.Fi Scanner** (multi-chain scam DB): https://de.fi

**Liquidity Lock Platforms**:
- **Unicrypt** (Ethereum): https://unicrypt.network
- **Team.Finance** (Ethereum, BSC): https://team.finance
- **PinkLock** (BSC): https://pinklock.app

### 11.2 Smart Contract Libraries

**OpenZeppelin** (secure contract templates):
- ERC-20: https://docs.openzeppelin.com/contracts/erc20
- Ownable: https://docs.openzeppelin.com/contracts/access-control
- Pausable: https://docs.openzeppelin.com/contracts/utils

**Uniswap V2**:
- Core contracts: https://github.com/Uniswap/v2-core
- Router: https://github.com/Uniswap/v2-periphery

**Uniswap V3**:
- Docs: https://docs.uniswap.org/contracts/v3/overview

### 11.3 CryptoRugMunch Documentation

**Related Skills**:
- `crypto-scam-analyst` - Solana risk scoring algorithm
- `solana-blockchain-specialist` - Solana-specific scam patterns
- `rugmunch-architect` - System architecture overview
- `security-auditor` - Smart contract security best practices

**Project Docs**:
- `/docs/03-TECHNICAL/integrations/blockchain-api-integration.md` - Solana API integration (reference for EVM)
- `/docs/03-TECHNICAL/integrations/telegram-bot-risk-algorithm.md` - Original 12-metric algorithm
- `/docs/06-ROADMAP/18-month-roadmap.md` - Multi-chain expansion timeline (Month 4-6)

---

## Summary

The **EVM & Multi-Chain Specialist** skill provides comprehensive expertise for expanding CryptoRugMunch to **Ethereum, Base, BSC, and Polygon** blockchains. Key capabilities:

1. **EVM vs Solana Understanding**: Account models, gas, finality, scam vectors
2. **Ethers.js Mastery**: Provider setup, ERC-20 interactions, multicall batching
3. **Scam Detection**: Honeypot, transfer tax, hidden mint, ownership analysis
4. **Liquidity Pool Analysis**: Uniswap V2/V3, PancakeSwap, lock verification
5. **Multi-Chain Architecture**: Unified API, chain-specific workers, database schema
6. **Performance Optimization**: RPC batching, aggressive caching, fallback providers
7. **Risk Scoring Adaptation**: EVM-specific metrics, weighted scoring, flags

**Timeline**: EVM expansion planned for **Month 4-6** post-Solana MVP launch.

**Next Steps**:
1. Implement Ethers.js provider setup with fallback
2. Create ERC-20 metadata fetching with caching
3. Integrate GoPlus API for honeypot/tax detection
4. Build Uniswap V2 pool analysis
5. Adapt risk scoring algorithm for EVM chains
6. Update Telegram bot for chain selection
7. Deploy chain-specific BullMQ workers

---

**Built for multi-chain scam detection** 🌐
**Protecting users across all major blockchains** 🛡️
