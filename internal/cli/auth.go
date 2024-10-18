// Copyright 2024 Nametag Inc.
//
// All information contained herein is the property of Nametag Inc.. The
// intellectual and technical concepts contained herein are proprietary, trade
// secrets, and/or confidential to Nametag, Inc. and may be covered by U.S.
// and Foreign Patents, patents in process, and are protected by trade secret or
// copyright law. Reproduction or distribution, in whole or in part, is
// forbidden except by express written permission of Nametag, Inc.

package cli

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nametaginc/cli/internal/api"
)

func init() {
	Root.PersistentFlags().StringP("access-token", "t", "", "Nametag API access token")
	Root.PersistentFlags().String("server", "", "Nametag server uri")
	Root.PersistentFlags().MarkHidden("server")
}

func getServerURL(cmd *cobra.Command) (string, error) {
	server, err := cmd.Flags().GetString("server")
	if err != nil {
		return "", err
	}
	if server == "" {
		server = os.Getenv("NAMETAG_SERVER")
	}
	if server == "" {
		server = "https://nametag.co"
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return "", fmt.Errorf("cannot parse nametag server url: %w", err)
	}
	switch {
	case serverURL.Host == "nametag.co":
	// ok
	case serverURL.Host == "nametagstaging.com":
	// ok
	case strings.HasSuffix(serverURL.Host, ".nametagdev.com"):
	// ok
	default:
		return "", fmt.Errorf("invalid nametag server url")
	}

	serverURL = &url.URL{
		Scheme: "https",
		Host:   serverURL.Host,
	}
	return serverURL.String(), nil
}

func GetAPIConfiguration(cmd *cobra.Command) (server string, authToken string, err error) {
	authToken, err = cmd.Flags().GetString("access-token")
	if err != nil {
		return "", "", err
	}
	if authToken == "" {
		authToken = os.Getenv("NAMETAG_AUTH_TOKEN")
	}
	if authToken == "" {
		return "", "", fmt.Errorf("You must specify an authentication token")
	}

	server, err = getServerURL(cmd)
	if err != nil {
		return "", "", err
	}
	return authToken, server, nil
}

func NewAPIClient(cmd *cobra.Command) (*api.ClientWithResponses, error) {
	server, authToken, err := GetAPIConfiguration(cmd)
	if err != nil {
		return nil, err
	}

	client, err := api.NewClientWithResponses(server, api.WithRequestEditorFn(
		func(ctx context.Context, req *http.Request) error {
			req.Header.Add("Authorization", "Bearer "+authToken)
			return nil
		}))
	if err != nil {
		return nil, err
	}

	return client, nil
}
