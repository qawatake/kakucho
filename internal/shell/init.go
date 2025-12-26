package shell

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/qawatake/kakucho/internal/config"
)

const bashTemplate = `# kakucho shell integration
# Add the following to your ~/.bashrc:
#   eval "$(kakucho init bash)"

{{range .Targets}}
{{.Name}}() {
    command kakucho exec {{.Name}} "$@"
}
{{end}}
`

const zshTemplate = `# kakucho shell integration
# Add the following to your ~/.zshrc:
#   eval "$(kakucho init zsh)"

{{range .Targets}}
{{.Name}}() {
    command kakucho exec {{.Name}} "$@"
}
{{end}}
`

type targetInfo struct {
	Name string
}

type templateData struct {
	Targets []targetInfo
}

// GenerateInit generates the shell initialization script
func GenerateInit(shellName string, cfg config.Config) (string, error) {
	var tmplStr string
	switch shellName {
	case "bash":
		tmplStr = bashTemplate
	case "zsh":
		tmplStr = zshTemplate
	default:
		return "", fmt.Errorf("unsupported shell: %s (supported: bash, zsh)", shellName)
	}

	tmpl, err := template.New("init").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	data := templateData{
		Targets: make([]targetInfo, 0, len(cfg)),
	}
	for name := range cfg {
		data.Targets = append(data.Targets, targetInfo{Name: name})
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
