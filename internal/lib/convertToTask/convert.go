package converttotask

import (
	"path/filepath"
	"time"

	"github.com/LashkaPashka/TaskDownloader/internal/payload"
	"github.com/LashkaPashka/TaskDownloader/internal/lib/random"
	"github.com/LashkaPashka/TaskDownloader/internal/models"
)
const (
	statusQueued = "queued"
)

func Convert(body *payload.SaveTaskRequest) models.Task {	
	var fl []models.File
	
	for index, url := range body.Urls {
		fl = append(fl, models.File{
			Index: index+1,
			Url: url,
			Filename: filepath.Base(url),
			Status: statusQueued,
			StartedAt: time.Time{},
			FinishedAt: time.Time{},
		})
	}

	return models.Task{
		ID: random.RandomString("task_", 10),
		ClientID:  body.ClientID,
		File: fl,
		Status: statusQueued,
		CreatedAt: time.Now(),
	}
}