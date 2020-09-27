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

type AndFilter struct {
	Filters []filters.GrepFilter `yaml:"filters,omitempty"`
}

type Filter struct {
	Filters map[string]AndFilter
}

func NewFilter(cfg *Config) (kio.Filter, error) {
	compare := func(a, b string) (int, error) {
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
	f := Filter{Filters: map[string]AndFilter{}}
	for name, fCfg := range cfg.Data {
		af := AndFilter{}
		err := yaml.Unmarshal([]byte(fCfg), &af)
		if err != nil {
			return nil, fmt.Errorf("can't unmarshal grep config: %v", err)
		}
		for _, gf := range af.Filters {
			gf.Compare = compare
		}
		f.Filters[name] = af
	}

	return &f, nil
}

func (f *AndFilter) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	for i, af := range f.Filters {
		x, err := af.Filter(items)
		if err != nil {
			return nil, fmt.Errorf("andfilter: %d element got error: %v", i, err)
		}
		items = x
	}
	return items, nil
}

func (f *Filter) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	r := map[*yaml.RNode]bool{}

	// we're implementing 'or' case. 'and' is possible to make with pipeline.
	for name, af := range f.Filters {
		out, err := af.Filter(items)
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
