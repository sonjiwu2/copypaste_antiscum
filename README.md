# Anti-Scam Trainer (`copypaste_antiscum`)

Интерактивный **антискам-тренажёр** для сделок на классифайдах: пользователь выбирает роль (покупатель / продавец), проходит учебный диалог, принимает решения и получает объяснения рисков.

Репозиторий — общий (frontend + backend). Этот документ — быстрый старт для жюри и команды.

## Стек (MVP)

| Слой | Статус |
|------|--------|
| **Backend** — Go 1.25, stdlib HTTP API | есть |
| **Сценарии** — JSON + `go:embed`, валидация при старте | есть |
| **Попытки** — in-memory (без БД) | есть |
| **PostgreSQL** | ещё не подключён (Backend №2) |
| **Frontend** — React 18 + TypeScript + Vite | ещё не в репозитории |
| **Docker Compose** | есть |
| **CI/CD** — GitHub Actions | есть |

Подробности backend: [`backend/README.md`](backend/README.md)
OpenAPI: [`docs/openapi.yaml`](docs/openapi.yaml)
Формат сценариев: [`docs/scenario-format.md`](docs/scenario-format.md)

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) + Docker Compose v2
- (опционально, без Docker) Go **1.25+**

## Быстрый запуск (жюри)

```bash
git clone https://github.com/sonjiwu2/copypaste_antiscum.git
cd copypaste_antiscum
git checkout dev   # или актуальная интеграционная ветка / tag

cp .env.example .env   # опционально — значения по умолчанию уже рабочие

docker compose up --build
```

Сервис API:

| | |
|--|--|
| Base URL | http://localhost:8080 |
| Health | http://localhost:8080/healthz |
| Scenarios | http://localhost:8080/api/v1/scenarios |

Проверка:

```bash
curl --fail http://localhost:8080/healthz
curl --fail http://localhost:8080/api/v1/scenarios
```

Остановка:

```bash
docker compose down
```

> **Ограничение MVP:** данные попыток живут в памяти процесса. После `docker compose down` / рестарта контейнера попытки пропадают. PostgreSQL будет добавлен отдельно.

## Архитектура Compose

```
┌─────────────────────────┐
│  backend (:8080)        │  Go API, embedded scenarios, in-memory attempts
│  non-root, healthcheck  │
└─────────────────────────┘
         ▲ future
         │
   postgres / frontend (закомментированы в compose.yaml)
```

Сервис сейчас один: **`backend`**. Заготовки `postgres` и `frontend` оставлены комментариями в `compose.yaml` — без фиктивных контейнеров.

## Локальная разработка backend (без Docker)

```bash
cd backend
go run ./cmd/api
```

### Тесты и качество

```bash
cd backend
gofmt -l .          # пустой вывод = OK
go vet ./...
go test ./...
go test -race ./...

# линтер: конфиг в корне репозитория
golangci-lint run --config=../.golangci.yaml ./...
```

На Windows для `-race` нужен C toolchain (например MinGW-w64).

## Переменные окружения

См. [`.env.example`](.env.example). Backend читает:

| Переменная | По умолчанию | Назначение |
|------------|--------------|------------|
| `HTTP_ADDR` | `:8080` | Адрес listen |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `BACKEND_PORT` | `8080` | Проброс порта хоста (Compose) |

Секреты в git не коммитятся (`.env` в `.gitignore`).

## Сценарии

- Файлы: `backend/scenarios/*.json`
- Встраиваются в бинарник (`go:embed`) — отдельный volume не нужен
- Битая фикстура **не даёт** приложению стартовать
- Как добавить сценарий: см. [`docs/scenario-format.md`](docs/scenario-format.md)

## CI/CD

| Workflow | Файл | Когда | Что делает |
|----------|------|-------|------------|
| **CI** | `.github/workflows/ci.yml` | PR/push в `dev`/`main` | gofmt, vet, test, race, coverage, golangci-lint, `docker compose` smoke |
| **Release** | `.github/workflows/release.yml` | tag `v*` или manual | push образа backend в **GHCR** |

Образ (после release):

```text
ghcr.io/<owner>/copypaste_antiscum/backend:<tag>
```

### Публичный деплой

Хостинг пока **не зафиксирован**. Release публикует контейнер; выкладка на VPS/Fly.io/Railway/Render — отдельный шаг:

1. Создать environment / secrets у провайдера
2. Тянуть образ из GHCR
3. Пробросить `HTTP_ADDR=:8080`
4. Smoke: `GET /healthz`

Публичная URL жюри появится здесь после реального деплоя.

## API (кратко)

| Метод | Путь | Назначение |
|-------|------|------------|
| `GET` | `/healthz` | Liveness |
| `GET` | `/api/v1/scenarios` | Каталог (`?role=buyer\|seller`) |
| `GET` | `/api/v1/scenarios/{id}` | Метаданные |
| `POST` | `/api/v1/attempts` | Старт попытки |
| `GET` | `/api/v1/attempts/{id}` | Состояние |
| `POST` | `/api/v1/attempts/{id}/choices` | Выбор |

Полный контракт: [`docs/openapi.yaml`](docs/openapi.yaml).

## Расширение стека

1. **PostgreSQL (Backend №2)** — раскомментировать сервис в `compose.yaml`, передать `DATABASE_URL`, миграции.
2. **Frontend** — `frontend/Dockerfile` (Vite build → nginx), proxy `/api` и `/healthz` на `backend`, `depends_on: service_healthy`.
3. **CD** — job deploy из `main` / tag с secrets выбранной платформы.

## Использование ИИ

Сценарии, Docker-конфигурация и CI/CD частично готовились с помощью AI-ассистентов и **проверялись** локальными тестами / (по возможности) Docker-сборкой командой. Ответственность за merge — у участников.

## Лицензия

MIT — см. [LICENSE](LICENSE).
