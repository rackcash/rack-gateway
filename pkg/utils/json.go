package utils

import "encoding/json"

func Unmarshal[T any](data []byte) (*T, error) {
	var unm T
	err := json.Unmarshal(data, &unm)
	if err != nil {
		return nil, err
	}
	return &unm, nil
}

// use only if u know that data is valid
func MustMarshal(v any) []byte {
	m, _ := json.Marshal(v)
	return m
}
