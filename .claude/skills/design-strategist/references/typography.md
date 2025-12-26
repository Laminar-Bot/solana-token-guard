# Typography Reference

Deep guidance on type selection, pairing, and typographic systems.

## Type Classification & Personality

### Serif Families
- **Old Style (Garamond, Palatino):** Classic, literary, trustworthy. Good for: editorial, luxury, traditional brands.
- **Transitional (Times, Baskerville):** Balanced, professional, neutral. Good for: news, corporate, formal documents.
- **Modern/Didone (Bodoni, Didot):** High contrast, elegant, fashion-forward. Good for: luxury, editorial headers.
- **Slab Serif (Rockwell, Clarendon):** Bold, sturdy, industrial. Good for: tech, startups, confident brands.

### Sans-Serif Families
- **Grotesque (Helvetica, Arial):** Neutral, ubiquitous, safe. Risk: generic, no personality.
- **Neo-Grotesque (Univers, Aktiv Grotesk):** Refined neutral, slightly more character.
- **Geometric (Futura, Avenir, Montserrat):** Modern, clean, constructed. Good for: tech, minimal design.
- **Humanist (Gill Sans, Frutiger, Open Sans):** Warm, readable, friendly. Good for: body text, accessible design.

### Display & Specialty
- **Script:** Elegant or casual handwriting. Use sparingly, never for body.
- **Display/Decorative:** Highly stylized. Headlines only, in small doses.
- **Monospace:** Technical, code-like. Good for: dev tools, data, retro aesthetics.

## Type Pairing Strategies

### The Safe Pair
One serif, one sans-serif from the same designer or era.
- **Example:** Adobe Caslon (serif) + Adobe Myriad (sans)
- **Why it works:** Similar proportions and stroke weights.

### The Contrast Pair
Opposites that complement.
- **Example:** Playfair Display (high-contrast serif) + Source Sans Pro (humanist sans)
- **Why it works:** Clear roles—display for impact, sans for readability.

### The Superfamily
Type families with serif, sans, and mono variants designed together.
- **Examples:** IBM Plex, Noto, PT Sans/Serif
- **Why it works:** Built-in harmony, no guesswork.

### The Modern Minimal
Two sans-serifs with different personalities.
- **Example:** Space Grotesk (geometric display) + Inter (neutral body)
- **Caution:** Needs careful weight/size differentiation or they'll blur.

## Type Pairing Anti-Patterns

- **Two decorative faces:** Visual chaos. One must be quiet.
- **Similar but different:** Two humanist sans or two transitional serifs. Too close to be intentional, too different to be harmonious.
- **Scripts with scripts:** Just don't.

## Building a Type Scale

Use a consistent ratio between sizes. Common ratios:
- **1.125 (Major Second):** Subtle, dense interfaces.
- **1.250 (Major Third):** Balanced, most applications.
- **1.333 (Perfect Fourth):** More dramatic hierarchy.
- **1.618 (Golden Ratio):** Editorial, expressive.

**Example scale at 1.250:**
- xs: 12px
- sm: 14px
- base: 16px
- lg: 20px
- xl: 25px
- 2xl: 31px
- 3xl: 39px

## Line Height Guidelines

- **Headings:** 1.1–1.3× (tight, punchy)
- **Body text:** 1.4–1.6× (comfortable reading)
- **Small text/captions:** 1.3–1.5× (tighter is often okay)

Line height should increase slightly as column width increases.

## Line Length (Measure)

- **Ideal:** 45–75 characters per line.
- **Minimum comfortable:** 40 characters.
- **Maximum comfortable:** 85 characters (gets exhausting).

**Practical tip:** If you can't control line length directly, use max-width on containers.

## Letter Spacing (Tracking)

- **Uppercase text:** Add 2–5% tracking. UPPERCASE IS HARD TO READ OTHERWISE.
- **Display headlines:** Often benefits from slight negative tracking (-1–2%).
- **Body text:** Leave default. Don't touch it.
- **Small text:** Slight positive tracking (1–2%) improves legibility.

## Font Weight Usage

- **Regular (400):** Body text default.
- **Medium (500):** Subtle emphasis, UI labels.
- **Semibold (600):** Strong emphasis, subheadings.
- **Bold (700):** Headlines, key actions.
- **Light (300):** Use sparingly; often hard to read on screens.

**Principle:** Contrast should be obvious. If you're using 400 vs 500, most users won't notice. Use 400 vs 700 for clear differentiation.

## Variable Fonts

Modern approach: one font file with adjustable axes (weight, width, slant).

**Benefits:**
- Single file, multiple weights = faster loading.
- Fine-tuned weight for optical sizing.
- Smooth animations between weights.

**Caution:** Not all browsers render consistently. Test.

## Typography Checklist

Before finalizing type choices:
- [ ] Does the typeface match the brand voice?
- [ ] Is body text readable at 16px on screens?
- [ ] Is hierarchy clear with at least 3 distinct levels?
- [ ] Does the scale work on mobile AND desktop?
- [ ] Are line lengths controlled (45-75 chars)?
- [ ] Is line height comfortable (1.4-1.6× for body)?
- [ ] Do pairings have clear roles (display vs body)?
- [ ] Have I tested with real content, not Lorem Ipsum?
