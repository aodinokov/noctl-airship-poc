package templater

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type FunctionConfig struct {
	// Values contains map with object parameters to render
	Values map[string]interface{} `json:"values,omitempty"`
	// Template field is used to specify actual go-template which is going
	// to be used to render the object defined in Spec field
	Template string `json:"template,omitempty"`
}

type Function struct {
	Config *FunctionConfig
}

func NewFunction(cfg *FunctionConfig) (*Function, error) {
	fn := Function{Config: cfg}
	return &fn, nil
}

func (f *Function) Exec(items []*yaml.RNode) ([]*yaml.RNode, error) {
	var out bytes.Buffer

	funcMap := sprig.TxtFuncMap()
	funcMap["toYaml"] = toYaml
	tmpl, err := template.New("tmpl").Funcs(funcMap).Parse(f.Config.Template)
	if err != nil {
		return nil, err
	}

	err = tmpl.Execute(&out, f.Config.Values)
	if err != nil {
		return nil, fmt.Errorf("template exec returned error: %v", err)
	}

	// Convert string to Rnodes
	p := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: bytes.NewBufferString(out.String())}},
		Outputs: []kio.Writer{&kio.PackageBuffer{}},
	}
	err = p.Execute()
	if err != nil {
		return nil, err
	}
	return append(items, p.Outputs[0].(*kio.PackageBuffer).Nodes...), nil
}

// Render input yaml as output yaml
func toYaml(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return string(data)
}
