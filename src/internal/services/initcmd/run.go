package initcmd

import (
	"errors"
	"fmt"
	"os"

	"pablo/pkg/ui"
)

var errInteractiveTerminalRequired = errors.New("--template requires an interactive terminal")

type Options struct {
	Version        string
	TemplateWizard bool
}

func Run(opts Options) error {
	templateType := TemplateStatic

	if opts.TemplateWizard {
		if !isInteractiveTerminal() {
			ui.Log("-", errInteractiveTerminalRequired.Error())
			return errInteractiveTerminalRequired
		}

		ui.Log(">", "Template wizard")
		selected, err := promptTemplateType(os.Stdin, os.Stdout)
		if err != nil {
			ui.Log("-", fmt.Sprintf("Template selection failed: %v", err))
			return err
		}

		templateType = selected
	} else {
		ui.Log(">", "Initializing sample pablo.yaml...")
	}

	content := templateYAML(templateType, opts.Version)
	if err := os.WriteFile(outputFileName, []byte(content), sampleFilePermission); err != nil {
		ui.Log("-", fmt.Sprintf("Failed to create sample: %v", err))
		return err
	}

	ui.Log("+", fmt.Sprintf("Sample %s created successfully (%s template).", outputFileName, templateType))
	return nil
}
