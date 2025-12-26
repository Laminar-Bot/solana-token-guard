# Stripe + Telegram Stars Payment Integration Patterns

## Pattern 1: Stripe Checkout Session Creation

```typescript
// src/modules/payment/stripe.service.ts
import Stripe from 'stripe';

const stripe = new Stripe(process.env.STRIPE_SECRET_KEY!, {
  apiVersion: '2024-12-18.acacia',
  maxNetworkRetries: 3,
  timeout: 10_000,
});

export async function createCheckoutSession(params: {
  userId: string;
  tier: 'premium' | 'enterprise';
  successUrl: string;
  cancelUrl: string;
  trialDays?: number;
}): Promise<string> {
  const { userId, tier, successUrl, cancelUrl, trialDays } = params;

  // Get or create Stripe customer
  let customer = await getStripeCustomer(userId);
  if (!customer) {
    customer = await stripe.customers.create({
      metadata: { userId, source: 'telegram' },
    });
    await userService.updateStripeCustomerId(userId, customer.id);
  }

  const priceId = tier === 'premium'
    ? process.env.STRIPE_PREMIUM_PRICE_ID
    : process.env.STRIPE_ENTERPRISE_PRICE_ID;

  const session = await stripe.checkout.sessions.create({
    customer: customer.id,
    mode: 'subscription',
    line_items: [{ price: priceId, quantity: 1 }],
    success_url: successUrl,
    cancel_url: cancelUrl,
    subscription_data: {
      trial_period_days: trialDays,
      metadata: { userId, tier },
    },
    metadata: { userId, tier },
  });

  return session.url!;
}
```

## Pattern 2: Webhook Signature Verification

```typescript
// src/modules/payment/stripe-webhook.controller.ts
import type { FastifyRequest, FastifyReply } from 'fastify';
import Stripe from 'stripe';

export async function handleStripeWebhook(
  request: FastifyRequest,
  reply: FastifyReply
) {
  const sig = request.headers['stripe-signature'] as string;
  const webhookSecret = process.env.STRIPE_WEBHOOK_SECRET!;

  let event: Stripe.Event;

  try {
    // CRITICAL: Verify webhook signature with raw body
    event = stripe.webhooks.constructEvent(
      request.rawBody!,
      sig,
      webhookSecret,
      300 // 5-minute tolerance
    );
  } catch (error: any) {
    logger.error({ error, signature: sig }, 'Webhook verification failed');
    return reply.status(400).send({ error: 'Invalid signature' });
  }

  // Process event asynchronously (don't block 200 response)
  handleStripeEvent(event).catch(error => {
    logger.error({ error, eventId: event.id }, 'Webhook handler failed');
    Sentry.captureException(error);
  });

  return reply.status(200).send({ received: true });
}
```

## Pattern 3: Subscription Lifecycle Events

```typescript
// src/modules/payment/stripe-event-handler.ts
export async function handleStripeEvent(event: Stripe.Event) {
  switch (event.type) {
    case 'customer.subscription.created': {
      const subscription = event.data.object as Stripe.Subscription;
      await handleSubscriptionCreated(subscription);
      break;
    }

    case 'customer.subscription.updated': {
      const subscription = event.data.object as Stripe.Subscription;
      await handleSubscriptionUpdated(subscription);
      break;
    }

    case 'invoice.paid': {
      const invoice = event.data.object as Stripe.Invoice;
      await handleInvoicePaid(invoice);
      break;
    }

    case 'invoice.payment_failed': {
      const invoice = event.data.object as Stripe.Invoice;
      await handlePaymentFailed(invoice);
      break;
    }

    case 'customer.subscription.deleted': {
      const subscription = event.data.object as Stripe.Subscription;
      await handleSubscriptionCanceled(subscription);
      break;
    }

    default:
      logger.debug({ eventType: event.type }, 'Unhandled webhook event');
  }
}

async function handleSubscriptionCreated(subscription: Stripe.Subscription) {
  const userId = subscription.metadata.userId;
  const tier = subscription.metadata.tier as SubscriptionTier;

  await userService.updateSubscription({
    userId,
    tier,
    status: subscription.status,
    stripeSubscriptionId: subscription.id,
    currentPeriodStart: new Date(subscription.current_period_start * 1000),
    currentPeriodEnd: new Date(subscription.current_period_end * 1000),
  });

  // Send Telegram notification
  await bot.api.sendMessage(
    userId,
    `🎉 Welcome to ${tier} tier!\\n\\nYour subscription is now active\\.`,
    { parse_mode: 'MarkdownV2' }
  );

  metrics.increment('subscription.created', 1, { tier });
}

async function handlePaymentFailed(invoice: Stripe.Invoice) {
  const userId = invoice.subscription_details?.metadata?.userId;
  const attemptCount = invoice.attempt_count || 0;

  if (attemptCount < 3) {
    // Retry (Stripe handles this automatically)
    logger.warn({ userId, attemptCount }, 'Payment failed - will retry');
  } else {
    // Final attempt failed - downgrade to free
    await userService.updateTier(userId, 'free');

    await bot.api.sendMessage(
      userId,
      `⚠️ Payment Failed\\n\\nYour subscription has been downgraded to Free tier\\.\\n\\n[Update Payment Method](${updatePaymentUrl})`,
      { parse_mode: 'MarkdownV2' }
    );
  }
}
```

## Pattern 4: Telegram Stars Payment (In-App Payments)

```typescript
// src/modules/telegram/commands/payment.command.ts
import { InlineKeyboard } from 'grammy';

export async function createTelegramStarsInvoice(ctx: MyContext) {
  const userId = ctx.from!.id.toString();

  // Telegram Stars pricing (100 Stars ≈ $1 USD)
  const prices = {
    premium_monthly: 1999, // 19.99 Stars = $19.99
    premium_yearly: 19999, // 199.99 Stars = $199.99
  };

  await ctx.replyWithInvoice(
    'Premium Subscription', // Title
    '50 scans/day, full analysis, PDF exports, email alerts', // Description
    JSON.stringify({ userId, tier: 'premium' }), // Payload
    '', // Provider token (empty for Telegram Stars)
    'XTR', // Currency (XTR = Telegram Stars)
    [
      {
        label: 'Premium Monthly',
        amount: prices.premium_monthly,
      },
    ],
    {
      photo_url: 'https://cryptorugmunch.com/images/premium-badge.png',
      need_email: true, // Optional: collect email
      send_email_to_provider: false,
    }
  );
}

// Handle successful payment
bot.on('pre_checkout_query', async (ctx) => {
  // Validate order
  const payload = JSON.parse(ctx.preCheckoutQuery.invoice_payload);
  const userId = payload.userId;

  // Check if user exists
  const user = await userService.find(userId);
  if (!user) {
    await ctx.answerPreCheckoutQuery(false, 'User not found');
    return;
  }

  await ctx.answerPreCheckoutQuery(true);
});

bot.on(':successful_payment', async (ctx) => {
  const payment = ctx.message!.successful_payment!;
  const payload = JSON.parse(payment.invoice_payload);

  // Upgrade user to premium
  await userService.updateTier(payload.userId, 'premium');

  await ctx.reply(
    '🎉 *Payment Successful!*\\n\\nYou now have Premium access\\.',
    { parse_mode: 'MarkdownV2' }
  );

  logger.info({ userId: payload.userId, amount: payment.total_amount }, 'Telegram Stars payment received');
});
```

## Pattern 5: Subscription Management (Cancel, Resume, Upgrade)

```typescript
// src/modules/payment/subscription.service.ts
export class SubscriptionService {
  async cancel(userId: string, immediately: boolean = false) {
    const user = await userService.find(userId);
    if (!user.stripeSubscriptionId) {
      throw new Error('No active subscription');
    }

    await stripe.subscriptions.update(user.stripeSubscriptionId, {
      cancel_at_period_end: !immediately,
    });

    if (immediately) {
      await userService.updateTier(userId, 'free');
    }

    return {
      message: immediately
        ? 'Subscription canceled immediately'
        : 'Subscription will cancel at end of billing period',
      effectiveDate: immediately ? new Date() : user.currentPeriodEnd,
    };
  }

  async resume(userId: string) {
    const user = await userService.find(userId);
    if (!user.stripeSubscriptionId) {
      throw new Error('No subscription to resume');
    }

    await stripe.subscriptions.update(user.stripeSubscriptionId, {
      cancel_at_period_end: false,
    });

    return { message: 'Subscription resumed' };
  }

  async upgrade(userId: string, newTier: 'premium' | 'enterprise') {
    const user = await userService.find(userId);
    const newPriceId = newTier === 'premium'
      ? process.env.STRIPE_PREMIUM_PRICE_ID
      : process.env.STRIPE_ENTERPRISE_PRICE_ID;

    if (user.stripeSubscriptionId) {
      // Upgrade existing subscription (prorated)
      const subscription = await stripe.subscriptions.retrieve(user.stripeSubscriptionId);

      await stripe.subscriptions.update(user.stripeSubscriptionId, {
        items: [
          {
            id: subscription.items.data[0].id,
            price: newPriceId,
          },
        ],
        proration_behavior: 'create_prorations',
      });
    } else {
      // Create new subscription
      const checkoutUrl = await createCheckoutSession({
        userId,
        tier: newTier,
        successUrl: `${process.env.APP_URL}/subscription/success`,
        cancelUrl: `${process.env.APP_URL}/subscription/cancel`,
      });

      return { checkoutUrl };
    }

    return { message: `Upgraded to ${newTier}` };
  }
}
```

## Pattern 6: Revenue Tracking & MRR Calculation

```typescript
// src/modules/payment/revenue-tracker.service.ts
export async function trackRevenue(invoice: Stripe.Invoice) {
  const amount = invoice.total / 100; // Convert cents to dollars
  const userId = invoice.subscription_details?.metadata?.userId;
  const tier = invoice.subscription_details?.metadata?.tier;

  await revenueRepository.create({
    userId,
    tier,
    amount,
    currency: invoice.currency,
    invoiceId: invoice.id,
    paidAt: new Date(invoice.status_transitions.paid_at! * 1000),
  });

  // Calculate MRR (Monthly Recurring Revenue)
  const mrr = await calculateMRR();

  metrics.gauge('revenue.mrr', mrr);
  metrics.increment('revenue.payment_received', amount, { tier });
}

async function calculateMRR(): Promise<number> {
  // Get all active subscriptions
  const activeSubscriptions = await stripe.subscriptions.list({
    status: 'active',
    limit: 100,
  });

  let totalMRR = 0;

  for (const sub of activeSubscriptions.data) {
    const planAmount = sub.items.data[0].price.unit_amount || 0;
    const interval = sub.items.data[0].price.recurring?.interval;

    // Normalize to monthly
    if (interval === 'month') {
      totalMRR += planAmount / 100;
    } else if (interval === 'year') {
      totalMRR += planAmount / 100 / 12;
    }
  }

  return totalMRR;
}
```

## Related Documentation

- Stripe Node.js SDK: https://stripe.com/docs/api/node
- Telegram Payments: https://core.telegram.org/bots/payments
- Stripe Webhooks: https://stripe.com/docs/webhooks
