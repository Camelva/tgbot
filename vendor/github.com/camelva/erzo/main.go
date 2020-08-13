package erzo

import (
	"github.com/camelva/erzo/engine"
	_ "github.com/camelva/erzo/loaders/ffmpeg"
	_ "github.com/camelva/erzo/parsers/soundcloud"
	_ "github.com/camelva/erzo/parsers/youtube"
)

// Get process given url and download song from it.
// @message - url to process
// @options:
// 		Truncate(true|false) - clear output folder before processing
//		Output(string)		 - change output folder
// Return filename or one of the following errors:
// ErrNotURL if there is no urls in your message
// ErrUnsupportedService if url belongs to unsupported service
// ErrUnsupportedType if service supported but certain type - not yet
// ErrCantFetchInfo if fatal error occurred while extracting info from url
// ErrUnsupportedProtocol if there is no downloader for this format
// ErrDownloadingError if fatal error occurred while downloading song
// ErrUndefined any other errors
func Get(message string, opts ...Option) (*engine.SongResult, error) {
	options := options{
		output:   "out",
		truncate: false,
	}
	for _, o := range opts {
		o.apply(&options)
	}
	e := engine.New(
		options.output,
		options.truncate,
	)
	r, err := e.Process(message)
	if err != nil {
		convertedErr := convertErr(err)
		return nil, convertedErr
	}
	return r, nil
}

func convertErr(err error) error {
	var convertedErr error
	switch err.(type) {
	case engine.ErrNotURL:
		convertedErr = ErrNotURL{err.(engine.ErrNotURL)}
	case engine.ErrUnsupportedService:
		convertedErr = ErrUnsupportedService{err.(engine.ErrUnsupportedService)}
	case engine.ErrUnsupportedType:
		convertedErr = ErrUnsupportedType{err.(engine.ErrUnsupportedType)}
	case engine.ErrCantFetchInfo:
		convertedErr = ErrCantFetchInfo{err.(engine.ErrCantFetchInfo)}
	case engine.ErrUnsupportedProtocol:
		convertedErr = ErrUnsupportedProtocol{err.(engine.ErrUnsupportedProtocol)}
	case engine.ErrDownloadingError:
		convertedErr = ErrDownloadingError{err.(engine.ErrDownloadingError)}
	case engine.ErrUndefined:
		convertedErr = ErrUndefined{err.(engine.ErrUndefined)}
	default:
		convertedErr = ErrUndefined{engine.ErrUndefined{}}
	}
	return convertedErr
}
