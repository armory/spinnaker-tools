package utils

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"fmt"
)

func RunCommand(verbose bool, command string, args ...string) (*bytes.Buffer, *bytes.Buffer, error) {
	if verbose {
		fmt.Println(command)
		fmt.Println(args)
	}
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

// Need better passback here
func RunCommandInput(verbose bool, command string, stdin string, args ...string) error {
	if verbose {
		fmt.Println(command)
		fmt.Println(args)
	}
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
