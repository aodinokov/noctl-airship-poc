package replacement

import (
	"k8s.io/apimachinery/pkg/labels"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type LabelFilter struct {
	Path     []string `yaml:"path,omitempty"`
	Selector string   `yaml:"selector,omitempty"`
}

func (f LabelFilter) Filter(input []*yaml.RNode) ([]*yaml.RNode, error) {
	var output kio.ResourceNodeSlice

	for i := range input {
		node := input[i]
		// log.Printf("path %v", f.Path)
		val, err := node.Pipe(yaml.Lookup(f.Path...))
		if err != nil {
			return nil, err
		}

		if val == nil || yaml.GetValue(val) == "" {
			// log.Printf("val is nil of content is empty")
			continue
		}

		// log.Printf("value: %s", yaml.GetValue(val))
		lset, err := labels.ConvertSelectorToLabelsMap(yaml.GetValue(val))
		if err != nil {
			return nil, err
		}

		// log.Printf("lset: %v", lset)
		s, err := labels.Parse(f.Selector)
		if err != nil {
			return nil, err
		}

		if s.Matches(lset) {
			output = append(output, node)
		}
	}
	return output, nil
}
