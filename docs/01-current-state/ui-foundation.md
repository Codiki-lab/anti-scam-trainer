# UI-основа тренажёра

Общие production-компоненты используют токены из `global.scss` и адаптивную сетку для 390, 768 и 1440 px. Интерактивные элементы имеют видимый `focus-visible` и минимальную область 44×44 px; `prefers-reduced-motion` отключает декоративную анимацию.

Визуальная документация доступна локально по `/preview/ui-states`. Она показывает кнопки `default`, `disabled` и `loading`, статусы `default`, `progress`, `success` и `locked`, skeleton, error, empty и confirm-dialog. Страница собирается из тех же компонентов `shared/ui-kit` и `shared/error-state`, которые используются продуктовыми экранами.

| Состояние | Общий компонент | Основное применение |
|---|---|---|
| loading | `LoadingState` | Dashboard, Темы, Теория, Quiz, каталог Уровней, Прохождение, Result, Прогресс и Достижения |
| empty | `EmptyState` | отсутствие завершённых Прохождений и другие пустые коллекции |
| error | `ErrorState` | ошибка HTTP с безопасным retry |
| status | `StatusBadge` | доступность и состояние Уровня |
| confirmation | `ConfirmDialog` | досрочное завершение и отказ от Прохождения |
