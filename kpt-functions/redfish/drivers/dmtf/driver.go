// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dmtf

import (
	//"context"
	"crypto/tls"
	//"fmt"
	"net/http"
	//"strings"
	//"time"

	redfishAPI "opendev.org/airship/go-redfish/api"
	redfishClient "opendev.org/airship/go-redfish/client"

	"github.com/aodinokov/noctl-airship-poc/kpt-functions/redfish"
)

type Driver struct {
	Config *redfish.DriverConfig
	Api redfishAPI.RedfishAPI
}

func NewDriver(config *redfish.DriverConfig) (redfish.Driver, error) {
	drv := Driver{Config: config}

	cfg := redfishClient.NewConfiguration()
	if config.UserAgent != nil {
		cfg.UserAgent = *config.UserAgent
	}

	// see https://github.com/golang/go/issues/26013
	// We clone the default transport to ensure when we customize the transport
	// that we are providing it sane timeouts and other defaults that we would
	// normally get when not overriding the transport
	defaultTransportCopy := http.DefaultTransport.(*http.Transport) //nolint:errcheck
	transport := defaultTransportCopy.Clone()

	if config.DisableCertificateVerification {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec
		}
	}

	if config.IgnoreProxySetting {
		transport.Proxy = nil
	}

	cfg.HTTPClient = &http.Client{
		Transport: transport,
	}


	drv.Api = redfishClient.NewAPIClient(cfg).DefaultApi

	return &drv, nil
}

func (d *Driver) IsOnline() (bool, error) {
	return false, nil
}

func (d *Driver) SyncPower(online bool) error {
	return nil
}

func (d *Driver) Reboot() error {
	return nil
}

func (d *Driver) EjectMedia() error {
	return nil
}

func (d *Driver) SetBootSource() error {
	return nil
}
