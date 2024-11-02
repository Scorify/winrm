package winrm

import (
	"context"
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
		return fmt.Errorf("server is required, got %q", conf.Server)
	}

	if conf.Port >= 65536 || conf.Port <= 0 {
		return fmt.Errorf("valid port is required, got %d", conf.Port)
	}

	if conf.Username == "" {
		return fmt.Errorf("username is required; got %q", conf.Username)
	}

	if conf.Command == "" {
		return fmt.Errorf("command is required; got %q", conf.Command)
	}

	if conf.ExpectedOutput == "" {
		return fmt.Errorf("expected_output is required; got %q", conf.ExpectedOutput)
	}

	return nil
}

func Run(ctx context.Context, config string) error {
	conf := Schema{}

	err := schema.Unmarshal([]byte(config), &conf)
	if err != nil {
		return err
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		return fmt.Errorf("failed to get context deadline")
	}
	timeout := time.Until(deadline)
	errChan := make(chan error, 1)

	go func() {
		endpoint := winrm.NewEndpoint(
			conf.Server,
			conf.Port,
			conf.HTTPS,
			conf.Insecure,
			[]byte{},
			[]byte{},
			[]byte{},
			timeout,
		)

		defer close(errChan)
		client, err := winrm.NewClient(endpoint, conf.Username, conf.Password)
		if err != nil {
			errChan <- fmt.Errorf("failed to create client: %v", err)
			return
		}

		stdout, stderr, _, err := client.RunCmdWithContext(ctx, conf.Command)
		if err != nil {
			errChan <- fmt.Errorf("failed to run command: %v", err)
			return
		}

		if stderr != "" {
			errChan <- fmt.Errorf("command returned error: %s", stderr)
			return
		}

		if strings.TrimSpace(stdout) != strings.TrimSpace(conf.ExpectedOutput) {
			errChan <- fmt.Errorf("expected output does not match actual output: %q != %q", conf.ExpectedOutput, stdout)
			return
		}

		errChan <- nil
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}
