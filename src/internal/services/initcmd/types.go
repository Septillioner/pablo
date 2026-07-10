package initcmd

const (
	outputFileName       = "pablo_sample.yaml"
	sampleFilePermission = 0o644
)

type TemplateType string

const (
	TemplateStatic  TemplateType = "static"
	TemplateBinary  TemplateType = "binary"
	TemplateDocker  TemplateType = "docker"
	TemplateGitSync TemplateType = "git-sync"
)
