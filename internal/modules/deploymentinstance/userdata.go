package deploymentinstance

import _ "embed"

//go:embed templates/bootstrap.sh.tmpl
var bootstrapTemplate string

func renderBootstrap() string {
	return bootstrapTemplate
}
