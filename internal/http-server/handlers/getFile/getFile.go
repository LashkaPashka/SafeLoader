package getfile

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/LashkaPashka/TaskDownloader/internal/models"
	"github.com/LashkaPashka/TaskDownloader/internal/payload"
)

type Storage interface {
	GetFileById(taskID string, fileID int) (models.File, error)
}

func New(storage Storage, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID := r.URL.Query().Get("task_id")
		strFileID := r.URL.Query().Get("file_id")
		
		fileID, _ := strconv.Atoi(strFileID)

		file, err := storage.GetFileById(taskID, fileID)
		if err != nil {
			w.WriteHeader(http.StatusConflict)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&payload.GetStatusOfFileResponse{
			Filename: file.Filename,
			Url: file.Url,
			Status: file.Status,
		})
	}
}