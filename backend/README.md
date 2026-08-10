# Backend антискам-тренажёра

Go 1.25 API интерактивного тренажёра. Сервис публикует каталог сценариев,
управляет прохождением, сохраняет анонимные профили и попытки в PostgreSQL,
считает прогресс и хранит неизменяемые версии сценариев.

## Запуск

Рекомендуемый путь из корня репозитория:

```bash
docker compose up --build
```

Compose поднимает PostgreSQL, применяет миграции отдельным одноразовым
сервисом и только после этого запускает API.

Для прямого запуска нужен доступный PostgreSQL:

```bash
cd backend
export DATABASE_URL='postgres://antiscam:change-me-locally@127.0.0.1:5432/antiscam?sslmode=disable'
go run ./cmd/migrate up
go run ./cmd/api
```

Проверка:

```bash
curl --fail http://localhost:8080/healthz
curl --fail http://localhost:8080/readyz
```

База обязательна в режиме по умолчанию. Автоматического fallback на память
нет: недоступный PostgreSQL завершает старт ошибкой. Для теста или временного
демо без базы memory-адаптер выбирается явно:

```bash
STORAGE_DRIVER=memory go run ./cmd/api
```

Данные этого режима пропадают при завершении процесса; production Compose его
не использует.

## Конфигурация

| Переменная | По умолчанию | Назначение |
|---|---|---|
| `HTTP_ADDR` | `:8080` | Адрес listener |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `HTTP_READ_HEADER_TIMEOUT` | `5s` | Таймаут чтения заголовков |
| `HTTP_READ_TIMEOUT` | `15s` | Таймаут чтения запроса |
| `HTTP_WRITE_TIMEOUT` | `15s` | Таймаут записи ответа |
| `HTTP_IDLE_TIMEOUT` | `60s` | Таймаут простоя соединения |
| `HTTP_SHUTDOWN_TIMEOUT` | `10s` | Graceful shutdown timeout |
| `HTTP_MAX_REQUEST_BYTES` | `65536` | Максимальный размер тела запроса |
| `STORAGE_DRIVER` | `postgres` | `postgres` или явный test/demo `memory` |
| `DATABASE_URL` | нет | Обязателен для PostgreSQL |
| `DATABASE_MAX_CONNS` | `10` | Максимум соединений пула |
| `DATABASE_MIN_CONNS` | `0` | Минимум соединений пула |
| `DATABASE_CONNECT_TIMEOUT` | `5s` | Подключение и стартовый ping |
| `DATABASE_QUERY_TIMEOUT` | `5s` | Таймаут запроса к базе |
| `DATABASE_HEALTH_TIMEOUT` | `2s` | Таймаут readiness ping |
| `COOKIE_NAME` | `ast_profile` | Имя cookie профиля |
| `COOKIE_MAX_AGE` | `4320h` | Срок жизни cookie (180 суток) |
| `COOKIE_SECURE` | `false` | Для публичного HTTPS должен быть `true` |
| `COOKIE_SAME_SITE` | `lax` | `lax`, `strict` или `none` (требует `COOKIE_SECURE=true`) |
| `CORS_ALLOWED_ORIGINS` | пусто | Список origin браузерного клиента через запятую |

Пример полного окружения: [../.env.example](../.env.example).

## Анонимный профиль и граница доступа

Учётных записей нет. Первый запрос к `/api/v1/**` создаёт серверный профиль и
выдаёт cookie `ast_profile` (`HttpOnly`, `SameSite=Lax`, `Path=/`). Отсутствующая
или синтаксически невалидная cookie заменяется новым случайным ID. Валидный ID
является bearer-секретом доступа к истории; его нельзя раскрывать или логировать.

- попытка доступна только профилю-владельцу;
- чужой профиль получает `403 ATTEMPT_FORBIDDEN`;
- `/healthz` и `/readyz` не создают профили;
- потеря cookie означает потерю доступа к анонимной истории.

## Доступ из браузера

Рекомендуемая схема — общий origin: frontend отдаётся тем же адресом, а его
nginx проксирует `/api` на backend. Cookie остаётся same-site, и открывать
межсайтовый доступ не требуется.

Для фронтенда на другом сайте доступ включается явно:

```bash
CORS_ALLOWED_ORIGINS=https://trainer.example,https://www.trainer.example
COOKIE_SAME_SITE=none
COOKIE_SECURE=true
```

- пустой список означает, что кросс-доменный доступ выключен;
- маска `*` отвергается: ответы привязаны к cookie профиля, и credentialed
  запрос с произвольного сайта означал бы чтение чужой истории;
- ответ всегда называет конкретный origin и сопровождается `Vary: Origin`;
- предварительный `OPTIONS` не доходит до профиля и не создаёт его;
- запрос без браузера (curl, мониторинг) работает независимо от этой настройки;
- `SameSite=None` без `COOKIE_SECURE=true` отвергается конфигурацией: браузер
  молча отбросил бы такую cookie, и каждый запрос заводил бы новый профиль.

Публичный deployment должен использовать HTTPS, `COOKIE_SECURE=true`,
production database credentials и rate limiting на внешнем proxy.

## Проверки состояния

| Путь | Назначение |
|---|---|
| `GET /healthz` | Процесс жив; внешние зависимости не проверяются |
| `GET /readyz` | PostgreSQL доступен и приложение готово обслуживать API |

Compose ориентируется на `/readyz`, поэтому недоступная база снимает backend с
готовности, но не подменяется memory-хранилищем.

## Миграции

SQL-файлы в `migrations/` версионируются goose и встроены в бинарник `migrate`
через `go:embed`.

```bash
cd backend
export DATABASE_URL='postgres://antiscam:change-me-locally@127.0.0.1:5432/antiscam?sslmode=disable'

go run ./cmd/migrate status
go run ./cmd/migrate up
go run ./cmd/migrate version
go run ./cmd/migrate down  # одна миграция назад; применять осознанно
```

Ошибка миграции возвращает ненулевой exit code. В Compose она блокирует запуск
API.

## HTTP API

Источник публичного контракта: [../docs/openapi.yaml](../docs/openapi.yaml).

| Метод | Путь | Назначение |
|---|---|---|
| `GET` | `/healthz` | Liveness |
| `GET` | `/readyz` | Readiness |
| `GET` | `/api/v1/scenarios` | Активный каталог, фильтр `?role=buyer\|seller` |
| `GET` | `/api/v1/scenarios/{scenarioId}` | Метаданные сценария |
| `POST` | `/api/v1/attempts` | Начать прохождение |
| `GET` | `/api/v1/attempts/{attemptId}` | Восстановить состояние |
| `POST` | `/api/v1/attempts/{attemptId}/choices` | Отправить выбор |
| `GET` | `/api/v1/progress` | Прогресс текущего профиля |

Все запросы `/api/v1/**` должны сохранять cookie. Пример начала попытки:

```bash
BASE=http://localhost:8080/api/v1
curl -sS -c cookies.txt -b cookies.txt \
  -H 'Content-Type: application/json' \
  -d '{"scenarioId":"buyer-fake-delivery"}' \
  "$BASE/attempts"
```

Клиент отправляет только `nodeId`, `choiceId` и непустой `idempotencyKey`.
Следующий узел, score, метки риска и эффекты навыков вычисляет сервер.
Повтор того же запроса с тем же ключом возвращает сохранённый результат и не
дублирует решение; другой payload с тем же ключом отвечает `409`.

## Прогресс

`GET /api/v1/progress` агрегирует только данные текущего профиля:

- завершённые попытки, average/best/latest score;
- активные попытки и прогресс по сценариям;
- эффекты навыков и слабые risk tags;
- последние завершения и рекомендации.

Результат завершённой попытки читается из сохранённой записи и не
пересчитывается отдельным scoring-алгоритмом. Незавершённые попытки не входят в
completed statistics. Новый профиль получает `200` с нулевой сводкой и
пустыми массивами.

## Архив версий сценариев

JSON-файлы в `scenarios/` — авторский источник. При старте PostgreSQL-режима
они декодируются строгим parser, проходят доменную проверку и синхронизируются
в `scenario_versions` одной транзакцией.

- `(scenario_id, version)` неизменяема;
- тот же canonical SHA-256 синхронизируется идемпотентно;
- другой hash требует увеличить `version` и блокирует запуск;
- каталог показывает активную embedded-версию;
- старая попытка читает точную архивную версию после deploy/restart;
- отсутствие или повреждение этой версии считается ошибкой целостности.

В canonical hash входят version, граф, тексты, переходы, scoring, risk tags и
skill effects. Порядок JSON-ключей и whitespace не влияют; `isActive`
исключён как изменяемая политика каталога. Подробнее:
[../docs/scenario-format.md](../docs/scenario-format.md).

## Архитектура

```text
httpapi       маршруты, DTO, middleware, перевод ошибок
    ↓
services      scenario, attempt, profile, progress
    ↓
domain        правила без HTTP и PostgreSQL
    ↑
storage       postgres (production) / memory (explicit tests and demo)
```

```text
backend/
├── cmd/api/                  сборка API и graceful shutdown
├── cmd/migrate/              CLI встроенных миграций
├── internal/app/             composition root
├── internal/scenario/        домен, strict fixture decoder, каталог
├── internal/scenarioarchive/ canonical hash и синхронизация версий
├── internal/attempt/         прохождение и идемпотентные переходы
├── internal/profile/         анонимные профили
├── internal/progress/        агрегация прогресса
├── internal/storage/         PostgreSQL и memory adapters
├── internal/httpapi/         HTTP transport
├── migrations/               embedded SQL migrations
└── scenarios/                embedded JSON fixtures
```

Optimistic locking по `attempt.version` защищает конкурентные обновления.
Уникальность `(attempt_id, idempotency_key)` и транзакционная запись попытки с
решением обеспечивают replay safety. Домен не импортирует HTTP DTO или pgx.

## Проверки

```bash
cd backend
gofmt -l .
go vet ./...
go test -count=1 ./...
go test -count=1 -race ./...
golangci-lint run --config=../.golangci.yaml ./...
```

На Windows `-race` требует C toolchain. PostgreSQL integration suite использует
build tag и отдельную базу; она пересоздаёт схему `public`, поэтому никогда не
указывайте рабочую базу.

```bash
cd backend
TEST_DATABASE_URL='postgres://antiscam:integration-only@127.0.0.1:5432/antiscam_test?sslmode=disable' \
  go test -count=1 -p 1 -tags=integration ./...
```

`-p 1` обязателен, потому что пакеты делят тестовую базу. Без
`TEST_DATABASE_URL` integration tests пропускаются с явным сообщением.

Полный Compose business flow проверяет `.github/scripts/smoke.sh`.

## Известные ограничения

- production hosting не входит в репозиторий; frontend лежит в `../frontend`;
- нет учётных записей и восстановления потерянной анонимной cookie;
- нет встроенного rate limiting;
- каталог мал и пока не требует пагинации;
- backup/restore PostgreSQL и TLS относятся к production-платформе.
