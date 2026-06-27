// Package integration tests the full pipeline.
package integration

import "encoding/json"

// jsonUnmarshal wraps json.Unmarshal for test use.
func jsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
