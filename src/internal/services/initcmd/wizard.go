package initcmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

var templateChoices = []struct {
	label string
	value TemplateType
}{
	{"static", TemplateStatic},
	{"binary", TemplateBinary},
	{"docker", TemplateDocker},
	{"git-sync", TemplateGitSync},
}

func isInteractiveTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

func promptTemplateType(in io.Reader, out io.Writer) (TemplateType, error) {
	reader := bufio.NewReader(in)

	for {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Select a template type:")
		for i, choice := range templateChoices {
			fmt.Fprintf(out, "  %d) %s\n", i+1, choice.label)
		}
		fmt.Fprint(out, "Choice [1-4]: ")

		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("read template choice: %w", err)
		}

		choice, ok := parseTemplateChoice(strings.TrimSpace(line))
		if ok {
			return choice, nil
		}

		fmt.Fprintln(out, "Invalid choice. Enter a number from 1 to 4.")
	}
}

func parseTemplateChoice(input string) (TemplateType, bool) {
	index, err := strconv.Atoi(input)
	if err != nil || index < 1 || index > len(templateChoices) {
		return "", false
	}

	return templateChoices[index-1].value, true
}
