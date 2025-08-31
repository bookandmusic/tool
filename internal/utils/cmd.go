package utils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/bookandmusic/tool/internal/logger"
)

func handlePipeOutput(r io.Reader, console logger.Logger) {
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			console.Debug(string(buf[:n]))
		}
		if err != nil {
			break
		}
	}
}

func buildEnv(env map[string]string) []string {
	out := os.Environ()
	for k, v := range env {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}
	return out
}

func RunCommand(console logger.Logger, sudo bool, env map[string]string, workdir string, args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command provided")
	}
	cmdStr := strings.Join(args, " ")
	console.Debug(fmt.Sprintf("[COMMAND] Executing command: %s\n", cmdStr))
	console.Debug(fmt.Sprintf("[COMMAND] Working directory: %s\n", workdir))
	if env == nil {
		console.Debug("[COMMAND] Environment variables: <nil>\n")
	} else {
		var envStrs []string
		for k, v := range env {
			envStrs = append(envStrs, fmt.Sprintf("%s=%s", k, v))
		}
		console.Debug(fmt.Sprintf("[COMMAND] Environment variables: %s\n", strings.Join(envStrs, " ")))
	}

	cmdName := args[0]
	cmdArgs := args[1:]

	if sudo {
		cmdArgs = append([]string{cmdName}, cmdArgs...)
		cmdName = "sudo"
		console.Debug("[COMMAND] Running with sudo\n")
	}

	cmd := exec.Command(cmdName, cmdArgs...) // #nosec G204
	cmd.Stdin = os.Stdin
	if workdir != "" {
		cmd.Dir = workdir
	}
	if env != nil {
		cmd.Env = buildEnv(env)
	}

	r, w := io.Pipe()
	cmd.Stdout = w
	cmd.Stderr = w

	done := make(chan struct{})
	go func() {
		handlePipeOutput(r, console)
		close(done)
	}()

	err := cmd.Run()
	if cerr := w.Close(); cerr != nil {
		console.Warning("[COMMAND] Failed to close pipe: %v\n", cerr)
	}
	<-done
	return err
}
