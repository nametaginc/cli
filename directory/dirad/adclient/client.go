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

// Package adclient interacts with AD domain controllers
package adclient

import (
	"fmt"

	"github.com/bhendo/go-powershell"
	"github.com/bhendo/go-powershell/backend"
)

var defaultPageSize = 250

// Client is the interface adapted by the client used for AD connection
type Client interface {
	Execute(cmd string) (string, error)
	Close() error
}

// PowershellClient is used for invoking commands in an underlying powershell process.
type PowershellClient struct {
	ps powershell.Shell
}

// New returns a new client
func New() (Client, error) {
	ps, err := powershell.New(&backend.Local{})
	if err != nil {
		return nil, err
	}

	// The underlying powershell is persistent. Load the required modules for subsequent commands.
	_, _, err = ps.Execute("Import-Module ActiveDirectory")
	if err != nil {
		return nil, err
	}

	return &PowershellClient{ps: ps}, nil
}

// Execute runs the cmd
func (s *PowershellClient) Execute(cmd string) (string, error) {
	stdout, _, err := s.ps.Execute(cmd)
	return stdout, err
}

// Close terminates the connection
func (s *PowershellClient) Close() error {
	s.ps.Exit()
	return nil
}

// MockClient is used for test cases. You can define the response for a query to powershell here.
type MockClient struct {
	ResponseMap map[string]Response
}

// Response indicates the return struct
type Response struct {
	Stdout string
	Err    error
}

// Execute runs the command
func (s *MockClient) Execute(cmd string) (string, error) {
	if s.ResponseMap == nil {
		return "", fmt.Errorf("no mock responses defined")
	}

	var response Response
	var ok bool
	if response, ok = s.ResponseMap[cmd]; !ok {
		return "", fmt.Errorf("no mock responses defined for command: %s", cmd)
	}

	if response.Err != nil {
		return "", response.Err
	}

	return response.Stdout, nil
}

// Close terminates the connection
func (s *MockClient) Close() error {
	return nil
}
