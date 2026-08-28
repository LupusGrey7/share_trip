# Share Trip Service — Project Context 

## What is it
Это приложение, которое обрабатывает поездки в Такси.
Состоит из нескольких компонентов:
1. Share Trip Service(этот проект) написана на Golang + goose+postgreSQL+kafka (в перспективеб пока не трогаем)
- принимает данные из REST API проводит обработку(прием заказа, поиск заказа, обновление и тд)
- возвращает быстрай ответ чтобы не задерживать потоки
- дает возможность проверить status или создать транзакцию
2. Share Trip Notification Service(app 2, stack Golang\goose\postgreSQL\kafka)
    - принимает данные из Кафка о поезках и уведомления и проводит оповещение сторон

3. Share Trip Contract Service (app 3, stack Golang\goose\postgreSQL)    

## Tech Stack
- **Backend:** Go 1.25 (сервисы на Go — отдельные модули)
- **Build:** Go
- **DB:** PostgreSQL, миграции через Goose
- **Tests:** Test
- **CI:** GitHub Actions
- **Контейнеризация:** Docker + docker-compose для локального окружения (deploy/)

### Description
chapter on construction

### build
chapter on construction

## Quick start
- Make
 for start build or start test etc
open terminal and 

```powershell
# 1) Postgres
make deploy

# 2) Migrations
make migrate-up

# 3) API on the host (loads .env.dev)
make run
```

Smoke check:

```powershell
curl http://127.0.0.1:8080/ready
```

```bash
make help
```
Also, you can select and specify a command separately, for example
for run all test files
```bash
# start on all test
make test
```
or 
run go format
```bash
make fmt
```

---

### Порты и URL

| Что | URL / host:port |
|-----|-----------------|
| **ShareTrip API** | http://localhost:8080 |
| **ShareTrip ready** | http://localhost:8080/ready |
| **ShareTrip metrics** | http://localhost:8080/metrics |
| **Contract API** (внешний сервис) | http://localhost:8081 — env `CONTRACT_SERVICE_URL` |
| **Contract availability** | `GET /api/v2/companies/{companyId}/services/{serviceCode}/availability` |
| **Jaeger UI** | http://localhost:16686 |
| **OTEL Collector** (экспорт с хоста) | `localhost:4319` (HTTP OTLP → Jaeger) |
| **ShareTrip PostgreSQL** | `localhost:6543` |
| **Keycloak** | http://localhost:8087 (realm: `/realms/sharetrip`) |
| **Grafana** | http://localhost:3000 (`admin` / `admin`) |
| **Prometheus** | http://localhost:9090 |
| **Kafka broker** (с хоста) | `localhost:9092` |
| **Kafka UI** (Kafdrop) | http://localhost:9000 |
| **Loki** | http://localhost:3100 |
| **postgres_exporter** | http://localhost:9187/metrics |

Инфра: `make up` / `make down` → `deploy/docker-compose.yml`.  
Подробнее про стек observability: [`.docs/cheatsheets/observability-cheatsheet.md`](.docs/cheatsheets/observability-cheatsheet.md).

### HTTP API ShareTrip Service

Базовый префикс: `/api/v2/trip`. Маршруты `trip/*` защищены Keycloak (Bearer + роль `client`).

```text
GET    /ready
GET    /metrics

GET    /api/v2/trip/:tripId
POST   /api/v2/trip/createTripDraft
PATCH  /api/v2/trip/moveTripDraft-ToPublish/:tripId
PATCH  /api/v2/trip/moveTripPublished-ToStarted/:tripId/company/:companyId/service/:serviceCode
```

Smoke после `make run`:

```powershell
curl http://127.0.0.1:8080/ready
```

Postman / JWT: [`.docs/cheatsheets/keycloak-cheatsheet.md`](.docs/cheatsheets/keycloak-cheatsheet.md).

### Трейсинг в Jaeger (ShareTrip)

| Endpoint | Spans в Jaeger |
|----------|----------------|
| `POST /api/v2/trip/createTripDraft` | ✅ handler → service → usecase → repository → `DB.Transaction` |
| `GET /api/v2/trip/:tripId` | ✅ полная цепочка |
| `PATCH /api/v2/trip/moveTripDraft-ToPublish/:tripId` | ✅ полная цепочка (+ outbox в tx) |
| `PATCH /api/v2/trip/moveTripPublished-ToStarted/...` | ✅ handler → Contract check → service → usecase → repository |
| `GET /ready` | ❌ |
| `GET /metrics` | ❌ |

В Jaeger выбери сервис **`share-trip`** (не `trip-api`, не `sharetrip-contract`).  
UI: http://localhost:16686 → Service `share-trip` → Find Traces.

Подробнее: [`.docs/cheatsheets/observability-cheatsheet.md`](.docs/cheatsheets/observability-cheatsheet.md) §4, §11.

### Grafana — дашборды (что / где / зачем)

UI: http://localhost:3000 → папка **`Share_Trip`** (автозагрузка из `deploy/grafana/dashboards_files/`).

| Дашборд | Файл | Зачем |
|---------|------|--------|
| **HTTP Status Codes** | `http_status_codes_go.json` | **Коды ответа**: error rate %, 2xx/4xx/5xx, top failing endpoints |
| Process Metrics | `app_metrics_go.json` | HTTP RPS/latency + бизнес trip (create/publish) + repository |
| Runtime Go | `runtime_go.json` | goroutines, heap, GC, CPU |
| PostgreSQL | `postgresql_go.json` | БД: up, connections, commits/rollbacks |

`app_metrics_go` показывает `status` в одной панели RPS
`http_status_codes_go.json` дашборд по кодам ответа — для ревью 4xx/5xx открывай в Графана **HTTP Status Codes** 

Подробно: [`.docs/cheatsheets/grafana-dashboards-cheatsheet.md`](.docs/cheatsheets/grafana-dashboards-cheatsheet.md).

---