// Copyright 2024 Nametag Inc.
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

package adclient

import (
	"encoding/json"
)

// GetADDomain returns domain information
func GetADDomain(s Client) (*Domain, error) {
	stdout, err := s.Execute("Get-ADDomain | ConvertTo-Json")
	if err != nil {
		return nil, err
	}

	var rv Domain
	if err := json.Unmarshal([]byte(stdout), &rv); err != nil {
		return nil, err
	}

	return &rv, nil
}

// Domain represents the domain response
type Domain struct {
	Forest      string `json:"Forest"`
	NetBIOSName string `json:"NetBIOSName"`
	DNSRoot     string `json:"DNSRoot"`
	Name        string `json:"Name"`
}
