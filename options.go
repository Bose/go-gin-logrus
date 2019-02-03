package ginlogrus

// Option - define options for WithTracing()
type Option func(*options)
type options struct {
	aggregateLogging bool
}

// defaultOptions - some defs options to NewJWTCache()
var defaultOptions = options{
	aggregateLogging: false,
}

// WithAggregateLogging - define an Option func for passing in an optional aggregateLogging
func WithAggregateLogging(a bool) Option {
	return func(o *options) {
		o.aggregateLogging = a
	}
}
