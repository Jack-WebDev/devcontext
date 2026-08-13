package cli

import (
	"errors"
	"os"
)

func isPermissionError(err error) bool {
	return os.IsPermission(err) || errors.Is(err, os.ErrPermission)
}
