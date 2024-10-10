package winrm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/masterzen/winrm"
	"github.com/scorify/schema"
)

type Schema struct {
	Server         string `key:"server"`
	Port           int    `key:"port" default:"5985"`
	Username       string `key:"username"`
	Password       string `key:"password"`
	Command        string `key:"command"`
	ExpectedOutput string `key:"expected_output"`
	HTTPS          bool   `key:"https"`
	Insecure       bool   `key:"insecure"`
}

func Validate(config string) error {
	conf := Schema{}

	err := schema.Unmarshal([]byte(config), &conf)
	if err != nil {
		return err
	}

	if conf.Server == "" {
		return fmt.Errorf("server is required, got %q", schema.Server)
	}

	if conf.Port >= 65536 || schema.Port <= 0 {
		return fmt.Errorf("valid port is required, got %d", schema.Port)
	}

	if conf.Username == "" {
		return fmt.Errorf("username is required; got %q", schema.Username)
	}

	if conf.Command == "" {
		return fmt.Errorf("command is required; got %q", schema.Command)
	}

	if conf.ExpectedOutput == "" {
		return fmt.Errorf("expected_output is required; got %q", schema.ExpectedOutput)
	}

	return nil
}

func Run(ctx context.Context, config string) error {
	schema := Schema{}

	err := json.Unmarshal([]byte(config), &schema)
	if err != nil {
		return err
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		return fmt.Errorf("failed to get context deadline")
	}

	timeout := time.Until(deadline)

	endpoint := winrm.NewEndpoint(
		schema.Target,
		schema.Port,
		schema.HTTPS,
		schema.Insecure,
		[]byte{},
		[]byte{},
		[]byte{},
		timeout,
	)

	client, err := winrm.NewClient(endpoint, schema.Username, schema.Password)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	stdout, stderr, _, err := client.RunCmdWithContext(ctx, schema.Command)
	if err != nil {
		return fmt.Errorf("failed to run command: %v", err)
	}

	if stderr != "" {
		return fmt.Errorf("command returned error: %s", stderr)
	}

	if strings.TrimSpace(stdout) != strings.TrimSpace(schema.ExpectedOutput) {
		return fmt.Errorf("expected output does not match actual output: %q != %q", schema.ExpectedOutput, stdout)
	}

	return nil
}
