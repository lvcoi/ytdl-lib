package youtube

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetChunks1(t *testing.T) {
	require := require.New(t)
	chunks := getChunks(13, 5)

	require.Len(chunks, 3)
	require.EqualValues(0, chunks[0].start)
	require.EqualValues(4, chunks[0].end)
	require.EqualValues(5, chunks[1].start)
	require.EqualValues(9, chunks[1].end)
	require.EqualValues(10, chunks[2].start)
	require.EqualValues(12, chunks[2].end)
}

func TestGetChunks_length(t *testing.T) {
	require := require.New(t)
	require.Len(getChunks(10, 9), 2)
	require.Len(getChunks(10, 10), 1)
	require.Len(getChunks(10, 11), 1)
}

func TestJsonHasContent(t *testing.T) {
	tests := []struct {
		name     string
		input    json.RawMessage
		expected bool
	}{
		{"nil", nil, false},
		{"null", json.RawMessage(`null`), false},
		{"empty string", json.RawMessage(`""`), false},
		{"empty array", json.RawMessage(`[]`), false},
		{"empty object", json.RawMessage(`{}`), false},
		{"true", json.RawMessage(`true`), true},
		{"false", json.RawMessage(`false`), true},
		{"short string", json.RawMessage(`"ab"`), true},
		{"number", json.RawMessage(`1234`), true},
		{"non-empty object", json.RawMessage(`{"a":1}`), true},
		{"non-empty array", json.RawMessage(`[1]`), true},
		{"longer string", json.RawMessage(`"hello"`), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, jsonHasContent(tt.input))
		})
	}
}

func TestJsonGetText_AlternativePaths(t *testing.T) {
	// Simulates YouTube metadata where "description" and "descriptionText"
	// are alternative field names at the same level.
	data := json.RawMessage(`{
		"description": {"text": "short desc"},
		"descriptionText": {"text": "full description"}
	}`)

	// Should pick the last matching alternative, not chain into description.descriptionText.
	result := jsonGetText(data, "description", "descriptionText")
	assert.Equal(t, "full description", result)
}

func TestJsonGetText_SinglePath(t *testing.T) {
	data := json.RawMessage(`{"title": "My Playlist"}`)
	assert.Equal(t, "My Playlist", jsonGetText(data, "title"))
}

func TestJsonGetText_RunsArray(t *testing.T) {
	data := json.RawMessage(`{"runs": [{"text": "hello "}, {"text": "world"}]}`)
	assert.Equal(t, "hello world", jsonGetText(data))
}
