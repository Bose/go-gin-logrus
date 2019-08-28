package ginlogrus

const DefaultBanner = "[GIN] --------------------------------------------------------------- GinLogrusWithTracing ----------------------------------------------------------------"

// LogBufferOption - define options for LogBuffer
type LogBufferOption func(*logBufferOptions)
type logBufferOptions struct {
	addBanner   bool
	withHeaders map[string]interface{}
	maxSize     uint
	banner      string
}

// DefaultLogBufferMaxSize - avg single spaced page contains 3k chars, so 100k == 33 pages which is a reasonable max
const DefaultLogBufferMaxSize = 100000

func defaultLogBufferOptions() logBufferOptions {
	return logBufferOptions{
		maxSize:   DefaultLogBufferMaxSize,
		banner:    DefaultBanner,
		addBanner: false,
	}
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
		if o.withHeaders == nil {
			o.withHeaders = make(map[string]interface{})
		}
		o.withHeaders[k] = v
	}
}

// WithMaxSize specifies the bounded max size the logBuffer can grow to
func WithMaxSize(s uint) LogBufferOption {
	return func(o *logBufferOptions) {
		o.maxSize = s
	}
}

// WithCustomBanner allows users to define their own custom banner
func WithCustomBanner(b string) LogBufferOption {
	return func(o *logBufferOptions) {
		o.banner = b
	}
}
