package payload

type SaveTaskRequest struct {
	Urls			[]string	`json:"urls" vaildate:"required"`
	ClientID		string		`json:"client_id" vaildate:"required"`
}

type GetStatusOfTaskResponse struct {
	Files			[]StatusOfFileResponse	`json:"files"`
	Status			string					`json:"status"`
	TaskID			string					`json:"task_id"`				
	ClientID		string					`json:"client_id"`
}

type StatusOfFileResponse struct {
	DownloadedBytes		int64				`json:"downloadedBytes"`
	Filename			string				`json:"filename"`
	Status 				string				`json:"status"`
}