# Design QA — Progress page from Figma

- Figma file: `Avito Anti-Scam Trainer`
- Figma page: `Screens`
- Target frame: `10:217` — `10 Прогресс / Progress`
- Source reference: `/var/folders/r5/ygzzxvl54vgcvmr136bwfzsh0000gn/T/TemporaryItems/NSIRD_screencaptureui_l7aIgJ/Снимок экрана 2026-08-12 в 23.50.00.png`
- Implementation screenshot: `frontend/design-qa-progress-implementation-1440.png`
- Target and implementation viewport: `1440 × 1024` CSS px
- State: preview Progress for the buyer role with one completed Прохождение and one earned Достижение.

## Full-view comparison evidence

The implementation follows the Figma composition: four equal summary cards, a wide content area, recent training history on the left and earned Achievements on the right. The existing completed-Topics panel is retained between the summary and the lower columns because it visualizes a real `ProgressSummary` field and was present in the product's current Progress design.

## Backend contract mapping

The Figma sample values `тренировок`, `лучший результат` and `достижения` are not all available from `GET /api/v1/progress`. They were not inferred from the ten-item recent history.

- `completed_levels` → `уровней пройдено`
- `average_score` → `средний результат`
- `stars` → `звёзд получено`
- `completed_topics` / `total_topics` → `тем завершено` and the Topics progress bar
- `recent_attempts` → linked recent Прохождения
- `GET /api/v1/achievements` → earned Achievements column

The frontend DTO, Zod schema and mapper already match these OpenAPI fields; no contract changes were required.

## Required fidelity surfaces

- Typography: existing Inter stack with the Figma hierarchy of 32px summary values and compact muted labels.
- Layout: four equal cards, 16px gaps, rounded neutral surfaces and a two-column content area matching the desktop frame.
- History: bordered 78px rows with topic, level, Stars and Score.
- Achievements: compact neutral cards with the existing Phosphor icon set; no handcrafted or substitute image assets.
- Responsive behavior: the lower columns stack below 840px and all summary cards stack on mobile.

## Interaction and technical QA

- Result link resolves to `/preview/sessions/9000/result`.
- `Все →` successfully navigates to `/preview/achievements` and browser back restores Progress.
- At `1440 × 1024`, `scrollWidth` equals `clientWidth` and the page has no horizontal overflow.
- Browser console contains no application errors.
- `npm run check` passed: formatting, ESLint, TypeScript, 38 Vitest tests and production build.

## Follow-up polish

- P3: the Figma sample uses three earned Achievement cards, while preview data contains one. The implementation correctly renders the real response length rather than inventing rewards.

final result: passed
