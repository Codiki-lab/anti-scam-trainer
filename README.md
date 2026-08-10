<div align="center">

# Anti-scam trainer

### Интерактивный тренажёр безопасного поведения в сделках и чатах классифайда

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)](backend/go.mod)
[![React](https://img.shields.io/badge/React-19-149ECA?logo=react&logoColor=white)](frontend/package.json)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.7-3178C6?logo=typescript&logoColor=white)](frontend/package.json)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-18-4169E1?logo=postgresql&logoColor=white)](deploy/docker-compose.yml)
[![Ollama](https://img.shields.io/badge/Ollama-qwen3%3A8b-111111)](docs/02-architecture/ai-provider.md)

[Быстрый запуск](#быстрый-запуск) · [Возможности](#что-реализовано-в-mvp) · [Архитектура](#архитектура) · [Документация](#документация) · [Команда](#команда)

</div>

## О проекте

Мошенничество в объявлениях часто выглядит как обычная сделка: собеседник просит перейти по ссылке, назвать код из SMS, оплатить товар вне сервиса или продолжить разговор в стороннем мессенджере. Пользователь может помнить правила безопасности, но растеряться в конкретном диалоге.

**Anti-scam trainer** переносит обучение в безопасную симуляцию сделки. Пользователь выбирает роль покупателя или продавца, изучает короткую Теорию, закрепляет знания в Quiz и принимает решения в диалоговых Сценариях. После Прохождения он получает Result с разбором рисков, безопасных действий, Баллами и Звёздами.

> Цель MVP — не просто рассказать правила, а сформировать устойчивую привычку распознавать мошеннический паттерн до реальной сделки.

Подробное одностраничное описание командного решения находится в [submission one-page](docs/submission/one-page.md).

## Что реализовано в MVP

- регистрация и вход по Логину и паролю; JWT хранится в защищённой `HttpOnly` cookie;
- отдельные ролевые ветки покупателя и продавца;
- 6 Тем безопасности для каждой роли;
- пять упорядоченных блоков Теории и Quiz из пяти вопросов для каждой Темы;
- четыре Уровня сложности с последовательным открытием по результату предыдущего Уровня;
- диалоговые тренировки с готовыми вариантами, смешанным и свободным вводом;
- Свободная игра, которая открывается после прохождения четвёртых Уровней всех Тем роли;
- Ежедневное задание «мошенник или нет» и Серия дней;
- Прогресс, Звёзды, Result, рекомендации следующего действия и Достижения;
- администраторский HTTP API для Тем, Теории, Quiz и Сценариев;
- web-админка для управления Темами и метаданными Сценариев через реальный backend;
- versioned OpenAPI, миграции PostgreSQL, Docker Compose и автоматические проверки.

## Пользовательский путь

1. Пользователь регистрируется и выбирает роль покупателя или продавца.
2. На главной странице видит рекомендуемое действие, Темы, Ежедневное задание и Свободную игру.
3. Читает Теорию выбранной Темы и проходит Quiz с порогом `80%`.
4. Проходит четыре Уровня диалоговой тренировки, получая до трёх Звёзд.
5. Изучает Result: решения по Шагам, замеченные сигналы риска и безопасные действия.
6. Следит за Прогрессом, Серией дней и Достижениями, а после завершения ветки запускает Свободную игру.

## Технологии

| Область               | Технологии                                                               |
| --------------------- | ------------------------------------------------------------------------ |
| Frontend              | React 19, TypeScript, Vite, React Router                                 |
| Данные и формы        | Redux Toolkit, RTK Query, React Hook Form, Zod                           |
| UI                    | SCSS Modules, Phosphor Icons                                             |
| Frontend-тестирование | Vitest, Testing Library, MSW, Playwright                                 |
| Backend               | Go, `net/http`, модульный монолит                                        |
| Данные                | PostgreSQL 18, SQL-миграции и seed-контент                               |
| AI                    | локальная Ollama, `qwen3:8b`, ограниченный контекст и строгий JSON-ответ |
| Контракты             | OpenAPI 3.1, HTTP API `/api/v1`                                          |
| Инфраструктура        | Docker Compose, Nginx, Make                                              |
| Качество              | ESLint, Prettier, Husky, lint-staged, Go tests                           |

## Архитектура

```mermaid
flowchart LR
    U[Пользователь] --> N[Nginx gateway :3000]
    N --> F[React frontend]
    F -->|/api/v1, HttpOnly cookie| B[Go backend]
    B --> P[(PostgreSQL)]
    B -->|Уровни 3–4 и Свободная игра| O[Ollama / qwen3:8b]
```

Backend построен как **модульный монолит**: feature-модули разделяют transport, service и repository. Frontend использует облегчённый FSD-подход с однонаправленными зависимостями:

```text
app → pages → widgets → features → entities → shared
```

- **RTK Query** владеет серверным состоянием и кэшем HTTP API.
- DTO проверяются через **Zod** и преобразуются в camelCase-модели до попадания в UI.
- Срезы имеют публичные `index.ts`; глубокие импорты между срезами запрещены.
- Стили компонентов colocated в **SCSS Modules**.
- OpenAPI-файл является источником истины для HTTP-контрактов.
- Модель вызывается только backend: prompt получает ограниченный предметный контекст, а не всю историю продукта.

Подробнее: [обзор архитектуры](docs/02-architecture/README.md) и [ADR](docs/02-architecture/adr/README.md).

## Структура проекта

```text
anti-scam-trainer/
├── backend/                       # Go API, доменная логика, миграции и OpenAPI
│   ├── cmd/api/                   # composition root приложения
│   ├── internal/core/             # общие доменные и инфраструктурные механизмы
│   ├── internal/features/         # auth, learning, attempts, scenarios
│   ├── internal/tests/            # контрактные и acceptance-тесты
│   ├── migrations/                # up/down SQL-миграции и seed-контент
│   └── openapi/v1/openapi.yaml    # источник истины HTTP API
├── frontend/                      # React-приложение и собственный toolchain
│   ├── src/app/                   # роутинг, store, providers и глобальные стили
│   ├── src/entities/              # пользователь, обучение, прогресс, admin-контент
│   ├── src/features/              # законченные пользовательские действия
│   ├── src/widgets/               # крупные составные блоки интерфейса
│   ├── src/pages/                 # пользовательские и административные страницы
│   ├── src/shared/                # независимые переиспользуемые срезы
│   └── e2e/                       # Playwright-сценарии
├── deploy/                        # Dockerfile, Compose и Nginx gateway
├── docs/                          # продуктовая, архитектурная и эксплуатационная база
├── CONTEXT.md                     # обязательный словарь предметной области
└── Makefile                       # единые команды разработки
```

## Быстрый запуск

### Требования

- Docker с поддержкой Compose;
- `make`;
- Go версии из [`backend/go.mod`](backend/go.mod);
- Node.js и npm для разработки frontend;
- свободные порты `3000`, `5432`, `8080`, а для локальной модели — `11434`.

### 1. Подготовка окружения

```bash
git clone git@github.com:Codiki-lab/anti-scam-trainer.git
cd anti-scam-trainer
make setup
```

`make setup` создаёт локальные `.env` из примеров и устанавливает зависимости. Перед запуском заполните в `backend/.env` обязательные секреты:

```dotenv
JWT_SECRET=<случайный длинный секрет>
ADMIN_USERNAME=<логин администратора>
ADMIN_PASSWORD=<пароль администратора>
SWAGGER_USERNAME=<логин Swagger UI>
SWAGGER_PASSWORD=<пароль Swagger UI>
```

Секреты не должны попадать в Git. При старте backend автоматически создаёт единственную учётную запись администратора из `ADMIN_USERNAME` и `ADMIN_PASSWORD`.

### 2. Сборка frontend и контейнеров

```bash
cd frontend
VITE_API_BASE_URL=/api npm run build
cd ..

make build
make up
```

После запуска доступны:

| Сервис               | Адрес                                 |
| -------------------- | ------------------------------------- |
| Приложение и gateway | http://localhost:3000                 |
| Backend API          | http://localhost:8080                 |
| Health check         | http://localhost:8080/api/v1/health   |
| Swagger UI           | http://localhost:8080/swagger/        |
| OpenAPI v1           | http://localhost:8080/openapi/v1.yaml |

Swagger UI и OpenAPI защищены отдельными `SWAGGER_USERNAME` и `SWAGGER_PASSWORD`.

### 3. Запуск с локальной моделью

```bash
make build-ollama
make up-ollama
make ollama-init
```

`make ollama-init` загружает модель из `OLLAMA_MODEL`; значение по умолчанию — `qwen3:8b`. Первый запуск потребует времени и свободного места для модели.

### Frontend отдельно в dev-режиме

```bash
cd frontend
npm install
npm run dev
```

Vite запускается на `http://localhost:5173`, а запросы отправляет на адрес из `VITE_API_BASE_URL`. Backend должен разрешать этот origin через `FRONTEND_ORIGINS`.

### Остановка

```bash
make down

# Если запускали конфигурацию с Ollama
make down-ollama
```

Полная инструкция, переменные окружения и диагностика: [локальная разработка](docs/03-operations/local-development.md).

## Проверки

```bash
# Backend
cd backend
go test ./...

# Frontend: format check, lint, types, unit-тесты и production build
cd frontend
npm run check

# E2E frontend
npm run test:e2e
```

Из корня также доступны:

```bash
make lint
make test
```

## Особенности реализации

- **Безопасная аутентификация.** JWT передаётся через `HttpOnly` cookie; frontend не хранит токен в `localStorage`.
- **Один origin.** Nginx отдаёт приложение и проксирует `/api`, поэтому cookie одинаково работает локально и в публичном демо.
- **Управляемая прогрессия.** Backend является источником истины для открытия Уровней, Баллов, Звёзд и завершения Темы.
- **Контроль AI.** Для модели ограничены контекстное окно и резерв ответа; backend проверяет JSON и применяет fallback без частичной записи состояния.
- **Редактируемый контент.** Темы, Теория, Quiz и Сценарии имеют lifecycle `draft → published → archived`.
- **Честная граница админки.** Web-панель полностью управляет Темами, Теорией и Quiz, а также метаданными и lifecycle Сценариев. Текущий backend не возвращает админскому клиенту сохранённый состав Шагов Сценария, поэтому их визуальное редактирование не имитируется на frontend.

## История разработки

Репозиторий сохраняет историю командной работы в [коммитах](https://github.com/Codiki-lab/anti-scam-trainer/commits/main/) и pull request’ах. Основные этапы:

1. инфраструктура, первичная схема данных и CRUD-основа;
2. доменная модель Тем, Сценариев, Прохождений и Прогресса;
3. аутентификация, контентный API, миграции и seed-данные;
4. AI provider, Свободная игра, Ежедневное задание и rate limiting;
5. frontend-архитектура, пользовательский путь и интеграция с API;
6. админский интерфейс и подготовка командной документации.

## Команда

| Участник                             | Ответственность                                                                                                                                                                                           |
| ------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Антон Загеев** — `w0rn3zz`, Wormix | Backend и интеграция: доменная модель, PostgreSQL и миграции, HTTP API, аутентификация, AI provider, Docker-инфраструктура, техническая документация, review и merge. `w0rn3zz` и Wormix — один участник. |
| **Илья Логинов**                     | Frontend: архитектура React-приложения, пользовательские и административные экраны, RTK Query, API-контракты, SCSS Modules, тесты и дизайн-QA.                                                            |
| **lopohlop**                         | Начальная подготовка backend-части и CRUD-основы проекта.                                                                                                                                                 |

## Документация

- [Submission one-page](docs/submission/one-page.md)
- [Глоссарий предметной области](CONTEXT.md)
- [Карта документации](docs/README.md)
- [Текущее состояние продукта](docs/01-current-state/README.md)
- [Описание продукта](docs/01-current-state/product.md)
- [Игровой дизайн](docs/01-current-state/game-design.md)
- [Архитектура](docs/02-architecture/README.md)
- [Архитектурные решения](docs/02-architecture/adr/README.md)
- [Frontend API-контракты](docs/08-frontend-api-contracts/current-http-api.md)
- [Локальная разработка](docs/03-operations/local-development.md)

## Ссылки для отправки

- Командный репозиторий: https://github.com/Codiki-lab/anti-scam-trainer
- One-page: [docs/submission/one-page.md](docs/submission/one-page.md)
