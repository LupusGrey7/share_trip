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

- Make
 for start build or start test etc
open terminal and 
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