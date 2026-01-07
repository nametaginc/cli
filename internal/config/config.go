// Copyright 2025 Nametag Inc.
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

// Package config is for the nametag cli config
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

// Config represents the format of the configuration file that
// contains the authentication token and other settings.
type Config struct {
	Version    string     `yaml:"version"`
	Server     string     `yaml:",omitempty"`
	Token      string     `yaml:"token"`
	LDAPConfig LDAPConfig `yaml:"LDAPConfig"`
}

// LDAPConfig represents the format of settings related to the LDAP agent functionality
type LDAPConfig struct {
	LDAPUrl                 string `yaml:"ldapURL"`
	BaseDN                  string `yaml:"baseDN"`
	BindDN                  string `yaml:"bindDN"`
	BindPassword            string `yaml:"bindPassword"`
	PageSize                uint32 `yaml:"pageSize"`
	DefaultPasswordPolicyDN string `yaml:"defaultPasswordPolicyDN"`
}

var (
	cachedConfig *Config
)

// ReadConfig returns the config from the file system
func ReadConfig(cmd *cobra.Command) (*Config, error) {
	if cachedConfig != nil {
		return cachedConfig, nil
	}
	path, err := GetPath(cmd)
	if err != nil {
		return nil, err
	}
	configBuf, err := os.ReadFile(path) //nolint:gosec  // no file inclusion vulnerability; this is a client side application where it's okay to specify a path
	if os.IsNotExist(err) {
		return &Config{}, nil
	} else if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(configBuf, &config); err != nil {
		return nil, fmt.Errorf("configuration file is not valid: %s: %w", path, err)
	}
	if config.Version != "1" {
		return nil, fmt.Errorf("unsupported configuration file version: %s", config.Version)
	}

	cachedConfig = &config
	return &config, nil
}

// GetPath returns the config path
func GetPath(cmd *cobra.Command) (string, error) {
	param, err := cmd.Flags().GetString("config")
	if err != nil {
		return "", err
	}
	if param != "" {
		return param, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configPath := filepath.Join(home, ".config", "nametag", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		return "", err
	}
	return configPath, nil
}

// ClearCachedConfig clears the cached config
func ClearCachedConfig() {
	cachedConfig = nil
}
