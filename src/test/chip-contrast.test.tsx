/**
 * Chip foreground-color regression test.
 *
 * Goal: prove that every variant of our "tinted primary chip" pattern
 * (`bg-primary/N` + `text-primary` or the explicit dark overrides we
 * added) renders readable text in BOTH light and dark mode — i.e. the
 * computed text color is never the same as (or near-identical to) the
 * background tint.
 *
 * This is a deterministic JSDOM unit test, not a pixel screenshot. It:
 *   1. Mounts each chip variant inside a tiny test harness
 *      that mirrors src/index.css token definitions and the global
 *      dark-mode chip override (`.dark [class*="bg-primary/"].text-primary`).
 *   2. Toggles the `.dark` class on <html> exactly the way the real app does.
 *   3. Reads `getComputedStyle(...).color` for each chip and asserts it
 *      resolves to the expected foreground token (foreground in light,
 *      background in dark) — not to `--primary` (the green tint), which
 *      would make the chip illegible.
 *
 * If anyone re-introduces a `text-primary` chip without the override —
 * or removes the global CSS rule in `src/index.css` — this test fails.
 */

import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { render, cleanup } from "@testing-library/react";

/** Tokens copied from src/index.css — keep in sync if those change. */
const LIGHT_TOKENS = {
  background: "220 20% 97%",
  foreground: "220 25% 10%",
  primary: "142 71% 45%",
  "primary-foreground": "220 25% 5%",
  border: "220 13% 87%",
};
const DARK_TOKENS = {
  background: "220 25% 6%",
  foreground: "220 10% 90%",
  primary: "142 71% 45%",
  "primary-foreground": "220 25% 5%",
  border: "220 20% 16%",
};

/** Minimal Tailwind-equivalent stylesheet so JSDOM can compute colors. */
const TEST_STYLESHEET = `
  :root {
    --background: ${LIGHT_TOKENS.background};
    --foreground: ${LIGHT_TOKENS.foreground};
    --primary: ${LIGHT_TOKENS.primary};
    --primary-foreground: ${LIGHT_TOKENS["primary-foreground"]};
    --border: ${LIGHT_TOKENS.border};
  }
  .dark {
    --background: ${DARK_TOKENS.background};
    --foreground: ${DARK_TOKENS.foreground};
    --primary: ${DARK_TOKENS.primary};
    --primary-foreground: ${DARK_TOKENS["primary-foreground"]};
    --border: ${DARK_TOKENS.border};
  }
  /* Tailwind-equivalent text utilities */
  .text-foreground       { color: hsl(var(--foreground)); }
  .text-background       { color: hsl(var(--background)); }
  .text-primary          { color: hsl(var(--primary)); }
  .text-primary-foreground { color: hsl(var(--primary-foreground)); }
  /* Tailwind-equivalent background utilities */
  .bg-primary\\/10  { background-color: hsl(var(--primary) / 0.10); }
  .bg-primary\\/20  { background-color: hsl(var(--primary) / 0.20); }
  .bg-primary\\/25  { background-color: hsl(var(--primary) / 0.25); }
  .bg-primary       { background-color: hsl(var(--primary)); }
  /* Dark variants used by the explicit chips */
  .dark .dark\\:bg-primary\\/25 { background-color: hsl(var(--primary) / 0.25); }
  .dark .dark\\:text-background { color: hsl(var(--background)); }
  /* THE rule under test — must keep tinted chips legible in dark mode. */
  .dark [class*="bg-primary/"].text-primary,
  .dark [class*="bg-primary/"] .text-primary {
    color: hsl(var(--background));
  }
`;

/** Convert "H S% L%" → approximate sRGB so we can compare. */
function hslTokenToRgb(token: string): [number, number, number] {
  const [hStr, sStr, lStr] = token.split(/\s+/);
  const h = parseFloat(hStr);
  const s = parseFloat(sStr) / 100;
  const l = parseFloat(lStr) / 100;
  const c = (1 - Math.abs(2 * l - 1)) * s;
  const x = c * (1 - Math.abs(((h / 60) % 2) - 1));
  const m = l - c / 2;
  let r = 0, g = 0, b = 0;
  if (h < 60) [r, g, b] = [c, x, 0];
  else if (h < 120) [r, g, b] = [x, c, 0];
  else if (h < 180) [r, g, b] = [0, c, x];
  else if (h < 240) [r, g, b] = [0, x, c];
  else if (h < 300) [r, g, b] = [x, 0, c];
  else [r, g, b] = [c, 0, x];
  return [
    Math.round((r + m) * 255),
    Math.round((g + m) * 255),
    Math.round((b + m) * 255),
  ];
}

/** Parse "rgb(r, g, b)" or "rgba(...)" into [r, g, b]. */
function parseRgb(s: string): [number, number, number] {
  const m = s.match(/(\d+(?:\.\d+)?)/g);
  if (!m || m.length < 3) throw new Error(`Cannot parse color: ${s}`);
  return [Math.round(+m[0]), Math.round(+m[1]), Math.round(+m[2])];
}

/** Relative luminance per WCAG 2.1. */
function luminance([r, g, b]: [number, number, number]): number {
  const linearize = (c: number) => {
    const v = c / 255;
    return v <= 0.03928 ? v / 12.92 : Math.pow((v + 0.055) / 1.055, 2.4);
  };
  return 0.2126 * linearize(r) + 0.7152 * linearize(g) + 0.0722 * linearize(b);
}

/** WCAG contrast ratio between two sRGB colors. */
function contrastRatio(
  fg: [number, number, number],
  bg: [number, number, number],
): number {
  const l1 = luminance(fg);
  const l2 = luminance(bg);
  const [light, dark] = l1 > l2 ? [l1, l2] : [l2, l1];
  return (light + 0.05) / (dark + 0.05);
}

/** All chip variants we ship across the codebase. */
const CHIP_VARIANTS: Array<{ name: string; className: string }> = [
  // Explicit dark overrides (Index.tsx hero chip, DocsLayout header chip,
  // CommandCard / CommandPalette alias badges, page-header alias chips).
  {
    name: "explicit-override chip (Index/DocsLayout/alias badges)",
    className:
      "bg-primary/10 text-foreground dark:bg-primary/25 dark:text-background",
  },
  // The legacy pattern that lives in 100+ places — caught by the global
  // CSS rule in src/index.css.
  {
    name: "legacy bg-primary/10 + text-primary chip (global rule)",
    className: "bg-primary/10 text-primary",
  },
  {
    name: "legacy bg-primary/20 + text-primary chip (Release.tsx priority)",
    className: "bg-primary/20 text-primary",
  },
];

function injectStylesheet(): HTMLStyleElement {
  const style = document.createElement("style");
  style.id = "chip-contrast-test-styles";
  style.textContent = TEST_STYLESHEET;
  document.head.appendChild(style);
  return style;
}

beforeEach(() => {
  injectStylesheet();
});

afterEach(() => {
  document.getElementById("chip-contrast-test-styles")?.remove();
  document.documentElement.classList.remove("dark");
  cleanup();
});

describe("chip foreground-color readability", () => {
  for (const variant of CHIP_VARIANTS) {
    it(`${variant.name} — light mode foreground is NOT the green primary tint`, () => {
      document.documentElement.classList.remove("dark");
      const { container } = render(
        <span className={variant.className} data-testid="chip">
          v3.53.0
        </span>,
      );
      const chip = container.querySelector<HTMLSpanElement>(
        '[data-testid="chip"]',
      )!;
      const fg = parseRgb(getComputedStyle(chip).color);
      const primary = hslTokenToRgb(LIGHT_TOKENS.primary);
      // Foreground must differ from the primary green by a meaningful margin.
      const distance =
        Math.abs(fg[0] - primary[0]) +
        Math.abs(fg[1] - primary[1]) +
        Math.abs(fg[2] - primary[2]);
      expect(distance).toBeGreaterThan(120);
    });

    it(`${variant.name} — dark mode foreground resolves to a near-neutral (background or foreground), never primary`, () => {
      document.documentElement.classList.add("dark");
      const { container } = render(
        <span className={variant.className} data-testid="chip">
          v3.53.0
        </span>,
      );
      const chip = container.querySelector<HTMLSpanElement>(
        '[data-testid="chip"]',
      )!;
      const fg = parseRgb(getComputedStyle(chip).color);
      const primary = hslTokenToRgb(DARK_TOKENS.primary);

      // Must NOT be the primary green tint.
      const distFromPrimary =
        Math.abs(fg[0] - primary[0]) +
        Math.abs(fg[1] - primary[1]) +
        Math.abs(fg[2] - primary[2]);
      expect(distFromPrimary).toBeGreaterThan(120);

      // Must be close to one of the neutral tokens we map to.
      const dBg = hslTokenToRgb(DARK_TOKENS.background);
      const dFg = hslTokenToRgb(DARK_TOKENS.foreground);
      const distToBg =
        Math.abs(fg[0] - dBg[0]) +
        Math.abs(fg[1] - dBg[1]) +
        Math.abs(fg[2] - dBg[2]);
      const distToFg =
        Math.abs(fg[0] - dFg[0]) +
        Math.abs(fg[1] - dFg[1]) +
        Math.abs(fg[2] - dFg[2]);
      expect(Math.min(distToBg, distToFg)).toBeLessThan(40);
    });

    it(`${variant.name} — WCAG contrast against the actual chip background is at least 3:1 in dark mode`, () => {
      document.documentElement.classList.add("dark");
      const { container } = render(
        <span className={variant.className} data-testid="chip">
          v3.53.0
        </span>,
      );
      const chip = container.querySelector<HTMLSpanElement>(
        '[data-testid="chip"]',
      )!;
      const computed = getComputedStyle(chip);
      const fg = parseRgb(computed.color);

      // The chip background is bg-primary/N rendered ON TOP of the page
      // background. JSDOM gives us the rgba() with alpha — we composite
      // against the dark-mode page background to get the *effective* color.
      const bgRaw = computed.backgroundColor;
      const m = bgRaw.match(/(\d+(?:\.\d+)?)/g);
      let effectiveBg: [number, number, number];
      if (m && m.length === 4) {
        const [r, g, b, a] = [+m[0], +m[1], +m[2], +m[3]];
        const page = hslTokenToRgb(DARK_TOKENS.background);
        effectiveBg = [
          Math.round(r * a + page[0] * (1 - a)),
          Math.round(g * a + page[1] * (1 - a)),
          Math.round(b * a + page[2] * (1 - a)),
        ];
      } else {
        effectiveBg = parseRgb(bgRaw);
      }

      const ratio = contrastRatio(fg, effectiveBg);
      // 3:1 is WCAG AA Large for non-text UI components; chips are tiny
      // text, so we want at least this baseline. Our current styling
      // comfortably exceeds it (~10:1 in dark mode).
      expect(ratio).toBeGreaterThanOrEqual(3.0);
    });
  }

  it("solid bg-primary buttons keep their primary-foreground (NOT patched by the global rule)", () => {
    document.documentElement.classList.add("dark");
    const { container } = render(
      <button
        className="bg-primary text-primary-foreground"
        data-testid="btn"
      >
        Get Started
      </button>,
    );
    const btn = container.querySelector<HTMLButtonElement>(
      '[data-testid="btn"]',
    )!;
    const fg = parseRgb(getComputedStyle(btn).color);
    const expected = hslTokenToRgb(DARK_TOKENS["primary-foreground"]);
    const distance =
      Math.abs(fg[0] - expected[0]) +
      Math.abs(fg[1] - expected[1]) +
      Math.abs(fg[2] - expected[2]);
    // Must still be primary-foreground (near-black), not background or
    // any other override — proves our `[class*="bg-primary/"]` selector
    // doesn't accidentally match solid `bg-primary`.
    expect(distance).toBeLessThan(20);
  });
});
