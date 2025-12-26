# Engagement Patterns & UX Heuristics

Patterns that drive user behavior, increase engagement, and improve conversion.

## Core UX Laws

### Hick's Law
**More choices = slower decisions (or no decision).**
- Limit options at each decision point.
- Use progressive disclosure to reveal complexity gradually.
- Default to sensible choices; let users change if needed.

### Fitts's Law
**Time to reach a target = distance / size.**
- Make important actions large and easily reachable.
- Place frequent actions in thumb-friendly zones on mobile.
- Increase clickable area beyond visual bounds where needed.

### Miller's Law
**Working memory holds ~7 (±2) items.**
- Chunk information into digestible groups.
- Navigation with 15 items fails; group into 5 categories of 3.
- Use visual grouping to reduce cognitive parsing.

### Jakob's Law
**Users spend most time on other sites.**
- Follow conventions unless you have strong reason not to.
- Logo top-left, search top-right, cart icon, hamburger menu—these are learned.
- Innovation in interaction patterns creates friction.

### Aesthetic-Usability Effect
**Beautiful things seem to work better.**
- Visual polish increases perceived usability and trust.
- Users are more forgiving of minor issues in beautiful interfaces.
- But: beauty without function is decoration, not design.

### Peak-End Rule
**Experiences are judged by peaks and endings, not averages.**
- Invest in delightful moments (first success, achievement unlocked).
- End flows on a high note (confirmation screens, thank you pages).
- One terrible moment can override many good ones.

## Engagement Patterns

### The Hook Model (Nir Eyal)
1. **Trigger:** External (notification) or internal (emotion/habit).
2. **Action:** Simple behavior anticipating reward.
3. **Variable Reward:** Unpredictable positive outcome.
4. **Investment:** User puts something in, increasing likelihood of return.

**Ethical use:** Build habits that benefit users, not just metrics.

### Progress & Completeness

**Endowed Progress Effect:** Users who feel they've started are more likely to finish.
- Show progress bars starting at 20%, not 0%.
- Celebrate small wins ("Profile 40% complete").
- Checklists with some items pre-checked outperform empty checklists.

**Goal Gradient Effect:** Effort increases as goal approaches.
- "3 items left to unlock your reward" drives completion.
- Progress bars accelerating toward the end feel faster.

### Social Proof Patterns

**Numbers:** "Join 50,000+ teams" (specificity = credibility).
**Testimonials:** Real names, photos, specific outcomes.
**Activity indicators:** "Sarah from Denver just signed up" (FOMO + social proof).
**Trust badges:** Security seals, press logos, certifications.

**Placement:** Near conversion points (pricing, checkout, signup forms).

### Scarcity & Urgency

**Scarcity:** "Only 3 left in stock" (limited supply).
**Urgency:** "Sale ends in 2:34:15" (limited time).

**Ethical use:** Real scarcity only. Fake urgency destroys trust when discovered.
**Dark pattern warning:** Perpetual countdown timers, fake stock levels.

### Reciprocity

Give before asking.
- Free trial, free tier, free resource.
- Valuable content before email capture.
- Help users succeed first, then ask for commitment.

## Conversion Optimization

### Landing Page Hierarchy

1. **Hero:** What is this? (Value proposition in <5 words)
2. **Subhead:** Why should I care? (Key benefit)
3. **Primary CTA:** What should I do? (Single clear action)
4. **Social proof:** Why should I trust you? (Testimonials, logos)
5. **Features/Benefits:** What do I get? (Outcomes, not features)
6. **Secondary CTA:** Still not sure? (Lower commitment option)
7. **FAQ/Objections:** What's stopping me? (Address concerns)

### CTA Optimization

**Copy:**
- Action-oriented ("Start free trial" > "Submit")
- Benefit-focused ("Get my free guide" > "Download")
- Low commitment where appropriate ("See how it works")

**Visual:**
- Contrast with surrounding elements.
- Sufficient size (Fitts's Law).
- Whitespace isolation.

**Placement:**
- Above the fold for primary action.
- Repeated after major content sections.
- Sticky CTAs for long pages (subtle, not obnoxious).

### Form Optimization

1. **Reduce fields.** Every field is friction. Ask only what's necessary.
2. **Smart defaults.** Pre-fill what you know (country from IP, etc.).
3. **Inline validation.** Immediate feedback > submit-then-error.
4. **Progress indication.** For multi-step forms, show where they are.
5. **Mobile-friendly inputs.** Proper keyboard types, large touch targets.

**Field reduction priority:**
- Required for function (email for account).
- Required for personalization (name).
- Optional enhancement (company size).
- Marketing nice-to-have (how did you hear about us).

Cut from the bottom up.

### Checkout Optimization

- **Guest checkout option.** Forcing account creation kills conversions.
- **Trust signals near payment.** Security badges, guarantees.
- **Transparent pricing.** No surprise fees at the end.
- **Multiple payment options.** Card, PayPal, Apple Pay, etc.
- **Order summary visible.** Users want to verify before committing.

## Friction Analysis Framework

### Good Friction (Keep or Add)
- **Confirmation for destructive actions:** "Delete account? This cannot be undone."
- **Progressive disclosure:** Hiding advanced options until needed.
- **Verification for security:** 2FA, email confirmation.
- **Pause points for big decisions:** "Review your order" before purchase.

### Bad Friction (Remove)
- **Unnecessary steps:** Why is this a 5-step wizard?
- **Redundant information requests:** You already know my email.
- **Poor defaults:** Making users configure everything.
- **Hidden information:** Burying pricing, requiring signup to see features.
- **Confusing navigation:** Users can't find what they need.
- **Error states without guidance:** "Something went wrong" helps no one.

## Micro-Interaction Patterns

### Feedback Loops
- Button press → visual change (color, scale, ripple).
- Form submission → loading state → success/error.
- Toggle → immediate state reflection.

**Principle:** Every action needs acknowledgment. Silence is confusing.

### Loading States
- Skeleton screens > spinners (perceived performance).
- Progress indicators for known duration.
- Optimistic UI for fast operations (assume success, rollback if failure).

### Empty States
- Explain why empty.
- Show what could be here.
- Provide clear action to fill it.

"No messages yet. Start a conversation →" beats "No data."

### Error States
- What went wrong (in human language).
- Why it might have happened.
- What to do about it.

"Your password must include a number" beats "Invalid input."

## Mobile-Specific Patterns

### Thumb Zone
- Primary actions in bottom third (easy reach).
- Destructive/rare actions in top corners (harder to hit accidentally).
- Avoid small targets at screen edges.

### Gesture Conventions
- Swipe right: reveal actions, go back.
- Swipe left: delete, secondary actions.
- Pull down: refresh.
- Swipe up: dismiss, access more.

Don't reinvent unless there's compelling reason.

### Bottom Navigation
- Max 5 items (more = confusion).
- Icons + labels (icons alone are ambiguous).
- Active state clearly differentiated.

## Measuring Engagement

### Key Metrics
- **Activation rate:** % of new users who complete key action.
- **Engagement rate:** Frequency/depth of interaction.
- **Retention rate:** % returning after Day 1, 7, 30.
- **Conversion rate:** % completing desired action.
- **Time to value:** How fast users get benefit.

### Qualitative Signals
- User interviews revealing confusion.
- Support tickets about specific flows.
- Rage clicks and dead clicks in analytics.
- Drop-off points in funnel analysis.

## Ethical Boundaries

**Persuasion ≠ Manipulation**

Design should:
- Help users achieve their goals.
- Present honest information.
- Respect user autonomy.
- Make it easy to leave or cancel.

Design should not:
- Trick users into unintended actions.
- Hide costs or commitments.
- Make cancellation deliberately difficult.
- Exploit psychological vulnerabilities.

The line: Would users feel deceived if they understood the pattern? If yes, don't do it.
