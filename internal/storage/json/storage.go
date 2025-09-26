package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/LashkaPashka/TaskDownloader/internal/lib/encode"
	"github.com/LashkaPashka/TaskDownloader/internal/models"
)

const (
	statusInProgress = "in_progress"
	statusRunning = "running"
	statusFailed = "failed"
	statusQueued = "queued"
	statusCompleted = "completed"
	statusDone = "done"
)

func searchTask(tasks []models.Task, taskID string) models.Task {
	for _, task := range tasks {
		if task.ID == taskID {
			return task
		}
	}
	return models.Task{}
}

func updateTask(tasks []models.Task, taskID string, file *models.File) []models.Task {	
	for ti := range tasks {
		if tasks[ti].ID != taskID {
			continue
		}

		for fi := range tasks[ti].File {
			if tasks[ti].File[fi].Index == file.Index {	
				tasks[ti].File[fi].DownloadedBytes = file.DownloadedBytes
				tasks[ti].File[fi].Size = file.Size
				tasks[ti].File[fi].Status = file.Status
			
				switch file.Status {
					case statusInProgress:
						tasks[ti].File[fi].StartedAt = time.Now()
						tasks[ti].Status = statusRunning
					case statusDone:
						tasks[ti].File[fi].FinishedAt = time.Now()
						tasks[ti].Status = statusCompleted
					default:
						tasks[ti].Status = statusFailed
				}

				return tasks
			}			
		}
	}

	return nil
}

type Storage struct {
	storagePath string
	logger *slog.Logger
}

func New(storagePath string, logger *slog.Logger) (*Storage, error) {
	if _, err := os.Stat(storagePath); err != nil {
		if os.IsNotExist(err) {
			logger.Error("Storage path does not exist",
				slog.String("path", storagePath),
				)
				return nil, fmt.Errorf("storage path does not exist: %s", storagePath)
			}
						logger.Error("Error checking storage path",
				slog.String("path", storagePath),
				slog.String("err", err.Error()),
			)
		return nil, err
	}

	return &Storage{
		storagePath: storagePath,
		logger: logger,
	}, nil
}

func (s *Storage) SaveTask(task models.Task) (success bool, err error) {
	const op = "TaskDonwloader.storage.methodsForJson.SaveTask"

	file, err := os.ReadFile(s.storagePath)
	if err != nil {
		s.logger.Error("Invalied readfile", 
			slog.String("op", op),
			slog.String("err", err.Error()),
		)
		return false, err
	}

	tasks := []models.Task{}

	json.Unmarshal(file, &tasks)

	tasks = append(tasks, task)

	encodeTasks, err := encode.Encode(tasks, s.logger)
	if err != nil {
		return false, err
	}

	if err = os.WriteFile(s.storagePath, encodeTasks, 0644); err != nil {
		s.logger.Error("Invalid save task in file",
			slog.String("op", op),
			slog.String("err", err.Error()),
		)
		return false, err
	}

	return true, nil
}

func (s *Storage) GetTask(taskID string) (models.Task, error) {
	const op = "TaskDonwloader.storage.methodsForJson.GetTask"
	
	file, err := os.Open(s.storagePath)
	if err != nil {
		s.logger.Error("Failed to open storage file",
        	slog.String("file", s.storagePath),
        	slog.String("err", err.Error()),
    	)
		return models.Task{}, err
	}
	
	scanner := bufio.NewScanner(file)	

	var b []byte
	for scanner.Scan() {
		b = append(b, []byte(scanner.Text()+"\n")...)
	}
	if err := scanner.Err(); err != nil {
		 s.logger.Error("Scanner encountered an error",
			slog.String("op", op),
			slog.String("err", err.Error()),
    	)
		return models.Task{}, err
	}

	var tasks []models.Task

	json.Unmarshal(b, &tasks)

	task := searchTask(tasks, taskID)

	return task, nil
}

func (s *Storage) SaveFile(taskID string, file *models.File) (success bool, err error) {
	const op = "TaskDonwloader.storage.methodsForJson.UpdateTask"

	f, err := os.Open(s.storagePath)
	if err != nil {
		return false, err
	}
	
	scanner := bufio.NewScanner(f)	

	var b []byte
	for scanner.Scan() {
		b = append(b, []byte(scanner.Text()+"\n")...)
	}
	if err := scanner.Err(); err != nil {
		s.logger.Error("Invalid scanner text", 
			slog.String("op", op),
			slog.String("err", err.Error()),
		)
		return false, err
	}

	var tasks []models.Task

	json.Unmarshal(b, &tasks)

	// TODO: update status of Task  
	updatedTasks := updateTask(tasks, taskID, file)

	// TODO: convert Task
	bytesTasks, err := encode.Encode(updatedTasks, s.logger)
	if err != nil {
		s.logger.Error("Invalid encode tasks", 
			slog.String("op", op),
			slog.String("err", err.Error()),
		)
		return false, err
	}

	if err := os.WriteFile(s.storagePath, bytesTasks, 0644); err != nil {
		panic(err)
	}

	return true, nil
}

func (s *Storage) ResetToQueued() (fileMp map[string][]models.File, err error) {
	const op = "TaskDonwloader.storage.methodsForJson.ResetToQueued"

	f, err := os.Open(s.storagePath)
	if err != nil {
		return nil, err
	}
	
	scanner := bufio.NewScanner(f)	

	var b []byte
	for scanner.Scan() {
		b = append(b, []byte(scanner.Text()+"\n")...)
	}
	if err := scanner.Err(); err != nil {
		s.logger.Error("Invalid scanner text", 
			slog.String("op", op),
			slog.String("err", err.Error()),
		)
		return nil, err
	}

	var tasks []models.Task

	json.Unmarshal(b, &tasks)

	// If length is equal 0 then return nil
	if len(tasks) == 0 {
		return nil, nil
	}

	fileMp = make(map[string][]models.File, len(tasks))

	for ti := range tasks {
		if tasks[ti].Status == statusCompleted {
			continue
		}

		for fi := range tasks[ti].File {
			if tasks[ti].File[fi].Status == statusInProgress || tasks[ti].File[fi].Status == statusQueued {
				tasks[ti].File[fi].Status = statusQueued
				fileMp[tasks[ti].ID] = append(fileMp[tasks[ti].ID], tasks[ti].File[fi])
			}
		}
	}

	bytesTasks, err := encode.Encode(tasks, s.logger)
	if err != nil {
		s.logger.Error("Invalid encode tasks", 
			slog.String("op", op),
			slog.String("Err", err.Error()),
		)
		return nil, err
	}

	if err := os.WriteFile(s.storagePath, bytesTasks, 0644); err != nil {
		panic(err)
	}

	return fileMp, nil
}

func (s *Storage) GetFileById(taskID string, fileID int) (models.File, error) {
	task, err := s.GetTask(taskID)
	if err != nil {
		return models.File{}, err
	}

	for _, file := range task.File {
		if file.Index == fileID {
			return file, nil
		}
	}

	return models.File{}, errors.New("not exist such file")
}