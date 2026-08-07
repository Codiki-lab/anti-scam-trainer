# Инструкции для агентов

## Документация проекта

Перед началом работы прочитайте документацию, относящуюся к задаче:

- [Глоссарий предметной области](CONTEXT.md) — обязательная терминология продукта.
- [Обзор документации](docs/README.md).
- [Текущее состояние продукта](docs/01-current-state/README.md), включая [описание продукта](docs/01-current-state/product.md) и [дизайн тренажёра](docs/01-current-state/game-design.md).
- [Архитектура](docs/02-architecture/README.md), [модель данных](docs/02-architecture/data-model.md) и применимые [архитектурные решения (ADR)](docs/02-architecture/adr/README.md).
- [Контракты frontend API](docs/08-frontend-api-contracts/current-http-api.md), если задача затрагивает клиент или HTTP API.
- [Локальная разработка](docs/03-operations/local-development.md), если задача затрагивает запуск, окружение или эксплуатацию.

Используйте термины из `CONTEXT.md`. Не нарушайте принятые ADR без явного обоснования в задаче или без обновления соответствующего ADR.

Если изменение затрагивает фактическое поведение продукта, API, модель данных, архитектуру, процессы запуска или терминологию, обновите относящуюся к нему документацию в той же работе. Если принимается новое архитектурное решение или пересматривается существующее — добавьте либо обновите ADR в `docs/02-architecture/adr/`.

## Agent skills

### Issue tracker

Задачи и PRD ведутся в GitHub Issues. См. [конфигурацию трекера](docs/agents/issue-tracker.md).

### Triage labels

Используются стандартные метки triage. См. [словарь меток](docs/agents/triage-labels.md).

### Domain docs

Одноконтекстная структура: [CONTEXT.md](CONTEXT.md) в корне и ADR в [docs/02-architecture/adr/](docs/02-architecture/adr/). См. [правила работы с доменной документацией](docs/agents/domain.md).
