package erzo

type options struct {
	output   string
	truncate bool
	//metadata bool
}

type Option interface {
	apply(*options)
}

type truncateOption bool

func (opt truncateOption) apply(opts *options) {
	opts.truncate = bool(opt)
}

func OptionTruncate(b bool) Option {
	return truncateOption(b)
}

type outputOption string

func (opt outputOption) apply(opts *options) {
	opts.output = string(opt)
}

func OptionOutput(s string) Option {
	return outputOption(s)
}

//type metadataOption bool
//
//func (opt metadataOption) apply(opts *options) {
//	opts.metadata = bool(opt)
//}
//
//func OptionMetadata(b bool) Option {
//	return metadataOption(b)
//}
