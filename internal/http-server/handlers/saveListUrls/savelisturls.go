package savelisturls

import (
	"log/slog"
	"net/http"

	"github.com/LashkaPashka/TaskDownloader/internal/payload"
	"github.com/LashkaPashka/TaskDownloader/internal/lib/req"
)

type Service interface {
	SaveTask(body payload.SaveTaskRequest) (success bool, err error)
}

func New(serv Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := req.HandleBody[payload.SaveTaskRequest](w, r, logger)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return 
		}

		// TODO: save task on Json
		if _, err := serv.SaveTask(body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}