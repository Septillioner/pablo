package initcmd

import "fmt"

func templateYAML(templateType TemplateType, version string) string {
	switch templateType {
	case TemplateBinary:
		return binaryTemplate(version)
	case TemplateDocker:
		return dockerTemplate(version)
	case TemplateGitSync:
		return gitSyncTemplate(version)
	default:
		return staticTemplate(version)
	}
}

func staticTemplate(version string) string {
	return fmt.Sprintf(`name: my-app
version: %s
profiles:
  default:
    type: static
    build:
      command: npm run build
      path: .
    environments:
      production:
        deploy:
          source:
            dir: ./dist
            include: ["**/*"]
          target_path: ./deploy-output
          strategy: overwrite
`, version)
}

func binaryTemplate(version string) string {
	return fmt.Sprintf(`name: my-app
version: %s
profiles:
  default:
    type: binary
    build:
      command: go build -o my-app .
      path: .
    environments:
      production:
        deploy:
          source:
            dir: .
            include: ["my-app"]
          target_path: ./deploy-output
          strategy: backup
          post_commands:
            - systemctl restart my-app
`, version)
}

func dockerTemplate(version string) string {
	return fmt.Sprintf(`name: my-app
version: %s
profiles:
  default:
    type: docker
    git:
      repo: git@github.com:user/repo.git
      branch: main
    environments:
      production:
        deploy:
          target_path: ./docker-stack
          strategy: overwrite
          docker:
            compose_file: docker-compose.yml
            build: true
`, version)
}

func gitSyncTemplate(version string) string {
	return fmt.Sprintf(`name: my-app
version: %s
profiles:
  default:
    type: git-sync
    git:
      repo: https://github.com/user/repo.git
      branch: main
    environments:
      production:
        deploy:
          target_path: ./docs-output
          strategy: overwrite
`, version)
}
