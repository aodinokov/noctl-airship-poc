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
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	redfishAPI "opendev.org/airship/go-redfish/api"
	redfishClient "opendev.org/airship/go-redfish/client"

	"github.com/aodinokov/noctl-airship-poc/kpt-functions/redfish"
)

type Driver struct {
	DrvConfig *redfish.DriverConfig
	Config    *redfishClient.Configuration
	Api       redfishAPI.RedfishAPI
	SystemId  string
	mgrId     string
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
	return ResourceId(url, "/redfish/v1/Systems/")
}

func ManagerId(url *url.URL) (string, error) {
	return ResourceId(url /*TODO:*/, "")
}

func MediaId(url *url.URL) (string, error) {
	return ResourceId(url /*TODO:*/, "")
}

func (d *Driver) Init(config *redfish.DriverConfig) error {
	if d.DrvConfig != nil {
		return fmt.Errorf("Driver is already initialized")
	}

	url, err := url.Parse(config.BMC.URL)
	if err != nil {
		return err
	}

	cfg := redfishClient.NewConfiguration()

	cfg.BasePath, err = BasePath(url)
	if err != nil {
		return err
	}

	d.SystemId, err = SystemId(url)
	if err != nil {
		return err
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

	d.Config = cfg
	d.Api = redfishClient.NewAPIClient(cfg).DefaultApi

	return nil
}

func NewDriver(config *redfish.DriverConfig) (redfish.Driver, error) {
	drv := Driver{}

	err := drv.Init(config)
	if err != nil {
		return nil, err
	}
	return &drv, nil
}

func (d *Driver) IsOnline() (bool, error) {
	cs, err := d.GetSystem()
	if err != nil {
		return false, err
	}
	return (cs.PowerState == redfishClient.POWERSTATE_ON), nil
}

func (d *Driver) ResetSystemAndEnsurePowerState(resetType redfishClient.ResetType,
	desiredPowerState redfishClient.PowerState) error {
	cs, err := d.GetSystem()
	if err != nil {
		return err
	}

	if cs.PowerState == desiredPowerState {
		return nil
	}

	req := redfishClient.ResetRequestBody{ResetType: resetType}
	err = d.ResetSystem(&req)
	if err != nil {
		return err
	}
	return d.EnsurePowerState(desiredPowerState)
}

func (d *Driver) EnsurePowerState(desiredPowerState redfishClient.PowerState) error {
	// TODO: add pollingInterval and systemReationTimeout
	for retry := 0; retry <= 60; retry++ {
		cs, err := d.GetSystem()
		if err != nil {
			return err
		}
		if cs.PowerState == desiredPowerState {
			return nil
		}

		time.Sleep(time.Second)

	}
	return fmt.Errorf("system hasn't reached desired power state %v", desiredPowerState)
}

func (d *Driver) SyncPower(online bool) error {
	var err error
	if !online {
		err = d.ResetSystemAndEnsurePowerState(redfishClient.RESETTYPE_FORCE_OFF, redfishClient.POWERSTATE_OFF)
	} else {
		err = d.ResetSystemAndEnsurePowerState(redfishClient.RESETTYPE_ON, redfishClient.POWERSTATE_ON)
	}
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) Reboot() error {
	cs, err := d.GetSystem()
	if err != nil {
		return err
	}
	if cs.PowerState == redfishClient.POWERSTATE_OFF {
		return fmt.Errorf("can't reboot system that is off")
	}

	err = d.SyncPower(false)
	if err != nil {
		return err
	}
	err = d.SyncPower(true)
	if err != nil {
		return err
	}
	return nil
}

func (d *Driver) ManagerId() (string, error) {
	if d.mgrId != "" {
		return d.mgrId, nil
	}

	cs, err := d.GetSystem()
	if err != nil {
		return "", err
	}

	url, err := url.Parse(cs.Links.ManagedBy[0].OdataId)
	if err != nil {
		return "", err
	}

	m, err := ManagerId(url)
	if err != nil {
		return "", err
	}

	d.mgrId = m
	return d.mgrId, nil
}

func (d *Driver) SetVirtualMediaImage(image string) error {
	err := d.EjectAllVirtualMedia()
	if err != nil {
		return err
	}

	cs, err := d.GetSystem()
	if err != nil {
		return nil
	}

	mediaTypesPrioOrder := []string{}
	for _, bootSource := range cs.Boot.BootSourceOverrideTargetRedfishAllowableValues {
		if bootSource == redfishClient.BOOTSOURCE_CD {
			mediaTypesPrioOrder = append(mediaTypesPrioOrder, []string{"DVD", "CD"}...)
			break
		}
	}
	if len(mediaTypesPrioOrder) == 0 {
		fmt.Errorf("bootsource %v isn't allowed", redfishClient.BOOTSOURCE_CD)
	}

	applicableVirtualMedia := map[string][]string{}
	for _, mt := range mediaTypesPrioOrder {
		applicableVirtualMedia[mt] = nil
	}

	// search for mdeiaId that fits to our CD/DVD mediaTypes
	mc, err := d.ListManagerVirtualMedia()
	if err != nil {
		return err
	}
	for _, mediaURI := range mc.Members {
		url, err := url.Parse(mediaURI.OdataId)
		if err != nil {
			return err
		}
		mediaId, err := MediaId(url)
		if err != nil {
			return err
		}

		vm, err := d.GetManagerVirtualMedia(mediaId)
		if err != nil {
			return err
		}

		for _, mediaType := range vm.MediaTypes {
			if _, ok := applicableVirtualMedia[mediaType]; ok {
				applicableVirtualMedia[mediaType] = append(applicableVirtualMedia[mediaType], mediaId)
			}
		}
	}

	var (
		mediaId   string
		mediaType string
	)
	for _, mt := range mediaTypesPrioOrder {
		if (len(applicableVirtualMedia[mt])) > 0 {
			mediaType = mt
			mediaId = applicableVirtualMedia[mediaType][0]
			break
		}
	}
	if mediaId == "" || mediaType == "" {
		return fmt.Errorf("wasn't able to find media with allowed mediatypes %v", mediaTypesPrioOrder)
	}

	mr := redfishClient.InsertMediaRequestBody{
		Image:    image,
		Inserted: true,
	}
	err = d.InsertVirtualMedia(mediaId, &mr)
	if err != nil {
		return err
	}
	return nil
}

func (d *Driver) AdjustBootOrder() error {
	sr := redfishClient.ComputerSystem{}
	sr.Boot.BootSourceOverrideTarget = redfishClient.BOOTSOURCE_CD
	_, err := d.SetSystem(&sr)
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) EjectAllVirtualMedia() error {
	mc, err := d.ListManagerVirtualMedia()
	if err != nil {
		return err
	}

	for _, mediaURI := range mc.Members {
		url, err := url.Parse(mediaURI.OdataId)
		if err != nil {
			return err
		}
		mediaId, err := MediaId(url)
		if err != nil {
			return err
		}

		vm, err := d.GetManagerVirtualMedia(mediaId)
		if err != nil {
			return err
		}

		if vm.Inserted != nil && *vm.Inserted {
			err := d.EjectVirtualMedia(mediaId)
			if err != nil {
				return err
			}
		}

		err = d.EnsureVirtualMediaInserted(mediaId, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Driver) EnsureVirtualMediaInserted(mediaId string, desiredInsertedValue bool) error {
	// TODO: add pollingInterval and systemReationTimeout
	for retry := 0; retry <= 60; retry++ {
		vm, err := d.GetManagerVirtualMedia(mediaId)
		if err != nil {
			return err
		}
		if vm.Inserted != nil && *vm.Inserted == desiredInsertedValue {
			return nil
		}

		time.Sleep(time.Second)

	}
	return fmt.Errorf("system hasn't reached desired inserted value %v", desiredInsertedValue)
}

// api wrappers
func (d *Driver) GetSystem() (*redfishClient.ComputerSystem, error) {
	ctx := d.UpdateContext(context.Background())

	system, httpResp, err := d.Api.GetSystem(ctx, d.SystemId)
	err = ResponseError(httpResp, err)
	if err != nil {
		return nil, err
	}
	return &system, nil
}

func (d *Driver) ResetSystem(r *redfishClient.ResetRequestBody) error {
	ctx := d.UpdateContext(context.Background())

	_, httpResp, err := d.Api.ResetSystem(ctx, d.SystemId, *r)
	err = ResponseError(httpResp, err)
	if err != nil {
		return err
	}
	return nil
}

func (d *Driver) SetSystem(r *redfishClient.ComputerSystem) (*redfishClient.ComputerSystem, error) {
	ctx := d.UpdateContext(context.Background())

	system, httpResp, err := d.Api.SetSystem(ctx, d.SystemId, *r)
	err = ResponseError(httpResp, err)
	if err != nil {
		return nil, err
	}
	return &system, nil
}

func (d *Driver) ListManagerVirtualMedia() (*redfishClient.Collection, error) {
	mgrId, err := d.ManagerId()
	if err != nil {
		return nil, err
	}

	ctx := d.UpdateContext(context.Background())

	mc, httpResp, err := d.Api.ListManagerVirtualMedia(ctx, mgrId)
	err = ResponseError(httpResp, err)
	if err != nil {
		return nil, err
	}
	return &mc, nil
}

func (d *Driver) GetManagerVirtualMedia(mediaId string) (*redfishClient.VirtualMedia, error) {
	mgrId, err := d.ManagerId()
	if err != nil {
		return nil, err
	}

	ctx := d.UpdateContext(context.Background())

	vm, httpResp, err := d.Api.GetManagerVirtualMedia(ctx, mgrId, mediaId)
	err = ResponseError(httpResp, err)
	if err != nil {
		return nil, err
	}
	return &vm, nil
}

func (d *Driver) EjectVirtualMedia(mediaId string) error {
	mgrId, err := d.ManagerId()
	if err != nil {
		return err
	}

	ctx := d.UpdateContext(context.Background())

	_, httpResp, err := d.Api.EjectVirtualMedia(ctx, mgrId, mediaId, map[string]interface{}{})
	err = ResponseError(httpResp, err)
	if err != nil {
		return err
	}
	return nil
}

func (d *Driver) InsertVirtualMedia(mediaId string, r *redfishClient.InsertMediaRequestBody) error {
	mgrId, err := d.ManagerId()
	if err != nil {
		return err
	}

	ctx := d.UpdateContext(context.Background())

	_, httpResp, err := d.Api.InsertVirtualMedia(ctx, mgrId, mediaId, *r)
	err = ResponseError(httpResp, err)
	if err != nil {
		return err
	}
	return nil
}

func (d *Driver) UpdateContext(ctx context.Context) context.Context {
	if d.DrvConfig.BMC.Username != "" && d.DrvConfig.BMC.Password != "" {
		ctx = context.WithValue(
			ctx,
			redfishClient.ContextBasicAuth,
			redfishClient.BasicAuth{
				UserName: d.DrvConfig.BMC.Username,
				Password: d.DrvConfig.BMC.Password},
		)
	}
	return ctx
}

// Error provides a detailed error message for end user consumption by inspecting all Redfish client
// responses and errors.
func ResponseError(httpResp *http.Response, clientErr error) error {
	if httpResp == nil {
		return fmt.Errorf("HTTP request failed. Redfish may be temporarily unavailable. Please try again.")
	}

	// NOTE(drewwalters96): The error, clientErr, may not be nil even though the request was successful. The HTTP
	// status code is the most reliable way to determine the result of a Redfish request using the go-redfish
	// library. The Redfish client uses HTTP codes 200 and 204 to indicate success.
	var finalError error
	switch httpResp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		finalError = fmt.Errorf("System not found. Correct the system name and try again.")
	case http.StatusBadRequest:
		finalError = fmt.Errorf("Invalid request. Verify the system name and try again.")
	case http.StatusMethodNotAllowed:
		finalError = fmt.Errorf("%s. BMC returned status '%s'.",
			"This operation is likely unsupported by the BMC Redfish version, or the BMC is busy",
			httpResp.Status)
	default:
		finalError = fmt.Errorf("BMC responded '%s'.", httpResp.Status)
		log.Printf("BMC responded '%s'. Attempting to unmarshal the raw BMC error response.", httpResp.Status)
	}

	/*
		TODO:
			// Retrieve the raw HTTP response body
			oAPIErr, ok := clientErr.(redfishClient.GenericOpenAPIError)
			if !ok {
				log.Print("Unable to decode BMC response.")
			}

			// Attempt to decode the BMC response from the raw HTTP response
			if bmcResponse, err := DecodeRawError(oAPIErr.Body()); err == nil {
				finalError = fmt.Errorf("%s BMC responded: '%s'", finalError.Message, bmcResponse)
			} else {
				log.Printf("Unable to decode BMC response. %q", err)
			}
	*/

	return finalError
}

/*
TODO:
// DecodeRawError decodes a raw Redfish HTTP response and retrieves the extended information and available resolutions
// returned by the BMC.
func DecodeRawError(rawResponse []byte) (string, error) {
	processExtendedInfo := func(extendedInfo map[string]interface{}) (string, error) {
		message, ok := extendedInfo["Message"]
		if !ok {
			return "", ErrUnrecognizedRedfishResponse{Key: "error.@Message.ExtendedInfo.Message"}
		}

		messageContent, ok := message.(string)
		if !ok {
			return "", ErrUnrecognizedRedfishResponse{Key: "error.@Message.ExtendedInfo.Message"}
		}

		// Resolution may be omitted in some responses
		if resolution, ok := extendedInfo["Resolution"]; ok {
			return fmt.Sprintf("%s %s", messageContent, resolution), nil
		}

		return messageContent, nil
	}

	// Unmarshal raw Redfish response as arbitrary JSON map
	var arbitraryJSON map[string]interface{}
	if err := json.Unmarshal(rawResponse, &arbitraryJSON); err != nil {
		return "", ErrUnrecognizedRedfishResponse{Key: "error"}
	}

	errObject, ok := arbitraryJSON["error"]
	if !ok {
		return "", ErrUnrecognizedRedfishResponse{Key: "error"}
	}

	errContent, ok := errObject.(map[string]interface{})
	if !ok {
		return "", ErrUnrecognizedRedfishResponse{Key: "error"}
	}

	extendedInfoContent, ok := errContent["@Message.ExtendedInfo"]
	if !ok {
		return "", ErrUnrecognizedRedfishResponse{Key: "error.@Message.ExtendedInfo"}
	}

	// NOTE(drewwalters96): The official specification dictates that "@Message.ExtendedInfo" should be a JSON array;
	// however, some BMCs have returned a single JSON dictionary. Handle both types here.
	switch extendedInfo := extendedInfoContent.(type) {
	case []interface{}:
		if len(extendedInfo) == 0 {
			return "", ErrUnrecognizedRedfishResponse{Key: "error.@MessageExtendedInfo"}
		}

		var errorMessage string
		for _, info := range extendedInfo {
			infoContent, ok := info.(map[string]interface{})
			if !ok {
				return "", ErrUnrecognizedRedfishResponse{Key: "error.@Message.ExtendedInfo"}
			}

			message, err := processExtendedInfo(infoContent)
			if err != nil {
				return "", err
			}

			errorMessage = fmt.Sprintf("%s\n%s", message, errorMessage)
		}

		return errorMessage, nil
	case map[string]interface{}:
		return processExtendedInfo(extendedInfo)
	default:
		return "", ErrUnrecognizedRedfishResponse{Key: "error.@Message.ExtendedInfo"}
	}
}
*/
