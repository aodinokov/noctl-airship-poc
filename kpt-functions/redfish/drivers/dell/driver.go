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

// Package dell wraps the standard Redfish client in order to provide additional functionality required to perform
// actions on iDRAC servers.
package dell

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	redfishClient "opendev.org/airship/go-redfish/client"

	"github.com/aodinokov/noctl-airship-poc/kpt-functions/redfish"
	"github.com/aodinokov/noctl-airship-poc/kpt-functions/redfish/drivers/dmtf"
)

// Dell specific part of client API
const (
	// ClientType is used by other packages as the identifier of the Redfish client.
	vCDBootRequestBody = `{
	    "ShareParameters": {
	        "Target": "ALL"
	    },
	    "ShutdownType": "NoReboot",
	    "ImportBuffer": "<SystemConfiguration>
	                       <Component FQDD=\"iDRAC.Embedded.1\">
	                         <Attribute Name=\"ServerBoot.1#BootOnce\">Enabled</Attribute>
	                         <Attribute Name=\"ServerBoot.1#FirstBootDevice\">VCD-DVD</Attribute>
	                       </Component>
	                     </SystemConfiguration>"
	}`
)

type iDRACAPIRespErr struct {
	Err iDRACAPIErr `json:"error"`
}

type iDRACAPIErr struct {
	ExtendedInfo []iDRACAPIExtendedInfo `json:"@Message.ExtendedInfo"`
	Code         string                 `json:"code"`
	Message      string                 `json:"message"`
}

type iDRACAPIExtendedInfo struct {
	Message    string `json:"Message"`
	Resolution string `json:"Resolution,omitempty"`
}

type Driver struct {
	dmtf.Driver
	BasePath string
}

func (d *Driver) ImportManagerSystemConfigurationForVCDDVD(managerId string) error {
	ctx := d.UpdateContext(context.Background())
	// NOTE(drewwalters96): Setting the boot device to a virtual media type requires an API request to the iDRAC
	// actions API. The request is made below using the same HTTP client used by the Redfish API and exposed by the
	// standard airshipctl Redfish client. Only iDRAC 9 >= 3.3 is supports this endpoint.
	url := fmt.Sprintf("%s/redfish/v1/Managers/%s/Actions/Oem/EID_674_Manager.ImportSystemConfiguration",
		d.BasePath,
		managerId)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(vCDBootRequestBody))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	if auth, ok := ctx.Value(redfishClient.ContextBasicAuth).(redfishClient.BasicAuth); ok {
		req.SetBasicAuth(auth.UserName, auth.Password)
	}

	httpResp, err := d.Config.HTTPClient.Do(req)
	if httpResp.StatusCode != http.StatusAccepted {
		body, ok := ioutil.ReadAll(httpResp.Body)
		if ok != nil {
			return fmt.Errorf("unable to set boot device. Malformed iDRAC response.")
		}
		var iDRACResp iDRACAPIRespErr
		ok = json.Unmarshal(body, &iDRACResp)
		if ok != nil {
			return fmt.Errorf("unable to set boot device. Can't unmarshal iDRAC response.")
		}
		return fmt.Errorf("unable to set boot device. %s", iDRACResp.Err.ExtendedInfo[0])
	} else if err != nil {
		return fmt.Errorf("Unable to set boot device. %v", err)
	}
	defer httpResp.Body.Close()
	return nil
}

// Overriding dmtf AdjustBootOrder fn
func (d *Driver) AdjustBootOrder() error {
	mgrId, err := d.ManagerId()
	if err != nil {
		return err
	}
	return d.ImportManagerSystemConfigurationForVCDDVD(mgrId)
}

func NewDriver(config *redfish.DriverConfig) (redfish.Driver, error) {
	d := Driver{}

	url, err := url.Parse(config.BMC.URL)
	if err != nil {
		return nil, err
	}

	d.BasePath, err = dmtf.BasePath(url)
	if err != nil {
		return nil, err
	}

	err = d.Driver.Init(config)
	if err != nil {
		return nil, err
	}

	return &d, nil
}
