---
name: Woles Life Admin
colors:
  surface: '#f8faf6'
  surface-dim: '#d8dbd7'
  surface-bright: '#f8faf6'
  surface-container-lowest: '#ffffff'
  surface-container-low: '#f2f4f1'
  surface-container: '#eceeeb'
  surface-container-high: '#e7e9e5'
  surface-container-highest: '#e1e3e0'
  on-surface: '#191c1b'
  on-surface-variant: '#404944'
  inverse-surface: '#2e312f'
  inverse-on-surface: '#eff1ee'
  outline: '#707974'
  outline-variant: '#bfc9c3'
  surface-tint: '#2b6954'
  primary: '#003527'
  on-primary: '#ffffff'
  primary-container: '#064e3b'
  on-primary-container: '#80bea6'
  inverse-primary: '#95d3ba'
  secondary: '#0058be'
  on-secondary: '#ffffff'
  secondary-container: '#2170e4'
  on-secondary-container: '#fefcff'
  tertiary: '#442800'
  on-tertiary: '#ffffff'
  tertiary-container: '#623c00'
  on-tertiary-container: '#f69f0d'
  error: '#ba1a1a'
  on-error: '#ffffff'
  error-container: '#ffdad6'
  on-error-container: '#93000a'
  primary-fixed: '#b0f0d6'
  primary-fixed-dim: '#95d3ba'
  on-primary-fixed: '#002117'
  on-primary-fixed-variant: '#0b513d'
  secondary-fixed: '#d8e2ff'
  secondary-fixed-dim: '#adc6ff'
  on-secondary-fixed: '#001a42'
  on-secondary-fixed-variant: '#004395'
  tertiary-fixed: '#ffddb8'
  tertiary-fixed-dim: '#ffb95f'
  on-tertiary-fixed: '#2a1700'
  on-tertiary-fixed-variant: '#653e00'
  background: '#f8faf6'
  on-background: '#191c1b'
  surface-variant: '#e1e3e0'
typography:
  display-lg:
    fontFamily: Geist
    fontSize: 48px
    fontWeight: '700'
    lineHeight: 56px
    letterSpacing: -0.02em
  headline-lg:
    fontFamily: Geist
    fontSize: 32px
    fontWeight: '600'
    lineHeight: 40px
    letterSpacing: -0.01em
  headline-lg-mobile:
    fontFamily: Geist
    fontSize: 28px
    fontWeight: '600'
    lineHeight: 36px
  headline-md:
    fontFamily: Geist
    fontSize: 24px
    fontWeight: '600'
    lineHeight: 32px
  title-lg:
    fontFamily: Inter
    fontSize: 20px
    fontWeight: '600'
    lineHeight: 28px
  body-lg:
    fontFamily: Inter
    fontSize: 18px
    fontWeight: '400'
    lineHeight: 28px
  body-md:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: '400'
    lineHeight: 24px
  label-md:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: '500'
    lineHeight: 20px
  label-sm:
    fontFamily: Inter
    fontSize: 12px
    fontWeight: '500'
    lineHeight: 16px
    letterSpacing: 0.02em
rounded:
  sm: 0.25rem
  DEFAULT: 0.5rem
  md: 0.75rem
  lg: 1rem
  xl: 1.5rem
  full: 9999px
spacing:
  base: 4px
  xs: 4px
  sm: 8px
  md: 16px
  lg: 24px
  xl: 32px
  2xl: 48px
  3xl: 64px
  container-margin-mobile: 16px
  container-margin-desktop: 40px
  gutter: 16px
---

## Brand & Style
The design system is built on the philosophy of "Woles"—an Indonesian term for being "slow" or "relaxed." It aims to transform the chaotic nature of personal life administration into a calm, structured, and effortless experience. The brand personality is **Calm, Trustworthy, Reliable, and Organized**, acting as a quiet partner in the user's daily life.

The aesthetic follows a **Modern SaaS** direction with a specific focus on the Indonesian consumer market, prioritizing a **mobile-first** experience. It draws inspiration from the utility of tools like Notion and the precision of Linear, while maintaining a friendlier, more approachable tone. The UI relies on generous whitespace, high-quality typography, and a "soft-touch" interface that feels premium yet accessible.

## Colors
The color strategy is anchored by **Deep Emerald Green**, representing stability and growth. This is complemented by a **Warm Blue** for secondary actions and a **Soft Amber** for highlights and gentle reminders. 

The background remains a clean **Off-white** to reduce eye strain, while interactive elements use white surfaces to create clear "islands" of information. High-contrast text is used against these light backgrounds to ensure maximum readability for the Indonesian consumer, who often navigates apps in varied lighting conditions.

## Typography
The typography system uses a pairing of **Geist** for headings and **Inter** for body text. Geist provides a technical, organized precision that feels reliable, while Inter ensures exceptional legibility across all device types.

Headlines should feel stable; we use tighter letter-spacing and heavier weights for titles to anchor the page. For body text, a generous 1.5x line height is maintained to ensure that long lists of tasks or administrative details do not feel overwhelming. Labels use a medium weight to distinguish them from standard body copy at smaller sizes.

## Layout & Spacing
The layout is based on a **4px/8px incremental grid**, ensuring consistent rhythm across all components. 

- **Mobile First:** A standard 4-column fluid grid with 16px side margins.
- **Tablet/Desktop:** Transitions to a 12-column fixed-width layout (max-width 1200px) centered on the screen.
- **Vertical Rhythm:** Elements are grouped using 8px (small), 16px (medium), or 24px (large) spacing to create clear semantic relationships between content blocks.

Whitespace is treated as a functional tool to prevent "information density anxiety."

## Elevation & Depth
This design system uses **Tonal Layering** combined with **Ambient Shadows**. Instead of heavy black shadows, we use soft, diffused shadows tinted with the primary emerald color to maintain a "natural" feel.

- **Level 0 (Base):** Background (#F9FAFB).
- **Level 1 (Cards):** Surface White (#FFFFFF) with a 1px border (#E5E7EB) or a very subtle 4px blur shadow.
- **Level 2 (Popovers/Modals):** Floating elements with a 12px blur, 10% opacity shadow (using a mix of Emerald-900 and Neutral colors).

Interactive elements slightly "lift" on hover/active states using a more pronounced shadow rather than a color change alone, emphasizing the tactile nature of the UI.

## Shapes
The shape language is **Rounded**, favoring friendliness and approachability. 

While the standard `rounded-md` (0.5rem/8px) is used for buttons and inputs to maintain a sense of order, container elements like cards use `rounded-lg` (1rem/12px) to soften the overall appearance of the screen. Chips and tags utilize a full "pill" shape to distinguish them as secondary interactive or metadata elements.

## Components
### Buttons
- **Primary:** Emerald-900 background, White text. 8px radius.
- **Secondary:** Blue-50 background, Blue-700 text. Used for supportive actions.
- **Ghost:** No background, Emerald-600 text. Used for tertiary navigation.

### Cards
Cards are the primary container. They must have a White background, 12px radius, and a 1px Subtle Gray border. In high-density views, shadows are removed in favor of borders to maintain "Notion-like" clarity.

### Input Fields
Inputs use an 8px radius with a 1px border. On focus, the border transitions to Primary Emerald with a soft 2px outer glow.

### Chips & Badges
Small, pill-shaped indicators. Status badges use low-saturation background tints (e.g., Amber-50 background with Amber-500 text for "Pending").

### Lists
List items should have 16px vertical padding with a subtle bottom divider. In mobile views, list items should have a minimum touch target height of 48px.

### Progress Indicators
Administrative tasks should use the Warm Blue for "in-progress" states and Emerald for "completed" states to provide a sense of psychological reward.