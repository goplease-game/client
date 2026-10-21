//go:build !js

package user

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func sessionFilePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "goplease", "session.json"), nil
}

type sessionFile struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

func loadPersistedSession() (token, username string, ok bool) {
	path, err := sessionFilePath()
	if err != nil {
		return "", "", false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", false
	}

	var sf sessionFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return "", "", false
	}

	return sf.Token, sf.Username, sf.Token != ""
}

func persistSession(token, username string) {
	path, err := sessionFilePath()
	if err != nil {
		return
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return
	}

	data, _ := json.Marshal(sessionFile{Token: token, Username: username})
	_ = os.WriteFile(path, data, 0600)
}

func clearPersistedSession() {
	path, err := sessionFilePath()
	if err != nil {
		return
	}
	_ = os.Remove(path)
}
