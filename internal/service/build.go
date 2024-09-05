package service

import (
	_ "embed"
	"html/template"
	"io"

	"github.com/itallix/go-metrics/internal/logger"
)

//go:embed build_info.tpl
var buildInfoTpl string

// PrintBuildInfo uses predefined template to output some build meta information using the passed writer.
// Use os.Stdout as w parameter to print information into console.
func PrintBuildInfo(version string, date string, commit string, w io.Writer) {
	data := struct {
		Version string
		Date    string
		Commit  string
	}{
		Version: version,
		Date:    date,
		Commit:  commit,
	}

	tmpl, err := template.New("buildInfo").Parse(buildInfoTpl)
	if err != nil {
		logger.Log().Errorf("Error parsing template file: %w", err)
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		logger.Log().Errorf("Error executing template: %w", err)
	}
}
