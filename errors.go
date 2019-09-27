package ip2loc

import (
	"fmt"
)

type ErrInvalidIP struct{}

func (e ErrInvalidIP) Error() string {
	return "Invalid IP Address"
}

type ErrUnsupportedFormat struct {
	badFormat uint8
}

func (e ErrUnsupportedFormat) Error() string {
	return fmt.Sprintf(
		"Unsupported database format; expected 1, found %d",
		e.badFormat,
	)
}

type ErrNoResults struct{}

func (e ErrNoResults) Error() string {
	return "No results found"
}
