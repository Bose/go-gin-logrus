package ginlogrus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// LogBuffer - implement io.Writer inferface to append to a string
type LogBuffer struct {
	Buff      strings.Builder
	Header    map[string]interface{}
	AddBanner bool
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
	if b.Header != nil {
		hdr, err := json.Marshal(b.Header)
		if err != nil {
			fmt.Println("Error encoding logBuffer JSON")
		}
		if len(hdr) > 0 {
			str.Write(hdr[1 : len(hdr)-1])
			str.WriteString(",")
		}
	}
	str.WriteString("\"entries\":[" + strings.TrimSuffix(b.Buff.String(), ",") + "]")
	if b.AddBanner {
		str.WriteString(",\"banner\":\"[GIN] --------------------------------------------------------------- GinLogrusWithTracing ----------------------------------------------------------------\"")
	}
	str.WriteString("}\n")
	return str.String()
}
