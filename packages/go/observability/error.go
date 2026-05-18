package observability

import (
	"fmt"
	"log/slog"
	"reflect"
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

func errorType(err error) string {
	if err == nil {
		return ""
	}
	errType := reflect.TypeOf(err)
	for errType.Kind() == reflect.Pointer {
		errType = errType.Elem()
	}
	if errType.PkgPath() == "" {
		return errType.Name()
	}
	return errType.PkgPath() + "." + errType.Name()
}
