package function

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type Config struct {
	Data map[string]string `json:"data,omitempty" yaml:"data,omitempty"`
}

type Filter struct {
	Filters map[string]filters.GrepFilter
}

func NewFilter(cfg *Config) (kio.Filter, error) {
	f := Filter{Filters: map[string]filters.GrepFilter{}}
	for name, grepCfg := range cfg.Data {
		gf := filters.GrepFilter{}
		gf.Compare = func(a, b string) (int, error) {
			qa, err := resource.ParseQuantity(a)
			if err != nil {
				return 0, fmt.Errorf("%s: %v", a, err)
			}
			qb, err := resource.ParseQuantity(b)
			if err != nil {
				return 0, err
			}

			return qa.Cmp(qb), err
		}
		err := yaml.Unmarshal([]byte(grepCfg), &gf)
		if err != nil {
			return nil, fmt.Errorf("can't unmarshal grep config: %v", err)
		}

		f.Filters[name] = gf
	}

	return &f, nil
}

func (f *Filter) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	r := map[*yaml.RNode]bool{}

	// we're implementing 'or' case. 'and' is possible to make with pipeline.
	for name, gf := range f.Filters {
		out, err := gf.Filter(items)
		if err != nil {
			return nil, fmt.Errorf("error executing filter %s: %v", name, err)
		}
		for _, p := range out {
			r[p] = true
		}
	}

	out := []*yaml.RNode{}
	for _, p := range items {
		if _, ok := r[p]; ok {
			out = append(out, p)
		}
	}
	return out, nil
}
