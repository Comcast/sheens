package testutil

import (
	"reflect"
	"testing"
)

type Person struct {
	Name string
	Age  int
}

func TestJS(t *testing.T) {
	tests := []struct {
		name string
		arg  interface{}
		want string
	}{
		{
			name: "simple struct",
			arg:  Person{"John Doe", 30},
			want: `{"Name":"John Doe","Age":30}`,
		},
		{
			name: "nested struct",
			arg: struct {
				Person Person
				ID     int
			}{Person{"Jane Doe", 25}, 1},
			want: `{"Person":{"Name":"Jane Doe","Age":25},"ID":1}`,
		},
		// It's difficult to test error handling since json.Marshal doesn't easily fail on simple types.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JS(tt.arg); got != tt.want {
				t.Errorf("JS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDwimjs(t *testing.T) {
	tests := []struct {
		name string
		arg  interface{}
		want interface{}
	}{
		{
			name: "valid JSON string",
			arg:  `{"name":"John Doe","age":30}`,
			want: map[string]interface{}{"name": "John Doe", "age": float64(30)}, // json.Unmarshal uses float64 for numbers
		},
		{
			name: "valid JSON bytes",
			arg:  []byte(`{"name":"Jane Doe","age":25}`),
			want: map[string]interface{}{"name": "Jane Doe", "age": float64(25)},
		},
		{
			name: "non-JSON string",
			arg:  "hello world",
			want: "hello world",
		},
		{
			name: "non-string, non-byte-slice type",
			arg:  12345,
			want: 12345,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Dwimjs(tt.arg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Dwimjs() = %v, want %v", got, tt.want)
			}
		})
	}
}
