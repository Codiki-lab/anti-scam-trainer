# Anti-scam trainer — целевой продуктовый и frontend-backend контракт

> Статус: целевая спецификация для согласования и реализации
> Дата: 2026-08-09
> Основа: дизайн из 11 экранов, пользовательский путь, актуальный main и OpenAPI v1
> Аудитория: backend, frontend, дизайн, контент и QA
> Важно: этот документ описывает целевое состояние. Фактически реализованные методы всегда сверяются с backend/openapi/v1/openapi.yaml.

## 0. Коротко: что backend должен подготовить

Чтобы frontend можно было собрать без ожидания и переделок, backend должен предоставить:

1. Локальную регистрацию и вход по логину и паролю через HttpOnly JWT-cookie.
2. Выбор и сохранение предпочтительной ролевой ветки: покупатель или продавец.
3. По шесть учебных тем для каждой ролевой ветки.
4. Теорию, quiz из 3–5 вопросов и результат quiz для каждой темы.
5. Четыре последовательных уровня тренировки внутри каждой темы.
6. Продуктовый контекст сделки и историю сообщений для экрана чат-тренировки.
7. Ответы готовым вариантом для уровней 1–2 и свободным текстом для уровней 3–4.
8. Итоговый Балл, 0–3 звезды, подробный разбор каждого Ответа пользователя.
9. Прогресс по темам и уровням, достижения и серию активных дней.
10. Стабильный OpenAPI-контракт, единые ошибки и seed-данные.

Целевой путь:

    регистрация или вход
        → выбор ролевой ветки
        → главная
        → темы теории
        → теория
        → quiz
        → выбор уровня тренировки
        → чат-тренировка
        → результат
        → прогресс
        → достижения и профиль

## 1. Источники истины и приоритет

Документ подготовлен после обновления локальной ветки до origin/main на коммите cc21c8b.

Использованы:

- backend/openapi/v1/openapi.yaml — фактически реализованный HTTP-контракт;
- docs/08-frontend-api-contracts/current-http-api.md — правила API для frontend;
- CONTEXT.md — обязательная терминология;
- docs/01-current-state/product.md — границы продукта;
- docs/01-current-state/game-design.md — четыре уровня, звёзды и Свободная игра;
- docs/02-architecture/data-model.md — текущая и целевая модель данных;
- принятые ADR по auth, атомарному завершению и OpenAPI;
- дизайн и уточнения по 11 экранам;
- docs/scripted-chat-scenarios.md — контент и ограничения модели.

При конфликте используется следующий приоритет:

1. Явно согласованное продуктовое решение.
2. Принятый ADR.
3. Канонический OpenAPI текущей версии.
4. Работающий код и тесты.
5. Этот целевой handoff-документ.
6. Старые схемы, чаты и черновики.

Любое изменение внешнего API одновременно меняет:

- backend/openapi/v1/openapi.yaml;
- HTTP handler и DTO;
- HTTP-contract tests;
- docs/08-frontend-api-contracts/current-http-api.md;
- этот документ, если меняется целевой пользовательский путь.

## 2. Терминология

Используются термины из CONTEXT.md:

- Пользователь — локальная учётная запись с логином и паролем.
- Роль доступа — user или admin.
- Ролевая ветка — отдельный прогресс покупателя или продавца.
- Роль пользователя — buyer или seller внутри сделки.
- Тема — учебный раздел о конкретном типе риска. Это расширение, требуемое дизайном.
- Уровень — одна из четырёх ступеней тренировки внутри темы.
- Сценарий — конкретная тренировочная ситуация.
- Прохождение — одна попытка завершить Сценарий.
- Шаг сценария — точка выбора или свободного ответа.
- Ответ пользователя — выбранный вариант или свободный текст.
- Сообщение — реплика в истории чата.
- Балл — число от 0 до 100 за Прохождение.
- Звезда — оценка завершённого уровня от 0 до 3.
- Открытый уровень — доступный для запуска уровень.
- Серия дней — число последовательных календарных дней с учебной активностью.

Не смешивать:

- access_role user/admin и training_role buyer/seller;
- Тему и Уровень;
- Сценарий и Прохождение;
- Ответ пользователя и Сообщение;
- Балл и Звезду.

## 3. Что уже реализовано в актуальном main

### Реализовано

| Область | Состояние |
| --- | --- |
| API | Версия /api/v1 |
| Документация API | OpenAPI 3.1 и Swagger |
| Регистрация | POST /api/v1/auth/register |
| Вход | POST /api/v1/auth/login |
| Выход | POST /api/v1/auth/logout |
| Текущий пользователь | GET /api/v1/auth/me |
| Auth | Подписанная HttpOnly JWT-cookie access_token на 7 дней |
| Уровни | GET /api/v1/training/levels?role=buyer или seller |
| Старт / продолжение | POST /api/v1/training/levels/{level}/start |
| Готовый ответ | POST /api/v1/attempts/{id}/answers |
| Отказ от Прохождения | POST /api/v1/attempts/{id}/abandon |
| Сценарии | draft, published, archived |
| Контент API | Административные методы Сценариев, Шагов и вариантов |
| Прогресс | Лучший Балл, звёзды, попытки по ролевой ветке и уровню в БД |
| Scoring | 0/25/50/75/100 за вариант; итог 0–100 |
| AI provider | Технический клиент Ollama с контролем контекста |

### Ещё не покрыто для дизайна

| Требование дизайна | Текущее состояние |
| --- | --- |
| 12 Тем | Сущности Темы нет |
| Теория | Нет таблиц и API |
| Quiz | Нет таблиц и API |
| Четыре уровня внутри каждой Темы | Сейчас один опубликованный Сценарий на роль + уровень |
| Выбор роли при регистрации | Регистрация принимает только username и password |
| Сохранение предпочтительной роли | Нет поля |
| Данные товара в игровом DTO | product_context есть в БД, но не возвращается игровым API |
| Реплика виртуального собеседника | Шаг возвращает goal, отдельной реплики нет |
| История чата для восстановления | Возвращаются только step_id и option_id |
| Свободный текст уровней 3–4 | answers принимает только option_id |
| Интеграция игрового сервиса с AI | AI provider создан, но сервис его не вызывает |
| Прогресс-экран | Нет публичного progress endpoint |
| Достижения | Таблицы старой схемы есть, публичной логики и API нет |
| Серия дней | Нет модели и API |
| Dashboard endpoint | Нет |
| Единая JSON-ошибка | Текущие ошибки в основном text/plain |
| Frontend | В frontend находится только .env.example |

## 4. Как объединяются дизайн и существующая игровая прогрессия

### Целевая структура

У каждой ролевой ветки шесть Тем. У каждой Темы четыре Уровня.

    Ролевая ветка
        └── 6 Тем
              ├── Теория
              ├── Quiz
              └── 4 Уровня
                    └── один опубликованный Сценарий на уровень

Итого целевая матрица:

| Сущность | Количество |
| --- | ---: |
| Ролевые ветки | 2 |
| Темы | 12 |
| Уровни на Тему | 4 |
| Сочетания Тема × Уровень | 48 |
| Quiz-вопросы | 36–60 |
| Опубликованные Сценарии полного продукта | 48 |

Для первого функционального релиза допускается:

- все 12 Тем;
- теория и quiz для всех Тем;
- уровни 1–2 для всех Тем: 24 работающих Сценария;
- уровни 3–4 отображаются как недоступные или «скоро», пока не готов AI-сервис;
- frontend получает явный availability_reason и не додумывает состояние.

Нельзя выдавать незавершённый уровень как работающий.

### Четыре звезды на карточке и 0–3 звезды результата

В дизайне карточка Темы содержит четыре маркера уровня в форме звёзд.

Каждый маркер соответствует Уровню 1–4:

- filled — уровень пройден хотя бы на одну звезду;
- current — уровень открыт, но не пройден;
- locked — предыдущий уровень не пройден;
- coming_soon — функция ещё не реализована backend.

Отдельно каждый завершённый уровень имеет earned_stars от 0 до 3. Эти 1–3 звезды показываются на экране результата и в подробном прогрессе.

Так frontend не смешивает:

- четыре позиции уровней;
- качество одного завершённого уровня от нуля до трёх звёзд.

### Открытие уровней

1. Теорию можно открыть сразу.
2. Quiz доступен после получения содержимого Темы.
3. Уровень 1 открывается после quiz_passed=true.
4. Уровень 2 открывается после минимум одной звезды уровня 1.
5. Уровень 3 открывается после минимум одной звезды уровня 2 и только если backend сообщает implemented=true.
6. Уровень 4 открывается после минимум одной звезды уровня 3 и только если backend сообщает implemented=true.
7. Три звезды не обязательны для продолжения.
8. Повторная попытка не уменьшает лучший Балл или звёзды.

## 5. Темы для подготовки контента

### Покупатель

| order | slug | Название | Основной риск |
| ---: | --- | --- | --- |
| 1 | buyer-phishing-links | Фишинговые ссылки | Переход на поддельный сайт |
| 2 | buyer-prepayment | Предоплата | Перевод денег до безопасного оформления |
| 3 | buyer-fake-delivery | Поддельная доставка | Фальшивая страница оформления |
| 4 | buyer-off-platform | Общение вне Avito | Увод в сторонний мессенджер |
| 5 | buyer-sms-codes | Коды из SMS | Передача кода подтверждения |
| 6 | buyer-too-good-offer | Слишком выгодное предложение | Давление низкой ценой и срочностью |

### Продавец

| order | slug | Название | Основной риск |
| ---: | --- | --- | --- |
| 1 | seller-fake-payment | Поддельная оплата | Фальшивое подтверждение платежа |
| 2 | seller-fake-delivery | Фальшивая доставка | Поддельное оформление получения |
| 3 | seller-external-links | Просьба перейти по ссылке | Фишинг |
| 4 | seller-confirmation-codes | Коды подтверждения | Получение доступа через код |
| 5 | seller-off-platform | Общение вне Avito | Увод из защищённого канала |
| 6 | seller-pressure | Давление и спешка | Социальная инженерия |

Названия и slugs становятся стабильными после согласования. Frontend использует id и данные API, а не ветвится по тексту title.

## 6. Общий UI-контракт

### Шапка авторизованных экранов

Показывается на экранах 03–11:

- логотип;
- навигация;
- текущая ролевая ветка;
- огонёк серии дней;
- имя Пользователя или меню профиля.

Не показывается на входе и регистрации, потому что Пользователь ещё не определён.

### Общие состояния

Каждый экран должен иметь:

- loading;
- success;
- empty;
- recoverable error с повтором;
- unauthorized с переходом на вход;
- unavailable / coming_soon;
- forbidden / locked;
- offline или timeout, если запрос не завершён.

Backend обязан различать эти состояния кодом ошибки или полями DTO.

### Общие правила отображения

- Frontend не вычисляет доступность уровня.
- Frontend не вычисляет Балл или звёзды.
- Frontend не хранит правильные ответы.
- Frontend не вызывает Ollama.
- Frontend не подставляет user_id в пользовательские endpoints.
- Учебные ссылки не становятся активными.
- theory_content рендерится только безопасным Markdown или структурированными блоками.

## 7. Спецификация 11 экранов

### Экран 01 — Вход / Login

Цель: войти в существующую локальную учётную запись.

Поля:

- username;
- password.

Действия:

- войти;
- перейти к регистрации.

API:

    POST /api/v1/auth/login

Запрос:

~~~json
{
  "username": "alex",
  "password": "secret"
}
~~~

Успех: 204 No Content, браузер получает HttpOnly cookie access_token.

После успеха frontend обязательно вызывает:

    GET /api/v1/auth/me

Frontend использует fetch с credentials: include.

Ошибки:

- AUTH_INVALID_CREDENTIALS;
- VALIDATION_ERROR;
- RATE_LIMITED;
- SERVICE_UNAVAILABLE.

### Экран 02 — Регистрация / Registration

Цель: создать Пользователя и выбрать начальную ролевую ветку.

Дизайн должен сохранить существующую backend-auth с паролем. Регистрация только по логину конфликтует с принятым ADR-0009 и актуальным main. Для быстрой сборки добавляются:

- username;
- password;
- training_role: buyer или seller.

Целевой запрос:

    POST /api/v1/auth/register

~~~json
{
  "username": "alex",
  "password": "secret",
  "training_role": "seller"
}
~~~

Целевой ответ 201:

~~~json
{
  "id": 1,
  "username": "alex",
  "access_role": "user",
  "training_role": "seller"
}
~~~

Текущая реализация игнорирует training_role, поэтому backend должен:

1. добавить поле предпочтительной ролевой ветки;
2. обновить OpenAPI;
3. обновить DTO, миграцию и тесты;
4. решить, выполняется ли автоматический login после регистрации.

Рекомендация: после регистрации frontend выполняет обычный login. Это сохраняет текущий контракт и не требует установки cookie в register.

Ошибки:

- USERNAME_TAKEN;
- VALIDATION_ERROR;
- INVALID_TRAINING_ROLE.

### Экран 03 — Главная / Dashboard

Цель: показать Пользователю, что делать дальше.

Данные:

- account;
- training_role;
- streak;
- summary;
- шесть Тем выбранной роли;
- четыре level markers каждой Темы;
- continue_action;
- daily_task;
- achievements_preview.

Целевой агрегированный endpoint:

    GET /api/v1/dashboard?role=seller

Ответ:

~~~json
{
  "account": {
    "id": 1,
    "username": "alex",
    "access_role": "user",
    "training_role": "seller"
  },
  "streak": {
    "current": 4,
    "longest": 9,
    "active_today": true
  },
  "summary": {
    "topics_completed": 2,
    "topics_total": 6,
    "levels_completed": 6,
    "levels_total": 24,
    "average_score": 76
  },
  "continue_action": {
    "type": "resume_attempt",
    "topic_id": 7,
    "level": 2,
    "attempt_id": 145
  },
  "topics": [],
  "daily_task": null,
  "achievements_preview": []
}
~~~

Почему нужен агрегированный endpoint:

- экран открывается одним согласованным снимком;
- frontend не склеивает пять разных моментов времени;
- уменьшается число запросов;
- continue_action вычисляет backend;
- проще мокировать и тестировать.

Если endpoint не готов, временно frontend может собрать экран из auth/me, topics, progress и achievements.

### Экран 04 — Темы теории / Lessons

Цель: показать шесть Тем выбранной ролевой ветки.

API:

    GET /api/v1/topics?role=buyer

Ответ:

~~~json
{
  "role": "buyer",
  "topics": [
    {
      "id": 1,
      "slug": "buyer-phishing-links",
      "title": "Фишинговые ссылки",
      "description": "Как распознавать поддельные ссылки",
      "icon": "link",
      "order": 1,
      "theory_read": true,
      "quiz": {
        "passed": true,
        "best_score": 80,
        "passing_score": 80
      },
      "levels": [
        {
          "number": 1,
          "opened": true,
          "implemented": true,
          "best_score": 100,
          "earned_stars": 3,
          "attempts": 1
        },
        {
          "number": 2,
          "opened": true,
          "implemented": true,
          "best_score": 0,
          "earned_stars": 0,
          "attempts": 0
        },
        {
          "number": 3,
          "opened": false,
          "implemented": false,
          "availability_reason": "coming_soon"
        },
        {
          "number": 4,
          "opened": false,
          "implemented": false,
          "availability_reason": "coming_soon"
        }
      ]
    }
  ]
}
~~~

Backend возвращает ровно Темы нужной роли, отсортированные по order.

### Экран 05 — Теория / Theory

Цель: дать материал по одной Теме.

API:

    GET /api/v1/topics/{topic_id}

Ответ:

~~~json
{
  "id": 1,
  "slug": "buyer-phishing-links",
  "role": "buyer",
  "title": "Фишинговые ссылки",
  "description": "Как распознавать поддельные ссылки",
  "theory": {
    "format": "blocks",
    "version": 1,
    "estimated_minutes": 4,
    "blocks": [
      {
        "type": "heading",
        "text": "Как распознать поддельную ссылку"
      },
      {
        "type": "callout",
        "tone": "info",
        "title": "Главное правило",
        "text": "Оформляйте оплату и доставку только внутри сервиса."
      },
      {
        "type": "steps",
        "items": [
          "Проверьте адрес",
          "Не спешите",
          "Не передавайте коды"
        ]
      }
    ]
  },
  "theory_read": false,
  "quiz_available": true
}
~~~

Предпочтителен структурированный blocks-формат. Он точно соответствует дизайну и безопаснее произвольного HTML.

Поддерживаемые block types MVP:

- heading;
- paragraph;
- callout;
- steps;
- checklist;
- warning;
- example.

Отметка прочтения:

    POST /api/v1/topics/{topic_id}/theory/read

Запрос можно сделать пустым. Backend записывает одну активность идемпотентно.

Ответ:

~~~json
{
  "topic_id": 1,
  "theory_read": true,
  "read_at": "2026-08-09T12:00:00Z",
  "streak": {
    "current": 4,
    "active_today": true
  }
}
~~~

### Экран 06 — Проверка знаний / Quiz

Цель: проверить теорию до запуска Уровня 1.

Получение:

    GET /api/v1/topics/{topic_id}/quiz

Frontend не получает is_correct.

~~~json
{
  "topic_id": 1,
  "passing_score": 80,
  "questions": [
    {
      "id": 101,
      "text": "Что безопаснее сделать со ссылкой из чата?",
      "order": 1,
      "options": [
        {
          "id": 1001,
          "text": "Перейти и проверить страницу",
          "order": 1
        },
        {
          "id": 1002,
          "text": "Открыть заказ внутри сервиса",
          "order": 2
        }
      ]
    }
  ]
}
~~~

Отправка:

    POST /api/v1/topics/{topic_id}/quiz/attempts

~~~json
{
  "answers": [
    {
      "question_id": 101,
      "option_id": 1002
    }
  ]
}
~~~

Ответ:

~~~json
{
  "attempt_id": 501,
  "topic_id": 1,
  "score": 80,
  "passed": true,
  "passing_score": 80,
  "correct_answers": 4,
  "total_questions": 5,
  "results": [
    {
      "question_id": 101,
      "option_id": 1002,
      "is_correct": true,
      "explanation": "Проверяйте заказ внутри сервиса."
    }
  ],
  "level_1_opened": true,
  "streak": {
    "current": 4,
    "active_today": true
  }
}
~~~

Правила:

- 3–5 вопросов;
- ровно один вариант на вопрос для первого MVP;
- проходной процент хранится на backend;
- лучший score не уменьшается;
- повторная попытка разрешена;
- правильные ответы показываются только после submit;
- уровень 1 открывается только при passed=true.

### Экран 07 — Выбор тренировки / Training

Цель: показать четыре Уровня выбранной Темы и контекст сделки.

API:

    GET /api/v1/training/levels?role=buyer&topic_id=1

Текущий endpoint уже существует, но должен принять topic_id и расширить DTO.

Ответ:

~~~json
{
  "topic": {
    "id": 1,
    "title": "Фишинговые ссылки",
    "role": "buyer"
  },
  "levels": [
    {
      "number": 1,
      "title": "Распознать очевидный риск",
      "response_type": "multiple_choice",
      "opened": true,
      "implemented": true,
      "scenario_id": 1001,
      "best_score": 75,
      "earned_stars": 2,
      "attempts": 1,
      "product_preview": {
        "title": "iPhone 14, 128 ГБ",
        "price": 42000,
        "currency": "RUB",
        "image_url": null,
        "category": "Электроника"
      }
    }
  ]
}
~~~

Старт или продолжение:

    POST /api/v1/training/levels/{level}/start?role=buyer&topic_id=1

Один незавершённый Attempt на Пользователя и Сценарий. Повторный start возвращает его, а не создаёт второй.

### Экран 08 — Чат-тренировка / Chat Training

Цель: показать историю, реплику виртуального собеседника и доступный способ ответа.

Целевой state:

~~~json
{
  "attempt_id": 145,
  "status": "IN_PROGRESS",
  "scenario": {
    "id": 1001,
    "topic_id": 1,
    "level": 1,
    "role": "buyer",
    "title": "Поддельная доставка",
    "description": "Продавец предлагает оформить доставку по ссылке",
    "product_context": {
      "title": "iPhone 14, 128 ГБ",
      "price": 42000,
      "currency": "RUB",
      "image_url": null,
      "seller_name": "Максим"
    }
  },
  "progress": {
    "current_step": 1,
    "total_steps": 3
  },
  "history": [
    {
      "id": "step-100",
      "speaker": "counterparty",
      "text": "Могу отправить доставкой, оформите по моей ссылке.",
      "created_at": null
    }
  ],
  "step": {
    "id": 100,
    "number": 1,
    "response_type": "multiple_choice",
    "options": [
      {
        "id": 10001,
        "text": "Перейти по ссылке"
      },
      {
        "id": 10002,
        "text": "Оформить всё внутри сервиса"
      }
    ],
    "free_text": {
      "allowed": false,
      "max_length": 0
    }
  }
}
~~~

Текущий backend возвращает goal вместо реплики. Для интерфейса необходимо добавить в Шаг сценария user-facing поле counterparty_message. Goal остаётся внутренней учебной целью и не показывается как сообщение.

Отправка готового варианта:

    POST /api/v1/attempts/{attempt_id}/answers

~~~json
{
  "step_id": 100,
  "option_id": 10002
}
~~~

Отправка свободного текста:

~~~json
{
  "step_id": 300,
  "free_text": "Я оформлю заказ только внутри сервиса."
}
~~~

Правило oneOf:

- option_id;
- free_text;
- никогда оба одновременно.

step_id нужен для защиты от устаревшего экрана после повторной отправки. Backend сверяет его с текущим Шагом.

Промежуточный ответ возвращает полный GameState следующего шага.

Последний ответ атомарно:

1. сохраняет Ответ пользователя;
2. завершает Прохождение;
3. вычисляет Балл и звёзды;
4. обновляет лучший Прогресс;
5. обновляет серию дней;
6. выдаёт достижения;
7. возвращает Result.

Отдельного complete endpoint нет.

Отказ:

    POST /api/v1/attempts/{attempt_id}/abandon

Возвращает 204. ABANDONED не меняет Балл, прогресс и серию дней.

### Экран 09 — Результат / Result

Цель: объяснить все решения, а не только показать число.

Ответ после последнего Шага:

~~~json
{
  "attempt_id": 145,
  "status": "COMPLETED",
  "topic": {
    "id": 1,
    "title": "Фишинговые ссылки"
  },
  "level": 1,
  "score": 75,
  "stars": 2,
  "best_score": 85,
  "best_stars": 3,
  "is_new_best": false,
  "summary": {
    "title": "Хороший результат",
    "safe_decisions": 2,
    "risky_decisions": 1,
    "detected_signals": [
      "внешняя ссылка",
      "давление срочностью"
    ],
    "safe_actions": [
      "Оформлять сделку внутри сервиса",
      "Не передавать коды"
    ]
  },
  "answers": [
    {
      "step_id": 100,
      "user_answer": {
        "type": "option",
        "option_id": 10002,
        "text": "Оформить всё внутри сервиса"
      },
      "points": 100,
      "max_points": 100,
      "explanation": "Это безопасное действие."
    }
  ],
  "next": {
    "next_level": 2,
    "opened": true
  },
  "new_achievements": [],
  "streak": {
    "current": 4,
    "longest": 9,
    "active_today": true,
    "increased_today": false
  }
}
~~~

Backend не возвращает is_correct для уровня с частичными 25/50/75 баллами. Frontend показывает качество решения по points/max_points и explanation.

### Экран 10 — Прогресс / Progress

Цель: показать сводку и подробный прогресс.

API:

    GET /api/v1/progress?role=seller

Ответ:

~~~json
{
  "role": "seller",
  "summary": {
    "topics_completed": 2,
    "topics_total": 6,
    "levels_completed": 6,
    "levels_total": 24,
    "attempts_completed": 12,
    "average_score": 76,
    "best_score": 100,
    "time_spent_seconds": 1860
  },
  "topics": [
    {
      "topic_id": 7,
      "title": "Поддельная оплата",
      "theory_read": true,
      "quiz_passed": true,
      "quiz_best_score": 100,
      "levels": [
        {
          "number": 1,
          "best_score": 100,
          "stars": 3,
          "attempts": 2,
          "passed_at": "2026-08-09T10:00:00Z"
        }
      ]
    }
  ],
  "recent_attempts": [
    {
      "attempt_id": 145,
      "topic_id": 7,
      "topic_title": "Поддельная оплата",
      "level": 1,
      "score": 75,
      "stars": 2,
      "finished_at": "2026-08-09T10:00:00Z"
    }
  ]
}
~~~

Средние значения считаются из COMPLETED Прохождений. ABANDONED не участвует в average_score.

### Экран 11 — Достижения и профиль / Achievements Profile

Цель: показать Пользователя, роли, серию дней и мотивацию.

Профиль:

    GET /api/v1/auth/me

Целевой расширенный ответ:

~~~json
{
  "id": 1,
  "username": "alex",
  "access_role": "user",
  "training_role": "seller",
  "streak": {
    "current": 4,
    "longest": 9,
    "active_today": true,
    "last_activity_date": "2026-08-09"
  }
}
~~~

Смена предпочтительной ветки:

    PATCH /api/v1/profile/preferences

~~~json
{
  "training_role": "buyer"
}
~~~

Смена не удаляет и не объединяет прогресс веток.

Достижения:

    GET /api/v1/achievements

~~~json
{
  "earned": [
    {
      "id": 1,
      "code": "first_training",
      "title": "Первая тренировка",
      "description": "Завершите первое Прохождение",
      "icon": "trophy",
      "earned_at": "2026-08-09T10:00:00Z"
    }
  ],
  "available": [
    {
      "id": 2,
      "code": "five_trainings",
      "title": "Пять тренировок",
      "description": "Завершите пять Прохождений",
      "icon": "medal",
      "progress": 3,
      "target": 5
    }
  ]
}
~~~

## 8. Целевой список пользовательских endpoints

### Уже реализованы

| Метод | Endpoint | Изменение для дизайна |
| --- | --- | --- |
| GET | /api/v1/health | Без изменений |
| POST | /api/v1/auth/register | Добавить training_role |
| POST | /api/v1/auth/login | Без изменений |
| POST | /api/v1/auth/logout | Без изменений |
| GET | /api/v1/auth/me | Добавить training_role и streak |
| GET | /api/v1/training/levels | Добавить topic_id и расширенный DTO |
| POST | /api/v1/training/levels/{level}/start | Добавить topic_id и расширенный GameState |
| POST | /api/v1/attempts/{id}/answers | Добавить step_id и free_text |
| POST | /api/v1/attempts/{id}/abandon | Без изменений |

### Требуется добавить

| Приоритет | Метод | Endpoint | Экран |
| --- | --- | --- | --- |
| P0 | GET | /api/v1/dashboard | 03 |
| P0 | GET | /api/v1/topics | 04 |
| P0 | GET | /api/v1/topics/{id} | 05 |
| P0 | POST | /api/v1/topics/{id}/theory/read | 05 |
| P0 | GET | /api/v1/topics/{id}/quiz | 06 |
| P0 | POST | /api/v1/topics/{id}/quiz/attempts | 06 |
| P0 | GET | /api/v1/progress | 10 |
| P0 | GET | /api/v1/achievements | 11 |
| P1 | PATCH | /api/v1/profile/preferences | 03, 11 |
| P1 | GET | /api/v1/attempts/{id} | Восстановление 08 |
| P1 | GET | /api/v1/attempts/{id}/result | Повторное открытие 09 |

GET attempt и result рекомендуются, даже если start уже возобновляет Прохождение:

- прямой URL после перезагрузки;
- повторное открытие результата;
- восстановление после потери последнего HTTP-ответа;
- E2E-тесты.

## 9. Общие правила HTTP API

### Base URL

Локально:

    http://localhost:8080/api/v1

Frontend:

    VITE_API_URL=http://localhost:8080/api/v1

При работе через разные origin backend должен настроить CORS:

- конкретный frontend origin;
- Access-Control-Allow-Credentials: true;
- разрешённые методы и Content-Type;
- wildcard origin запрещён с credentials.

Предпочтительный production-вариант — один origin через Nginx.

### Auth

- cookie access_token;
- HttpOnly;
- SameSite=Lax;
- Path=/;
- Secure на HTTPS;
- срок 7 дней;
- frontend не читает и не хранит JWT;
- каждый пользовательский запрос отправляется с credentials: include;
- user_id не принимается от клиента.

### Формат данных

- JSON UTF-8;
- snake_case;
- timestamps в ISO 8601 UTC;
- date-only серия дней в YYYY-MM-DD;
- money: целое число минимальных единиц либо целые рубли по единому решению;
- enum в нижнем регистре, кроме сохранённого AttemptStatus;
- списки возвращаются в объекте с именованным полем, если к ним добавляется metadata.

### Успешные коды

| Код | Использование |
| ---: | --- |
| 200 | Чтение и действие с телом |
| 201 | Создание сущности |
| 204 | Успех без тела |

### Целевая ошибка

Текущий API использует text/plain. До активной frontend-интеграции рекомендуется перейти на:

~~~json
{
  "error": {
    "code": "LEVEL_LOCKED",
    "message": "Сначала пройдите предыдущий уровень",
    "details": {
      "required_level": 1
    },
    "request_id": "req-123"
  }
}
~~~

Обязательные codes:

- VALIDATION_ERROR;
- UNAUTHORIZED;
- FORBIDDEN;
- USERNAME_TAKEN;
- INVALID_TRAINING_ROLE;
- TOPIC_NOT_FOUND;
- QUIZ_NOT_PASSED;
- LEVEL_LOCKED;
- LEVEL_NOT_IMPLEMENTED;
- SCENARIO_NOT_FOUND;
- ATTEMPT_NOT_FOUND;
- ATTEMPT_ALREADY_COMPLETED;
- STALE_STEP;
- INVALID_OPTION;
- FREE_TEXT_TOO_LONG;
- AI_UNAVAILABLE;
- RATE_LIMITED;
- INTERNAL_ERROR.

Каждый ответ сохраняет X-Request-ID.

### Идемпотентность и гонки

- theory/read можно повторять.
- Повторный start возвращает IN_PROGRESS Attempt.
- Один session_answer на Attempt + Step.
- Backend сверяет step_id и текущий номер.
- Двойной submit не начисляет очки дважды.
- Последний submit атомарен.
- Повторный GET result возвращает сохранённый результат.
- Achievement выдаётся один раз.
- Серия дней увеличивается максимум один раз в день.

## 10. Целевая модель данных

### 10.1 Users

Существующее:

- id;
- username;
- password_hash;
- access_role.

Добавить:

- preferred_training_role buyer/seller;
- created_at;
- updated_at.

Не хранить текущий user_id из клиента.

### 10.2 Topics

Новая таблица topics:

| Поле | Тип | Правило |
| --- | --- | --- |
| id | BIGSERIAL | PK |
| slug | VARCHAR | UNIQUE |
| user_role | VARCHAR | buyer/seller |
| title | VARCHAR | required |
| description | TEXT | required |
| icon | VARCHAR | nullable |
| sort_order | INT | unique внутри role |
| status | VARCHAR | draft/published/archived |
| created_at | TIMESTAMPTZ | required |
| updated_at | TIMESTAMPTZ | required |

Уникальность:

- slug;
- user_role + sort_order.

### 10.3 Theory

Рекомендуемая таблица topic_theory:

| Поле | Тип |
| --- | --- |
| topic_id | FK UNIQUE |
| version | INT |
| format | VARCHAR |
| content | JSONB |
| estimated_minutes | INT |
| updated_at | TIMESTAMPTZ |

content содержит валидированные blocks.

### 10.4 Quiz

Таблицы:

- quiz_questions;
- quiz_options;
- quiz_attempts;
- quiz_answers.

quiz_questions:

- id;
- topic_id;
- text;
- explanation;
- sort_order.

quiz_options:

- id;
- question_id;
- text;
- is_correct;
- sort_order.

quiz_attempts:

- id;
- user_id;
- topic_id;
- score;
- passed;
- created_at.

quiz_answers:

- attempt_id;
- question_id;
- option_id;
- is_correct.

Один правильный вариант на вопрос для первого MVP.

### 10.5 User topic progress

Новая таблица user_topic_progress:

- user_id;
- topic_id;
- theory_read_at;
- quiz_best_score;
- quiz_passed_at;
- completed_at;
- UNIQUE user_id + topic_id.

Не смешивать quiz score с игровым Баллом.

### 10.6 Levels and Scenarios

Таблица levels остаётся общей конфигурацией 1–4.

В chats добавить обязательный topic_id.

Текущий уникальный индекс один published Scenario на user_role + level_id несовместим с шестью Темами.

Заменить на:

    UNIQUE topic_id + level_id
    WHERE content_status = published AND archived_at IS NULL

user_role у Сценария должен совпадать с topics.user_role.

product_context JSONB должен иметь схему:

~~~json
{
  "title": "iPhone 14, 128 ГБ",
  "price": 42000,
  "currency": "RUB",
  "image_url": null,
  "category": "Электроника",
  "location": "Москва",
  "counterparty_name": "Максим"
}
~~~

### 10.7 Steps

В chat_steps требуются:

- id;
- chat_id;
- step_number;
- response_type;
- counterparty_message;
- step_goal;
- ai_instruction;
- fallback_message;
- max_points.

counterparty_message — интерфейсная реплика.

step_goal — внутренняя учебная цель.

Для уровней 1–2:

- counterparty_message required;
- минимум два варианта;
- ai_instruction не нужен.

Для уровня 3:

- counterparty_message required;
- mixed;
- ai_instruction и fallback_message required.

Для уровня 4:

- free_text;
- ai_instruction и fallback_message required.

### 10.8 Options

Существующая модель подходит:

- step_id;
- option_text;
- points: 0/25/50/75/100;
- explanation;
- sort_order.

Не возвращать points и explanation во время IN_PROGRESS.

### 10.9 Attempts and Answers

Существующие chat_sessions и session_answers сохраняются.

Добавить или обеспечить:

- topic определяется через Scenario;
- awarded_points;
- user-facing answer text доступен для результата;
- AI evaluation хранит структурированный итог или JSONB;
- история доступна после перезагрузки;
- duration можно вычислить из started_at и finished_at.

### 10.10 Streak

Рекомендуется отдельная таблица user_daily_activity:

- user_id;
- activity_date;
- first_activity_at;
- activity_types JSONB или отдельные события;
- UNIQUE user_id + activity_date.

Текущая серия вычисляется по последовательным датам. Для быстрого чтения можно кэшировать в user_streaks:

- user_id;
- current_streak;
- longest_streak;
- last_activity_date.

Источник истины — daily activity, если используются обе таблицы.

События, которые считаются активностью:

- первое завершённое чтение теории;
- отправленная попытка quiz;
- COMPLETED Прохождение.

ABANDONED, login и простое открытие экрана не увеличивают серию.

Часовой пояс MVP: Europe/Moscow через APP_TIMEZONE. В БД хранится activity_date и UTC timestamps.

### 10.11 Achievements

Старые achievements и user_achievements можно развить.

Минимальные codes:

- first_training;
- five_trainings;
- perfect_score;
- first_topic_completed;
- all_buyer_topics;
- all_seller_topics;
- streak_3;
- streak_7;

Условие вычисляет backend после атомарного события.

### 10.12 Статистика

Отдельная таблица общей statistics не обязательна:

- attempts_completed — COUNT COMPLETED;
- average_score — AVG score COMPLETED;
- best_score — MAX score;
- time_spent — SUM finished_at - started_at;
- topic completion — user_topic_progress;
- stars — user_level_progress.

Если производительность потребует денормализацию, она оформляется отдельным решением.

## 11. Scoring

### Варианты

Допустимые points:

- 0 — опасное действие;
- 25 — преимущественно опасное;
- 50 — сомнительное;
- 75 — безопасное, но неполное;
- 100 — безопасное и полное.

### Итог

    score = round_half_up(sum(awarded_points) / sum(step.max_points) × 100)

Диапазон 0–100.

### Звёзды

| Балл | Звёзды |
| ---: | ---: |
| 0–54 | 0 |
| 55–69 | 1 |
| 70–84 | 2 |
| 85–100 | 3 |

### Прогресс

- best_score = max(previous, current);
- stars = max(previous, current);
- attempts увеличивается на каждое COMPLETED;
- passed_at устанавливается при первой попытке минимум на одну звезду;
- худшая повторная попытка не уменьшает результат.

## 12. AI и Qwen

### Фактическое состояние

В main есть технический Ollama provider, но:

- он не подключён к игровому сервису;
- используется llama3.2:3b;
- estimator откалиброван под llama3.2:3b;
- целевой сценарный документ рассчитан на qwen3:8b.

Смена модели требует:

1. изменить OLLAMA_MODEL;
2. переименовать или заменить model-specific estimator;
3. провести preflight-калибровку на русских примерах;
4. обновить тесты и docs/02-architecture/ai-provider.md;
5. согласовать ADR или документ модели.

### Использование по уровням

| Уровень | Модель |
| --- | --- |
| 1 | Не используется |
| 2 | Не используется |
| 3 | Оценивает свободный Ответ пользователя и формирует контролируемую реакцию |
| 4 | Генерирует реплики виртуального собеседника в рамках Сценария |
| Свободная игра | Не входит в обязательные 11 экранов первого релиза |

### Бюджет Qwen3-8B

- num_ctx: 8192;
- целевой input: не более 3300 токенов;
- evaluator output: не более 240 токенов;
- generator output: не более 120 токенов;
- thinking: false;
- temperature evaluator: 0;
- полная документация не попадает в prompt;
- передаётся один role-модуль, одна Тема и один Сценарий;
- rolling history не более трёх пар;
- после двух невалидных ответов используется fallback.

### Structured evaluator result

~~~json
{
  "score": 75,
  "risk_type": "phishing",
  "detected_signals": [
    "внешняя ссылка"
  ],
  "evaluation": "Пользователь отказался от ссылки, но не назвал безопасную альтернативу.",
  "safe_action": "Оформить сделку внутри сервиса.",
  "reply": "Хорошо, тогда оформим иначе."
}
~~~

Backend валидирует:

- JSON schema;
- score;
- risk_type;
- длину полей;
- отсутствие URL, телефонов, карт и новых денежных значений;
- соответствие текущей Теме и фазе;
- контекстный бюджет.

Невалидный ответ модели никогда не передаётся frontend.

## 13. Подготовка контента backend-командой

### Для каждой Темы

Нужно подготовить:

- slug;
- user_role;
- title;
- description;
- icon;
- order;
- 3–7 theory blocks;
- 3–5 quiz questions;
- 3–4 options на вопрос;
- explanation каждого вопроса;
- passing_score.

### Для каждого Сценария

Нужно подготовить:

- topic_id;
- level_id;
- title;
- description;
- scam_scheme;
- product_context;
- content_status;
- 3 Шага для уровней 1–2;
- до двух free-text точек уровня 3;
- ограничения уровня 4;
- fallback messages.

### Для каждого Шага

- counterparty_message;
- step_goal;
- response_type;
- max_points;
- варианты и explanations;
- ai_instruction при необходимости;
- fallback_message при необходимости.

### Контентные объёмы

Полный продукт:

- 12 Тем;
- 36–60 quiz-вопросов;
- 48 Сценариев;
- минимум 72 управляемых Шага только для уровней 1–2;
- при четырёх вариантах — минимум 288 вариантов уровня 1–2.

Первый release slice:

- одна Тема покупателя;
- одна Тема продавца;
- теория;
- quiz;
- уровни 1–2;
- прогресс;
- один end-to-end flow.

После проверки slice контент масштабируется до 12 Тем.

### Валидация публикации

Сценарий нельзя опубликовать, если:

- нет Темы или Уровня;
- роль не совпадает с Темой;
- нет product_context;
- нет Шагов;
- номера Шагов имеют пропуск;
- max_points равен нулю;
- для choice нет минимум двух вариантов;
- points не входят в allowlist;
- нет explanation;
- для AI-уровня нет instruction или fallback;
- уже опубликован другой Сценарий этой Темы и Уровня.

## 14. Безопасность

Backend:

- не доверяет user_id, access_role, training_role, score, stars и points клиента;
- bcrypt для пароля;
- JWT secret обязателен;
- cookie HttpOnly;
- rate limit login/register/AI;
- SQL только параметризованный;
- длина username, password и free_text ограничена;
- логирование не содержит password, JWT и полный свободный ответ;
- admin endpoints требуют access_role=admin;
- Swagger Basic Auth не совпадает с учёткой Пользователя;
- CORS разрешает только известный origin;
- правильные ответы не выдаются до submit;
- published content неизменяем.

Frontend:

- credentials: include;
- JWT не хранится;
- theory blocks рендерятся по allowlist;
- URL из сообщений не кликабельны;
- 401 очищает локальное состояние;
- текст backend не вставляется как HTML;
- submit блокируется до ответа;
- пользовательские ошибки не показывают stack trace.

## 15. Нефункциональные требования

### Производительность

- p95 обычного API до 500 мс локально без AI;
- dashboard не более двух backend-запросов;
- список Тем возвращается одним запросом;
- AI timeout не более 30 секунд;
- все списочные SQL-запросы без N+1;
- индексы на role/order, topic/level, user/progress, session/status.

### Надёжность

- миграции обратимы;
- последняя отправка ответа атомарна;
- AI failure имеет fallback;
- незавершённое Прохождение восстанавливается;
- результат повторно читается;
- пустая БД имеет seed или понятный empty state.

### Наблюдаемость

- X-Request-ID;
- structured logs;
- latency;
- status code;
- attempt_id без текста ответа;
- AI usage и fallback rate;
- ошибки публикации контента;
- health PostgreSQL и опциональный AI status.

### Доступность интерфейса

Backend не возвращает смысл только цветом:

- status enum;
- title;
- availability_reason;
- объяснение;
- alt/label для content images при наличии.

## 16. Порядок backend-работ

### P0 — блокирует frontend

1. Зафиксировать в OpenAPI DTO текущего игрового цикла, а не только descriptions.
2. Добавить Темы и связь Scenario → Topic.
3. Изменить уникальность published Scenario на topic + level.
4. Добавить theory и quiz.
5. Добавить preferred_training_role.
6. Расширить GameState: Scenario, product_context, counterparty_message, response_type, total_steps, history.
7. Реализовать dashboard/topics/progress/achievements.
8. Сделать единый JSON error.
9. Добавить seed-данные первого vertical slice.

### P1 — полный управляемый MVP

1. Заполнить 12 Тем.
2. Подготовить уровни 1–2 всех Тем.
3. Добавить GET Attempt и Result.
4. Добавить streak.
5. Добавить достижения.
6. Завершить content validation.

### P2 — AI-уровни

1. Выбрать и откалибровать модель.
2. Реализовать evaluator уровня 3.
3. Добавить free_text oneOf.
4. Реализовать generator уровня 4.
5. Добавить AI metrics и fallback tests.

Свободная игра не должна задерживать первые 11 экранов.

## 17. Что frontend может начать до полной готовности backend

Backend публикует OpenAPI с target DTO. Затем frontend:

1. генерирует или вручную фиксирует API types;
2. создаёт RTK Query base API;
3. использует MSW с примерами из OpenAPI;
4. собирает app shell и auth;
5. собирает экран Тем;
6. собирает theory и quiz;
7. собирает Training, Chat и Result;
8. меняет MSW на реальный API без изменения UI types.

Backend и frontend не должны независимо придумывать поля.

## 18. Обязательные тестовые сценарии

### Auth

- регистрация;
- занятый username;
- неверный пароль;
- me без cookie;
- logout;
- истёкший JWT.

### Темы и quiz

- шесть Тем каждой роли;
- чужая роль;
- отсутствующая Тема;
- правильный quiz;
- непрошедший quiz;
- повторная попытка;
- попытка открыть Уровень 1 без quiz.

### Игровой цикл

- старт Уровня 1;
- повторный start возвращает тот же Attempt;
- закрытый Уровень;
- неверный option_id;
- устаревший step_id;
- двойной submit;
- последний ответ;
- score rounding;
- stars thresholds 55/70/85;
- худшая повторная попытка не уменьшает прогресс;
- abandon;
- доступ к чужому Attempt.

### Прогресс

- независимость buyer/seller;
- независимость Тем;
- opening следующего Уровня;
- средний Балл только COMPLETED;
- recent attempts order.

### Streak

- первая активность дня;
- две активности одного дня;
- следующий день;
- пропущенный день;
- longest streak;
- граница дня Europe/Moscow.

### AI

- valid structured output;
- malformed JSON;
- timeout;
- context limit;
- retry;
- fallback;
- prompt injection в free_text;
- активная ссылка в output;
- смена Темы или товара моделью.

## 19. Definition of Done backend

Backend готов для frontend-интеграции, если:

- миграции применяются на чистой БД;
- seed создаёт полный vertical slice;
- OpenAPI описывает request и response schemas;
- Swagger открывает актуальную v1;
- auth работает через cookie;
- CORS проверен;
- все P0 endpoints доступны;
- ошибки имеют стабильные codes;
- topic/quiz/level access проверяется на сервере;
- GameState восстанавливается;
- последний submit атомарен;
- progress и streak обновляются один раз;
- правильные answers не утекают до завершения;
- go test ./... проходит;
- HTTP-contract tests покрывают OpenAPI;
- секреты отсутствуют в Git.

## 20. Definition of Done frontend

Frontend готов, если:

- реализованы 11 экранов;
- есть protected routes;
- auth восстанавливается через me;
- роль сохраняется и переключается;
- шесть Тем отображаются из API;
- четыре markers отображают состояние уровней;
- theory blocks безопасны;
- quiz работает без локальных правильных ответов;
- locked и coming_soon различаются;
- чат восстанавливается;
- двойной submit исключён;
- Result использует backend score;
- Progress и Achievements используют API;
- streak виден на экранах 03–11;
- все loading/empty/error states реализованы;
- lint/typecheck/tests/build проходят.

## 21. Решения, которые нужно подтвердить перед реализацией

| Вопрос | Рекомендация |
| --- | --- |
| Оставлять ли пароль в дизайне | Да, соответствует актуальному main и ADR |
| Хранить ли preferred role | Да, но не ограничивать другую ветку |
| Тема содержит четыре уровня | Да |
| Quiz открывает Уровень 1 | Да |
| Один правильный вариант quiz | Да для первого MVP |
| Структура theory | JSON blocks |
| Маркеры на карточке | Четыре звезды по числу уровней |
| Rating уровня | Отдельно 0–3 звезды |
| Уровни 3–4 первого релиза | coming_soon до готовности AI |
| Свободная игра | После основных 11 экранов |
| Модель | Отдельно согласовать; main сейчас llama3.2:3b, целевая спецификация — qwen3:8b |
| Error format | JSON envelope |
| Timezone streak | Europe/Moscow |

## 22. Передача данных между командами

Backend передаёт frontend:

- URL тестового API;
- OpenAPI YAML;
- Swagger credentials;
- тестовый user;
- тестовый admin при необходимости;
- seed-описание;
- список error codes;
- CORS origin;
- перечень implemented/coming_soon;
- дату готовности каждого P0 endpoint.

Frontend передаёт backend:

- origin dev-сервера;
- точные route URLs приложения;
- перечень используемых DTO;
- найденные расхождения OpenAPI и реального ответа;
- X-Request-ID каждого воспроизводимого backend-сбоя.

Контент-команда передаёт backend:

- 12 Тем;
- theory blocks;
- quiz;
- Сценарии;
- product contexts;
- Шаги;
- варианты, points и explanations;
- AI instructions и fallback.

## 23. Самая короткая версия для backend-разработчика

Нужно расширить текущий игровой цикл main, а не заменить его. Между ролевой веткой и четырьмя Уровнями добавляется Тема. У каждой роли шесть Тем; у каждой Темы есть теория, quiz и четыре Уровня. Quiz открывает первый Уровень, следующая ступень открывается после одной звезды. Backend возвращает product_context, реплики, историю, готовые варианты или free_text, сам считает Балл и звёзды, обновляет прогресс, достижения и серию дней. Frontend не знает правильных ответов и не вызывает AI. Фактический контракт фиксируется в OpenAPI до начала реализации экранов.
