package req

import (
	"encoding/json"
	"io"
	"log/slog"
)

func decode[T any](body io.Reader, logger *slog.Logger) (T, error) {
	const op = "AuthService.pkg.req.decode.go"

	var payload T

	err := json.NewDecoder(body).Decode(&payload)

	if err != nil {
		logger.Error("Invalid decode body request", 
			slog.String("op", op),
			slog.String("err", err.Error()),
		)
		return payload, err
	}

	return payload, nil
}