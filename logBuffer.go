package ginlogrus

import (
	"bytes"
	"strings"
)

// StringWriter - implement io.Writer inferface to append to a string
type logBuffer struct {
	Buff strings.Builder
}

// Write - simply append to the strings.Buffer but add a comma too
func (b *logBuffer) Write(data []byte) (n int, err error) {
	newData := bytes.TrimSuffix(data, []byte("\n"))
	return b.Buff.Write(append(newData, []byte(",")...))
}

// String - output the strings.Builder as one aggregated JSON object
func (b *logBuffer) String() string {
	return "{\"entries\":[" + strings.TrimSuffix(b.Buff.String(), ",") + "]}\n"
}
