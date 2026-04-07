//go:build !windows

package elevate

import (
	"errors"
)

func Elevate() error {
	return errors.New("elevation not supported on this platform")
}