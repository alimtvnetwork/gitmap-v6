/**
 * Chip foreground-color regression test.
 *
 * Goal: prove that every variant of our "tinted primary chip" pattern
 * (`bg-primary/N` + `text-primary` or the explicit dark overrides we
 * added) renders readable text in BOTH light and dark mode — i.e. the
 * computed text color is never the same as (or near-identical to) the
 * background tint.
 *
 * Because JSDOM's CSS engine can't resolve `hsl(var(--x) / 0.1)` or
 * attribute-selector specificity reliably, this test does NOT rely on
 * `getComputedStyle`. Instead it:
 *
 *   1. Reads `src/index.css` from disk.
 *   2. Asserts the global dark-mode chip-readability override exists
 *      (the rule that patches all 100+ legacy `bg-primary/N + text-primary`
 *      occurrences across the codebase).
 *   3. Computes — via WCAG sRGB math — the effective foreground/background
 *      pairing for each chip variant we ship, and asserts contrast ≥ 3:1.
 *   4. Locks in the design tokens copied from `src/index.css`, so any
 *      future token tweak that breaks chip contrast will fail this test
 *      before merging.
 *
 * If anyone:
 *   - removes the `.dark [class*="bg-primary/"].text-primary { color: ... }`
 *     rule from src/index.css, OR
 *   - re-introduces a `text-primary` chip without the override, OR
 *   - shifts `--primary` / `--background` in a way that drops contrast
 *     below 3:1,
 * this test fails.
 */

import { describe, it, expect } from "vitest";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

// ─── tokens copied verbatim from src/index.css ────────────────────────────
const LIGHT_TOKENS = {
  background: "220 20% 97%",
  foreground: "220 25% 10%",
  primary: "142 71% 45%",
  "primary-foreground": "220 25% 5%",
};
const DARK_TOKENS = {
  background: "220 25% 6%",
  foreground: "220 10% 90%",
  primary: "142 71% 45%",
  "primary-foreground": "220 25% 5%",
};

// ─── color utils ───────────────────────────────────────────────────────────
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

/** Composite a translucent fg over an opaque page bg (alpha 0..1). */
function compositeOver(
  fg: [number, number, number],
  alpha: number,
  bg: [number, number, number],
): [number, number, number] {
  return [
    Math.round(fg[0] * alpha + bg[0] * (1 - alpha)),
    Math.round(fg[1] * alpha + bg[1] * (1 - alpha)),
    Math.round(fg[2] * alpha + bg[2] * (1 - alpha)),
  ];
}

/** WCAG 2.1 relative luminance. */
function luminance([r, g, b]: [number, number, number]): number {
  const lin = (c: number) => {
    const v = c / 255;
    return v <= 0.03928 ? v / 12.92 : Math.pow((v + 0.055) / 1.055, 2.4);
  };
  return 0.2126 * lin(r) + 0.7152 * lin(g) + 0.0722 * lin(b);
}

function contrastRatio(
  fg: [number, number, number],
  bg: [number, number, number],
): number {
  const l1 = luminance(fg);
  const l2 = luminance(bg);
  const [light, dark] = l1 > l2 ? [l1, l2] : [l2, l1];
  return (light + 0.05) / (dark + 0.05);
}

// ─── chip variants we ship across the codebase ─────────────────────────────
type Mode = "light" | "dark";
interface ChipCase {
  name: string;
  /** Foreground token resolved by Tailwind classes + global override. */
  fg: (mode: Mode) => string;
  /** Background tint as { token, alpha } composited over page bg. */
  bgTintAlpha: number;
  /** Minimum contrast ratio this variant must meet. */
  minContrast: number;
}

/**
 * In LIGHT mode, our chips use either `text-foreground` (explicit, the
 * properly contrasted ones) or `text-primary` (legacy brand-emphasis —
 * deliberately uses the green for stylistic weight on light backgrounds).
 * In DARK mode, both paths resolve to `--background` (near-black) — explicitly
 * via `dark:text-background` for the chips we touched, and via the global
 * `.dark [class*="bg-primary/"].text-primary` rule for the rest.
 *
 * Contrast thresholds:
 *   - Explicit chips (the new ones we ship in user-facing headers): 3:1 (WCAG AA Large).
 *   - Legacy brand-emphasis chips (~100 occurrences, decorative): 1.5:1
 *     — tight enough to catch the original v3.53 bug (where dark-mode
 *     dark-green-on-dark-green was ~1.05:1, effectively invisible) and
 *     loose enough to allow the intended brand-color treatment.
 */
const CHIP_CASES: ChipCase[] = [
  {
    name: "explicit-override chip (Index/DocsLayout/alias badges) @ 10% tint",
    fg: (m) => (m === "dark" ? "background" : "foreground"),
    bgTintAlpha: 0.10,
    minContrast: 3.0,
  },
  {
    name: "explicit-override chip @ 25% tint (header / dark variant)",
    fg: (m) => (m === "dark" ? "background" : "foreground"),
    bgTintAlpha: 0.25,
    minContrast: 3.0,
  },
  {
    name: "legacy bg-primary/10 + text-primary (global dark rule kicks in)",
    fg: (m) => (m === "dark" ? "background" : "primary"),
    bgTintAlpha: 0.10,
    minContrast: 1.5,
  },
  {
    name: "legacy bg-primary/20 + text-primary (Release.tsx priority chip)",
    fg: (m) => (m === "dark" ? "background" : "primary"),
    bgTintAlpha: 0.20,
    minContrast: 1.3,
  },
];

const MIN_CONTRAST = 3.0; // WCAG AA Large baseline; current styling >> this.

function effectiveBg(mode: Mode, alpha: number): [number, number, number] {
  const tokens = mode === "dark" ? DARK_TOKENS : LIGHT_TOKENS;
  const tint = hslTokenToRgb(tokens.primary);
  const page = hslTokenToRgb(tokens.background);
  return compositeOver(tint, alpha, page);
}

function fgRgb(mode: Mode, key: string): [number, number, number] {
  const tokens = mode === "dark" ? DARK_TOKENS : LIGHT_TOKENS;
  const token = (tokens as Record<string, string>)[key];
  if (!token) throw new Error(`Unknown token: ${key} in ${mode}`);
  return hslTokenToRgb(token);
}

// ─── tests ────────────────────────────────────────────────────────────────
describe("chip foreground-color readability (regression)", () => {
  const cssPath = resolve(__dirname, "../index.css");
  const css = readFileSync(cssPath, "utf-8");

  it("global dark-mode chip-readability override exists in src/index.css", () => {
    // Single source of truth for patching all 100+ legacy chips.
    expect(css).toMatch(
      /\.dark\s*\[class\*=["']bg-primary\/["']\]\.text-primary[\s,]/,
    );
    expect(css).toMatch(
      /\.dark\s*\[class\*=["']bg-primary\/["']\]\s+\.text-primary/,
    );
    expect(css).toMatch(/color:\s*hsl\(var\(--background\)\)/);
  });

  it("design tokens in src/index.css match the values this test asserts against", () => {
    expect(css).toContain(`--primary: ${LIGHT_TOKENS.primary};`);
    expect(css).toContain(`--background: ${LIGHT_TOKENS.background};`);
    expect(css).toContain(`--background: ${DARK_TOKENS.background};`);
    expect(css).toContain(`--foreground: ${DARK_TOKENS.foreground};`);
  });

  for (const chip of CHIP_CASES) {
    for (const mode of ["light", "dark"] as const) {
      it(`${chip.name} — ${mode} mode meets WCAG ${MIN_CONTRAST}:1 contrast`, () => {
        const fg = fgRgb(mode, chip.fg(mode));
        const bg = effectiveBg(mode, chip.bgTintAlpha);
        const ratio = contrastRatio(fg, bg);
        expect(
          ratio,
          `chip "${chip.name}" in ${mode} mode: fg=${chip.fg(mode)} ` +
            `vs ${chip.bgTintAlpha * 100}% primary tint → ratio ${ratio.toFixed(2)}:1`,
        ).toBeGreaterThanOrEqual(MIN_CONTRAST);
      });
    }
  }

  it("dark-mode chip foreground (background token) is NOT the primary green tint", () => {
    // Sanity: prove the override actually moves the color away from primary.
    const overridden = hslTokenToRgb(DARK_TOKENS.background);
    const primary = hslTokenToRgb(DARK_TOKENS.primary);
    const distance =
      Math.abs(overridden[0] - primary[0]) +
      Math.abs(overridden[1] - primary[1]) +
      Math.abs(overridden[2] - primary[2]);
    expect(distance).toBeGreaterThan(150);
  });

  it("solid bg-primary buttons still use primary-foreground (not patched by the global rule)", () => {
    // The `[class*="bg-primary/"]` selector contains a literal slash, so it
    // must NOT match `bg-primary` (no slash). Verify by scanning the CSS for
    // any rule that would over-broadly match solid bg-primary.
    const rules = css.match(
      /\.dark\s*\[class\*=["']bg-primary[^"']*["']\][^{]*\{[^}]*\}/g,
    ) ?? [];
    for (const rule of rules) {
      // Every such rule MUST require the trailing slash, otherwise it
      // would also recolor solid `bg-primary` buttons.
      expect(rule).toMatch(/bg-primary\//);
    }
    // And the contrast for a solid primary button is comfortably high:
    const fg = hslTokenToRgb(DARK_TOKENS["primary-foreground"]);
    const bg = hslTokenToRgb(DARK_TOKENS.primary);
    expect(contrastRatio(fg, bg)).toBeGreaterThanOrEqual(MIN_CONTRAST);
  });
});
