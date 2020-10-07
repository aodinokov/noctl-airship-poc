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
	Values map[string]interface{} `json:"values,omitempty" yaml:"values,omitempty"`
	// Template field is used to specify actual go-template which is going
	// to be used to render the object defined in Spec field
	Template string `json:"template,omitempty" yaml:"template,omitempty"`
	// Remove all documents before adding the generated one
	CleanPipeline bool `json:"cleanPipeline,omitempty" yaml:"cleanPipeline,omitempty"`
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
	pb := kio.PackageBuffer{}
	p := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: bytes.NewBufferString(out.String())}},
		Outputs: []kio.Writer{&pb},
	}
	err = p.Execute()
	if err != nil {
		return nil, err
	}
	if f.Config.CleanPipeline {
		return pb.Nodes, nil
	}
	return append(items, pb.Nodes...), nil
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
