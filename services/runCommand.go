package services

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func RunCommand(ctx context.Context, command ...string) (string, error) {
	if len(command) == 0 {
		return "", fmt.Errorf("command is empty")
	}

	cmd := exec.CommandContext(ctx, command[0], command[1:]...)

	out, err := cmd.CombinedOutput()
	outStr := strings.TrimSpace(string(out))

	if err != nil {
		return outStr, fmt.Errorf(
			"command failed: %s\noutput:\n%s",
			strings.Join(command, " "),
			outStr,
		)
	}

	return outStr, nil
}