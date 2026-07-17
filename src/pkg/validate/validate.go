package validate

import (
	"fmt"
	"pablo/pkg/config"
	"pablo/pkg/domain"
	"pablo/pkg/schema"
	"pablo/pkg/target"
	"strings"

	"gopkg.in/yaml.v3"
)

type Severity int

const (
	SeverityError Severity = iota
	SeverityWarning
)

type Diagnostic struct {
	Path      string
	Line      int
	Column    int
	EndLine   int
	EndColumn int
	Message   string
	Severity  Severity
}

var (
	profileTypes             = []string{"static", "binary", "docker", "git-sync"}
	credentialTypes          = []string{"ssh", "token", "basic"}
	deployStrategies         = []string{"overwrite", "backup", "recreate", "rename-replace"}
	deployTransferModes      = []string{"tar", "legacy"}
	hostKeyVerificationModes = []string{"on", "off"}
	trustOnFirstUseModes     = []string{"on", "off"}
	registerPathScopes       = []string{"user", "system"}
)

func ValidateYAML(content []byte, baseDir string) ([]Diagnostic, *domain.Config, error) {
	loader := config.NewLoader()
	doc, err := loader.ParseDocument(content)
	if err != nil {
		return []Diagnostic{syntaxDiagnostic(err)}, nil, nil
	}

	cfg, err := loader.LoadFromBytes(content, baseDir)
	if err != nil {
		return []Diagnostic{syntaxDiagnostic(err)}, nil, nil
	}

	positions := buildPositionIndex(doc)
	diags := Validate(cfg, positions)
	diags = append(diags, checkUnknownKeys(doc, schema.Root, "")...)
	return diags, cfg, nil
}

func Validate(cfg *domain.Config, positions map[string]*yaml.Node) []Diagnostic {
	if positions == nil {
		positions = map[string]*yaml.Node{}
	}

	var diags []Diagnostic

	if cfg.Name == "" {
		diags = append(diags, requiredDiagnostic(positions, "name", "name is required"))
	}

	if len(cfg.Profiles) == 0 {
		diags = append(diags, requiredDiagnostic(positions, "profiles", "profiles is required"))
	}

	for credName, cred := range cfg.Credentials {
		base := "credentials." + credName
		if cred.Type == "" {
			diags = append(diags, requiredDiagnostic(positions, base+".type", "credential type is required"))
		} else if !contains(credentialTypes, cred.Type) {
			diags = append(diags, enumDiagnostic(positions, base+".type", cred.Type, credentialTypes))
		}
	}

	for profileName, profile := range cfg.Profiles {
		base := "profiles." + profileName
		if profile.Type == "" {
			diags = append(diags, requiredDiagnostic(positions, base+".type", "profile type is required"))
		} else if !contains(profileTypes, profile.Type) {
			diags = append(diags, enumDiagnostic(positions, base+".type", profile.Type, profileTypes))
		}

		if profile.Git != nil && profile.Git.Credential != "" {
			if _, ok := cfg.Credentials[profile.Git.Credential]; !ok {
				diags = append(diags, refDiagnostic(positions, base+".git.credential", fmt.Sprintf("credential %q not found", profile.Git.Credential)))
			}
		}

		switch profile.Type {
		case "docker", "git-sync":
			if profile.Git == nil || profile.Git.Repo == "" {
				diags = append(diags, requiredDiagnostic(positions, base+".git", "git.repo is required for "+profile.Type+" profiles"))
			}
		case "static", "binary":
			if profile.Git != nil {
				diags = append(diags, nodeDiagnostic(positions, base+".git", "git is not allowed for "+profile.Type+" profiles", SeverityError))
			}
		}

		if len(profile.Environments) == 0 {
			diags = append(diags, requiredDiagnostic(positions, base+".environments", "profile must define at least one environment"))
		}

		for envName, env := range profile.Environments {
			envBase := base + ".environments." + envName
			diags = append(diags, validateEnvironment(cfg, profile, env, envBase, positions)...)
		}
	}

	for seqName, steps := range cfg.Sequences {
		seqBase := "sequences." + seqName
		if len(steps) == 0 {
			diags = append(diags, requiredDiagnostic(positions, seqBase, fmt.Sprintf("sequence %q must have at least one step", seqName)))
			continue
		}

		for i, step := range steps {
			stepPath := fmt.Sprintf("%s[%d]", seqBase, i)
			profileName, envName, err := target.Parse(step)
			if err != nil {
				diags = append(diags, nodeDiagnostic(positions, seqBase, fmt.Sprintf("sequence %q step %d: %v", seqName, i+1, err), SeverityError))
				continue
			}

			profile, _ := cfg.GetProfile(profileName)
			if profile == nil {
				diags = append(diags, refDiagnostic(positions, stepPath, fmt.Sprintf("profile %q not found", profileName)))
				continue
			}

			if _, ok := profile.Environments[envName]; !ok {
				diags = append(diags, refDiagnostic(positions, stepPath, fmt.Sprintf("environment %q not found in profile %q", envName, profileName)))
			}
		}
	}

	return diags
}

func validateEnvironment(cfg *domain.Config, profile domain.Profile, env domain.Environment, envBase string, positions map[string]*yaml.Node) []Diagnostic {
	var diags []Diagnostic

	if env.Deploy.TargetPath == "" {
		diags = append(diags, requiredDiagnostic(positions, envBase+".deploy.target_path", "deploy.target_path is required"))
	}

	if env.Deploy.Strategy != "" && !contains(deployStrategies, env.Deploy.Strategy) {
		diags = append(diags, enumDiagnostic(positions, envBase+".deploy.strategy", env.Deploy.Strategy, deployStrategies))
	}

	if env.Deploy.Transfer != "" && !contains(deployTransferModes, env.Deploy.Transfer) {
		diags = append(diags, enumDiagnostic(positions, envBase+".deploy.transfer", env.Deploy.Transfer, deployTransferModes))
	}

	if env.Remote != nil {
		if env.Remote.Host == "" {
			diags = append(diags, requiredDiagnostic(positions, envBase+".remote.host", "remote.host is required"))
		}
		if env.Remote.Credential == "" {
			diags = append(diags, requiredDiagnostic(positions, envBase+".remote.credential", "remote.credential is required"))
		} else if _, ok := cfg.Credentials[env.Remote.Credential]; !ok {
			diags = append(diags, refDiagnostic(positions, envBase+".remote.credential", fmt.Sprintf("credential %q not found", env.Remote.Credential)))
		}
		if env.Remote.HostKeyVerification != "" && !contains(hostKeyVerificationModes, env.Remote.HostKeyVerification) {
			diags = append(diags, enumDiagnostic(positions, envBase+".remote.host_key_verification", env.Remote.HostKeyVerification, hostKeyVerificationModes))
		}
		if env.Remote.TrustOnFirstUse != "" && !contains(trustOnFirstUseModes, env.Remote.TrustOnFirstUse) {
			diags = append(diags, enumDiagnostic(positions, envBase+".remote.trust_on_first_use", env.Remote.TrustOnFirstUse, trustOnFirstUseModes))
		}
	}

	if env.RegisterPath != nil {
		if profile.Type != "binary" {
			diags = append(diags, nodeDiagnostic(positions, envBase+".register_path", "register_path is only allowed for binary profiles", SeverityError))
		}
		if env.RegisterPath.Scope != "" && !contains(registerPathScopes, env.RegisterPath.Scope) {
			diags = append(diags, enumDiagnostic(positions, envBase+".register_path.scope", env.RegisterPath.Scope, registerPathScopes))
		}
	}

	switch profile.Type {
	case "static", "binary":
		if env.Deploy.Source == nil || env.Deploy.Source.Dir == "" {
			diags = append(diags, requiredDiagnostic(positions, envBase+".deploy.source", "deploy.source.dir is required for "+profile.Type+" profiles"))
		}
		if env.Deploy.Docker != nil {
			diags = append(diags, nodeDiagnostic(positions, envBase+".deploy.docker", "deploy.docker is not allowed for "+profile.Type+" profiles", SeverityError))
		}
		if profile.Type == "binary" {
			if env.Build == nil || env.Build.Command == "" {
				diags = append(diags, requiredDiagnostic(positions, envBase+".build", "build.command is required for binary profiles (set on profile or environment)"))
			}
		}
	case "docker":
		if env.Deploy.Source != nil {
			diags = append(diags, nodeDiagnostic(positions, envBase+".deploy.source", "deploy.source is not allowed for docker profiles", SeverityError))
		}
		if env.Deploy.Docker == nil || env.Deploy.Docker.ComposeFile == "" {
			diags = append(diags, requiredDiagnostic(positions, envBase+".deploy.docker", "deploy.docker.compose_file is required for docker profiles"))
		}
	case "git-sync":
		if env.Deploy.Source != nil {
			diags = append(diags, nodeDiagnostic(positions, envBase+".deploy.source", "deploy.source is not allowed for git-sync profiles", SeverityError))
		}
		if env.Deploy.Docker != nil {
			diags = append(diags, nodeDiagnostic(positions, envBase+".deploy.docker", "deploy.docker is not allowed for git-sync profiles", SeverityError))
		}
	}

	return diags
}

func checkUnknownKeys(node *yaml.Node, field *schema.Field, path string) []Diagnostic {
	if node == nil || field == nil || field.Children == nil {
		return nil
	}
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		return checkUnknownKeys(node.Content[0], field, path)
	}
	if node.Kind != yaml.MappingNode {
		return nil
	}

	var diags []Diagnostic
	for i := 0; i+1 < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		key := keyNode.Value

		childPath := key
		if path != "" {
			childPath = path + "." + key
		}

		childField, ok := field.Children[key]
		if !ok {
			childField, ok = field.Children["*"]
		}
		if !ok {
			diags = append(diags, Diagnostic{
				Path:      childPath,
				Line:      keyNode.Line,
				Column:    keyNode.Column,
				EndLine:   keyNode.Line,
				EndColumn: keyNode.Column + len(key),
				Message:   fmt.Sprintf("unknown field %q", key),
				Severity:  SeverityError,
			})
			continue
		}

		if childField != nil && childField.Children != nil {
			diags = append(diags, checkUnknownKeys(valueNode, childField, childPath)...)
		}
	}
	return diags
}

func HasErrors(diags []Diagnostic) bool {
	for _, d := range diags {
		if d.Severity == SeverityError {
			return true
		}
	}
	return false
}

func FormatDiagnostic(filename string, d Diagnostic) string {
	severity := "error"
	if d.Severity == SeverityWarning {
		severity = "warning"
	}
	return fmt.Sprintf("%s:%d:%d: %s: %s", filename, d.Line, d.Column, severity, d.Message)
}

func buildPositionIndex(doc *yaml.Node) map[string]*yaml.Node {
	index := make(map[string]*yaml.Node)
	if doc == nil {
		return index
	}
	walkMapping(doc, nil, index)
	return index
}

func walkMapping(node *yaml.Node, path []string, index map[string]*yaml.Node) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			walkMapping(child, path, index)
		}
	case yaml.MappingNode:
		for i := 0; i+1 < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]
			keyPath := append(path, keyNode.Value)
			pathKey := strings.Join(keyPath, ".")
			index[pathKey] = valueNode
			index[pathKey+".key"] = keyNode

			switch valueNode.Kind {
			case yaml.MappingNode:
				walkMapping(valueNode, keyPath, index)
			case yaml.SequenceNode:
				for _, item := range valueNode.Content {
					if item.Kind == yaml.MappingNode {
						walkMapping(item, keyPath, index)
					}
				}
			}
		}
	}
}

func requiredDiagnostic(index map[string]*yaml.Node, path, message string) Diagnostic {
	return nodeDiagnostic(index, path, message, SeverityError)
}

func enumDiagnostic(index map[string]*yaml.Node, path, value string, allowed []string) Diagnostic {
	message := fmt.Sprintf("invalid value %q; expected one of: %s", value, strings.Join(allowed, ", "))
	return nodeDiagnostic(index, path, message, SeverityError)
}

func refDiagnostic(index map[string]*yaml.Node, path, message string) Diagnostic {
	return nodeDiagnostic(index, path, message, SeverityError)
}

func nodeDiagnostic(index map[string]*yaml.Node, path, message string, severity Severity) Diagnostic {
	d := Diagnostic{
		Path:      path,
		Line:      1,
		Column:    1,
		EndLine:   1,
		EndColumn: 1,
		Message:   message,
		Severity:  severity,
	}

	node := index[path]
	if node == nil {
		node = index[path+".key"]
	}
	if node != nil {
		d.Line = node.Line
		d.Column = node.Column
		d.EndLine = node.Line
		d.EndColumn = node.Column + len(node.Value)
		if d.EndColumn <= d.Column {
			d.EndColumn = d.Column + 1
		}
	}

	return d
}

func syntaxDiagnostic(err error) Diagnostic {
	line, column := 1, 1
	msg := err.Error()
	if strings.Contains(msg, "line ") {
		fmt.Sscanf(msg, "yaml: line %d:", &line)
	}
	return Diagnostic{
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: column + 1,
		Message:   msg,
		Severity:  SeverityError,
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
