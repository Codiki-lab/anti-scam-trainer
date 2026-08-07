# Локальная разработка

## Что потребуется

- Go версии, совместимой с `go 1.26.5` из `backend/go.mod`;
- Docker с поддержкой Compose;
- `make`;
- свободные локальные порты `5432`, `8080` и, при работе с моделью, `11434`.

Node.js пока не требуется: фронтенд-приложения в репозитории нет.

## Подготовка

Из корня репозитория:

```bash
make setup
```

Команда создаёт `backend/.env` и `frontend/.env` из примеров и загружает Go-модули. Если `frontend/package.json` отсутствует, установка фронтенда пропускается.

Перед запуском дополните `backend/.env` значениями, которые читает текущий Go-код:

```dotenv
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_NAME=antiscam
PORT=8080
```

`POSTGRES_NAME` должен совпадать с `POSTGRES_DB`. Это временное дублирование связано с расхождением имён переменных в коде и Compose.

## PostgreSQL

Запуск контейнера:

```bash
make up
```

PostgreSQL доступен приложению на `localhost:5432` через контейнер `port-forwarder`. Данные сохраняются в `out/pgdata`.

Текущий `init.sql` не оформлен как миграция `golang-migrate`. До исправления формата схему можно применить вручную, подставив значения из `backend/.env`:

```bash
docker compose -f deploy/docker-compose.yml \
  exec -T anti-scam-trainer-postgres \
  psql -U antiscam -d antiscam < backend/migrations/init.sql
```

Повторный запуск команды завершится ошибкой, потому что SQL использует обычный `CREATE TABLE` без `IF NOT EXISTS`.

## Бэкенд

Бэкенд запускается на машине разработчика, а не в Docker Compose:

```bash
cd backend
go run .
```

Проверка:

```bash
curl http://localhost:8080/health
```

Ожидаемый ответ:

```json
{"status":"ok"}
```

## Ollama

PostgreSQL и Ollama можно поднять вместе:

```bash
make up-ollama
make ollama-init
```

`ollama-init` загружает модель из `OLLAMA_MODEL`. По умолчанию в примере окружения указана `llama3.2:3b`.

```bash
make logs-ollama
```

Пока бэкенд не обращается к Ollama, поэтому успешный запуск контейнера проверяет только инфраструктуру.

## Остановка

```bash
make down
```

Для конфигурации с Ollama:

```bash
make down-ollama
```

`make clean` удаляет контейнеры и именованные Docker volumes, но каталог `out/pgdata` подключён как bind mount и остаётся на диске.

## Проверки

Рабочая на текущем состоянии команда:

```bash
cd backend
go test ./...
```

Общий `make test` пока завершится на шаге фронтенда из-за отсутствия `frontend/package.json`.
