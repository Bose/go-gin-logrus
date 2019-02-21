package ginlogrus

// LogBufferOption - define options for LogBuffer
type LogBufferOption func(*logBufferOptions)
type logBufferOptions struct {
	addBanner   bool
	withHeaders map[string]interface{}
}

// defaultBufferOptions - some defs options to logBuffer
var defaultLogBufferOptions = logBufferOptions{
	addBanner:   false,
	withHeaders: make(map[string]interface{}),
}

// WithBanner - define an Option func for passing in an optional add Banner
func WithBanner(a bool) LogBufferOption {
	return func(o *logBufferOptions) {
		o.addBanner = a
	}
}

// WithHeader - define an Option func for passing in a set of optional header
func WithHeader(k string, v interface{}) LogBufferOption {
	return func(o *logBufferOptions) {
		o.withHeaders[k] = v
	}
}
