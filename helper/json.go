package helper

import (
	"encoding/json"
	"fmt"
	"os"
)

func ReadJSON(path string, v interface{}) error {
	fl, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file failed: %w", err)
	}
	defer fl.Close()

	dec := json.NewDecoder(fl)

	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("decode JSON content failed: %w", err)
	}

	return nil
}

func WriteJSON(path string, v interface{}) error {
	fl, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file failed: %w", err)
	}
	defer fl.Close()

	enc := json.NewEncoder(fl)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encoding the JSON content failed: %w", err)
	}

	return nil
}
