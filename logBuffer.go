package ginlogrus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/mitchellh/copystructure"
)

// LogBuffer - implement io.Writer inferface to append to a string
type LogBuffer struct {
	Buff      strings.Builder
	header    map[string]interface{}
	headerMU  *sync.RWMutex
	AddBanner bool
}

// NewLogBuffer - create a LogBuffer and initialize it
func NewLogBuffer(opt ...LogBufferOption) LogBuffer {
	opts := defaultLogBufferOptions
	dup, _ := copystructure.Copy(defaultLogBufferOptions)
	opts = dup.(logBufferOptions)
	if opts.withHeaders == nil {
		opts.withHeaders = make(map[string]interface{})
	}
	for _, o := range opt {
		o(&opts)
	}
	return LogBuffer{
		header:    opts.withHeaders,
		headerMU:  &sync.RWMutex{},
		AddBanner: opts.addBanner}
}

// StoreHeader - store a header
func (b *LogBuffer) StoreHeader(k string, v interface{}) {
	b.headerMU.Lock()
	b.header[k] = v
	b.headerMU.Unlock()
}

// DeleteHeader - delete a header
func (b *LogBuffer) DeleteHeader(k string) {
	b.headerMU.Lock()
	delete(b.header, k)
	b.headerMU.Unlock()
}

// GetHeader - get a header
func (b *LogBuffer) GetHeader(k string) (interface{}, bool) {
	b.headerMU.RLock()
	r, ok := b.header[k]
	b.headerMU.RUnlock()
	return r, ok
}

// GetAllHeaders - return all the headers
func (b *LogBuffer) GetAllHeaders() (map[string]interface{}, error) {
	b.headerMU.RLock()
	dup, err := copystructure.Copy(b.header)
	b.headerMU.RUnlock()
	if err != nil {
		return nil, err
	}
	return dup.(map[string]interface{}), nil
}

// CopyHeader - copy a header
func CopyHeader(dst *LogBuffer, src *LogBuffer) {
	src.headerMU.Lock()
	dup, err := copystructure.Copy(src.header)
	dupBanner := src.AddBanner
	src.headerMU.Unlock()

	dst.headerMU.Lock()
	if err != nil {
		dst.header = map[string]interface{}{}
	} else {
		dst.header = dup.(map[string]interface{})
	}
	dst.AddBanner = dupBanner
	dst.headerMU.Unlock()
}

// Write - simply append to the strings.Buffer but add a comma too
func (b *LogBuffer) Write(data []byte) (n int, err error) {
	newData := bytes.TrimSuffix(data, []byte("\n"))
	return b.Buff.Write(append(newData, []byte(",")...))
}

// String - output the strings.Builder as one aggregated JSON object
func (b *LogBuffer) String() string {
	var str strings.Builder
	str.WriteString("{")
	if b.header != nil && len(b.header) != 0 {
		b.headerMU.RLock()
		hdr, err := json.Marshal(b.header)
		b.headerMU.RUnlock()
		if err != nil {
			fmt.Println("Error encoding logBuffer JSON")
		}
		str.Write(hdr[1 : len(hdr)-1])
		str.WriteString(",")
	}
	str.WriteString("\"entries\":[" + strings.TrimSuffix(b.Buff.String(), ",") + "]")
	if b.AddBanner {
		str.WriteString(",\"banner\":\"[GIN] --------------------------------------------------------------- GinLogrusWithTracing ----------------------------------------------------------------\"")
	}
	str.WriteString("}\n")
	return str.String()
}
