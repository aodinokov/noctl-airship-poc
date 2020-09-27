package function

import (
	"testing"

	"bytes"
	"os"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestTemplates(t *testing.T) {

	os.Setenv("TESTTEMPLATE", "testtemplatevalue")

	tc := []struct {
		cfg         string
		in          string
		expectedOut string
		expectedErr bool
	}{
		{
			in: `
apiVersion: kpt.dev/v1alpha1
kind: Kptfile
metadata:
  name: cluster-exm01a
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cfg
data:
  x: y
`,
			cfg: `
data:
  flt1: |
    filters:
    - path:
      - kind
      value: Kptfile
`,
			expectedOut: `apiVersion: kpt.dev/v1alpha1
kind: Kptfile
metadata:
  name: cluster-exm01a
`,
		},
		{
			in: `
apiVersion: kpt.dev/v1alpha1
kind: Kptfile
metadata:
  name: cluster-exm02a
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cfg
data:
  x: y
`,
			cfg: `
data:
  flt1: |
    filters:
    - path:
      - kind
      value: Kptfile
      invertMatch: true
`,
			expectedOut: `apiVersion: v1
kind: ConfigMap
metadata:
  name: cfg
data:
  x: y
`,
		},
	}

	for i, ti := range tc {
		fcfg := Config{}
		err := yaml.Unmarshal([]byte(ti.cfg), &fcfg)
		if err != nil {
			t.Errorf("can't unmarshal config %s: %v. continue", ti.cfg, err)
			continue
		}

		nodes, err := (&kio.ByteReader{Reader: bytes.NewBufferString(ti.in)}).Read()
		if err != nil {
			t.Errorf("can't unmarshal in yamls %s: %v. continue", ti.in, err)
			continue
		}

		f, err := NewFilter(&fcfg)
		if err != nil {
			if !ti.expectedErr {
				t.Errorf("can't create filter for config %s: %v. continue", ti.in, err)
			}
			continue
		}
		nodes, err = f.Filter(nodes)
		if err != nil && !ti.expectedErr {
			t.Errorf("exec %d returned unexpected error %v for %s", i, err, ti.cfg)
			continue
		}
		out := &bytes.Buffer{}
		err = kio.ByteWriter{Writer: out}.Write(nodes)
		if err != nil {
			t.Errorf("write returned unexpected error %v for %s", err, ti.cfg)
			continue
		}
		if out.String() != ti.expectedOut {
			t.Errorf("expected %s, got %s", ti.expectedOut, out.String())
		}
	}

}
