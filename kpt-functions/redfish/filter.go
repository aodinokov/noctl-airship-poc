package redfish

import (
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type complexFilter struct {
	Filters []kio.Filter
}

func (cf *complexFilter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {

	p := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.PackageBuffer{Nodes: nodes}},
		Filters: cf.Filters,
		Outputs: []kio.Writer{&kio.PackageBuffer{}},
	}

	err := p.Execute()
	if err != nil {
		return nil, err
	}

	return p.Outputs[0].(*kio.PackageBuffer).Nodes, nil
}
