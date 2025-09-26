package storage

import (
	"log/slog"
	"os"
	"testing"

	"github.com/LashkaPashka/TaskDownloader/internal/payload"
	converttotask "github.com/LashkaPashka/TaskDownloader/internal/lib/convertToTask"
)

const storagePath = "../tasks/tasks.json"
var logger = setupLogger("local")
var st *Storage

func TestMain(m *testing.M) {
	st = &Storage{
		storagePath: storagePath,
		logger: logger,
	}

	m.Run()
}

func TestGetTask(t *testing.T) {
	task, err := st.GetTask("task_no9oYjS4J0")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	t.Log(task)
}

func TestSaveTask(t *testing.T) {
	body := payload.SaveTaskRequest{
			Urls: []string{
				"https://getsamplefiles.com/download/zip/sample-1.zip",
				"https://getsamplefiles.com/download/zip/sample-4.zip",
			},
			ClientID: "3f3f32f2",

	}
	
	task := converttotask.Convert(&body)

	if _, err := st.SaveTask(task); err != nil {
		t.Fatalf("Error: %v", err)
	}

	t.Log(task)
}

func TestUpdateTask(t *testing.T) {
	//success, err := st.SaveFile("queued", 1, "task_p4hk5g0KaC")

	// if err != nil {
	// 	t.Fatalf("Error: %v", err)
	// }

	// t.Log(success)
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