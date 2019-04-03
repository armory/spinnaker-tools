package utils

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

func RunCommand(command string, args ...string) (*bytes.Buffer, *bytes.Buffer, error) {
	cmd := exec.Command(command, args...)
	out := &bytes.Buffer{}
	serr := &bytes.Buffer{}
	cmd.Stdout = out
	cmd.Stderr = serr
	err := cmd.Run()
	if err != nil {
		return nil, serr, err
	}
	return out, serr, nil
}

func RunCommandInput(command string, stdin string, args ...string) error {
	c := exec.Command(command, args...)
	s, err := c.StdinPipe()
	if err != nil {
		return err
	}
	c.Stderr = os.Stderr
	err = c.Start()
	if err != nil {
		return err
	}
	_, err = io.WriteString(s, stdin)
	if err != nil {
		return err
	}
	s.Close()
	return c.Wait()
}
