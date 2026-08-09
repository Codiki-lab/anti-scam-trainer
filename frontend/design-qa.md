# Design QA

- Source visual truth: `/var/folders/r5/ygzzxvl54vgcvmr136bwfzsh0000gn/T/codex-clipboard-5c54a4f5-6287-480f-9e9b-f6dfeb9c00cb.png`
- Training source: `/var/folders/r5/ygzzxvl54vgcvmr136bwfzsh0000gn/T/TemporaryItems/NSIRD_screencaptureui_gpdRJe/Снимок экрана 2026-08-09 в 22.30.10.png`
- Implementation: `http://127.0.0.1:5173/preview/dashboard`
- Implementation captures were taken locally during verification and are intentionally not committed to the repository.
- Browser viewport: 1280 × 720 CSS px, device density 1.
- Source pixels: 1280 × 832. Implementation dashboard pixels: 1280 × 1542.
- Normalization: compared at equal 1280 px width; browser chrome in the source and the longer implementation page were excluded from layout judgments below the matching content region.
- State: preview user, buyer role, three-day streak.

## Full-view comparison evidence

The implementation preserves the source hierarchy, navigation, page width, typography scale, topic grid, progress panel and training rows. The dashboard now uses two independent columns: the topic grid follows the role selector without inheriting the height of the daily-task and free-play cards. Training lock states are clearer than in the source while retaining its quiet neutral treatment.

## Focused region comparison evidence

- Header: the former approximate CSS mark was replaced by the official vector wordmark; it remains sharp and aligned at header size.
- Role control: the native select was replaced by a two-option segmented control with visible selected, hover and disabled states.
- Topic access: all six topic cards are links and no topic-level lock is rendered.
- Free play: the primary action navigates to `/preview/sessions/free-play` and renders a free-text adaptive dialogue.
- Topic routing: every topic link resolves its own numeric id or stable slug; selecting «Предоплата» renders the «Предоплата» theory at `/preview/lessons/2`.
- Training availability: levels 3 and 4 are visibly muted, carry lock icons, explain the opening condition and use disabled buttons.
- Chat and result: their main containers use centered max-width layouts.
- Achievements: earned and in-progress items are separated into readable groups; incomplete achievements expose progress bars and counts.

## Required fidelity surfaces

- Fonts and typography: existing system font stack and heading hierarchy preserved; no visible wrapping or truncation regression.
- Spacing and layout rhythm: original grid and section rhythm preserved. The extra free-play card intentionally increases the first section height.
- Colors and visual tokens: existing accent, purple daily-task and pale-blue training palette retained with adequate contrast.
- Image quality and assets: official SVG Avito logo is used; no raster scaling artifacts are visible.
- Copy and content: daily task, role choice, topic states, training count and adaptive free-play behavior are explained in product language.

## Interaction and runtime checks

- Segmented role buttons rendered and remained interactive.
- Free-play button opened the adaptive free-text chat.
- All six lesson cards were independently accessible.
- A non-first lesson opened its own theory instead of falling back to phishing links.
- Two unavailable preview levels rendered as disabled, visibly locked actions.
- No browser console errors were reported.
- Formatting, ESLint, TypeScript, unit tests and production build passed.

## Findings

No actionable P0, P1 or P2 findings remain.

## Follow-up polish

- P3: after real backend data is connected, tune the dashboard's free-play description using the actual difficulty label returned or inferred by the server.

## Comparison history

The dashboard gap, preview topic routing and ambiguous training lock states were corrected. The final browser comparison and interaction checks passed without P0/P1/P2 findings.

final result: passed
