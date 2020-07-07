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
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	//"time"

	redfishAPI "opendev.org/airship/go-redfish/api"
	redfishClient "opendev.org/airship/go-redfish/client"

	"github.com/aodinokov/noctl-airship-poc/kpt-functions/redfish"
)

type Driver struct {
	Config   *redfish.DriverConfig
	Api      redfishAPI.RedfishAPI
	SystemId string
}

func BasePath(url *url.URL) (string, error) {
	var scheme string

	// for possible options
	// see https://github.com/metal3-io/baremetal-operator/blob/master/docs/api.md#spec-fields
	switch url.Scheme {
	case "redfish", "https", "redfish+https":
		scheme = "https"
	case "redfish+http", "http":
		scheme = "http"
	default:
		return "", fmt.Errorf("the scheme %s isn't supported", url.Scheme)
	}
	return fmt.Sprintf("%s://%s", scheme, url.Host), nil
}

func ResourceId(url *url.URL, expectedPath string) (string, error) {
	if expectedPath != "" {
		if strings.TrimSuffix(expectedPath, "/") != path.Dir(strings.TrimSuffix(url.Path, "/")) {
			return "", fmt.Errorf("Invalid resource type: expected %s, got %s",
				strings.TrimSuffix(expectedPath, "/"),
				path.Dir(strings.TrimSuffix(url.Path, "/")))
		}
	}
	return path.Base(url.Path), nil
}

func SystemId(url *url.URL) (string, error) {
	return ResourceId(url, "redfish/v1/Systems/")
}

func NewDriver(config *redfish.DriverConfig) (redfish.Driver, error) {
	drv := Driver{Config: config}

	url, err := url.Parse(config.BMC.URL)
	if err != nil {
		return nil, err
	}

	cfg := redfishClient.NewConfiguration()

	cfg.BasePath, err = BasePath(url)
	if err != nil {
		return nil, err
	}

	drv.SystemId, err = SystemId(url)
	if err != nil {
		return nil, err
	}

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

func (d *Driver) updateContext(ctx context.Context) context.Context {
	if d.Config.BMC.Username != "" && d.Config.BMC.Password != "" {
		ctx = context.WithValue(
			ctx,
			redfishClient.ContextBasicAuth,
			redfishClient.BasicAuth{
				UserName: d.Config.BMC.Username,
				Password: d.Config.BMC.Password},
		)
	}
	return ctx
}
