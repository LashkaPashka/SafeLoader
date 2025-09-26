package req

import (
	"log/slog"

	"github.com/go-playground/validator/v10"
)


func isValid[T any](payload T, logger *slog.Logger) error {
	const op = "AuthService.pkg.req.isValidate.go"

	validate := validator.New()
	err := validate.Struct(payload)
	
	if err != nil {
		logger.Error("Error validation", 
			slog.String("op", op),
			slog.String("err", err.Error()),
		)
		return err
	}

	return nil
}