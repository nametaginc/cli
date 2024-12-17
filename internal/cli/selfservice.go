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
	"fmt"
	"time"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/nametaginc/cli/internal/api"
)

func newSelfServiceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "self-service",
		Short: "Commands for managing self-service microsite",
	}
	cmd.AddCommand(newSelfServicePresignCmd())
	return cmd
}

func newSelfServicePresignCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "presign",
		Short: "Generate a presigned self-service microsite URL",
		RunE: func(cmd *cobra.Command, args []string) error {
			envID, err := cmd.Flags().GetString("env")
			if err != nil || envID == "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "the --env flag is required\n")
				return fmt.Errorf("the --env flag is required")
			}

			client, err := NewAPIClient(cmd)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "cannot initialize API client: %s\n", err)
				return err
			}

			req := api.RecoveryMicrositePresignRequest{}
			if v, err := cmd.Flags().GetString("email"); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "cannot read email: %s\n", err)
				return err
			} else if v != "" {
				req.Email = &v
			} else {
				req.Email = lo.ToPtr("*")
			}
			if v, err := cmd.Flags().GetStringArray("operation"); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "cannot read operation: %s\n", err)
				return err
			} else if len(v) > 0 {
				req.Operations = lo.ToPtr(make([]api.RecoveryMicrositeOperation, 0, len(v)))
				for _, opStr := range v {
					op := api.RecoveryMicrositeOperation(opStr)
					*req.Operations = append(*req.Operations, op)
				}
			}
			if v, err := cmd.Flags().GetStringArray("directory"); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "cannot read directory: %s\n", err)
				return err
			} else if len(v) > 0 {
				req.Directories = &v
			}
			if v, err := cmd.Flags().GetString("flow"); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "cannot read flow: %s\n", err)
				return err
			} else {
				req.Flow = lo.ToPtr(api.RecoveryMicrositeFlow(v))
			}
			if v, err := cmd.Flags().GetDuration("ttl"); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "cannot read ttl: %s\n", err)
				return err
			} else if v > 0 {
				req.Ttl = v
			} else {
				req.Ttl = time.Hour
			}

			resp, err := client.PresignRecoveryMicrositeURLWithResponse(cmd.Context(), envID,
				req)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "cannot initialize generate presigned URL: %s\n", err)
				return err
			}
			if resp.StatusCode() != 200 {
				fmt.Fprintf(cmd.ErrOrStderr(), "cannot initialize generate presigned URL: %s\n", resp.Status())
				return fmt.Errorf("%s", resp.Status())
			}
			fmt.Fprintln(cmd.OutOrStdout(), resp.JSON200.Url)
			return nil
		},
	}
	cmd.Flags().StringP("env", "e", "", "The `environment` identifier of the self-service site.")
	cmd.Flags().String("email", "", "The `identifier` of the account to recover, or '*' to allow recovery for any account.")
	cmd.Flags().StringArray("operation", []string{},
		"Which `operation`s to allow for recovery. If not specified, all operations are allowed. "+
			"One of 'mfa', 'password', 'unlock', or 'temporary-access-pass'")
	cmd.Flags().StringArray("directory", []string{},
		"The `directory` to allow for recovery. If not specified, all directories are allowed. Can be specified multiple times to allow multiple directories.")
	cmd.Flags().String("flow", "",
		"Which `flow` to use for recovery. If not specified, the default flow 'recover' is used. "+
			"One of 'recover' or 'enroll'")
	cmd.Flags().Duration("ttl", 0, "How long the presigned URL should be valid. "+
		"If not specified, the default of 1 hour is used.")
	return cmd
}
