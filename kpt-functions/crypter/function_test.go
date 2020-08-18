package crypter

import (
	"testing"

	"bytes"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestCrypter(t *testing.T) {

	tc := []struct {
		cfg         string
		in          string
		expectedOut string
		expectedErr bool
	}{
		{
			cfg: `
password: testpass
refs:
- objref:
    kind: VariableCatalogue
    name: host-catalogue
  fieldrefs:
  - hosts.m3.node01.bmcPassword
  - hosts.m3.node02.bmcPassword
`,
			in: `
apiVersion: airshipit.org/v1alpha1
kind: VariableCatalogue
metadata:
  name: host-catalogue
hosts:
  m3:
    node01:
      macAddress: 52:54:00:b6:ed:31
      bmcAddress: redfish+http://10.23.25.1:8000/redfish/v1/Systems/air-target-1
      bmcUsername: root
      bmcPassword: Z0FBQUFBQmZPMWthSGVxU18tWmdTS242eTBkUWQ5UVhfMEt2bEhzSEFJdDNTRUl2eXpDY0N0MWJxODFpRDlkbnBRU05qZ2o3cjBxRDIyRExSaGZCX2pkMFJkbnI5SzdZMnc9PQ==
      ipAddresses:
        oam-ipv4: 10.23.25.102
        pxe-ipv4: 10.23.24.102
      macAddresses:
        oam: 52:54:00:9b:27:4c
        pxe: 52:54:00:b6:ed:31
    node02:
      macAddress: 00:3b:8b:0c:ec:8b
      bmcAddress: redfish+http://10.23.25.2:8000/redfish/v1/Systems/air-target-2
      bmcUsername: username
      bmcPassword: Z0FBQUFBQmZPMWp0anFJcnpDNElXa2dEOFZOd19DNDNiODdPSlF3UVNERFk2cjdydmFpU3BEckxpVkY3S2VUVmpjQUpEdUZMT2x3RjQ1NnBYa2p5cFpKX1dHY1B3UFVmQVE9PQ==
      ipAddresses:
        oam-ipv4: 10.23.25.101
        pxe-ipv4: 10.23.24.101
`,
			expectedOut: `
apiVersion: airshipit.org/v1alpha1
kind: VariableCatalogue
metadata:
  name: host-catalogue
hosts:
  m3:
    node01:
      macAddress: 52:54:00:b6:ed:31
      bmcAddress: redfish+http://10.23.25.1:8000/redfish/v1/Systems/air-target-1
      bmcUsername: root
      bmcPassword: r00tme
      ipAddresses:
        oam-ipv4: 10.23.25.102
        pxe-ipv4: 10.23.24.102
      macAddresses:
        oam: 52:54:00:9b:27:4c
        pxe: 52:54:00:b6:ed:31
    node02:
      macAddress: 00:3b:8b:0c:ec:8b
      bmcAddress: redfish+http://10.23.25.2:8000/redfish/v1/Systems/air-target-2
      bmcUsername: username
      bmcPassword: password
      ipAddresses:
        oam-ipv4: 10.23.25.101
        pxe-ipv4: 10.23.24.101
`,
		},
	}

	for i, ti := range tc {
		fcfg := FunctionConfig{}
		err := yaml.Unmarshal([]byte(ti.cfg), &fcfg)
		if err != nil {
			t.Errorf("can't unmarshal config %s\nerr: %v continue", ti.cfg, err)
			continue
		}

		nodes, err := (&kio.ByteReader{Reader: bytes.NewBufferString(ti.in)}).Read()
		if err != nil {
			t.Errorf("can't unmarshal config %s. continue", ti.in)
			continue
		}

		f := Function{Config: &fcfg}
		nodes, err = f.Exec(nodes)
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
		if out.String() != ti.expectedOut[1:] {
			t.Errorf("expected %s, got %s", ti.expectedOut, out.String())
		}
	}

}
