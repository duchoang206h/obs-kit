package observability

import (
	"fmt"
	"log/slog"
)

func ErrorAttrs(err error) []any {
	if err == nil {
		return nil
	}

	return []any{
		slog.String("error.type", fmt.Sprintf("%T", err)),
		slog.String("error.message", err.Error()),
	}
}
