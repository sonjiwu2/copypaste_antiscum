# Anti-Scam Trainer (`copypaste_antiscum`)

Интерактивный антискам-тренажёр для сделок на классифайдах. Пользователь
выбирает роль покупателя или продавца, проходит учебный диалог, принимает
решения и получает объяснения рисков. Backend сохраняет анонимный профиль,
попытки и прогресс в PostgreSQL.

## Текущее состояние

| Слой | Реализация |
|---|---|
| Backend | Go 1.25, `net/http`, JSON API |
| Сценарии | JSON + `go:embed`, строгая проверка графа при старте |
| Данные | PostgreSQL 16: профили, попытки, решения, прогресс |
| Версии сценариев | Неизменяемый архив в PostgreSQL; попытка закреплена за точной версией |
| Docker | PostgreSQL → миграции → backend с readiness healthcheck |
| CI/CD | unit/race/lint/integration/Compose smoke; публикация образа в GHCR |
| Frontend | Пока отсутствует в репозитории |

Документация backend: [backend/README.md](backend/README.md)  
OpenAPI: [docs/openapi.yaml](docs/openapi.yaml)  
Формат сценариев: [docs/scenario-format.md](docs/scenario-format.md)

## Быстрый запуск

Нужны Docker и Docker Compose v2.

```bash
cp .env.example .env  # необязательно: локальные значения уже имеют defaults
docker compose up --build
```

Порядок запуска контролируется Compose:

1. `postgres` становится healthy;
2. одноразовый сервис `migrate` применяет встроенные миграции;
3. `backend` синхронизирует архив версий сценариев и начинает отвечать на
   `/readyz` только после готовности базы.

Проверка:

```bash
curl --fail http://localhost:8080/healthz
curl --fail http://localhost:8080/readyz
curl --fail -c cookies.txt -b cookies.txt \
  http://localhost:8080/api/v1/scenarios
```

| URL | Назначение |
|---|---|
| `http://localhost:8080/healthz` | Liveness процесса |
| `http://localhost:8080/readyz` | Готовность API и PostgreSQL |
| `http://localhost:8080/api/v1/scenarios` | Каталог сценариев |

### Остановка и сохранность данных

```bash
docker compose down      # контейнеры удаляются, PostgreSQL volume сохраняется
docker compose down -v   # полный и необратимый сброс локальной базы
```

Обычный restart или `docker compose down` не удаляет прогресс. Удаление
именованного volume через `down -v` предназначено только для явного сброса.

## Сервисы Compose

```text
postgres (healthy)
       ↓
migrate (успешно завершён)
       ↓
backend :8080 (/readyz healthy)
```

PostgreSQL не публикует порт на host и доступен только внутри Compose-сети.
Backend и мигратор работают non-root, с `no-new-privileges` и read-only
filesystem. Миграции встроены в тот же образ, что и API.

## Локальная разработка backend

По умолчанию backend требует PostgreSQL и не переключается на память при
ошибке подключения.

```bash
cd backend
export DATABASE_URL='postgres://antiscam:change-me-locally@127.0.0.1:5432/antiscam?sslmode=disable'
go run ./cmd/migrate up
go run ./cmd/api
```

Для изолированных тестов или временного демо без базы режим памяти включается
только явно:

```bash
cd backend
STORAGE_DRIVER=memory go run ./cmd/api
```

В memory-режиме данные теряются при завершении процесса. Этот режим не
используется production Compose.

## Конфигурация

Полный пример находится в [.env.example](.env.example).

| Переменная | По умолчанию | Назначение |
|---|---|---|
| `BACKEND_PORT` | `8080` | Host-порт backend в Compose |
| `HTTP_ADDR` | `:8080` | Адрес HTTP listener при прямом запуске |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `STORAGE_DRIVER` | `postgres` | `postgres` или явный demo/test `memory` |
| `DATABASE_URL` | нет | Обязателен при `STORAGE_DRIVER=postgres` |
| `POSTGRES_DB` | `antiscam` | Локальная база Compose |
| `POSTGRES_USER` | `antiscam` | Локальный пользователь Compose |
| `POSTGRES_PASSWORD` | `change-me-locally` | Только локальный default, не production secret |
| `COOKIE_NAME` | `ast_profile` | Cookie анонимного профиля |
| `COOKIE_MAX_AGE` | `4320h` | Срок жизни cookie |
| `COOKIE_SECURE` | `false` | Для публичного HTTPS должен быть `true` |

Параметры пула, database timeouts и HTTP timeouts перечислены в
[backend/README.md](backend/README.md). Реальные секреты должны поступать из
секрет-хранилища платформы; `.env` исключён из Git.

## Миграции

```bash
cd backend
go run ./cmd/migrate status
go run ./cmd/migrate up
go run ./cmd/migrate version
go run ./cmd/migrate down  # откатывает одну миграцию; применять осознанно
```

`DATABASE_URL` должен быть задан. В Compose миграции выполняются автоматически
до запуска API; ошибка миграции блокирует backend.

## API и анонимный профиль

| Метод | Путь | Назначение |
|---|---|---|
| `GET` | `/healthz` | Liveness |
| `GET` | `/readyz` | Readiness PostgreSQL-backed приложения |
| `GET` | `/api/v1/scenarios` | Каталог, фильтр `?role=buyer\|seller` |
| `GET` | `/api/v1/scenarios/{scenarioId}` | Метаданные сценария |
| `POST` | `/api/v1/attempts` | Начать попытку |
| `GET` | `/api/v1/attempts/{attemptId}` | Восстановить состояние попытки |
| `POST` | `/api/v1/attempts/{attemptId}/choices` | Применить решение идемпотентно |
| `GET` | `/api/v1/progress` | Сводка и история текущего профиля |

При первом запросе к `/api/v1/**` сервер выдаёт cookie `ast_profile` с
`HttpOnly`, `SameSite=Lax` и `Path=/`. Все последующие запросы пользователя
должны возвращать эту cookie. Чужая попытка отвечает `403
ATTEMPT_FORBIDDEN`; потерянную анонимную cookie восстановить нельзя.
Валидный идентификатор профиля является bearer-секретом доступа к этой истории.

Текущая конфигурация рассчитана на общий origin через reverse proxy. CORS для
отдельного frontend origin не настроен. Публичный deployment обязан
использовать HTTPS, `COOKIE_SECURE=true`, production credentials и внешние
ограничения частоты запросов.

Полный контракт, схемы ответов и коды ошибок: [docs/openapi.yaml](docs/openapi.yaml).

## Сценарии и их версии

- Авторский источник: `backend/scenarios/*.json`.
- Файлы встроены в бинарник и валидируются до старта API.
- PostgreSQL хранит неизменяемую копию каждой пары `(scenario_id, version)`.
- Повтор той же версии с другим SHA-256 останавливает запуск: содержимое можно
  менять только с увеличением `version`.
- Старые попытки продолжают читать точную архивную версию после нового deploy.

Подробности формата, канонического hash и versioning policy описаны в
[docs/scenario-format.md](docs/scenario-format.md).

## Проверки

```bash
cd backend
gofmt -l .                 # пустой вывод = OK
go vet ./...
go test -count=1 ./...
go test -count=1 -race ./...
golangci-lint run --config=../.golangci.yaml ./...
```

PostgreSQL integration tests требуют отдельную тестовую базу: они пересоздают
в ней схему `public`.

```bash
cd backend
TEST_DATABASE_URL='postgres://antiscam:integration-only@127.0.0.1:5432/antiscam_test?sslmode=disable' \
  go test -count=1 -p 1 -tags=integration ./...
```

Для полного локального smoke:

```bash
docker compose config --quiet
docker compose up --build -d
bash .github/scripts/smoke.sh
```

Smoke проходит безопасный сценарий, проверяет идемпотентный replay, изоляцию
профилей, прогресс, повторный запуск миграций и сохранность после restart.

## CI/CD и deployment

| Workflow | Когда | Проверки |
|---|---|---|
| `.github/workflows/ci.yml` | PR/push в `dev`/`main`, manual | format, vet, unit, race, coverage, lint, PostgreSQL integration, Compose business smoke |
| `.github/workflows/release.yml` | tag `v*`, manual | build и публикация backend image в GHCR |

Release workflow только публикует образ. Production hosting, frontend,
TLS/reverse proxy, database backups и secret management остаются задачами
целевой платформы; фиктивного deploy job в репозитории нет.

## Использование ИИ

Часть кода, сценариев и инфраструктурных файлов создавалась с помощью
AI-ассистентов. Результат проверяется автоматическими тестами, PostgreSQL
integration suite и Docker smoke; ответственность за review и merge остаётся
у команды.

## Лицензия

MIT — см. [LICENSE](LICENSE).
