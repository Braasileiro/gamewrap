package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

func newLogger() (*log.Logger, func()) {
	if value, configured := os.LookupEnv("GW_LOG"); configured && loggingDisabled(value) {
		return discardLogger(), func() {}
	}

	path := os.Getenv("GW_LOG_FILE")
	if path == "" {
		state := os.Getenv("XDG_STATE_HOME")
		if state == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return discardLogger(), func() {}
			}

			state = filepath.Join(home, ".local", "state")
		}

		path = filepath.Join(state, "gamewrap", "gamewrap.log")
	}

	// Logging failures must never prevent the game from starting.
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return discardLogger(), func() {}
	}

	if info, err := os.Stat(path); err == nil && info.Size() >= 1<<20 {
		_ = os.Rename(path, path+".1")
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return discardLogger(), func() {}
	}

	return log.New(file, "", log.Ldate|log.Ltime|log.Lmsgprefix), func() {
		_ = file.Close()
	}
}

func loggingDisabled(value string) bool {
	return value == "0"
}

func discardLogger() *log.Logger {
	return log.New(io.Discard, "", 0)
}
