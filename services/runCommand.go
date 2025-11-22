package services

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func RunCommand(ctx context.Context, command []string) error {

	if len(command) == 0 {
		return fmt.Errorf("command is empty")
	}

	cmd := exec.CommandContext(ctx,  command[0], command[1:]...)
	
	out, err := cmd.CombinedOutput() 
	outStr := strings.TrimSpace(string(out))

	if err != nil {

		return fmt.Errorf("\ncommand: '%s'\noutput: %s", strings.Join(command, " "), outStr)
	}

	fmt.Printf("\nCommand: '%s' successfully done.\n Output:\n%s\n", strings.Join(command, " "), string(outStr))

	return nil
}

func RunCommandWithOutput(ctx context.Context, command []string) (string, error) {
    if len(command) == 0 {
        return "", fmt.Errorf("command is empty")
    }

    c := exec.CommandContext(ctx, command[0], command[1:]...)

    out, err := c.CombinedOutput()
    if err != nil {
        return string(out), fmt.Errorf("%s failed: %w", strings.Join(command, " "), err)
    }

    return string(out), nil
}