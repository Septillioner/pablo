package main

type Field struct {
	Description string
	Enum        []string
	Children    map[string]*Field
}

var PabloSchema = &Field{
	Children: map[string]*Field{
		"name": {
			Description: "The unique name of your project.",
		},
		"version": {
			Description: "The version of the configuration schema (e.g., '1.4.2').",
		},
		"credentials": {
			Description: "Define reusable credentials (SSH keys, tokens).",
			Children: map[string]*Field{
				"*": {
					Children: map[string]*Field{
						"type": {
							Description: "Type of credential.",
							Enum:        []string{"ssh", "token", "basic"},
						},
						"username":   {Description: "Username for basic auth or SSH."},
						"password":   {Description: "Password for basic auth."},
						"key":        {Description: "SSH private key content or path."},
						"passphrase": {Description: "Passphrase for the SSH key."},
					},
				},
			},
		},
		"profiles": {
			Description: "Define one or more deployment profiles. Each profile defines build and packaging strategy.",
			Children: map[string]*Field{
				"*": {
					Children: map[string]*Field{
						"type": {
							Description: "Defines how the application is processed.",
							Enum:        []string{"static", "binary", "docker", "git-sync"},
						},
						"build": {
							Description: "Configuration for the build command.",
							Children: map[string]*Field{
								"command":   {Description: "The shell command to execute (e.g., 'npm run build')."},
								"path":      {Description: "The working directory for the command."},
								"variables": {Description: "Environment variables to inject during build."},
								"env_file":  {Description: "File to write variables to before building."},
							},
						},
						"output_dir": {
							Description: "Defines where the build artifacts are located.",
							Children: map[string]*Field{
								"dir":     {Description: "The directory containing artifacts."},
								"include": {Description: "Patterns of files to include."},
								"exclude": {Description: "Patterns of files to exclude."},
							},
						},
						"git": {
							Description: "Git repository settings.",
							Children: map[string]*Field{
								"repo":       {Description: "Git repository URL."},
								"branch":     {Description: "Git branch to use."},
								"credential": {Description: "Name of the credential to use for Git."},
							},
						},
						"hooks": {
							Description: "Lifecycle hooks.",
							Children: map[string]*Field{
								"pre":  {Description: "Command to run before deployment."},
								"post": {Description: "Command to run after deployment."},
							},
						},
						"pipeline": {
							Description: "Pipeline orchestration settings.",
							Children: map[string]*Field{
								"on_success":   {Description: "Command to run on success."},
								"on_failure":   {Description: "Command to run on failure."},
								"health_check": {Description: "URL or command for health check."},
							},
						},
						"environments": {
							Description: "Configuration for different deployment environments.",
							Children: map[string]*Field{
								"*": {
									Children: map[string]*Field{
										"variables": {Description: "Runtime variables for this environment."},
										"build":     {Description: "Override profile build settings."},
										"remote": {
											Description: "Connection to a remote server.",
											Children: map[string]*Field{
												"method":     {Description: "Connection method.", Enum: []string{"ssh"}},
												"host":       {Description: "Remote host address."},
												"credential": {Description: "Credential name for remote access."},
											},
										},
										"deploy": {
											Description: "Physical deployment parameters.",
											Children: map[string]*Field{
												"target_path": {Description: "Absolute path on the target machine."},
												"strategy":    {Description: "Deployment strategy.", Enum: []string{"overwrite", "backup", "recreate"}},
												"remote":      {Description: "Transfer method.", Enum: []string{"tar", "legacy"}},
												"docker": {
													Description: "Docker Compose specific settings.",
													Children: map[string]*Field{
														"compose_file": {Description: "Path to docker-compose file."},
														"build":        {Description: "Whether to build images."},
														"command":      {Description: "Custom docker-compose command."},
													},
												},
												"service": {
													Description: "Service management settings.",
													Children: map[string]*Field{
														"type":    {Description: "Service manager type.", Enum: []string{"systemd", "pm2"}},
														"name":    {Description: "Service name."},
														"restart": {Description: "Whether to restart the service."},
													},
												},
												"pre_commands":  {Description: "Commands to run before deployment artifacts are transferred."},
												"post_commands": {Description: "Commands to run after deployment artifacts are transferred."},
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
