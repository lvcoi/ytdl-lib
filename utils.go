package youtube

import (
	"encoding/base64"
	"encoding/json"
)

type chunk struct {
	start int64
	end   int64
	data  chan []byte
}

func getChunks(totalSize, chunkSize int64) []chunk {
	var chunks []chunk

	for start := int64(0); start < totalSize; start += chunkSize {
		end := chunkSize + start - 1
		if end > totalSize-1 {
			end = totalSize - 1
		}

		chunks = append(chunks, chunk{start, end, make(chan []byte, 1)})
	}

	return chunks
}

// jsonGet navigates a json.RawMessage by object keys.
// Returns nil if the path doesn't exist or the data isn't an object.
func jsonGet(data json.RawMessage, keys ...string) json.RawMessage {
	current := data
	for _, key := range keys {
		var obj map[string]json.RawMessage
		if json.Unmarshal(current, &obj) != nil {
			return nil
		}
		val, ok := obj[key]
		if !ok {
			return nil
		}
		current = val
	}
	return current
}

// jsonGetIndex gets an element from a JSON array by index.
// Returns nil if the data isn't an array or the index is out of bounds.
func jsonGetIndex(data json.RawMessage, idx int) json.RawMessage {
	var arr []json.RawMessage
	if json.Unmarshal(data, &arr) != nil || idx < 0 || idx >= len(arr) {
		return nil
	}
	return arr[idx]
}

// jsonString extracts a string from a json.RawMessage.
// Returns empty string if the data isn't a JSON string.
func jsonString(data json.RawMessage) string {
	var s string
	if json.Unmarshal(data, &s) != nil {
		return ""
	}
	return s
}

// jsonHasContent returns true if the data is non-nil and represents
// something more substantial than null, empty string, empty array, or empty object.
func jsonHasContent(data json.RawMessage) bool {
	if len(data) <= 4 {
		return false
	}
	return true
}

// jsonFirstKey returns the value of the first key in a JSON object.
// Returns the original data if it's not an object.
func jsonFirstKey(data json.RawMessage) json.RawMessage {
	var obj map[string]json.RawMessage
	if json.Unmarshal(data, &obj) != nil {
		return data
	}
	for _, v := range obj {
		return v
	}
	return data
}

// jsonGetText tries to extract text from a YouTube JSON node.
// It traverses the given paths, then looks for a plain string, "text" field, or "runs" array.
func jsonGetText(data json.RawMessage, paths ...string) string {
	current := data
	for _, path := range paths {
		if next := jsonGet(current, path); jsonHasContent(next) {
			current = next
		}
	}

	// Try direct string
	if s := jsonString(current); s != "" {
		return s
	}

	// Try .text
	if textNode := jsonGet(current, "text"); jsonHasContent(textNode) {
		return jsonString(textNode)
	}

	// Try .runs[].text
	if runsNode := jsonGet(current, "runs"); jsonHasContent(runsNode) {
		var runs []json.RawMessage
		if json.Unmarshal(runsNode, &runs) == nil {
			var text string
			for _, run := range runs {
				if t := jsonGet(run, "text"); jsonHasContent(t) {
					text += jsonString(t)
				}
			}
			return text
		}
	}

	return ""
}

// jsonGetContinuation extracts the continuation token from a YouTube JSON node.
func jsonGetContinuation(data json.RawMessage) string {
	return jsonString(jsonGet(jsonGetIndex(jsonGet(data, "continuations"), 0), "nextContinuationData", "continuation"))
}

func base64PadEnc(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func base64Enc(str string) string {
	return base64.RawStdEncoding.EncodeToString([]byte(str))
}
