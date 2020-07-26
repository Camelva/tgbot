package engine

import (
	"fmt"
	"github.com/camelva/erzo/parsers"
)

type ErrNotURL struct{}

func (ErrNotURL) Error() string {
	return "there is no valid url"
}

type ErrUndefined struct{}

func (ErrUndefined) Error() string {
	return "undefined error"
}

// parsers errors
type ErrUnsupportedService struct {
	Service string
}

func (e ErrUnsupportedService) Error() string {
	return fmt.Sprintf("%s unsupported yet", e.Service)
}

type ErrUnsupportedType struct {
	parsers.ErrFormatNotSupported
}

type ErrCantFetchInfo struct {
	parsers.ErrCantContinue
}

// loaders errors
type ErrUnsupportedProtocol struct {
	Protocol string
}

func (ErrUnsupportedProtocol) Error() string {
	return "current loaders don't work with this protocol"
}

type ErrDownloadingError struct {
	Reason string
}

func (e ErrDownloadingError) Error() string {
	return fmt.Sprintf("can't download this song: %s", e.Reason)
}
