package common

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func GetEnvOrDefault(env string, defaultValue int) int {
	if env == "" || os.Getenv(env) == "" {
		return defaultValue
	}
	num, err := strconv.Atoi(os.Getenv(env))
	if err != nil {
		SysError(fmt.Sprintf("failed to parse %s: %s, using default value: %d", env, err.Error(), defaultValue))
		return defaultValue
	}
	return num
}

func GetEnvOrDefaultString(env string, defaultValue string) string {
	if env == "" || os.Getenv(env) == "" {
		return defaultValue
	}
	return os.Getenv(env)
}

func GetEnvOrDefaultBool(env string, defaultValue bool) bool {
	if env == "" || os.Getenv(env) == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(os.Getenv(env))
	if err != nil {
		SysError(fmt.Sprintf("failed to parse %s: %s, using default value: %t", env, err.Error(), defaultValue))
		return defaultValue
	}
	return b
}

func GetEnvOrDefaultFloat64(env string, defaultValue float64) float64 {
	if env == "" || os.Getenv(env) == "" {
		return defaultValue
	}
	num, err := strconv.ParseFloat(os.Getenv(env), 64)
	if err != nil {
		SysError(fmt.Sprintf("failed to parse %s: %s, using default value: %f", env, err.Error(), defaultValue))
		return defaultValue
	}
	return num
}

func GetEnvOrDefaultIntMap(env string, defaultValue map[int]int) map[int]int {
	value := strings.TrimSpace(os.Getenv(env))
	if env == "" || value == "" {
		return cloneIntMap(defaultValue)
	}
	raw := map[string]int{}
	if err := json.Unmarshal([]byte(value), &raw); err != nil {
		SysError(fmt.Sprintf("failed to parse %s: %s, using default value", env, err.Error()))
		return cloneIntMap(defaultValue)
	}
	result := make(map[int]int, len(raw))
	for key, val := range raw {
		intKey, err := strconv.Atoi(strings.TrimSpace(key))
		if err != nil {
			SysError(fmt.Sprintf("failed to parse %s key %q: %s, using default value", env, key, err.Error()))
			return cloneIntMap(defaultValue)
		}
		result[intKey] = val
	}
	return result
}

func cloneIntMap(src map[int]int) map[int]int {
	dst := make(map[int]int, len(src))
	for key, val := range src {
		dst[key] = val
	}
	return dst
}
