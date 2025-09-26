package models

import "time"


type Task struct {
	File 			[]File		`json:"files"` 	
	CreatedAt		time.Time	`json:"created_at"`
	ID				string		`json:"id"`
	ClientID		string		`json:"client_id"`
	Status			string		`json:"status"`
}

type File struct {
	DownloadedBytes		int64		`json:"downloadedBytes"`
	Size				int64		`json:"size"`
	Index				int			`json:"index"`
	Url					string		`json:"url"`
	Filename			string		`json:"filename"`
	Status				string		`json:"status"`
	StartedAt			time.Time	`json:"started_at"`
	FinishedAt			time.Time	`json:"finished_at"`
}

type EventData struct {
	ClientID		string		`json:"client_id"`
	TaskID			string		`json:"task_id"`
}