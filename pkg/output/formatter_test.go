package output

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONFormatter(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		data     interface{}
		pretty   bool
		validate func(t *testing.T, output string)
	}{
		{
			name: "Simple Object JSON",
			data: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				Name: "John Doe",
				Age:  30,
			},
			pretty: false,
			validate: func(t *testing.T, output string) {
				var parsed map[string]interface{}
				err := json.Unmarshal([]byte(output), &parsed)
				require.NoError(t, err)
				assert.Equal(t, "John Doe", parsed["name"])
				assert.Equal(t, float64(30), parsed["age"])
			},
		},
		{
			name: "Pretty Printed JSON",
			data: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				Name: "Jane Smith",
				Age:  25,
			},
			pretty: true,
			validate: func(t *testing.T, output string) {
				var parsed map[string]interface{}
				err := json.Unmarshal([]byte(output), &parsed)
				require.NoError(t, err)
				assert.Equal(t, "Jane Smith", parsed["name"])
				assert.Equal(t, float64(25), parsed["age"])
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formatter := NewJSONFormatter(tc.pretty)
			output, err := formatter.Format(tc.data)
			require.NoError(t, err)
			tc.validate(t, output)
		})
	}
}

func TestTableFormatter(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		headers  []string
		data     [][]string
		expected string
	}{
		{
			name:    "Simple Table",
			headers: []string{"Name", "Age"},
			data: [][]string{
				{"John Doe", "30"},
				{"Jane Smith", "25"},
			},
			expected: "Name\tAge\n" +
				"--------------------------------------------------------------------------------\n" +
				"John Doe\t30\n" +
				"Jane Smith\t25\n",
		},
		{
			name:    "Empty Table",
			headers: []string{"Name", "Age"},
			data:    [][]string{},
			expected: "Name\tAge\n" +
				"--------------------------------------------------------------------------------\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formatter := NewTableFormatter(tc.headers)
			output, err := formatter.Format(tc.data)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, output)
		})
	}
}

func TestFormatterErrors(t *testing.T) {
	t.Parallel()

	t.Run("JSON Formatter with Unsupported Type", func(t *testing.T) {
		formatter := NewJSONFormatter(false)
		// Use a channel which can't be marshaled
		_, err := formatter.Format(make(chan int))
		assert.Error(t, err)
	})

	t.Run("Table Formatter with Invalid Type", func(t *testing.T) {
		formatter := NewTableFormatter([]string{"Name", "Age"})
		// Pass a string instead of [][]string
		_, err := formatter.Format("invalid data")
		assert.Error(t, err)
	})
}
