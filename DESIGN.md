# Design System Strategy: The Sun-Drenched Library

## 1. Overview & Creative North Star
The Creative North Star for this design system is **"The Digital Vellum."**

We are moving away from the cold, sterile "app" aesthetic and toward a high-end editorial experience that feels like a sun-drenched library at mid-afternoon. The goal is to facilitate "Deep Work" and "Long-Form Thought" by reducing cognitive load and visual friction.

By leveraging a low-contrast, warm-toned palette and sophisticated Newsreader typography, we create an environment that feels organic, tactile, and premium. We break the "template" look by favoring intentional asymmetry, generous whitespace (using the `16` and `24` spacing tokens), and a total rejection of traditional structural borders. This is not a utility tool; it is a digital sanctuary.

---

## 2. Colors & Surface Philosophy
The palette is rooted in the warmth of parchment and soft sand, designed to be easy on the eyes during extended sessions.

* **Primary (`#8b4d3f`):** A muted terracotta used for focal points and calls to action.
* **Secondary (`#5a694f`):** A deep moss green for secondary actions and organic highlights.
* **Neutral Surfaces:** The background (`#fffcf7`) and surface tiers (`#f6f4ec` to `#e4e3d7`) form the "Vellum" layers.

### The "No-Line" Rule
**Explicit Instruction:** Traditional 1px solid borders are strictly prohibited for sectioning. Structural boundaries must be defined solely through background color shifts.
* *Example:* A sidebar should be `surface_container_low` against a main content area of `surface`. If you feel the need for a line, increase the spacing (`spacing.4` or `spacing.6`) or shift the tonal value instead.

### Surface Hierarchy & Nesting
Treat the UI as a series of physical layers—like stacked sheets of fine paper.
* **Layer 0 (Base):** `surface` (`#fffcf7`)
* **Layer 1 (Cards/Sections):** `surface_container_low` (`#fcf9f3`)
* **Layer 2 (Embedded Elements):** `surface_container` (`#f6f4ec`)
* **Layer 3 (Highest Interaction):** `surface_container_highest` (`#eae8de`)

### The "Glass & Gradient" Rule
To add "soul," use subtle linear gradients (e.g., `primary` to `primary_container`) for large CTAs. For floating menus, employ **Glassmorphism**: use `surface` with 80% opacity and a `backdrop-blur` of 12px to allow the warmth of the library to bleed through the navigation.

---

## 3. Typography: The Editorial Voice
The typography is the backbone of the "Digital Vellum" concept. We use **Newsreader**, a serif designed for high legibility and character, alongside **Work Sans** for functional labels.

* **Display & Headlines:** Use `display-lg` (3.5rem) and `headline-lg` (2rem) to create an editorial feel. These should be set with tighter letter-spacing (-0.02em) to feel intentional and "inked."
* **Body Copy:** All long-form text must use `body-lg` (Newsreader). The low contrast between `on_surface` (`#383831`) and `surface` ensures that the "ink" doesn't vibrate against the "paper."
* **Functional Labels:** Use `label-md` (Work Sans) for buttons and navigation. This sans-serif provides a modern, functional counterpoint to the romanticism of the serif body text.

---

## 4. Elevation & Depth: Tonal Layering
Depth in this system is achieved through "Tonal Layering" rather than traditional drop shadows.

* **The Layering Principle:** To lift an element, move it one step up the surface-container scale. A `surface_container_highest` card on a `surface` background provides a soft, natural lift.
* **Ambient Shadows:** If a floating effect is required (e.g., a modal), use an extra-diffused shadow: `box-shadow: 0 12px 40px rgba(56, 56, 49, 0.06);`. Note the shadow color uses `on_surface` at a very low opacity to mimic natural ambient light.
* **The "Ghost Border" Fallback:** For accessibility in inputs, use the `outline_variant` token at 20% opacity. 100% opaque borders are forbidden as they "cut" the parchment texture.

---

## 5. Components

### Buttons
* **Primary:** Background `primary`, text `on_primary`. Roundedness: `md` (0.375rem). No shadow.
* **Secondary:** Background `secondary_container`, text `on_secondary_container`.
* **Tertiary:** Text-only using `primary`, but with a subtle `surface_container_high` background on hover.

### Input Fields
* **Styling:** Use a "filled" style with `surface_container_low`.
* **Indicator:** Instead of a bottom border, use a 2px vertical accent of `primary` on the left side during the `focus` state to signify the "insertion point" of the user's thought.

### Cards & Lists
* **Rule:** Forbid divider lines.
* **Implementation:** Use `spacing.4` to separate list items. For cards, use a slight tonal shift (`surface_container_low`) and a `md` corner radius.

### Signature Component: The "Vellum Scroll"
For long-form writing interfaces, use a centered container with ultra-wide margins (`spacing.24`). The background should have a subtle, non-repeating noise texture (1% opacity) to simulate the tactile grain of paper.

---

## 6. Do's and Don'ts

### Do:
* **Do** use asymmetrical layouts. Place a headline on the left and the body text offset to the right to create a "custom-built" feel.
* **Do** prioritize vertical whitespace. If the content feels crowded, increase spacing using the `spacing.10` or `spacing.12` tokens.
* **Do** use `secondary` (Moss Green) for "success" or "organic" interactions rather than a bright functional green.

### Don't:
* **Don't** use 1px solid black or grey borders. This destroys the "Digital Vellum" illusion.
* **Don't** use pure white (`#ffffff`) for backgrounds unless it is the `surface_container_lowest` for a specific floating element.
* **Don't** use high-speed animations. Transitions should be slow and easing (e.g., 300ms ease-out) to match the "Sun-Drenched Library" pace.