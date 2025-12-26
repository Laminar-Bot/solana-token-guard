# $CRM Token Economics & DeFi Patterns

## Staking Tiers & Rewards

```typescript
export const STAKING_TIERS = {
  bronze: { minStake: 10_000, scansPerDay: 5, rewardRate: 0.05 },
  silver: { minStake: 50_000, scansPerDay: 25, rewardRate: 0.08 },
  gold: { minStake: 200_000, scansPerDay: 100, rewardRate: 0.12 },
  platinum: { minStake: 1_000_000, scansPerDay: -1, rewardRate: 0.15 },
};

// Calculate staking rewards (continuous compounding)
export function calculateStakingRewards(
  stakeAmount: number,
  apr: number,
  daysStaked: number
): number {
  const secondsPerYear = 365 * 24 * 60 * 60;
  const elapsedSeconds = daysStaked * 24 * 60 * 60;
  return (stakeAmount * apr * elapsedSeconds) / secondsPerYear;
}
```

## Automated Token Burns (Rust/Anchor)

```rust
#[program]
pub mod crm_burns {
    pub fn execute_monthly_burn(ctx: Context<ExecuteBurn>, amount: u64) -> Result<()> {
        // Burn 10% of monthly revenue
        let burn_account = &ctx.accounts.burn_account;

        token::burn(
            CpiContext::new(
                ctx.accounts.token_program.to_account_info(),
                token::Burn {
                    mint: ctx.accounts.crm_mint.to_account_info(),
                    from: ctx.accounts.treasury.to_account_info(),
                    authority: ctx.accounts.authority.to_account_info(),
                },
            ),
            amount,
        )?;

        emit!(BurnEvent { amount, timestamp: Clock::get()?.unix_timestamp });
        Ok(())
    }
}
```

## DAO Governance Voting

```typescript
// Voting power = staked $CRM * multiplier
export function calculateVotingPower(userId: string): number {
  const user = await userRepository.findById(userId);
  const stakedAmount = user.stakedCRM;
  const tier = getStakingTier(stakedAmount);

  const multipliers = { bronze: 1, silver: 2, gold: 3, platinum: 5 };
  return stakedAmount * (multipliers[tier] || 1);
}

// Submit governance proposal
export async function createProposal(params: {
  title: string;
  description: string;
  votingPeriodDays: number;
}) {
  const proposal = await proposalRepository.create({
    ...params,
    status: 'active',
    votesFor: 0,
    votesAgainst: 0,
    createdAt: new Date(),
  });

  return proposal.id;
}
```
