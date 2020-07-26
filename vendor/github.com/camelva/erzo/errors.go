package erzo

import "github.com/camelva/erzo/engine"

type ErrNotURL struct {
	engine.ErrNotURL
}
type ErrUnsupportedService struct {
	engine.ErrUnsupportedService
}
type ErrUnsupportedType struct {
	engine.ErrUnsupportedType
}
type ErrCantFetchInfo struct {
	engine.ErrCantFetchInfo
}
type ErrUnsupportedProtocol struct {
	engine.ErrUnsupportedProtocol
}
type ErrDownloadingError struct {
	engine.ErrDownloadingError
}
type ErrUndefined struct {
	engine.ErrUndefined
}
