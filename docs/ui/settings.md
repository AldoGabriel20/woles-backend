---
name: Emerald Concierge
colors:
  surface: '#f8faf6'
  surface-dim: '#d8dbd7'
  surface-bright: '#f8faf6'
  surface-container-lowest: '#ffffff'
  surface-container-low: '#f2f4f0'
  surface-container: '#eceeeb'
  surface-container-high: '#e7e9e5'
  surface-container-highest: '#e1e3df'
  on-surface: '#191c1a'
  on-surface-variant: '#404944'
  inverse-surface: '#2e312f'
  inverse-on-surface: '#eff1ed'
  outline: '#707974'
  outline-variant: '#bfc9c3'
  surface-tint: '#2b6954'
  primary: '#003527'
  on-primary: '#ffffff'
  primary-container: '#064e3b'
  on-primary-container: '#80bea6'
  inverse-primary: '#95d3ba'
  secondary: '#006c49'
  on-secondary: '#ffffff'
  secondary-container: '#6cf8bb'
  on-secondary-container: '#00714d'
  tertiary: '#2a2f2d'
  on-tertiary: '#ffffff'
  tertiary-container: '#404543'
  on-tertiary-container: '#aeb2af'
  error: '#ba1a1a'
  on-error: '#ffffff'
  error-container: '#ffdad6'
  on-error-container: '#93000a'
  primary-fixed: '#b0f0d6'
  primary-fixed-dim: '#95d3ba'
  on-primary-fixed: '#002117'
  on-primary-fixed-variant: '#0b513d'
  secondary-fixed: '#6ffbbe'
  secondary-fixed-dim: '#4edea3'
  on-secondary-fixed: '#002113'
  on-secondary-fixed-variant: '#005236'
  tertiary-fixed: '#dfe3e0'
  tertiary-fixed-dim: '#c3c8c4'
  on-tertiary-fixed: '#181d1a'
  on-tertiary-fixed-variant: '#434845'
  background: '#f8faf6'
  on-background: '#191c1a'
  surface-variant: '#e1e3df'
typography:
  display:
    fontFamily: Geist
    fontSize: 48px
    fontWeight: '600'
    lineHeight: '1.1'
    letterSpacing: -0.04em
  headline-lg:
    fontFamily: Geist
    fontSize: 32px
    fontWeight: '600'
    lineHeight: '1.2'
    letterSpacing: -0.03em
  headline-lg-mobile:
    fontFamily: Geist
    fontSize: 24px
    fontWeight: '600'
    lineHeight: '1.2'
    letterSpacing: -0.02em
  headline-md:
    fontFamily: Geist
    fontSize: 20px
    fontWeight: '600'
    lineHeight: '1.4'
    letterSpacing: -0.01em
  body-lg:
    fontFamily: Geist
    fontSize: 16px
    fontWeight: '400'
    lineHeight: '1.6'
    letterSpacing: -0.01em
  body-md:
    fontFamily: Geist
    fontSize: 14px
    fontWeight: '400'
    lineHeight: '1.5'
    letterSpacing: '0'
  label-md:
    fontFamily: Geist
    fontSize: 12px
    fontWeight: '500'
    lineHeight: '1.2'
    letterSpacing: 0.02em
rounded:
  sm: 0.25rem
  DEFAULT: 0.5rem
  md: 0.75rem
  lg: 1rem
  xl: 1.5rem
  full: 9999px
spacing:
  margin-page: 2rem
  gutter: 1.5rem
  card-padding: 1.5rem
  stack-gap: 1rem
  sidebar-width: 280px
---

## Brand & Style
The design system is built on the narrative of a high-end personal concierge. It transforms administrative tasks into a premium experience through a "Sophisticated Modern" aesthetic. The system utilizes a blend of **Minimalism** and **Glassmorphism** to evoke a sense of clarity, calm, and exclusive service.

The target audience expects precision and elegance. By utilizing deep emerald tones against ethereal off-white backgrounds, the UI feels breathable yet authoritative. Visual weight is managed through layered transparency and soft blurs, ensuring the interface feels like a lightweight digital assistant rather than a heavy utility tool.

## Colors
This design system uses a sophisticated, low-fatigue palette. 
- **Primary**: Deep Emerald (#064E3B) is used for brand moments, primary actions, and active navigation states.
- **Surface**: The core background is an off-white Sage (#F8FAF6), complemented by subtle mesh gradients of slightly warmer and cooler greens to provide depth without clutter.
- **Glass**: Semi-transparent whites (rgba 255, 255, 255, 0.7) are used for elevated cards to create the glassmorphic effect.
- **Accents**: Subtle emerald tints are used for success states and interactive highlights.

## Typography
The system utilizes **Geist** for its technical precision and clean, geometric proportions. 
- **Tracking**: Tighten letter-spacing on headlines to create a customized, high-end editorial feel.
- **Weight**: Leverage Semibold (600) for structural headers to establish a clear hierarchy.
- **Hierarchy**: Display and Headline styles should be used sparingly for page titles and summary stats to maintain an airy, "concierge" feel.

## Layout & Spacing
The layout follows a **Fluid Grid** model with generous safe areas.
- **Sidebar**: A fixed 280px vertical sidebar on desktop, utilizing a glassmorphic background blur (20px) to separate it from the content area.
- **Content**: A 12-column system for the main stage. Use wider gutters (24px) to ensure elements don't feel cramped.
- **Responsive**: On mobile, the sidebar collapses into a bottom navigation bar or a full-screen glass overlay. Page margins reduce to 1rem to maximize screen real estate.

## Elevation & Depth
Depth is achieved through **Glassmorphism** and soft ambient light:
- **Surface**: Use a backdrop filter (blur: 16px) combined with a high-transparency white fill for all primary cards and the sidebar.
- **Shadows**: Only one level of shadow is used (Elevation-1): `0px 10px 30px rgba(6, 78, 59, 0.04)`. This creates a soft lift without looking heavy.
- **Borders**: All glass elements feature a 1px solid border (#E2E8F0) to define edges against the mesh background.

## Shapes
The shape language is "Generous and Soft." 
- **Base Radius**: 16px (1rem) for all standard cards and containers.
- **Inner Elements**: Buttons and inputs should maintain 12px (0.75rem) to sit harmoniously within the larger containers.
- **Interactive States**: Hovering over elements triggers a subtle `scale(0.98)` transition, providing a tactile, "squishy" feedback loop that feels high-end.

## Components
- **Sidebar**: A unified column featuring thin-line icons. Active states use a Deep Emerald (#064E3B) glyph and a soft tinted background capsule.
- **Cards**: Feature a semi-transparent white background, 16px radius, and a 1px #E2E8F0 border. Headers within cards should use `label-md` for metadata.
- **Buttons**:
  - *Primary*: Deep Emerald fill with white text. No shadow, flat, 12px radius.
  - *Ghost*: Transparent with 1px border.
- **Inputs**: Large, 48px height fields with off-white fills and soft borders. Focus state uses a 2px Emerald ring.
- **Chips**: Small, pill-shaped tags used for status (e.g., "Paid", "Pending") with high-contrast text and desaturated background tints.
- **Concierge Widget**: A signature floating action component in the bottom right, using a glassmorphic circle with a primary emerald icon.