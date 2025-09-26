package service

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	eventbus "github.com/LashkaPashka/TaskDownloader/internal/lib/eventBus"
	"github.com/LashkaPashka/TaskDownloader/internal/payload"
	storage "github.com/LashkaPashka/TaskDownloader/internal/storage/json"
)
var body = payload.SaveTaskRequest{
		Urls: []string{
			"https://getsamplefiles.com/download/zip/sample-1.zip",
			"https://getsamplefiles.com/download/zip/sample-4.zip",
		},
		ClientID: "3f3f32f2",
}

const filename = "../../storage/tasks/tasks.json"
const localStoragePath = "./storageFiles/"
var ft *GoFetchService



func TestMain(m *testing.M) {
	logger := setupLogger("local")
	
	eventBus := eventbus.NewEventBus()

	storage, _ := storage.New(filename, logger)

	ft = &GoFetchService{
		logger: logger,
		eventBus: eventBus,
		storage: storage,
	}	

	m.Run()
}

func TestDownloadWithResume(t *testing.T) {
	taskID := "task_9UnHfCqGN8"
	
	tasks, err := ft.storage.GetTask(taskID)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	file := tasks.File[0]

	if err := ft.DownloadWithResume(nil, taskID, &file); err != nil {
		t.Fatalf("Error: %v", err)
	}

	t.Log("Ok!")
}

func TestGourotine(t *testing.T) {
	go ft.eventBus.Publish(eventbus.Event{
		Type: eventbus.EventCreateTask,
		Data: "123",
	})
	
	go ft.eventBus.Publish(eventbus.Event{
		Type: eventbus.EventCreateTask,
		Data: "35",
	})


	for msg := range ft.eventBus.Subscribe() {
		if msg.Type == eventbus.EventCreateTask {
			t.Log(msg.Data)
		}
	}
	
	
}

func TestFilePath(t *testing.T) {
	fileUrl := "https://getsamplefiles.com/download/zip/sample-1.zip"

	filename := filepath.Base(fileUrl)

	path := filepath.Join(localStoragePath, "123134", filename)

	t.Log(path)
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case "local":
		log  = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case "dev":
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case "prod":
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}