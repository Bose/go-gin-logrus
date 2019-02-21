package ginlogrus

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestLogBuffer_String(t *testing.T) {

	tests := []struct {
		name     string
		buff     LogBuffer
		write    []byte
		contains string
	}{
		{
			name:     "hey",
			buff:     NewLogBuffer(WithBanner(true), WithHeader("id1", "val1"), WithHeader("id2", "id2")),
			write:    []byte("\"msg\":\"hey-one\""),
			contains: "hey",
		},
		{
			name:     "hey-now",
			buff:     NewLogBuffer(WithHeader("hey", "now")),
			write:    []byte("\"msg\":\"hey-two\""),
			contains: "hey",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.buff.Write(tt.write)
			fmt.Println("buff == ", tt.buff.String())
			if !strings.Contains(tt.buff.String(), tt.contains) {
				t.Errorf("LogBuffer.String() = %v, want %v", tt.buff.String(), tt.contains)
			}
		})
	}
}

func TestNewLogBuffer(t *testing.T) {
	tests := []struct {
		name string
		opt  []LogBufferOption
		want LogBuffer
	}{
		{
			name: "one",
			opt:  []LogBufferOption{WithBanner(true), WithHeader("1", true)},
			want: LogBuffer{AddBanner: true, header: map[string]interface{}{"1": true}, headerMU: &sync.RWMutex{}},
		},
		{
			name: "two",
			opt:  []LogBufferOption{WithHeader("1", "one"), WithHeader("2", true)},
			want: LogBuffer{AddBanner: false, header: map[string]interface{}{"1": "one", "2": true}, headerMU: &sync.RWMutex{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewLogBuffer(tt.opt...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLogBuffer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogBuffer_StoreHeader_DeleteHeader_GetHeader_GetAllHeaders_CopyHeader(t *testing.T) {
	tests := []struct {
		name     string
		buff     LogBuffer
		k        string
		v        interface{}
		contains string
	}{
		{
			name:     "1",
			buff:     NewLogBuffer(WithBanner(true)),
			k:        "test-hdr",
			v:        "test-value",
			contains: "test-value",
		},
		{
			name:     "2",
			buff:     NewLogBuffer(WithBanner(true)),
			k:        "test-hdr-2",
			v:        true,
			contains: "test-hdr-2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.buff.StoreHeader(tt.k, tt.v)
			if !strings.Contains(tt.buff.String(), tt.contains) {
				t.Errorf("expected %v and got %v", tt.contains, tt.buff.String())
			}
			if got, _ := tt.buff.GetHeader(tt.k); !reflect.DeepEqual(got, tt.v) {
				t.Errorf("GetHeader() = %v, want %v", got, tt.v)
			}
			got, _ := tt.buff.GetAllHeaders()
			if reflect.DeepEqual(got, tt.v) {
				t.Errorf("GetAllHeaders() = %v, want %v", got, tt.v)
			}
			newBuff := NewLogBuffer()
			CopyHeader(&newBuff, &tt.buff)
			if reflect.DeepEqual(newBuff, tt.buff) != true {
				t.Errorf("CopyHeader() = %v, want %v", newBuff, tt.buff)
			}
			tt.buff.DeleteHeader(tt.k)
			if strings.Contains(tt.buff.String(), tt.contains) {
				t.Errorf("expected %v to be deleted and got %v", tt.contains, tt.buff.String())
			}

		})
	}
}
