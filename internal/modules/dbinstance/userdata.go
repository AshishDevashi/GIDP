package dbinstance

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"
)

//go:embed templates/bootstrap.sh.tmpl
var bootstrapTemplateSource string

var bootstrapTemplate = template.Must(template.New("bootstrap").Parse(bootstrapTemplateSource))

// bootstrapParams are the values injected into the cloud-init script. They all
// originate from application config, never from client input.
type bootstrapParams struct {
	Region          string
	AdminSecretName string
	AdminUsername   string
	PostgresImage   string
	PostgresPort    int32
	DataDeviceName  string
	DataMountPoint  string
	ContainerName   string
}

func renderBootstrap(params bootstrapParams) (string, error) {
	var buf bytes.Buffer
	if err := bootstrapTemplate.Execute(&buf, params); err != nil {
		return "", fmt.Errorf("dbinstance: render bootstrap script: %w", err)
	}
	return buf.String(), nil
}
