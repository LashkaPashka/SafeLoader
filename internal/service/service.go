package service

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	converttotask "github.com/LashkaPashka/TaskDownloader/internal/lib/convertToTask"
	eventbus "github.com/LashkaPashka/TaskDownloader/internal/lib/eventBus"
	"github.com/LashkaPashka/TaskDownloader/internal/models"
	"github.com/LashkaPashka/TaskDownloader/internal/payload"
)

const (
	statusDone = "done"
	statusFailed = "failed"
	statusInProgress = "in_progress"
)

type Storage interface {
	SaveTask(task models.Task) (success bool, err error)
	GetTask(taskID string) (models.Task, error)
	SaveFile(taskID string, file *models.File) (success bool, err error)
	GetFileById(taskID string, fileID int) (models.File, error)
	ResetToQueued() (fileMp map[string][]models.File, err error)
}

type GoFetchService struct {
	logger *slog.Logger
	eventBus *eventbus.EventBus
	localStoragePath string
	storage Storage
}

func New(
	storage Storage,
	localStoragePath string,
	eventBus *eventbus.EventBus, 
	logger *slog.Logger,
) (*GoFetchService, error) {
	return &GoFetchService{
		logger: logger,
		eventBus: eventBus,
		localStoragePath: localStoragePath,
		storage: storage,
	}, nil
}

func (g *GoFetchService) SaveTask(body payload.SaveTaskRequest) (success bool, err error) {
	const op = "TaskDownloader.service.goFetch.SaveTask"

	// TODO: convert to modelTask
	task := converttotask.Convert(&body)

	// TODO: call method saveTask in file json
	if success, err := g.storage.SaveTask(task); err != nil {
		g.logger.Error("Invalid method of storage SaveTask", 
			slog.String("op", op),
			slog.String("err", err.Error()),
		)
		return success, err
	}

	// TODO: create event in queue
	go g.eventBus.Publish(eventbus.Event{
		Type: eventbus.EventCreateTask,
		Data: models.EventData{
			ClientID: task.ClientID,
			TaskID: task.ID,
		},
	})

	return true, nil
}

func (g *GoFetchService) CompleteTask() {
	const op = "TaskDownloader.service.goFetch.CompleteTask"

	for msg := range g.eventBus.Subscribe() {
		if msg.Type == eventbus.EventCreateTask {
			eventData, ok := msg.Data.(models.EventData)
			if !ok {
				log.Fatalln("Wrong data")
				continue
			}
			
			task, err := g.storage.GetTask(eventData.TaskID)
			if err != nil {
				g.logger.Error("Error get task", slog.String("op", op))
				return
			}

			var wg sync.WaitGroup
			var mux sync.Mutex

			wg.Add(len(task.File))
			for i := range task.File {
				file := &task.File[i]

				go func (taskID string, file *models.File) {
					defer wg.Done()

					if err := g.DownloadWithResume(&mux, taskID, file); err != nil {
						g.logger.Error("Ivalid download file", 
							slog.String("err", err.Error()), 
							slog.String("op", op),
						)

						file.Status = statusFailed

						if _, err := g.storage.SaveFile(taskID, file); err != nil {
								g.logger.Error("Failed to save file status",
									slog.String("err", err.Error()),
									slog.String("op", op),
									slog.String("task_id", taskID),
								)
						}
					}
				}(eventData.TaskID, file)
			}
			
			wg.Wait()

		} else if msg.Type == eventbus.EventUnfinishedTask {
			data, ok := msg.Data.(map[string][]models.File)
			if !ok {
				log.Fatalln("Wrong data")
				continue
			}

			var wg sync.WaitGroup
			var mux sync.Mutex

			for taskID, fileList := range data {
				for i := range fileList {
					wg.Add(1)
					file := &fileList[i]
					go func (taskID string, file *models.File) {
						defer wg.Done()

						if err := g.DownloadWithResume(&mux, taskID, file); err != nil {
							g.logger.Error("Invalid download file",
								slog.String("err", err.Error()),
								slog.String("op", op),
							)
	
							file.Status = statusFailed

							if _, err := g.storage.SaveFile(taskID, file); err != nil {
								g.logger.Error("Failed to save file status",
									slog.String("err", err.Error()),
									slog.String("op", op),
									slog.String("task_id", taskID),
								)
							}
						}
					}(taskID, file)
				}
				wg.Wait()
			}
		}
	}
}

func (g *GoFetchService) SearchQueuedAndComplete() (error) {
	// TODO: serach task where status = in_progress and set up queued
	files, err := g.storage.ResetToQueued()
	if err != nil {
		return err
	}

	go g.eventBus.Publish(eventbus.Event{
		Type: eventbus.EventUnfinishedTask,
		Data: files,
	})
	
	return nil
}

func (g *GoFetchService) DownloadWithResume(mux *sync.Mutex, taskID string, file *models.File) (error) {
	const op = "TaskDownloader.service.DownloadWithResume"
	
	path := filepath.Join(g.localStoragePath, taskID, file.Filename)
	
	if err := os.MkdirAll(filepath.Join(g.localStoragePath, taskID), os.ModePerm); err != nil {
        g.logger.Error("Invalid mkdir folder", 
			slog.String("op", op),
			slog.String("err", err.Error()),
		)
		return err
    }

	tmpPath := path + ".part"

	out, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		g.logger.Error("Error open file",
		 	slog.String("op", op),
			slog.String("err", err.Error()),
		)
		return err
	}

	defer out.Close()

	if _, err := out.Seek(file.DownloadedBytes, io.SeekStart); err != nil {
		g.logger.Error("Error search file", 
			slog.String("op", op),
			slog.String("err", err.Error()),
		)
		return err
	}

	req, _ := http.NewRequest("GET", file.Url, nil)
	if file.DownloadedBytes > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", file.DownloadedBytes))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
			g.logger.Error("Failed to make HTTP request",
			slog.String("url", req.URL.String()),
			slog.String("method", req.Method),
			slog.String("op", op),
			slog.String("err", err.Error()),
		)
		return err
	}
	defer resp.Body.Close()

	if file.Size == 0 && resp.ContentLength > 0 {
		file.Size = file.DownloadedBytes + resp.ContentLength
	}

	buf := make([]byte, 32*1024)
	file.Status = statusInProgress
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, err := out.Write(buf[:n]); err != nil {
				return err
			}

			file.DownloadedBytes += int64(n)
			mux.Lock()
			if _, err := g.storage.SaveFile(taskID, file); err != nil {
				g.logger.Error("Failed to save file status",
					slog.String("err", err.Error()),
					slog.String("task_id", taskID),
				)
			}
			mux.Unlock()
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			g.logger.Error("Failed to write chunk to file",
            slog.String("file", tmpPath),
            slog.Int("bytes_written", n),
            slog.String("err", err.Error()),
        )
			return err
		}
	}

	file.Status = statusDone
	if _, err := g.storage.SaveFile(taskID, file); err != nil {
		g.logger.Error("Failed to save file status",
			slog.String("err", err.Error()),
			slog.String("task_id", taskID),
		)
	}

	return os.Rename(tmpPath, path)
}


func (g *GoFetchService) GetTaskByID(taskID string) (models.Task, error) {
	const op = "TaskDownloader.service.goFetch.GetTaskByID"
	
	task, err := g.storage.GetTask(taskID)
	if err != nil {
		g.logger.Error("Invalid get task", slog.String("op", op))
		return models.Task{}, err
	}

	return task, nil
}


func (g *GoFetchService) GetFileById(taskID string, fileID int) (models.File, error) {
	const op = "TaskDownloader.service.goFetch.GetFileById"
	
	file, err := g.storage.GetFileById(taskID, fileID)
	if err != nil {
		g.logger.Error("Invalid get file", slog.String("Reason", err.Error()),  slog.String("op", op))
		return models.File{}, err
	}

	return file, nil
}