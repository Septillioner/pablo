package pipeline

import (
	"fmt"
	"strings"

	"pablo/internal/services/hooks"
	"pablo/pkg/domain"
	"pablo/pkg/ui"
)

const (
	envPabloTargetSlot   = "PABLO_TARGET_SLOT"
	envPabloPreviousSlot = "PABLO_PREVIOUS_SLOT"
)

type slotSelection struct {
	Active   bool
	Target   string // configured slot path written this run
	Previous string // detected slot path; empty on first deploy
	Switch   string // resolved switch command
}

func (s *Service) detectSlot(env domain.Environment, cfg *domain.Config) (*slotSelection, error) {
	bg := env.Deploy.BlueGreen
	if bg == nil {
		return &slotSelection{}, nil
	}

	out, err := s.runDetectCommand(bg.DetectCommand, env, cfg)
	if err != nil {
		return nil, fmt.Errorf("blue_green.detect_command failed: %w", err)
	}

	active := strings.TrimSpace(out)
	if strings.Contains(active, "\n") {
		return nil, fmt.Errorf("blue_green.detect_command returned multiple lines; expected a single slot path")
	}

	slots := bg.Slots
	if active == "" {
		sel := &slotSelection{
			Active:   true,
			Target:   slots[0].Path,
			Previous: "",
			Switch:   resolveSlotSwitch(slots[0], bg),
		}
		ui.Log("*", fmt.Sprintf("Blue-green: no active slot; deploying to %s", sel.Target))
		return sel, nil
	}

	activeIndex := -1
	for i, slot := range slots {
		if resolveSlotKey(slot) == active {
			activeIndex = i
			break
		}
	}
	if activeIndex < 0 {
		expected := make([]string, len(slots))
		for i, slot := range slots {
			expected[i] = fmt.Sprintf("%q", resolveSlotKey(slot))
		}
		return nil, fmt.Errorf("blue_green.detect_command returned %q; expected one of: %s", active, strings.Join(expected, ", "))
	}

	idleIndex := 1 - activeIndex
	sel := &slotSelection{
		Active:   true,
		Target:   slots[idleIndex].Path,
		Previous: slots[activeIndex].Path,
		Switch:   resolveSlotSwitch(slots[idleIndex], bg),
	}
	ui.Log("*", fmt.Sprintf("Blue-green: active %s; deploying to %s", sel.Previous, sel.Target))
	return sel, nil
}

func resolveSlotKey(slot domain.SlotConfig) string {
	if slot.Key != "" {
		return slot.Key
	}
	return slot.Path
}

func resolveSlotSwitch(slot domain.SlotConfig, bg *domain.BlueGreenConfig) string {
	if slot.SwitchCommand != "" {
		return slot.SwitchCommand
	}
	return bg.SwitchCommand
}

func (s *Service) runDetectCommand(command string, env domain.Environment, cfg *domain.Config) (string, error) {
	if env.Remote != nil {
		sshClient, err := s.getSSHClient(env, cfg)
		if err != nil {
			return "", err
		}
		defer sshClient.Close()
		return s.deployer.ExecuteRemoteCommandStdout(sshClient, command)
	}
	// Local: run relative to the manifest directory so relative paths in the
	// command resolve against the project, not the process cwd.
	return hooks.Capture(command, cfg.BaseDir, nil)
}

func (s *Service) runSwitch(sel *slotSelection, env domain.Environment, cfg *domain.Config) error {
	if sel == nil || !sel.Active {
		return nil
	}

	ui.Section("Slot Switch")
	ui.Log(">", sel.Switch)

	slotEnv := slotCommandEnv(sel)
	if env.Remote != nil {
		sshClient, err := s.getSSHClient(env, cfg)
		if err != nil {
			return err
		}
		defer sshClient.Close()

		var envPrefix strings.Builder
		for k, v := range slotEnv {
			envPrefix.WriteString(fmt.Sprintf("%s='%s' ", k, shellEscapeSingle(v)))
		}
		// Prefer cd into target_path when it exists; otherwise run in place (absolute paths).
		fullCmd := fmt.Sprintf("if [ -d %s ]; then cd %s; fi; %s%s",
			env.Deploy.TargetPath, env.Deploy.TargetPath, envPrefix.String(), sel.Switch)
		if _, err := s.deployer.ExecuteRemoteCommand(sshClient, fullCmd); err != nil {
			return err
		}
		ui.Log("+", "Slot switch completed")
		return nil
	}

	// Local: same cwd as detect_command (manifest dir) so relative paths
	// like .\send-cmd.ps1 resolve against the project.
	if err := hooks.Execute(sel.Switch, cfg.BaseDir, slotEnv); err != nil {
		return err
	}
	ui.Log("+", "Slot switch completed")
	return nil
}

func slotCommandEnv(sel *slotSelection) map[string]string {
	if sel == nil || !sel.Active {
		return nil
	}
	return map[string]string{
		envPabloTargetSlot:   sel.Target,
		envPabloPreviousSlot: sel.Previous,
	}
}

func shellEscapeSingle(v string) string {
	return strings.ReplaceAll(v, "'", `'"'"'`)
}

func mergeCommandEnv(vars map[string]string, slotEnv map[string]string) map[string]string {
	if len(slotEnv) == 0 {
		return vars
	}
	merged := make(map[string]string, len(vars)+len(slotEnv))
	for k, v := range vars {
		merged[k] = v
	}
	for k, v := range slotEnv {
		merged[k] = v
	}
	return merged
}

func defaultDeployStrategy(env domain.Environment) string {
	if env.Deploy.Strategy != "" {
		return env.Deploy.Strategy
	}
	if env.Deploy.BlueGreen != nil {
		return "recreate"
	}
	return "overwrite"
}
