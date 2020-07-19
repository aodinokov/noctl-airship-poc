package replacement

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var (
	// substring substitutions are appended to paths as: ...%VARNAME%
	substringPatternRegex = regexp.MustCompile(`(\S+)%(\S+)%$`)
)

// Gvk identifies a Kubernetes API type.
// https://github.com/kubernetes/community/blob/master/contributors/design-proposals/api-machinery/api-group.md
type Gvk struct {
	Group   string `json:"group,omitempty" yaml:"group,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	Kind    string `json:"kind,omitempty" yaml:"kind,omitempty"`
}

// Selector specifies a set of resources.
// Any resource that matches intersection of all conditions
// is included in this set.
type Selector struct {
	Gvk       `json:",inline,omitempty" yaml:",inline,omitempty"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`

	// AnnotationSelector is a string that follows the label selection expression
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#api
	// It matches with the resource annotations.
	AnnotationSelector string `json:"annotationSelector,omitempty" yaml:"annotationSelector,omitempty"`

	// LabelSelector is a string that follows the label selection expression
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#api
	// It matches with the resource labels.
	LabelSelector string `json:"labelSelector,omitempty" yaml:"labelSelector,omitempty"`
}

// Target refers to a kubernetes object by Group, Version, Kind and Name
// gvk.Gvk contains Group, Version and Kind
// APIVersion is added to keep the backward compatibility of using ObjectReference
// for Var.ObjRef
type SourceObjRef struct {
	APIVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Gvk        `json:",inline,omitempty" yaml:",inline,omitempty"`
	Name       string `json:"name" yaml:"name"`
	Namespace  string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

type MultiSourceObjRef struct {
	Refs []struct {
		ObjRef   *SourceObjRef `json:"objref" yaml:"objref"`
		FieldRef string        `json:"fieldref" yaml:"fiedldref"`
	} `json:"refs,omitempty" yaml:"refs,omitempty"`
	Template string `json:"template" yaml:"template"`
}

// Source defines where a substitution is from
// It can from two different kinds of sources
//  - from a field of one resource
//  - from a several sources and go template
//  - from a string
type Source struct {
	Value string `json:"value,omitempty" yaml:"value,omitempty"`

	ObjRef   *SourceObjRef `json:"objref,omitempty" yaml:"objref,omitempty"`
	FieldRef string        `json:"fieldref,omitempty" yaml:"fiedldref,omitempty"`

	MultiRef *MultiSourceObjRef `json:"multiref,omitempty" yaml:"multiref,omitempty"`
}

// ReplTarget defines where a substitution is to.
type Target struct {
	ObjRef    *Selector `json:"objref,omitempty" yaml:"objref,omitempty"`
	FieldRefs []string  `json:"fieldrefs,omitempty" yaml:"fieldrefs,omitempty"`
}

type Replacement struct {
	Source *Source `json:"source" yaml:"source"`
	Target *Target `json:"target" yaml:"target"`
}

type FunctionConfig struct {
	Replacements []Replacement `json:"replacements,omitempty" yaml:"replacements,omitempty"`
}

type Function struct {
	Config *FunctionConfig
}

func NewFunction(cfg *FunctionConfig) (*Function, error) {
	for _, r := range cfg.Replacements {
		if r.Source == nil {
			return nil, fmt.Errorf("`from` must be specified in one replacement")
		}
		if r.Target == nil {
			return nil, fmt.Errorf("`to` must be specified in one replacement")
		}
		count := 0
		if r.Source.ObjRef != nil {
			count += 1
		}
		if r.Source.Value != "" {
			count += 1
		}
		if r.Source.MultiRef != nil {
			count += 1
		}
		if count > 1 {
			return nil, fmt.Errorf("only one of fieldref and value is allowed in one replacement")
		}
	}

	fn := Function{Config: cfg}
	return &fn, nil
}

func (f *Function) Exec(items []*yaml.RNode) error {
	for _, r := range f.Config.Replacements {
		value, err := prepareValue(items, r.Source)
		if err != nil {
			return err
		}

		fmt.Printf("The value is %s\n", value)

		err = apply(items, r.Target, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func prepareValue(items []*yaml.RNode, s *Source) (string, error) {
	if s.Value != "" {
		return s.Value, nil
	}
	if s.ObjRef != nil {
		return prepareValueFromObjRefFieldRef(items, s.ObjRef, s.FieldRef)
	}
	if s.MultiRef != nil {
		return prepareValueFromMultiRef(items, s.MultiRef)
	}
	return "", nil
}

func prepareValueFromObjRefFieldRef(
	items []*yaml.RNode, objRef *SourceObjRef, fieldRef string) (string, error) {

	node, err := objRef.FindOne(items)
	if err != nil {
		return "", err
	}
	if fieldRef == "" {
		fieldRef = ".metadata.name"
	}
	v, err := getFieldValue(node, fieldRef)
	if err != nil {
		return "", err
	}
	return v, nil
}

func prepareValueFromMultiRef(items []*yaml.RNode, m *MultiSourceObjRef) (string, error) {
	data := struct {
		Values []string
	}{
		Values: make([]string, 0, len(m.Refs)),
	}
	for i := range m.Refs {
		v, err := prepareValueFromObjRefFieldRef(
			items, m.Refs[i].ObjRef, m.Refs[i].FieldRef)
		if err != nil {
			return "", fmt.Errorf("error preparing multiref %v ref %d: %w", m, i, err)
		}
		data.Values = append(data.Values, v)
	}

	var out bytes.Buffer
	tmpl, err := template.New("tmpl").Funcs(sprig.TxtFuncMap()).Parse(m.Template)
	if err != nil {
		return "", fmt.Errorf("error parsing template %s: %w", m.Template, err)
	}

	err = tmpl.Execute(&out, data)
	if err != nil {
		return "", fmt.Errorf("error executing template %s, data %v: %w",
			m.Template, data, err)
	}

	return out.String(), nil
}

func apply(items []*yaml.RNode, t *Target, value string) error {
	matching, err := t.ObjRef.Filter(items)
	if err != nil {
		return fmt.Errorf("error filtering by objref %v: %w", t.ObjRef, err)
	}
	for _, node := range matching {
		for _, fieldref := range t.FieldRefs {
			err := setFieldValue(node, fieldref, value)
			if err != nil {
				return fmt.Errorf("error setting value for objref %v, fieldref %s, value %s, node %v: %w",
					t.ObjRef, fieldref, value, node, err)
			}
		}
	}
	return nil
}

func (g *Gvk) Filters() ([]kio.Filter, error) {
	flts := []kio.Filter{}
	if g.Group != "" {
		flts = append(flts, filters.GrepFilter{Path: []string{"group"}, Value: g.Group})
	}
	if g.Version != "" {
		flts = append(flts, filters.GrepFilter{Path: []string{"version"}, Value: g.Version})
	}
	if g.Kind != "" {
		flts = append(flts, filters.GrepFilter{Path: []string{"kind"}, Value: g.Kind})
	}
	return flts, nil
}

func (s *Selector) Filters() ([]kio.Filter, error) {
	flts, err := s.Gvk.Filters()
	if err != nil {
		return nil, err
	}
	if s.Name != "" {
		flts = append(flts, filters.GrepFilter{Path: []string{"metadata", "name"}, Value: s.Name})
	}
	if s.Namespace != "" {
		flts = append(flts, filters.GrepFilter{Path: []string{"metadata", "namespace"}, Value: s.Namespace})
	}
	if s.AnnotationSelector != "" {
		//TODO: flts = append(flts, LabelFilter{Path: []string{"metadata", "annotations"}, Value: s.AnnotationSelector})
	}
	if s.LabelSelector != "" {
		//TODO: flts = append(flts, LabelFilter{Path: []string{"metadata", "labels"}, Value: s.LabelSelector})
	}
	return flts, nil
}

func (s *Selector) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	flts, err := s.Filters()
	if err != nil {
		return nil, err
	}
	return findMatching(items, flts)
}

func (s *SourceObjRef) Filters() ([]kio.Filter, error) {
	flts := []kio.Filter{}
	if s.APIVersion != "" {
		flts = append(flts, filters.GrepFilter{Path: []string{"apiVersion"}, Value: s.APIVersion})
	}

	gflts, err := s.Gvk.Filters()
	if err != nil {
		return nil, err
	}
	flts = append(flts, gflts...)

	if s.Name != "" {
		flts = append(flts, filters.GrepFilter{Path: []string{"metadata", "name"}, Value: s.Name})
	}
	if s.Namespace != "" {
		flts = append(flts, filters.GrepFilter{Path: []string{"metadata", "namespace"}, Value: s.Namespace})
	}
	return flts, nil
}

func (s *SourceObjRef) FindOne(items []*yaml.RNode) (*yaml.RNode, error) {
	flts, err := s.Filters()
	if err != nil {
		return nil, err
	}

	matching, err := findMatching(items, flts)
	if err != nil {
		return nil, err
	}

	if len(matching) > 1 {
		return nil, fmt.Errorf("found more than one resources matching from %v", s)
	}

	if len(matching) == 0 {
		return nil, fmt.Errorf("failed to find one resource matching from %v", s)
	}

	return matching[0], nil
}

func findMatching(items []*yaml.RNode, flts []kio.Filter) ([]*yaml.RNode, error) {
	p := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.PackageBuffer{Nodes: items}},
		Filters: flts,
		Outputs: []kio.Writer{&kio.PackageBuffer{}},
	}

	err := p.Execute()
	if err != nil {
		return nil, err
	}
	return p.Outputs[0].(*kio.PackageBuffer).Nodes, nil
}

func getFieldValue(node *yaml.RNode, fieldRef string) (string, error) {
	var value string
	parts := strings.Split(fieldRef, "|")
	for i, path := range parts {
		v, err := node.Pipe(yaml.PathGetter{Path: strings.Split(path, ".")})
		if err != nil {
			return "", err
		}
		if v != nil {
			value = yaml.GetValue(v)
			if i+1 < len(parts) {
				node, err = yaml.Parse(value)
				if err != nil {
					return "", err
				}
			}
		}
	}
	return value, nil
}

func setFieldValue(node *yaml.RNode, fieldRef string, value string) error {
	// fieldref can contain substtringpattern for regexp - we need to get it
	substringPattern := ""
	groups := substringPatternRegex.FindStringSubmatch(fieldRef)
	if len(groups) == 3 {
		fieldRef = groups[1]
		substringPattern = groups[2]

		// calculate real value
		v, err := getFieldValue(node, fieldRef)
		if err != nil {
			return fmt.Errorf("wasn't able to get value for node %v, fieldref %s",
				node, fieldRef)
		}

		p := regexp.MustCompile(substringPattern)
		if !p.MatchString(v) {
			return fmt.Errorf("wasn't able to match pattern %s with value %s",
				substringPattern, v)
		}
		value = p.ReplaceAllString(v, value)
	}

	return setFieldValueImpl(node, strings.Split(fieldRef, "|"), value)
}

func setFieldValueImpl(node *yaml.RNode, fieldRefPart []string, value string) error {
	if len(fieldRefPart) > 1 {
		// this can be done only for string field
		v, err := node.Pipe(yaml.Lookup(strings.Split(fieldRefPart[0], ".")...))
		if err != nil {
			return err
		}
		if v == nil {
			return fmt.Errorf("wasn't able to find value for fieldref %s", fieldRefPart[0])
		}
		value = yaml.GetValue(v)
		includedNode, err := yaml.Parse(value)
		if err != nil {
			return fmt.Errorf("wasn't able to parse yaml value for fieldref %s", fieldRefPart[0])
		}
		err = setFieldValueImpl(includedNode, fieldRefPart[1:], value)
		if err != nil {
			return err
		}
		s, err := includedNode.String()
		if err != nil {
			return err
		}
		return node.PipeE(yaml.FieldSetter{StringValue: s})

	}
	v, err := node.Pipe(yaml.LookupCreate(yaml.ScalarNode, strings.Split(fieldRefPart[0], ".")...))
	if err != nil {
		return err
	}
	return v.PipeE(yaml.FieldSetter{StringValue: value})
}
