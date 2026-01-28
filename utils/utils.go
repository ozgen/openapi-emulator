package utils

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func GetEnvAsBool(key string, defaultVal bool) bool {
	val := strings.ToLower(os.Getenv(key))
	if val == "" {
		return defaultVal
	}
	return val == "1" || val == "true" || val == "yes"
}

func WriteJSON(w http.ResponseWriter, status int, obj any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	b, _ := json.Marshal(obj)
	_, _ = w.Write(b)
}
