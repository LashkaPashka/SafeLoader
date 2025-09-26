package gettask

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/LashkaPashka/TaskDownloader/internal/models"
	"github.com/LashkaPashka/TaskDownloader/internal/payload"
)

type Service interface {
	GetTaskByID(taskID string) (models.Task, error)
}

func New(service Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID := r.URL.Query().Get("task_id")

		task, err := service.GetTaskByID(taskID)
		if err != nil {
			w.WriteHeader(http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-type", "application/json")
		json.NewEncoder(w).Encode(&payload.GetStatusOfTaskResponse{
			ClientID: task.ClientID,
			TaskID: task.ID,
			Files: getFileResponse(task),
			Status: task.Status,
		})
	}
}

func getFileResponse(task models.Task) (fileResponse []payload.StatusOfFileResponse) {
	for _, file := range task.File {
		fileResponse = append(fileResponse, payload.StatusOfFileResponse{
			Filename: file.Filename,
			Status: file.Status,
			DownloadedBytes: file.DownloadedBytes,
		})
	}
	
	return fileResponse
}