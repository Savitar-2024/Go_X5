---

## Результаты выполнения

Реализован backend-сервис управления заказами на Go с разделением на слои `handler`, `service`, `repository`.

Реализованы все шесть HTTP-эндпоинтов, включая создание заказов, получение списка, изменение статусов и просмотр истории.

Данные хранятся в потокобезопасном in-memory-репозитории. Для конкурентного доступа используется `sync.RWMutex`.

HTTP-запросы логируются через `log/slog`. Логи содержат метод, путь, статус, длительность и идентификатор запроса. Middleware также обрабатывает паники.

Swagger-документация доступна по адресу `http://localhost:9090/swagger/index.html` при запуске на порту 9090.

### Запуск из PowerShell

```powershell
cd control_1
go mod download
$env:HTTP_ADDR = ":9090"
go run ./cmd/api
```

По умолчанию сервер использует порт 8080.

### Проверка

```powershell
go test ./...
go test ./internal/service -cover
go build ./...
go vet ./...
```

Покрытие бизнес-логики сервиса — 97%, что превышает требуемые 80%.

Swagger-документация генерируется командой:

```powershell
swag init -g cmd/api/main.go -o docs --parseInternal
```