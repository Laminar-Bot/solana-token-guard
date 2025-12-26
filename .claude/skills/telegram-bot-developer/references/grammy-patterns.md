# Grammy.js Implementation Patterns for CryptoRugMunch

## Pattern 1: Command Handler with Rate Limiting

```typescript
// src/modules/telegram/commands/scan.command.ts
import { Composer } from 'grammy';
import type { MyContext } from '../types';
import { rateLimiter } from '@/shared/rate-limiter';

export const scanCommand = new Composer<MyContext>();

scanCommand.command('scan', async (ctx) => {
  const userId = ctx.from!.id.toString();

  // Check rate limit based on tier
  const user = await userService.find(userId);
  const canScan = await rateLimiter.check(user.id, user.tier);

  if (!canScan) {
    await ctx.reply(
      '⚠️ *Rate limit reached*\\n\\n' +
      `${user.tier === 'free' ? 'Free tier: 1 scan/day' : `${getTierLimits(user.tier).scansPerDay} scans/day used`}\\n` +
      'Upgrade to Premium for 50 scans/day',
      { parse_mode: 'MarkdownV2' }
    );
    return;
  }

  // Set conversation state
  await ctx.conversation.enter('scan-flow');
});
```

## Pattern 2: Conversation Flow (Multi-Step Input)

```typescript
// src/modules/telegram/conversations/scan-flow.ts
import { createConversation } from '@grammyjs/conversations';
import type { MyContext, MyConversation } from '../types';

export async function scanFlow(conversation: MyConversation, ctx: MyContext) {
  await ctx.reply('🔍 *Token Scanner*\\n\\nSend me a Solana token address:', {
    parse_mode: 'MarkdownV2',
  });

  // Wait for user input
  const { message } = await conversation.wait();
  const tokenAddress = message?.text?.trim();

  if (!tokenAddress || !isValidSolanaAddress(tokenAddress)) {
    await ctx.reply('❌ Invalid Solana address\\. Try again:');
    return scanFlow(conversation, ctx); // Retry
  }

  // Queue scan job
  await ctx.reply('⏳ Scanning\\.\\.\\. \\(usually takes 2\\-3 seconds\\)');

  await scanQueue.add('token-scan', {
    tokenAddress,
    userId: ctx.from!.id.toString(),
    chatId: ctx.chat!.id.toString(),
    tier: ctx.session.user.tier,
  });
}
```

## Pattern 3: Inline Keyboards with Callback Queries

```typescript
// src/modules/telegram/keyboards/subscription.keyboard.ts
import { InlineKeyboard } from 'grammy';

export function createSubscriptionKeyboard(currentTier: string) {
  const keyboard = new InlineKeyboard();

  if (currentTier === 'free') {
    keyboard
      .text('💎 Upgrade to Premium ($19.99/mo)', 'upgrade_premium')
      .row()
      .text('🏢 Enterprise Plans', 'upgrade_enterprise')
      .row()
      .text('🪙 Stake $CRM Tokens', 'stake_crm');
  } else if (currentTier === 'premium') {
    keyboard
      .text('📊 View Usage', 'view_usage')
      .row()
      .text('🏢 Upgrade to Enterprise', 'upgrade_enterprise')
      .row()
      .text('❌ Cancel Subscription', 'cancel_subscription');
  }

  return keyboard;
}

// Handle callback queries
bot.callbackQuery('upgrade_premium', async (ctx) => {
  await ctx.answerCallbackQuery(); // Remove loading state

  const checkoutUrl = await stripeService.createCheckoutSession({
    userId: ctx.from!.id.toString(),
    tier: 'premium',
    successUrl: `https://t.me/${bot.botInfo.username}?start=payment_success`,
    cancelUrl: `https://t.me/${bot.botInfo.username}?start=payment_cancel`,
  });

  await ctx.editMessageText(
    '💎 *Premium Subscription*\\n\\n' +
    '✅ 50 scans per day\\n' +
    '✅ Full 12\\-metric analysis\\n' +
    '✅ PDF exports\\n' +
    '✅ Email alerts\\n\\n' +
    `[Complete Purchase \\(opens Stripe\\)](${checkoutUrl})`,
    {
      parse_mode: 'MarkdownV2',
      link_preview_options: { is_disabled: true },
    }
  );
});
```

## Pattern 4: Webhook vs Polling Mode

```typescript
// src/modules/telegram/bot.ts
import { Bot, webhookCallback } from 'grammy';
import type { FastifyInstance } from 'fastify';

const bot = new Bot(process.env.TELEGRAM_BOT_TOKEN!);

// Development: Polling mode
if (process.env.NODE_ENV === 'development') {
  bot.start({
    onStart: () => console.log('✅ Bot started (polling mode)'),
  });
}

// Production: Webhook mode
export function registerTelegramWebhook(app: FastifyInstance) {
  const webhookPath = `/webhook/telegram/${process.env.TELEGRAM_WEBHOOK_SECRET}`;

  app.post(webhookPath, webhookCallback(bot, 'fastify'));

  // Set webhook URL
  bot.api.setWebhook(`${process.env.APP_URL}${webhookPath}`, {
    drop_pending_updates: true,
    max_connections: 100,
  });
}
```

## Pattern 5: Message Formatting (MarkdownV2 Escaping)

```typescript
// src/modules/telegram/formatters/scan-result.formatter.ts
export function formatScanResult(scan: Scan): string {
  const riskEmoji = {
    LOW_RISK: '✅',
    MEDIUM_RISK: '⚠️',
    HIGH_RISK: '🚨',
  }[scan.riskLevel];

  // IMPORTANT: Escape special characters for MarkdownV2
  const escapeMarkdown = (text: string) =>
    text.replace(/[_*[\]()~`>#+\-=|{}.!]/g, '\\\\$&');

  return `
${riskEmoji} *Risk Score: ${scan.riskScore}/100* \\(${escapeMarkdown(scan.riskLevel.replace('_', ' '))}\\)

*Token:* ${escapeMarkdown(scan.tokenName)} \\(${escapeMarkdown(scan.tokenSymbol)}\\)
*Address:* \`${scan.tokenAddress}\`

*Top Risks:*
${scan.redFlags.slice(0, 3).map(f => `• ${escapeMarkdown(f.message)}`).join('\\n')}

*Key Metrics:*
• Liquidity: \\$${scan.metrics.liquidity.toLocaleString()}
• Holder Concentration: ${scan.metrics.holderConcentration}%
• LP Lock: ${scan.metrics.lpLockDays} days
• Mint Authority: ${scan.metrics.mintAuthority ? '⚠️ Active' : '✅ Revoked'}

${scan.tier === 'free' ? '\\n💎 Upgrade to Premium for detailed reports' : ''}
  `.trim();
}
```

## Pattern 6: GDPR Compliance (Data Export & Deletion)

```typescript
// src/modules/telegram/commands/gdpr.commands.ts
import { Composer, InputFile } from 'grammy';

export const gdprCommands = new Composer<MyContext>();

// /export command
gdprCommands.command('export', async (ctx) => {
  const userId = ctx.from!.id.toString();
  const userData = await userService.exportData(userId);

  // Generate JSON file
  const jsonData = JSON.stringify(userData, null, 2);
  const buffer = Buffer.from(jsonData, 'utf-8');

  await ctx.replyWithDocument(
    new InputFile(buffer, 'my-data.json'),
    { caption: '📦 Your complete data export (GDPR compliant)' }
  );

  logger.info({ userId }, 'GDPR data export completed');
});

// /delete command with confirmation
gdprCommands.command('delete', async (ctx) => {
  const keyboard = new InlineKeyboard()
    .text('❌ Yes, delete my account', 'gdpr_delete_confirm')
    .row()
    .text('Cancel', 'gdpr_delete_cancel');

  await ctx.reply(
    '⚠️ *Delete Account*\\n\\n' +
    'This will permanently delete:\\n' +
    '• Your scan history\\n' +
    '• Your subscription\\n' +
    '• All personal data\\n\\n' +
    '*This cannot be undone\\.*',
    {
      parse_mode: 'MarkdownV2',
      reply_markup: keyboard,
    }
  );
});

gdprCommands.callbackQuery('gdpr_delete_confirm', async (ctx) => {
  await ctx.answerCallbackQuery();

  const userId = ctx.from!.id.toString();
  await userService.deleteAccount(userId);

  await ctx.editMessageText('✅ Your account has been deleted\\.');
  logger.info({ userId }, 'GDPR account deletion completed');
});
```

## Pattern 7: Error Handling & Logging

```typescript
// src/modules/telegram/middleware/error-handler.ts
import type { MyContext } from '../types';
import { logger } from '@/shared/logger';
import Sentry from '@sentry/node';

export async function errorHandler(ctx: MyContext, next: () => Promise<void>) {
  try {
    await next();
  } catch (error: any) {
    logger.error(
      {
        error,
        userId: ctx.from?.id,
        chatId: ctx.chat?.id,
        updateType: ctx.update.message ? 'message' : ctx.update.callback_query ? 'callback' : 'other',
      },
      'Telegram bot error'
    );

    Sentry.captureException(error, {
      tags: { handler: 'telegram-bot' },
      extra: {
        userId: ctx.from?.id,
        chatId: ctx.chat?.id,
        update: ctx.update,
      },
    });

    // User-friendly error message
    try {
      await ctx.reply(
        '❌ Sorry, something went wrong\\. Our team has been notified\\. Please try again later\\.',
        { parse_mode: 'MarkdownV2' }
      );
    } catch (replyError) {
      // Failed to send error message - log but don't throw
      logger.error({ replyError }, 'Failed to send error message to user');
    }
  }
}

// Register middleware
bot.use(errorHandler);
```

## Pattern 8: Job Queue Integration (BullMQ + Telegram)

```typescript
// src/modules/scan/scan.worker.ts
import { Worker } from 'bullmq';
import { bot } from '@/modules/telegram/bot';

export const scanWorker = new Worker(
  'token-scan',
  async (job) => {
    const { tokenAddress, userId, chatId, tier } = job.data;

    try {
      // Perform scan
      const scan = await scanService.analyzeFull(tokenAddress, tier);

      // Send result back to Telegram
      await bot.api.sendMessage(
        chatId,
        formatScanResult(scan),
        { parse_mode: 'MarkdownV2' }
      );

      return { success: true, scanId: scan.id };
    } catch (error) {
      // Send error to user
      await bot.api.sendMessage(
        chatId,
        `❌ Scan failed: ${error.message}\\n\\nPlease try again\\.`,
        { parse_mode: 'MarkdownV2' }
      );

      throw error;
    }
  },
  {
    connection: redisConnection,
    concurrency: 6,
  }
);
```

## Related Documentation

- Grammy.js Docs: https://grammy.dev
- Telegram Bot API: https://core.telegram.org/bots/api
- MarkdownV2 Syntax: https://core.telegram.org/bots/api#markdownv2-style
