package utils

import (
	"errors"
	"strings"
)

// GetValueFromMap() - returns a value from a map via dot-notation
func GetValueFromMap(m map[string]interface{}, key string, defaultValue interface{}) (interface{}, error) {
	// split dot-notation
	keys := strings.Split(key, ".")

	var current interface{} = m
	for _, k := range keys {
		if current == nil {
			return defaultValue, nil // default value
		}

		// ensure the current value is a map[string]interface{}
		if currentMap, ok := current.(map[string]interface{}); ok {
			if val, exists := currentMap[k]; exists {
				current = val
			} else {
				return defaultValue, nil // default value
			}
		} else {
			return nil, errors.New("not type of map[string]interface{}")
		}
	}
	return current, nil
}
