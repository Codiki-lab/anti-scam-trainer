# Модель данных

## Реализованная схема

Текущий `backend/migrations/init.sql` создаёт десять таблиц.

```mermaid
erDiagram
    USERS ||--o{ CHAT_SESSIONS : проходит
    CHATS ||--o{ CHAT_SESSIONS : используется_в
    CHATS ||--o{ CHAT_STEPS : содержит
    CHATS ||--o{ CHAT_OPTIONS : содержит
    CHAT_SESSIONS ||--o{ MESSAGES : содержит
    USERS ||--|| STATISTICS : имеет
    USERS ||--|| LEADERBOARD : имеет
    USERS ||--o{ USER_ACHIEVEMENTS : получает
    ACHIEVEMENTS ||--o{ USER_ACHIEVEMENTS : выдается

    USERS {
        serial id PK
        varchar user_id UK
        varchar username
        integer completed_chats
    }
    CHATS {
        serial id PK
        varchar title
        text description
        varchar difficulty
        varchar role
        boolean is_active
    }
    CHAT_SESSIONS {
        serial id PK
        integer user_id FK
        integer chat_id FK
        varchar status
        timestamptz started_at
        timestamptz finished_at
        integer score
    }
    CHAT_STEPS {
        serial id PK
        integer chat_id FK
        integer step_number
        varchar role
        text message_text
        varchar response_type
    }
    CHAT_OPTIONS {
        serial id PK
        integer chat_id FK
        integer step_number
        text option_text
        boolean is_correct
        text explanation
        integer points
    }
    MESSAGES {
        serial id PK
        integer session_id FK
        varchar role
        text message
        timestamptz created_at
    }
    STATISTICS {
        integer user_id PK,FK
        integer total_chats
        integer completed_chats
        integer failed_chats
        integer total_messages
        integer time_spent
        numeric success_rate
    }
    LEADERBOARD {
        integer user_id PK,FK
        integer rank
        integer score
        timestamptz updated_at
    }
    ACHIEVEMENTS {
        serial id PK
        varchar title
        text description
        varchar icon
        varchar condition_type
        varchar condition_value
    }
    USER_ACHIEVEMENTS {
        serial id PK
        integer user_id FK
        integer achievement_id FK
        timestamptz received_at
    }
```

Go-модели и CRUD сейчас существуют только для `users`, `chats` и `chat_sessions`.

## Принятая целевая модель

Текущий SQL появился до согласования игрового цикла и будет заменён новой миграцией. Целевая модель включает уровни, отдельный прогресс по ролям, ответы внутри прохождения и корректное владение вариантами ответа.

```mermaid
erDiagram
    USERS ||--o{ USER_LEVEL_PROGRESS : накапливает
    LEVELS ||--o{ USER_LEVEL_PROGRESS : оценивается_в
    LEVELS o|--o{ CHATS : содержит
    USERS ||--o{ CHAT_SESSIONS : проходит
    CHATS ||--o{ CHAT_SESSIONS : запускается_как
    CHATS ||--o{ CHAT_STEPS : содержит
    CHAT_STEPS ||--o{ CHAT_OPTIONS : предлагает
    CHAT_SESSIONS ||--o{ SESSION_ANSWERS : содержит
    CHAT_STEPS ||--o{ SESSION_ANSWERS : получает
    CHAT_OPTIONS o|--o{ SESSION_ANSWERS : выбирается
    CHAT_SESSIONS ||--o{ MESSAGES : содержит
    USERS ||--o{ USER_ACHIEVEMENTS : получает
    ACHIEVEMENTS ||--o{ USER_ACHIEVEMENTS : выдается

    LEVELS {
        integer id PK
        integer level_number UK
        varchar response_type
        varchar ai_mode
        boolean is_active
    }
    USER_LEVEL_PROGRESS {
        integer id PK
        integer user_id FK
        integer level_id FK
        varchar user_role
        integer best_score
        integer stars
        integer attempts
        timestamptz passed_at
    }
    CHATS {
        integer id PK
        integer level_id FK
        varchar title
        text description
        varchar user_role
        varchar mode
        varchar scam_scheme
        jsonb product_context
        boolean is_active
    }
    CHAT_STEPS {
        integer id PK
        integer chat_id FK
        integer step_number
        varchar response_type
        text step_goal
        text ai_instruction
        text fallback_message
        integer max_points
    }
    CHAT_OPTIONS {
        integer id PK
        integer step_id FK
        text option_text
        boolean is_correct
        integer points
        text explanation
        integer sort_order
    }
    CHAT_SESSIONS {
        integer id PK
        integer user_id FK
        integer chat_id FK
        varchar mode
        varchar status
        integer current_step_number
        boolean is_scam
        integer score
        integer max_score
        timestamptz started_at
        timestamptz finished_at
    }
    SESSION_ANSWERS {
        integer id PK
        integer session_id FK
        integer step_id FK
        integer option_id FK
        text free_text
        boolean is_correct
        integer awarded_points
        text ai_evaluation
        text explanation
        timestamptz created_at
    }
```

На схеме опущены уже принятые поля `users`, `messages`, `achievements` и `user_achievements`, которые не меняют механику уровней.

### Режимы уровней

| Уровень | `response_type` | `ai_mode` |
| --- | --- | --- |
| 1 | `multiple_choice` | `disabled` |
| 2 | `similar_choice` | `disabled` |
| 3 | `mixed` | `evaluate_and_reply` |
| 4 | `free_text` | `generate` |

Свободная игра не является записью пятого уровня. Она запускается как отдельный режим сессии. Поле `is_scam` для такой сессии выбирается при старте и после этого не изменяется.

### Правила целевой модели

- `chat_options` принадлежит конкретному `chat_step` через `step_id`.
- `session_answers` хранит результат шага: либо `option_id`, либо `free_text`.
- Пара `(session_id, step_id)` в `session_answers` уникальна.
- Пара `(chat_id, step_number)` в `chat_steps` уникальна.
- Пара `(step_id, sort_order)` в `chat_options` уникальна.
- Прогресс уникален по тройке `(user_id, user_role, level_id)`.
- Лучшие звёзды не уменьшаются после неудачного повторного прохождения.
- `is_locked` не хранится: доступ вычисляется по прогрессу предыдущего уровня в той же ролевой ветке.
- Физические таблицы статистики и рейтинга не нужны для MVP; значения можно вычислять из прохождений и прогресса.

## Инварианты, уже выраженные в SQL

- внешний `users.user_id` уникален;
- номер шага уникален внутри сценария;
- пользователь не может дважды получить одно достижение;
- сессия ссылается на существующих пользователя и сценарий;
- сообщения удаляются вместе с сессией.

Индекс `chat_options(chat_id, step_number)` сейчас также уникален. Для нескольких вариантов на одном шаге это ограничение ошибочно, поэтому на него нельзя опираться как на предметное правило.

## Что ещё требует решения перед миграцией

- точные SQL-типы и ограничения для перечислений;
- может ли `chat_id` отсутствовать у полностью генерируемой свободной игры или для неё всегда создаётся сценарий-конфигурация;
- перечень статусов прохождения;
- пороги звёзд и правила критических ошибок;
- нужна ли история изменения лучшего прогресса помимо сохранённых сессий.
