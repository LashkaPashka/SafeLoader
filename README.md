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

## URL-запросы (API)

Создание новой задачи
POST /tasks

Тело запроса (JSON):
```json
{
  "urls": [
    "https://example.com/file1.jpg",
    "https://example.com/file2.png"
  ],
  "client_id": "12345"
}
```

Получение статуса задачи
GET /tasks/{task_id}

Пример ответа:
```json
{
  "task_id": "task_1dsa3r",
  "status": "in_progress",
  "files": [
    {
      "url": "https://example.com/file1.jpg",
      "status": "done",
      "downloaded_bytes": 800
    },
    {
      "url": "https://example.com/file2.png",
      "status": "in_progress",
      "downloaded_bytes": 203
    }
  ]
}
```

## EventBus

Для обработки задач используется паттерн EventBus.

Реализация находится в lib/eventbus/eventbus.go.

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
