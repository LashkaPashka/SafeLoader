package payload

type SaveTaskRequest struct {
	Urls			[]string	`json:"urls" vaildate:"required"`
	ClientID		string		`json:"client_id" vaildate:"required"`
}

type GetStatusOfTaskResponse struct {
	Status			string		`json:"status"`
	TaskID			string		`json:"task_id"`				
	ClientID		string		`json:"client_id"`
}

type GetStatusOfFileResponse struct {
	Filename		string		`json:"filename"`
	Url				string		`json:"url"`
	Status 			string		`json:"status"`
}