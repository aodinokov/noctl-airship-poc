package replacement

import (
	"testing"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestParseFieldRef(t *testing.T) {
	ts := []struct {
		In  string
		Out []string
		Err bool
	}{
		{
			In:  `a.b.c`,
			Out: []string{"a", "b", "c"},
		},
		{
			In:  `a.b[name=x].c`,
			Out: []string{"a", "b", "[name=x]", "c"},
		},
		{
			In:  `a.b[name=x.y].c`,
			Out: []string{"a", "b", "[name=x.y]", "c"},
		},
		{
			In:  `a.b.[name=x.y].c`,
			Out: []string{"a", "b", "[name=x.y]", "c"},
		},
		{
			In:  `a.b[2].c`,
			Out: []string{"a", "b", "[2]", "c"},
		},
		{
			In:  `a.b.2.c`,
			Out: []string{"a", "b", "2", "c"},
		},
		{
			In:  `a.b[2.c`,
			Err: true,
		},
	}
	for _, ti := range ts {
		x, err := parseFieldRef(ti.In)
		if err != nil && !ti.Err {
			t.Errorf("unexpected error for %s: %v", ti.In, err)
			continue
		}

		eq := true
		if len(x) != len(ti.Out) {
			eq = false
		}
		if eq {
			for i := range x {
				if x[i] != ti.Out[i] {
					eq = false
					break
				}
			}
		}
		if !eq {
			t.Errorf("for %s expected %v, got %v", ti.In, ti.Out, x)
		}
	}
}

func TestGetFieldValue(t *testing.T) {
	ts := []struct {
		InYaml        string
		InField       string
		ExpectedError bool
		ExpectedVal   string
	}{
		{
			InYaml: `
a:
  b:
    c: value
`,
			InField:     "a.b.c",
			ExpectedVal: "value",
		},
		{
			InYaml: `
a:
  b:
    c: value
`,
			InField:     ".a.b.c",
			ExpectedVal: "value",
		},
		{
			InYaml: `
a:
  b:
  - c: value1
    d: data1
  - c: value2
    d: data2
`,
			InField:     "a.b[c=value1].d",
			ExpectedVal: "data1",
		},
		{
			InYaml: `
a:
  b:
  - c: value.1
    d: data1
  - c: value.2
    d: data2
`,
			InField:     "a.b[c=value.1].d",
			ExpectedVal: "data1",
		},
		{
			InYaml: `
a:
  b:
  - c: value1
    d: data1
  - c: value2
    d: data2
`,
			InField:     "a.b[1].d",
			ExpectedVal: "data2",
		},
		{
			InYaml: `
a:
  b:
  - c: value1
    d: data1
  - c: value2
    d: data2
`,
			InField:     "a.b.[1].d",
			ExpectedVal: "data2",
		},
		{
			InYaml: `
a:
  b:
  - c: value1
    d: data1
  - c: value2
    d: data2
`,
			InField:     "a.b.1.d",
			ExpectedVal: "data2",
		},
		{
			InYaml: `
a:
  b: |
    c:
      d: innerValue
`,
			InField:     ".a.b|c.d",
			ExpectedVal: "innerValue",
		},
		{
			InYaml: `
a:
  b: |
    c:
      d: innerValue1
      e: "f: innerValue2"
`,
			InField:     "a.b|c.e|f",
			ExpectedVal: "innerValue2",
		},
	}

	for _, ti := range ts {
		node, err := yaml.Parse(ti.InYaml)
		if err != nil {
			t.Errorf("wasn't able to parse inYaml %s: %v, trying to continue", ti.InYaml, err)
			continue
		}

		val, err := getFieldValue(node, ti.InField)
		if err != nil && ti.ExpectedError {
			continue
		}
		if err != nil {
			t.Errorf("didn't expect error for field: %s yaml %s: %v", ti.InField, ti.InYaml, err)
		}
		if val != ti.ExpectedVal {
			t.Errorf("unexpected value %s for field: %s yaml %s. Expected %s",
				val, ti.InField, ti.InYaml, ti.ExpectedVal)
		}
	}

}

func TestSetFieldValue(t *testing.T) {
	ts := []struct {
		InYaml        string
		InField       string
		InValueString string
		InValueYaml   string
		ExpectedError bool
		ExpectedYaml  string
	}{
		{
			InYaml: `
a:
  b:
    c: value
`,
			InField:       "a.b.c",
			InValueString: "newvalue",
			ExpectedYaml: `
a:
  b:
    c: newvalue
`,
		},
		{
			InYaml: `
a:
  b:
    c: value
`,
			InField:       "a.b.d",
			InValueString: "newvalue",
			ExpectedYaml: `
a:
  b:
    c: value
    d: newvalue
`,
		},
		{
			InYaml: `
a:
  x: test
`,
			InField: "a.b.d",
			InValueYaml: `
e:
  - name: ng0
    value: vg0
  - name: ng1
    value: vg1
`,
			ExpectedYaml: `
a:
  x: test
  b:
    d:
      e:
      - name: ng0
        value: vg0
      - name: ng1
        value: vg1
`,
		},
		{
			InYaml: `
a:
  b:
    c: |
      d:
        e: value
`,
			InField:       "a.b.c|d.e",
			InValueString: "newvalue",
			ExpectedYaml: `
a:
  b:
    c: |
      d:
        e: newvalue
`,
		},
	}

	for _, ti := range ts {
		if ti.InValueString != "" && ti.InValueYaml != "" {
			t.Errorf("ambigious test with value string %s and yaml %s - skipping", ti.InValueString, ti.InValueYaml)
			continue
		}

		var value interface{}
		if ti.InValueString != "" {
			value = ti.InValueString
		}
		if ti.InValueYaml != "" {
			node, err := yaml.Parse(ti.InValueYaml)
			if err != nil {
				t.Errorf("wasn't able to parse value yaml %s, %v", ti.InValueYaml, err)
				continue
			}
			value = node
		}

		node, err := yaml.Parse(ti.InYaml)
		if err != nil {
			t.Errorf("wasn't able to parse inYaml %s: %v, trying to continue", ti.InYaml, err)
			continue
		}

		err = setFieldValue(node, ti.InField, value)
		if err != nil && ti.ExpectedError {
			continue
		}
		if err != nil {
			t.Errorf("didn't expect error for field: %s yaml %s: %v", ti.InField, ti.InYaml, err)
		}
		resYaml, err := node.String()
		if err != nil {
			t.Errorf("got unexpected error converting node back for inYaml %s: %v", ti.InYaml, err)
			continue
		}
		if resYaml != ti.ExpectedYaml[1:] {
			t.Errorf("expected \n%s, got \n%s", ti.ExpectedYaml, resYaml)
		}
	}
}
