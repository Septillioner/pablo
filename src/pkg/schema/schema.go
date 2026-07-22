package schema

type Field struct {
	Description string
	Enum        []string
	Children    map[string]*Field
}

var Root = &Field{
	Children: map[string]*Field{
		"name": {
			Description: "Project name.",
		},
		"version": {
			Description: "Project version.",
		},
		"sequences": {
			Description: "Named ordered lists of profile/environment targets. List order is execution order.",
			Children: map[string]*Field{
				"*": {
					Description: "Sequence name. Each item is a profile/environment target (e.g. app/linux-remote).",
				},
			},
		},
		"credentials": {
			Description: "Named reusable credentials. Referenced by string from remote.credential or git.credential.",
			Children: map[string]*Field{
				"*": {
					Children: map[string]*Field{
						"type": {
							Description: "Credential type.",
							Enum:        []string{"ssh", "token", "basic"},
						},
						"username":   {Description: "Username (ssh, basic)."},
						"password":   {Description: "Password (basic, or ssh password auth)."},
						"key":        {Description: "SSH private key path."},
						"passphrase": {Description: "SSH key passphrase."},
						"value":      {Description: "Token value (token type)."},
					},
				},
			},
		},
		"profiles": {
			Description: "Application profiles. Profile = what you build; environments = where it runs.",
			Children: map[string]*Field{
				"*": {
					Children: map[string]*Field{
						"type": {
							Description: "Profile type. Gates which fields are allowed.",
							Enum:        []string{"static", "binary", "docker", "git-sync"},
						},
						"variables": {Description: "Canonical variables inherited by every environment unless overridden."},
						"env_file":  {Description: "Default deploy env file name for environments."},
						"build": {
							Description: "Build command. Required for binary; optional for static.",
							Children: map[string]*Field{
								"command":   {Description: "Shell command to build."},
								"path":      {Description: "Working directory for the build command."},
								"variables": {Description: "Optional build-only overlay on environment variables (pre-build file + process env)."},
								"env_file":  {Description: "Write environment variables (plus optional build.variables) to this file under build.path before building."},
							},
						},
						"git": {
							Description: "Git repository settings (docker and git-sync).",
							Children: map[string]*Field{
								"repo":       {Description: "Git repository URL."},
								"branch":     {Description: "Git branch."},
								"credential": {Description: "Credential name for Git access."},
							},
						},
						"environments": {
							Description: "Deployment environments for this profile.",
							Children: map[string]*Field{
								"*": {
									Children: map[string]*Field{
										"variables": {Description: "Canonical variables for this environment (deploy env_file, and pre-build file when build.env_file is set)."},
										"env_file":  {Description: "Env file name written into the deploy target."},
										"build": {
											Description: "Override profile build for this environment.",
											Children: map[string]*Field{
												"command":   {Description: "Shell command to build."},
												"path":      {Description: "Working directory for the build command."},
												"variables": {Description: "Optional build-only overlay on environment variables (pre-build file + process env)."},
												"env_file":  {Description: "Write environment variables (plus optional build.variables) to this file under build.path before building."},
											},
										},
										"remote": {
											Description: "SSH connection. Present = remote deploy; absent = local.",
											Children: map[string]*Field{
												"host":                  {Description: "Remote host (host or host:port)."},
												"credential":            {Description: "Credential name for SSH."},
												"host_key_verification": {Description: "Verify host key against known_hosts. Default: on.", Enum: []string{"on", "off"}},
												"trust_on_first_use":    {Description: "Record unknown host key on first connect. Default: off.", Enum: []string{"on", "off"}},
											},
										},
										"register_path": {
											Description: "Register target_path on PATH (binary only).",
											Children: map[string]*Field{
												"scope": {Description: "PATH scope.", Enum: []string{"user", "system"}},
											},
										},
										"deploy": {
											Description: "How artifacts land on the target.",
											Children: map[string]*Field{
												"source": {
													Description: "Artifact location (static/binary). Object only.",
													Children: map[string]*Field{
														"dir":     {Description: "Directory containing artifacts."},
														"include": {Description: "Include patterns."},
														"exclude": {Description: "Exclude patterns."},
													},
												},
												"target_path":     {Description: "Absolute or relative path on the target machine."},
												"strategy":        {Description: "Deployment strategy.", Enum: []string{"overwrite", "backup", "recreate", "rename-replace"}},
												"transfer":        {Description: "Remote transfer method. Default: tar.", Enum: []string{"tar", "legacy"}},
												"verify_checksum": {Description: "After remote static/binary deploy, verify SHA-256 (default false)."},
												"pre_commands":    {Description: "Commands to run before artifacts are transferred."},
												"post_commands":   {Description: "Commands to run after artifacts are transferred."},
												"docker": {
													Description: "Docker Compose settings (docker type).",
													Children: map[string]*Field{
														"compose_file":     {Description: "Path to docker-compose file."},
														"build":            {Description: "Whether to build images on up."},
														"stop_before_sync": {Description: "If true (default), stop a running Compose stack before git sync."},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	},
}

func GetFieldAtPath(path []string) *Field {
	current := Root
	for _, segment := range path {
		if current.Children == nil {
			return nil
		}
		if next, ok := current.Children[segment]; ok {
			current = next
		} else if wildcard, ok := current.Children["*"]; ok {
			current = wildcard
		} else {
			return nil
		}
	}
	return current
}
