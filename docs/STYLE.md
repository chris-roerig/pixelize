# pixelize — site style guide

The landing page (`docs/index.html`) shares the single-file, monospace,
centered-column layout of the sibling project pages, but uses its own
identity. Keep changes consistent with the tokens below.

## Concept

Retro-arcade / CRT: warm phosphor tones on a near-black background, faint
scanline texture, blocky pixel-art logo. The look should feel like the output
the tool produces.

## Color tokens (CSS variables)

| Token | Value | Use |
|-------|-------|-----|
| `--bg` | `#11131a` | page background (with scanline overlay) |
| `--surface` | `#191c26` | cards / panels |
| `--surface-2` | `#1f2430` | raised panels |
| `--border` | `#2d3342` | hairline borders |
| `--text` | `#f4f1e8` | primary text (warm cream) |
| `--muted` | `#8b93a7` | secondary text |
| `--accent` | `#ffb347` | amber — primary accent, prompts-after, links |
| `--accent-2` | `#ff5d8f` | magenta — secondary accent, section labels |
| `--green` | `#7be081` | phosphor green — shell prompts, success output |
| `--blue` | `#58c7f3` | occasional accent |
| `--purple` | `#b98cff` | code comments in examples |
| `--code-bg` | `#0c0e14` | code blocks |

## Type

Monospace stack (`SF Mono`, `Fira Code`, `Menlo`, …), base 14px, line-height
1.6. Section labels are uppercase with `letter-spacing: 2px`.

## Layout

- Centered column, `max-width: 820px`, `24px` side padding.
- Section order: logo + wordmark → tagline/sub → install (click-to-copy) →
  preset chips → features grid (2-col) → before/after gallery → terminal
  example → footer.
- `image-rendering: pixelated` on all sample images and the logo so scaled
  pixel art stays crisp.

## Logo

Inline SVG wordmark, `shape-rendering="crispEdges"`, built from 20×20 `rect`
tiles in amber → magenta → green forming a stylized pixel "P". Edit the `<g>`
groups in `index.html` to adjust.

## Assets

Gallery images live in `docs/img/` (copied from `examples/`) so they resolve
when the site is served from the `/docs` folder. Regenerate with
`pixelize examples/cassette.png --preset all` and recopy the referenced files.

## Deploy

GitHub Pages → Settings → Pages → Source: `main` branch, `/docs` folder.
`.nojekyll` disables Jekyll processing for this plain static site.
