# Трекер задач: GitHub

Задачи и PRD этого репозитория ведутся в GitHub Issues. Для операций используйте CLI `gh`.

## Правила работы

- Создание: `gh issue create --title "..." --body "..."`.
- Чтение: `gh issue view <number> --comments`.
- Список: `gh issue list --state open` с подходящими фильтрами.
- Комментарий: `gh issue comment <number> --body "..."`.
- Метки: `gh issue edit <number> --add-label "..."` или `--remove-label "..."`.
- Закрытие: `gh issue close <number> --comment "..."`.

Репозиторий определяется по `origin` автоматически при запуске команд из этой рабочей копии.

## Pull request как источник triage

**PR как источник triage: нет.** Внешние pull request не попадают в очередь triage, пока это явно не будет изменено в этом документе.

## Соответствие инструкциям навыков

- «Опубликовать в трекере задач» означает создать GitHub Issue.
- «Получить релевантный тикет» означает выполнить `gh issue view <number> --comments`.
