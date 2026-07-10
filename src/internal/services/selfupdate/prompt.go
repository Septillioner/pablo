package selfupdate

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
	"pablo/pkg/ui"
)

var errUpdateCancelled = errors.New("update cancelled")

func isInteractiveTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

func ensureExecutableReplaceable(targetPath string) error {
	processes, err := findProcessesUsingExecutable(targetPath)
	if err != nil {
		return err
	}
	if len(processes) == 0 {
		return nil
	}

	if !isInteractiveTerminal() {
		return nil
	}

	ui.Log("!", "Pablo is in use by other processes:")
	for _, process := range processes {
		ui.Log("*", fmt.Sprintf("%d %s", process.PID, process.Name))
	}

	confirmed, err := promptCloseProcesses(os.Stdin, os.Stdout)
	if err != nil {
		return err
	}
	if !confirmed {
		ui.Log("-", "Update cancelled")
		return errUpdateCancelled
	}

	ui.Log(">", "Closing processes and continuing update")
	if err := terminateProcesses(processes); err != nil {
		return err
	}

	return nil
}

func promptCloseProcesses(in io.Reader, out io.Writer) (bool, error) {
	reader := bufio.NewReader(in)
	fmt.Fprint(out, "Close these processes and continue? [y/N]: ")

	line, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("read confirmation: %w", err)
	}

	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}

func wrapReplaceError(targetPath string, err error) error {
	if err == nil {
		return nil
	}

	processes, findErr := findProcessesUsingExecutable(targetPath)
	if findErr != nil || len(processes) == 0 {
		return err
	}

	return fmt.Errorf("%w\n%s", err, formatProcessesInUseMessage(processes))
}
