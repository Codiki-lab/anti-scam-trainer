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

## Training topic selector — 2026-08-13

- Source surface: `/var/folders/r5/ygzzxvl54vgcvmr136bwfzsh0000gn/T/TemporaryItems/NSIRD_screencaptureui_oFIGam/Снимок экрана 2026-08-13 в 19.21.13.png`.
- Implementation: `http://127.0.0.1:5173/preview/chats`.
- Visual verification: the unstyled native select was replaced by an Avito-like trigger and a structured six-item Тема menu with numbering, learning state and selected indicator.
- Interaction verification: selection updates the `topic` search parameter and the four displayed Уровня; the menu closes after selection, with Escape and after an outside pointer press.
- Keyboard verification: Arrow Up/Down, Home and End move focus through options; the trigger opens the list from the keyboard.
- Responsive verification: checked at 390 × 844 CSS px; the selector fills the available width and the menu stays within the viewport with bounded scrolling.
- Runtime verification: no console warnings or errors; the full check pipeline passed with 15 test files and 30 tests.

final result: passed

## Profile menu and quiz copy — 2026-08-13

- Source surface: `/var/folders/r5/ygzzxvl54vgcvmr136bwfzsh0000gn/T/TemporaryItems/NSIRD_screencaptureui_vB5DON/Снимок экрана 2026-08-13 в 19.11.16.png`.
- Implementation: shared `AppHeader` profile menu and the preview Theory/quiz flow.
- Visual verification: the plain account popup was replaced by a structured profile header, navigation items with Phosphor icons and a separated destructive sign-out action.
- Interaction verification: the menu closes with Escape, an outside pointer press and navigation through a menu item; `aria-expanded` follows the visible state.
- Copy verification: visible `Quiz` labels were translated to «квиз» while TypeScript names, cache tags and HTTP routes remain unchanged.
- Runtime verification: no console warnings or errors; the full check pipeline passed with 14 test files and 29 tests.

final result: passed

## Progress redesign — 2026-08-13

- Source surface: `/var/folders/r5/ygzzxvl54vgcvmr136bwfzsh0000gn/T/TemporaryItems/NSIRD_screencaptureui_kCZtvO/Снимок экрана 2026-08-13 в 18.49.11.png`.
- Implementation: `http://127.0.0.1:5173/preview/progress`.
- Data truth: `GET /api/v1/progress`; completion, Level totals, average Балл, Звёзды, per-Тема state and recent Прохождения are rendered from the response model.
- Desktop verification: the flat metric row was replaced with one primary progress surface, three supporting metrics, six linked Тема cards and a compact recent-history list.
- Mobile verification: checked at 390 × 844 CSS px; the summary, metric cards, Тема cards, stars and score badges stack without horizontal overflow.
- Accessibility verification: overall and per-Тема bars expose current and maximum values; visual stars have a text alternative with the earned count.
- Runtime verification: no console warnings or errors; the full `npm run check` pipeline passed with 13 test files and 28 tests.

final result: passed

## Achievements redesign — 2026-08-13

- Source surface: `/var/folders/r5/ygzzxvl54vgcvmr136bwfzsh0000gn/T/TemporaryItems/NSIRD_screencaptureui_FGUJT7/Снимок экрана 2026-08-13 в 18.39.36.png`.
- Implementation: `http://127.0.0.1:5173/preview/achievements`.
- Visual direction: the existing Avito-like product language from the header, Dashboard and shared blue/green/purple/coral palette; no new brand assets were introduced.
- Data truth: `GET /api/v1/achievements` and the OpenAPI `Achievement` schema. The UI renders server `title`, `description`, `earned_at` and `progress`; frontend presentation maps the server `icon` key to an existing Phosphor icon.
- Contract coverage: preview mode now contains all eight achievement codes declared in OpenAPI and preserves the backend split into `earned` and `available`.
- Desktop verification: the oversized gray profile panel was replaced by a compact profile strip and collection summary; earned and available cards remain visually distinct without hiding progress.
- Mobile verification: checked at 390 × 844 CSS px; hero, profile strip, summary and all cards stack without horizontal overflow or clipped text.
- Accessibility verification: incomplete cards expose named progress bars with current and target values; decorative icons are hidden from assistive technology.
- Runtime verification: no console warnings or errors; ESLint, TypeScript, 27 Vitest tests and production build passed.

final result: passed

## Topic catalog refinement — 2026-08-13

- Source surface: `/var/folders/r5/ygzzxvl54vgcvmr136bwfzsh0000gn/T/TemporaryItems/NSIRD_screencaptureui_vINDd8/Снимок экрана 2026-08-13 в 19.29.50.png`.
- Implementation: `http://127.0.0.1:5173/preview/lessons`.
- Visual direction: the existing two-column structure was preserved; the flat cards gained a restrained state accent, a compact numbered marker, one transition affordance and a thin progress line. No new illustrations or decorative surfaces were added.
- Data truth: every state is derived from the current `Topic` response (`is_theory_read`, `is_quiz_passed`, `is_completed` and `levels[].stars`). Progress represents Theory, Quiz and the four Levels without introducing client-only completion data.
- Desktop verification: checked at 1440 × 1000 CSS px; all six cards retain equal rhythm, long titles fit, and only the completed Topic receives the green completion treatment.
- Mobile verification: checked at 390 × 844 CSS px; cards collapse to one column, status and count remain on one row, and no horizontal overflow or clipping is present.
- Accessibility verification: each progress line is exposed as a named progressbar with current and maximum values; decorative icons are hidden from assistive technology.
- Runtime verification: the focused TopicList test passed.

final result: passed

## Training cards redesign — 2026-08-13

- Source surface: `/var/folders/r5/ygzzxvl54vgcvmr136bwfzsh0000gn/T/TemporaryItems/NSIRD_screencaptureui_wfJil8/Снимок экрана 2026-08-13 в 19.40.22.png`.
- Implementation: `http://127.0.0.1:5173/preview/chats`.
- Visual direction: the four long repeated rows were replaced with a restrained two-column card grid. Each card keeps one level marker, one response-mode label, scenario content, result and primary action; no decorative imagery was introduced.
- Backend contract: `LevelState` now consumes the complete OpenAPI response (`scenario_title`, `scenario_description`, `response_type`, `in_progress_attempt_id`) in addition to level availability and scenario id. The card copy and «Продолжить» state are therefore server-driven.
- Progress truth: Stars, best Балл, open state and completed state continue to come from the Topic level progress response. Preview data was corrected so a closed Level cannot simultaneously contain Stars.
- Desktop verification: checked at 1440 × 1100 CSS px; all four cards fit above the fold in a balanced 2 × 2 grid, long copy remains readable, and completed/open/closed states are distinct without strong decoration.
- Mobile verification: checked at 390 × 844 CSS px; cards collapse to one column, the selector stays in the viewport, and the action becomes full-width without text clipping.
- Accessibility verification: Stars retain a text alternative, decorative Phosphor icons are hidden from assistive technology, and unavailable actions remain native disabled buttons.
- Runtime verification: focused mapper and TrainingList tests passed.

final result: passed
