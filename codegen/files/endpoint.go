package files

import (
	"path/filepath"
	"text/template"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design"
)

type (
	// endpointData contains the data necessary to render the endpoint
	// template.
	endpointData struct {
		// Name is the service name.
		Name string
		// VarName is the endpoint struct name.
		VarName string
		// Methods lists the endpoint struct methods.
		Methods []*endpointMethod
	}

	// endpointMethod describes a single endpoint method.
	endpointMethod struct {
		// Name is the method name.
		Name string
		// PayloadType is name of the payload Go type if any.
		PayloadType string
		// HasPayload is true if the payload type is not empty.
		HasPayload bool
	}

	// endpointFile is the codgen file for a given service.
	endpointFile struct {
		service *design.ServiceExpr
	}
)

// endpointTmpl is the template used to render the body of the endpoint file.
var endpointTmpl = template.Must(template.New("endpoint").Parse(endpointT))

// Endpoint returns the endpoint file for the given service.
func Endpoint(service *design.ServiceExpr) codegen.File {
	return &endpointFile{service}
}

// Sections returns the endpoint file sections.
func (e *endpointFile) Sections(genPkg string) []*codegen.Section {
	var (
		data *endpointData
	)
	{
		methods := make([]*endpointMethod, len(e.service.Endpoints))
		for i, v := range e.service.Endpoints {
			methods[i] = &endpointMethod{
				Name:        codegen.Goify(v.Name, true),
				PayloadType: codegen.Goify(v.Payload.Name(), true),
				HasPayload:  v.Payload != design.Empty,
			}
		}
		data = &endpointData{
			Name:    e.service.Name,
			VarName: codegen.Goify(e.service.Name, true),
			Methods: methods,
		}
	}

	var (
		header, body *codegen.Section
	)
	{
		header = codegen.Header(e.service.Name+"Endpoints", "endpoints",
			[]*codegen.ImportSpec{
				&codegen.ImportSpec{Path: "context"},
				&codegen.ImportSpec{Path: "goa.design/goa.v2"},
				&codegen.ImportSpec{Path: genPkg + "/services"},
			})
		body = &codegen.Section{
			Template: endpointTmpl,
			Data:     data,
		}
	}

	return []*codegen.Section{header, body}
}

// OutputPath is the path to the generated endpoint file relative to the output
// directory.
func (e *endpointFile) OutputPath(reserved map[string]bool) string {
	svc := codegen.SnakeCase(e.service.Name)
	return UniquePath(filepath.Join("endpoints", svc+"%d.go"), reserved)
}

// endpointT is the template used to write an endpoint definition.
const endpointT = `type (
	// {{ .VarName }} lists the {{ .Name }} service endpoints.
	{{ .VarName }} struct {
{{ range .Methods }}		{{ .Name }} goa.Endpoint
{{ end }}	}
)

// New{{ .VarName }} wraps the methods of a {{ .Name }} service with endpoints.
func New{{ .VarName }}(s services.{{ .VarName }}) *{{ .VarName }} {
	ep := &{{ .VarName }}{}
{{ range .Methods }}
	ep.{{ .Name }} = func(ctx context.Context, req interface{}) (interface{}, error) {
{{- if .HasPayload }}
		p := req.(*services.{{ .PayloadType }})
{{- end }}
		return s.{{ .Name }}(ctx, {{ if .HasPayload }}p{{ else }}nil{{ end }})
	}
{{ end }}
	return ep
}`
