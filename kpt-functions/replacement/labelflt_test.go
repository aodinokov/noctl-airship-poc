package replacement

import (
	"bytes"

	"sigs.k8s.io/kustomize/kyaml/kio"

	"testing"
)

func TestLabelFilter(t *testing.T) {
	ts := []struct {
		InNodes    string
		InPath     []string
		InSelector string
		OutNodes   string
		OutError   bool
	}{
		{
			InNodes: `
kind: test
metadata:
  name: testname1
  labels: "x=y,z=a"
---
kind: test
metadata:
  name: testname2
  labels: "x=r,z=xxx"
`,
			InPath:     []string{"metadata", "labels"},
			InSelector: "x in (r), z in (xxx)",
			OutNodes: `
kind: test
metadata:
  name: testname2
  labels: "x=r,z=xxx"
`,
		},
	}

	for _, ti := range ts {
		out := &bytes.Buffer{}
		err := kio.Pipeline{
			Inputs:  []kio.Reader{&kio.ByteReader{Reader: bytes.NewBufferString(ti.InNodes)}},
			Filters: []kio.Filter{LabelFilter{Path: ti.InPath, Selector: ti.InSelector}},
			Outputs: []kio.Writer{kio.ByteWriter{Writer: out}},
		}.Execute()
		if err != nil && ti.OutError {
			t.Errorf("got unexpected error when filtering %s with path %v, selector %s",
				ti.InNodes, ti.InPath, ti.InSelector)
			continue
		}
		if ti.OutNodes[1:] != out.String() {
			t.Errorf("got unexpected result when filtering\n%s\nwith path %v, selector %s. got:\n%s\nexpected:\n%s\n",
				ti.InNodes, ti.InPath, ti.InSelector, out.String(), ti.OutNodes)
		}
	}
}
