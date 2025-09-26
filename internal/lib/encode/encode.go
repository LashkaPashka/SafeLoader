package encode

import (
	"encoding/json"
	"log/slog"
)


func Encode[T any](payload T, logger *slog.Logger) ([]byte, error){
	const op = "TaskDownloader.lib.encode"

	enData, err := json.Marshal(payload)
	if err != nil {
		logger.Error("Error encoding payload", 
		slog.String("op", op),
		slog.String("err", err.Error()),
	)
		return nil, err
	}

	return enData, nil
}