# Выполнил тестовое задание 26.09.2025. 

# TaskDownloader

## Общая информация

TaskDownloader — сервис на Go, который принимает список URL файлов, скачивает их и сохраняет в локальной папке `storage`.

Проект автономный, не использует внешнюю инфраструктуру. Цель сервиса: обработка задач по скачиванию файлов, работа через EventBus, докачка файлов при перезапуске и безопасное параллельное скачивание.

---

## Конфигурация

Пример файла `config/local.yaml`:

```yaml
env: "local"
storage_path: "./tasks/tasks.json"
local_path_storage: "./storage/"
http_server:
  address: "0.0.0.0:8080"
  timeout: 4s
  idle_timeout: 30s
```
Пояснение полей:
 1. env — среда запуска (local)
 2. storage_path — путь к JSON-файлу с задачами и файлами
 3. local_path_storage — папка для скачанных файлов
 4. http_server.address — адрес и порт HTTP-сервера
 5. http_server.timeout — таймаут чтения/записи HTTP-запроса
 6. http_server.idle_timeout — таймаут простоя соединения

## Запуск проекта

Для запуска понадобится написать следующие команды в консоли:
```bash
git clone https://github.com/LashkaPashka/SafeLoader.git
cd SafeLoader
export CONFIG="./config/local.yaml"
go run cmd/taskdownloader/main.go
```

## URL-запросы (API)

Создание новой задачи
POST /tasks

Тело запроса (JSON):
```json
{
	"urls": [
    "https://echo.epa.gov/files/echodownloads/pipeline_caa_downloads.zip",
    "https://echo.epa.gov/files/echodownloads/npdes_outfalls_layer.zip"
	],
	"client_id": "u_342fvr5"
}
```

Получение статуса задачи
GET /tasks/{task_id}

Пример ответа:
```json
{
	"files": [
		{
			"downloadedBytes": 5650509,
			"filename": "pipeline_caa_downloads.zip",
			"status": "done"
		},
		{
			"downloadedBytes": 52891687,
			"filename": "npdes_outfalls_layer.zip",
			"status": "done"
		}
	],
	"status": "completed",
	"task_id": "task_YQuKr2fRF0",
	"client_id": "u_342fvr5"
}
```

## EventBus

Для обработки задач используется паттерн EventBus.

Реализация находится в `lib/eventbus/eventbus.go`.

Когда пользователь создаёт задачу через handler, данные отправляются в EventBus как событие task.created.

События обрабатываются в отдельной горутине в main.

Если задача не завершена (например, сервис перезапущен), файлы со статусом in_progress переводятся в queued и повторно отправляются как task.unfinished.

EventBus реализован через каналы (chan) для очередей событий.


## Обработка сигналов ОС и остановка сервера
```go
done := make(chan os.Signal, 1)
signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

srv := &http.Server{
    Addr: cfg.Address,
    Handler: router,
    ReadTimeout: cfg.HTTPServer.Timeout,
    WriteTimeout: cfg.HTTPServer.Timeout,
    IdleTimeout: cfg.HTTPServer.IdleTimeout,
}

go func() {
    if err := srv.ListenAndServe(); err != nil {
        logger.Error("failed to stop server")
    }
}()

logger.Info("server started")

<-done
logger.Info("stopping server")

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

if err := srv.Shutdown(ctx); err != nil {
    logger.Error("failed to stop server")
    return
}
```
Пояснение кода на Go:
  1. done := make(chan os.Signal, 1) — канал для перехвата сигналов ОС.
  2. signal.Notify подписывает канал на стандартные сигналы остановки (Ctrl+C, SIGINT, SIGTERM).
  3. srv.ListenAndServe() запускает HTTP-сервер в отдельной горутине.
  4. <-done — основной поток ждёт сигнал остановки.
  5. srv.Shutdown(ctx) аккуратно завершает работу сервера, давая 10 секунд на завершение текущих запросов.



## Обработка незавершённых задач

При запуске проверяются все задачи со статусом running.

Файлы со статусом in_progress переводятся в queued.

Они добавляются обратно в EventBus как task.unfinished для докачки.

## Параллельное скачивание и безопасное сохранение

Каждый файл скачивается в отдельной горутине.

Для ожидания завершения всех файлов используется sync.WaitGroup.

Для безопасного сохранения прогресса в task.json применяется sync.Mutex.

Прогресс сохраняется по мере скачивания, включая поле downloaded_bytes.

## Логирование и статус

Все действия логируются через logger.

Статусы задач и файлов:

- queued — готов к скачиванию

- in_progress — скачивание в процессе

- failed — ошибка

- done — скачивание завершено

## Архитектура и SOLID

В проекте используется принцип DIP (Dependency Inversion Principle) из SOLID.

Бизнес-логика (сервис) не зависит напрямую от конкретной реализации хранилища или логгера.

Вместо этого в коде применяются интерфейсы.

Конкретные реализации (Storage, Service и т.д.) подставляются при инициализации приложения.

Это позволяет легко расширять проект, подменять зависимости (например, заменить файловое хранилище на БД) и упрощает тестирование.

## Тестирование

В проекте написаны unit-тесты:

Для storage — проверка корректности сохранения, загрузки и обновления задач в JSON-файле.

Для service — проверка бизнес-логики обработки задач, изменения статусов файлов и повторного запуска незавершённых загрузок.
