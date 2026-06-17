package pgxx

import "encoding/json"

// MarshalJSONBSlice encodes a slice for a nullable jsonb column: an empty/nil
// slice marshals to nil so the column stores SQL NULL rather than "[]".
func MarshalJSONBSlice[T any](v []T) ([]byte, error) {
	if len(v) == 0 {
		return nil, nil
	}
	return json.Marshal(v)
}

// UnmarshalJSONBSlice decodes a nullable jsonb column into a slice. Empty/nil
// input yields a nil slice (no error).
func UnmarshalJSONBSlice[T any](b []byte) ([]T, error) {
	if len(b) == 0 {
		return nil, nil
	}
	var out []T
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}
