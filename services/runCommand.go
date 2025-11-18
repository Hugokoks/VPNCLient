package services

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func RunCommand(ctx context.Context, command []string) error {

	if len(command) == 0 {

		return fmt.Errorf("Error: Field command is empty")

	}

	cmdName := command[0]
	cmdArg := command[1:]

	cmd := exec.CommandContext(ctx, cmdName, cmdArg...)

	output, err := cmd.Output()

	if err != nil {

		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr != nil {
			// Vypíšeme stderr, pokud je k dispozici
			return fmt.Errorf("Command '%s' failed with error: %s. Output: %s",
				strings.Join(command, " "), exitErr.Error(), string(exitErr.Stderr))
		}
		return fmt.Errorf("Command '%s' failed: %w", strings.Join(command, " "), err)

	}

	fmt.Printf("Command '%s' successfully done. Output:\n%s\n", strings.Join(command, " "), string(output))

	return nil
}
