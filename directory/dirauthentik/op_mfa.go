// Copyright 2026 Nametag Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dirauthentik

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/nametaginc/cli/diragentapi"
	"github.com/nametaginc/cli/directory"
)

func (p *Provider) performOperationRemoveAllMfa(ctx context.Context, req diragentapi.DirAgentPerformOperationRequest) (*diragentapi.DirAgentPerformOperationResponse, error) {
	user, err := p.lookupUserByImmutableID(ctx, req.AccountImmutableID)
	if err != nil {
		return nil, err
	}

	devices, err := p.fetchUserDevices(ctx, user.PK)
	if err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		return nil, directory.CodedError{
			Code:    diragentapi.UnsupportedAccountState,
			Message: "no MFA factors to remove",
		}
	}

	if req.DryRun != nil && *req.DryRun {
		return &diragentapi.DirAgentPerformOperationResponse{}, nil
	}

	for _, device := range devices {
		path, err := deviceDeletePath(device)
		if err != nil {
			return nil, err
		}
		if err := p.doJSON(ctx, http.MethodDelete, path, nil, nil, nil); err != nil {
			return nil, err
		}
	}

	return &diragentapi.DirAgentPerformOperationResponse{}, nil
}

func (p *Provider) fetchUserDevices(ctx context.Context, userPK int) ([]apiDevice, error) {
	query := url.Values{}
	query.Set("user", strconv.Itoa(userPK))

	var devices []apiDevice
	if err := p.doJSON(ctx, http.MethodGet, "authenticators/admin/all/", query, nil, &devices); err != nil {
		return nil, err
	}
	return devices, nil
}

func deviceDeletePath(device apiDevice) (string, error) {
	if device.PK == "" {
		return "", fmt.Errorf("authentik: device missing identifier")
	}
	segment, ok := deviceTypeSegment(device)
	if !ok {
		return "", fmt.Errorf("authentik: unsupported device type %q", device.Type)
	}
	return fmt.Sprintf("authenticators/admin/%s/%s/", segment, device.PK), nil
}

func deviceTypeSegment(device apiDevice) (string, bool) {
	rawType := strings.ToLower(strings.TrimSpace(device.Type))
	if rawType == "" {
		rawType = strings.ToLower(strings.TrimSpace(device.MetaModelName))
	}

	switch rawType {
	case "totp", "totpdevice":
		return "totp", true
	case "webauthn", "webauthndevice":
		return "webauthn", true
	case "sms", "smsdevice":
		return "sms", true
	case "email", "emaildevice":
		return "email", true
	case "static", "staticdevice":
		return "static", true
	case "duo", "duodevice":
		return "duo", true
	case "endpoint", "endpointdevice", "googleendpointdevice":
		return "endpoint", true
	case "authentik_stages_authenticator_totp.totpdevice":
		return "totp", true
	case "authentik_stages_authenticator_webauthn.webauthndevice":
		return "webauthn", true
	case "authentik_stages_authenticator_sms.smsdevice":
		return "sms", true
	case "authentik_stages_authenticator_email.emaildevice":
		return "email", true
	case "authentik_stages_authenticator_static.staticdevice":
		return "static", true
	case "authentik_stages_authenticator_duo.duodevice":
		return "duo", true
	case "authentik_stages_authenticator_endpoint.googleendpointdevice":
		return "endpoint", true
	default:
		return "", false
	}
}
