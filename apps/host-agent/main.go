package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type powerRequest struct {
	Action string `json:"action"`
}

func main() {
	token := strings.TrimSpace(os.Getenv("HOST_AGENT_TOKEN"))
	if token == "" {
		log.Fatal("HOST_AGENT_TOKEN is required")
	}

	port := strings.TrimSpace(os.Getenv("HOST_AGENT_PORT"))
	if port == "" {
		port = "8765"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /api/power", func(w http.ResponseWriter, r *http.Request) {
		if !validBearerToken(r, token) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "Unauthorized"})
			return
		}

		var data powerRequest
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid JSON"})
			return
		}

		if data.Action != "shutdown" && data.Action != "sleep" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid power action"})
			return
		}

		if err := runPowerCommand(r.Context(), data.Action); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"message": "Failed to run power command",
				"error":   err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusAccepted, map[string]string{"message": "Power action accepted"})
	})

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("host-agent listening on :%s", port)
	log.Fatal(server.ListenAndServe())
}

func validBearerToken(r *http.Request, expectedToken string) bool {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return false
	}

	actualToken := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	return subtle.ConstantTimeCompare([]byte(actualToken), []byte(expectedToken)) == 1
}

func runPowerCommand(_ context.Context, action string) error {
	commandText := commandFromEnv(action)
	if commandText == "" {
		return errors.New("no command configured for " + action)
	}

	command := shellCommand(commandText)
	return command.Start()
}

func shellCommand(commandText string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd", "/C", commandText)
	}

	return exec.Command("sh", "-c", commandText)
}

func commandFromEnv(action string) string {
	switch action {
	case "shutdown":
		if command := strings.TrimSpace(os.Getenv("HOST_AGENT_SHUTDOWN_CMD")); command != "" {
			return command
		}
	case "sleep":
		if command := strings.TrimSpace(os.Getenv("HOST_AGENT_SLEEP_CMD")); command != "" {
			return command
		}
	}

	switch runtime.GOOS {
	case "windows":
		if action == "shutdown" {
			return "shutdown /s /t 0"
		}
		if action == "sleep" {
			return "rundll32.exe powrprof.dll,SetSuspendState 0,1,0"
		}
	case "linux":
		if action == "shutdown" {
			return "systemctl poweroff"
		}
		if action == "sleep" {
			return "systemctl suspend"
		}
	case "darwin":
		if action == "shutdown" {
			return "osascript -e tell app System Events to shut down"
		}
		if action == "sleep" {
			return "pmset sleepnow"
		}
	}

	return ""
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
