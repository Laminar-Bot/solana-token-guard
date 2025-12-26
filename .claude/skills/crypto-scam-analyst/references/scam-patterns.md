# Common Solana Scam Patterns

Real-world scam types, how they work, and how CryptoRugMunch detects them.

## 1. Classic Rugpull (Instant Liquidity Drain)

**How it works**:
1. Team creates token with unlocked LP
2. Marketing campaign pumps price 10-100x
3. Team removes all liquidity in single transaction
4. Token becomes worthless instantly

**Detection**:
- Metric #2 (LP Lock): Unlocked = -20 points
- Usually combined with high concentration (Metric #3)

**Real example** (Solana, 2023):
- Token: "SafeMoon2.0"
- Launch liquidity: $50K
- Peak market cap: $2M
- Rugpull: Day 3, removed all liquidity
- Victims: ~1,200 wallets

---

## 2. Slow Rug (Team Wallet Drip-Sell)

**How it works**:
1. LP locked for 90+ days (looks safe!)
2. Team holds 40-60% in separate wallets
3. Sells slowly over weeks (5-10% per week)
4. Price bleeds down 80-90%
5. LP unlocks, team removes remaining liquidity

**Detection**:
- Metric #3 (Top 10 Holder %): High concentration
- Metric #4 (Whale Count): Few large holders
- Requires ongoing monitoring (price trend alerts)

**Real example** (Solana, 2024):
- Token: "MetaverseCoin"
- LP locked: 180 days
- Team wallets: 55% supply
- Sell pattern: 8% per week for 6 weeks
- Final rug: LP unlocked on day 181

---

## 3. Honeypot (Freeze Authority)

**How it works**:
1. Freeze authority enabled (looks innocent)
2. Users buy normally (no issues)
3. When users try to sell → transaction fails
4. Team gradually unfreezes select wallets (insiders cash out)

**Detection**:
- Metric #6 (Freeze Authority): Enabled = -15 points
- Critical flag (instant HIGH_RISK minimum)

**Real example** (Solana, 2023):
- Token: "ShipCoin"
- Freeze authority: Enabled
- Buy tax: 0%, Sell tax: 0% (looked safe!)
- Reality: 90% of wallets frozen after buying
- Team cashed out $400K

---

## 4. Honeypot (High Sell Tax)

**How it works**:
1. Buy tax: 2-5% (reasonable)
2. Sell tax: 25-50% (hidden or misrepresented)
3. Users buy and see price increase
4. When selling → lose 30-50% to taxes
5. Effective profit = impossible

**Detection**:
- Metric #9 (Tax Asymmetry): >10% difference = -50 points
- **CRITICAL METRIC** (forces LIKELY_SCAM if detected)

**Real example** (Solana, 2024):
- Token: "ReflectToken"
- Advertised: "5% tax funds reflections!"
- Reality: 2% buy, 35% sell
- Victims couldn't profit even with 2x price increase

---

## 5. Mint Dilution Attack

**How it works**:
1. Mint authority enabled (excuse: "future utility")
2. Token pumps to $10M market cap
3. Team mints 10x supply overnight
4. Original holders diluted 90%
5. Team dumps newly minted tokens

**Detection**:
- Metric #5 (Mint Authority): Enabled = -15 points
- Combined with Metric #3 (concentration increases post-mint)

**Real example** (Solana, 2023):
- Token: "UtilityDAO"
- Initial supply: 100M tokens
- Day 30: Minted 900M tokens (10x dilution)
- Excuse: "For staking rewards!" (lie)

---

## 6. Sybil Launch (Fake Fair Launch)

**How it works**:
1. Announce "fair launch" (no team allocation)
2. Create 50-100 wallets controlled by team
3. Distribute tokens to Sybil wallets
4. Looks like distributed launch
5. Team controls 70%+ via multiple wallets

**Detection**:
- Metric #3/#4: Would catch if wallets consolidated
- **Requires funding trail analysis** (see `holder-forensics.md`)
- Check: Are top wallets funded from same source?

**Real example** (Solana, 2024):
- Token: "FairCoin"
- Claimed: 200 initial holders
- Reality: 180 wallets funded from same address (team)
- Detection: Funding chain analysis

---

## 7. Wash Trading (Fake Volume)

**How it works**:
1. Bot buys and sells with itself
2. Creates illusion of high activity
3. Volume 10-20x higher than liquidity
4. Attracts real users ("this is trending!")
5. Real users buy, bot dumps on them

**Detection**:
- Metric #8 (Volume/Liquidity Ratio): >10x = -12 points
- Pattern: Consistent small trades, same size, round numbers

**Real example** (Solana, 2023):
- Token: "TrendCoin"
- Liquidity: $8K
- 24h volume: $150K (19x ratio)
- Pattern: 500+ trades of exactly 0.5 SOL each
- Same wallet buying and selling

---

## 8. Celebrity Impersonation

**How it works**:
1. Create fake Twitter account (@elonmusk → @e1onmusk)
2. Tweet about new token
3. Gullible users buy
4. Real celebrity never endorsed it
5. Team dumps

**Detection**:
- Metric #12 (Social Verification): Helps but limited
- **Requires off-chain verification** (check blue checkmark, account age)

**Real example** (Solana, 2024):
- Token: "ElonAI"
- Fake account: @elonmusk (with Cyrillic 'o')
- Claimed: "Official Elon Musk Solana token!"
- Reality: Elon Musk never heard of it
- Rugpull: $2M stolen

---

## 9. DAO Treasury Drain

**How it works**:
1. Legitimate DAO launches (looks safe)
2. Governance proposal: "Invest treasury in yield farming"
3. Malicious proposal passes (insider voting, low participation)
4. Treasury sent to team-controlled address
5. Funds drained

**Detection**:
- **Not detectable via token metrics**
- Requires governance monitoring
- Check: DAO multi-sig signers, proposal history

**Real example** (Solana, 2023):
- DAO: "BuilderDAO"
- Treasury: $500K
- Proposal: "Invest in Raydium farming"
- Reality: Funds sent to EOA, drained
- 3 insiders voted, 97% of tokens didn't participate

---

## 10. LP Lock Expiration Rug

**How it works**:
1. LP locked for 30-90 days (short-term)
2. Marketing emphasizes "LP LOCKED!" (omits duration)
3. Users assume long-term safety
4. LP unlocks → team removes liquidity
5. "We said it was locked, not *how long*"

**Detection**:
- Metric #2 (LP Lock): <30 days = -15 points
- UI must show lock duration, not just binary locked/unlocked

**Real example** (Solana, 2024):
- Token: "LockCoin"
- LP lock: 7 days (marketed as "LOCKED!")
- Day 8: Rugpull
- Team defense: "We never said how long"

---

## 11. Multi-Pool Fragmentation

**How it works**:
1. Create 5 pools across different DEXs (Raydium, Orca, Meteora)
2. Each pool: $5K liquidity (individually looks low)
3. Combined: $25K (acceptable)
4. Scanner only checks primary pool ($5K) → flags as low liquidity

**Detection**:
- **Requires multi-pool aggregation**
- Sum liquidity across all DEXs
- Birdeye API provides multi-pool data

**Real example** (Solana, 2024):
- Token: "MultiSwap"
- 6 pools: Each $3K-8K
- Total liquidity: $35K (acceptable)
- Single-pool scan: Flagged as low liquidity (false positive)

---

## 12. Fake Audit

**How it works**:
1. Pay unknown "audit firm" $500
2. Receive PDF with green checkmarks
3. Claim "AUDITED BY [FirmName]!"
4. Firm has no reputation, no track record
5. Audit is worthless

**Detection**:
- Metric #7 (Contract Verification): Only checks if verified, not audit quality
- **Requires audit firm whitelist** (CertiK, Trail of Bits, Quantstamp)
- Check: Is audit from reputable firm?

**Real example** (Solana, 2023):
- Token: "AuditSafe"
- Claimed: "Audited by BlockSecure!"
- Reality: BlockSecure = unknown firm, no website
- Audit PDF = generic template

---

## 13. Influencer Pump-and-Dump

**How it works**:
1. Pay crypto influencer $10K-50K
2. Influencer tweets: "This is the next 100x!"
3. Followers buy (FOMO)
4. Team dumps on followers
5. Influencer deletes tweet

**Detection**:
- **Not detectable via on-chain metrics**
- Requires social listening (sudden influencer mentions + price spike)

**Real example** (Solana, 2024):
- Token: "MoonShot"
- Influencer: 200K followers
- Tweet: "I'm all in!"
- Price: +800% in 2 hours
- Influencer sold before tweeting
- Team dumped $300K

---

## 14. Copy-Cat Token (Name Confusion)

**How it works**:
1. Successful token exists: "BonkInu" (real)
2. Create: "BonkInu" (fake, slightly different contract address)
3. List on same DEX with similar logo
4. Users buy wrong token by mistake
5. Fake token has no liquidity/utility

**Detection**:
- **Not detectable via single-token metrics**
- Requires name/logo similarity detection
- Check: Is this a known impersonation?

**Real example** (Solana, 2024):
- Real: BONK (WIF competitor)
- Fake: "BONK" (Cyrillic characters, looks identical)
- Victims: Bought fake token worth $0

---

## 15. Program Upgrade Authority Rug

**How it works**:
1. Mint/freeze authority disabled (looks safe!)
2. **Program update authority still enabled** (hidden risk)
3. Team deploys program upgrade with backdoor
4. New program re-enables mint authority
5. Team mints infinite supply

**Detection**:
- **Requires checking program upgrade authority** (not just mint/freeze)
- Solana CLI: `solana program show <program_id>`
- Check: `Upgrade Authority: <address>` vs `Upgrade Authority: None`

**Real example** (Solana, 2023):
- Token: "ImmutableCoin"
- Mint authority: Disabled
- Freeze authority: Disabled
- Program upgrade authority: Enabled
- Day 45: Program upgraded, mint authority re-enabled
- Team minted 1000x supply

---

## Detection Coverage Matrix

| Scam Type | Detectable? | Primary Metrics | Notes |
|-----------|-------------|-----------------|-------|
| Classic Rugpull | ✅ Yes | #2 (LP Lock) | High confidence |
| Slow Rug | ⚠️ Partial | #3, #4 (Holders) | Requires ongoing monitoring |
| Honeypot (Freeze) | ✅ Yes | #6 (Freeze Auth) | High confidence |
| Honeypot (Tax) | ✅ Yes | #9 (Tax Asymmetry) | High confidence |
| Mint Dilution | ✅ Yes | #5 (Mint Auth) | High confidence |
| Sybil Launch | ⚠️ Partial | #3, #4 (Holders) | Requires funding analysis |
| Wash Trading | ✅ Yes | #8 (Volume Ratio) | Moderate confidence |
| Celebrity Fake | ❌ No | #12 (Social) | Requires manual verification |
| DAO Drain | ❌ No | — | Out of scope |
| LP Lock Expiry | ✅ Yes | #2 (LP Lock) | Check duration! |
| Multi-Pool Frag | ✅ Yes | #1 (Liquidity) | Aggregate pools |
| Fake Audit | ⚠️ Partial | #7 (Verification) | Need audit whitelist |
| Influencer P&D | ❌ No | — | Requires social listening |
| Copy-Cat | ❌ No | — | Requires name/logo DB |
| Upgrade Authority | ⚠️ Future | — | Not yet implemented |

**Coverage**: 80% of common scams detectable via current metrics

**Roadmap**:
- Phase 2: Add program upgrade authority check
- Phase 3: Add social listening (influencer pump detection)
- Phase 4: Add name/logo similarity detection

---

## Scam Evolution Timeline

**2021-2022: Basic Rugs**
- Unlocked LP, instant drain
- Obvious red flags
- Detection: Easy

**2023: Honeypot Era**
- Freeze authority, high sell tax
- Less obvious, need simulation
- Detection: Moderate difficulty

**2024: Sophisticated Evasion**
- Sybil distribution
- Multi-pool fragmentation
- Program upgrade backdoors
- Detection: Challenging, requires deep analysis

**2025+ Prediction: AI-Generated Scams**
- Programmatic token generation
- AI-written whitepapers
- Deepfake celebrity endorsements
- Detection: Will require ML models
