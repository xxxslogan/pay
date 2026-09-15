package internal

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

func CORS(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func CheckAPIToken(r *http.Request) bool {
	token := strings.TrimSpace(os.Getenv("API_TOKEN"))
	if token == "" {
		return false
	}
	got := strings.TrimSpace(r.Header.Get("Authorization"))
	return got == "Bearer "+token
}

func PackageOK(id string) bool {
	switch id {
	case "rimix-perm-399", "ace-perm-399":
		return true
	default:
		return false
	}
}

func PackageMoney() string {
	m := strings.TrimSpace(os.Getenv("PACKAGE_PRICE"))
	if m == "" {
		return "399.00"
	}
	return m
}

func PackageName() string {
	n := strings.TrimSpace(os.Getenv("PACKAGE_NAME"))
	if n == "" {
		return "RimixFac永久授权"
	}
	return n
}
