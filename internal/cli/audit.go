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

package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/spf13/cobra"
)

func newAuditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audit",
		Short: "View audit logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, err := url.Parse(getServer(cmd))
			if err != nil {
				return err
			}
			wsURL.Scheme = "wss"
			wsURL.Path = "/api/audit"

			authToken, err := getAuthToken(cmd)
			if err != nil {
				return err
			}

			conn, resp, err := websocket.Dial(cmd.Context(), wsURL.String(), &websocket.DialOptions{
				HTTPHeader: http.Header{
					"Authorization": []string{"Bearer " + authToken},
				},
			})
			if err != nil {
				if resp != nil {
					return fmt.Errorf("websocket connection failed: %s", resp.Status)
				}
				return fmt.Errorf("websocket connection failed: %w", err)
			}

			for {
				var msg map[string]interface{}
				if err := wsjson.Read(cmd.Context(), conn, &msg); err != nil {
					_ = conn.Close(websocket.StatusAbnormalClosure, err.Error())
					return err
				}

				msgBuf, err := json.Marshal(msg)
				if err != nil {
					_ = conn.Close(websocket.StatusAbnormalClosure, err.Error())
					return err
				}

				_, err = cmd.OutOrStdout().Write(append(msgBuf, '\n'))
				if err != nil {
					return err
				}
			}
		},
	}
	return cmd
}
