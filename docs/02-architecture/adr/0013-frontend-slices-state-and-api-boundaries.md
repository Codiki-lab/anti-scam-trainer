# ADR-0013: frontend slices, state ownership и API boundary

- Статус: Принято
- Дата: 2026-08-09

## Контекст

Frontend должен одновременно поддерживать 11 продуктовых экранов, работу с
реальным backend и автономный preview дизайна. Без явных границ роль
пользователя дублировалась бы между Redux, Account и localStorage, DTO backend
попадали бы в UI, а `shared` превращался бы в набор технических корзин.

## Решение

1. Использовать light FSD с направлением зависимостей
   `app → pages → widgets → features → entities → shared`.
2. На каждом слое группировать код в связные slices. Внутри slice создавать
   только нужные сегменты `api`, `model`, `lib`, `ui` и локальные тесты.
3. Экспортировать внешний контракт slice через его `index.ts`; deep imports
   между slices запрещаются ESLint.
4. В `shared` хранить независимые инфраструктурные или визуальные slices, а не
   глобальные корзины `shared/ui`, `shared/lib` и `shared/model`.
5. Server state принадлежит RTK Query, формы — React Hook Form, route state —
   React Router, локальные визуальные состояния — компонентам.
6. Account является единственным runtime-источником роли. Нижние слои получают
   его через `CurrentAccountProvider` и узкий контракт смены роли.
7. OpenAPI DTO проверяются Zod и преобразуются mapper-функциями в camelCase
   модели до передачи в UI.
8. Preview fixtures принадлежат `app/preview` и внедряются сверху; `shared` и
   features не импортируют preview-данные.
9. Стили UI реализуются через colocated SCSS Modules. Глобальный SCSS содержит
   только design tokens, reset и базовые правила HTML, а повторяемые визуальные
   примитивы экспортируются публичным API `shared/ui-kit`.

## Последствия

- Нарушения направления зависимостей и Public API останавливают lint.
- Изменение snake_case DTO локализуется в `entities/*/api` и mapper-функциях.
- Переключение роли не создаёт рассинхронизацию нескольких client stores.
- Не каждый slice имеет все возможные сегменты: пустые папки запрещены, потому
  что не выражают ответственность.
- Preview остаётся отдельным data source и не попадает в production API layer.
- Стили страницы или виджета не могут неявно переопределить другой slice через
  глобальный селектор класса.
