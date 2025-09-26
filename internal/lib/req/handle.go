package req

import (
	"log/slog"
	"net/http"
)


func HandleBody[T any](w http.ResponseWriter, r *http.Request, logger *slog.Logger) (T, error) {
	// TODO: decode body	
	body, err := decode[T](r.Body, logger)
	if err != nil {
		return body, err
	}

	// TODO: validate decode body
	if err := isValid(body, logger); err != nil {
		return body, err
	}

	return body, nil
}