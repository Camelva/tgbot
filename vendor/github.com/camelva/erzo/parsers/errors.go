package parsers

import "fmt"

type ErrFormatNotSupported struct {
	Format string
}

func (e ErrFormatNotSupported) Error() string {
	return fmt.Sprintf("format %s not supported yet", e.Format)
}

type ErrCantContinue struct {
	Reason string
}

func (e ErrCantContinue) Error() string {
	return fmt.Sprintf("error %s interrupted process", e.Reason)
}
