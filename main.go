package winrm

import (
	"context"
	"encoding/json"
)

type Schema struct {
	Target         string `json:"target"`
	Port           int    `json:"port"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	Command        string `json:"command"`
	ExpectedOutput string `json:"expected_output"`
	HTTPS          bool   `json:"https"`
	Insecure       bool   `json:"insecure"`
}

func Run(ctx context.Context, config string) error {
	schema := Schema{}

	err := json.Unmarshal([]byte(config), &schema)
	if err != nil {
		return err
	}

	return nil
}
