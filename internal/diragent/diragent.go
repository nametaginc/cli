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

// Package diragent implements the client part of the directory
// agent protocol. It handles connecting to the server via websocket,
// authentication, and relaying messages from the server to a child
// process (implemented by RunWorker) and back.
package diragent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/jpillora/backoff"
	"github.com/samber/lo"

	"github.com/nametaginc/cli/diragentapi"
)

// Service runs the parent process for the directory agent. It connects to the
// server and relays messages between the server (via websocket) and the child process
// (via stdin/stdout).
type Service struct {
	Server           string
	AuthToken        string
	DirID            string
	Command          string
	Env              map[string]string
	Stderr           io.Writer
	HTTPClient       *http.Client
	cmd              *exec.Cmd
	cmdStdin         io.WriteCloser
	cmdStdout        io.ReadCloser
	cmdStdinEncoder  *json.Encoder
	cmdStdoutDecoder *json.Decoder
}

// Run runs the directory agent service. It connects to the server
// and relays messages. If the connection fails, it retries. It returns
// when the parent ctx is closed.
func (s *Service) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	var err error
	if runtime.GOOS == "windows" {
		s.cmd = exec.CommandContext(ctx, "cmd", "/c", s.Command) //nolint:gosec
	} else {
		s.cmd = exec.CommandContext(ctx, "/bin/sh", "-c", s.Command) //nolint:gosec
	}
	s.cmd.Env = append(os.Environ(), "NAMETAG_AGENT_WORKER=true")
	for k, v := range s.Env {
		s.cmd.Env = append(s.cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	s.cmdStdin, err = s.cmd.StdinPipe()
	if err != nil {
		return err
	}
	s.cmdStdout, err = s.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	s.cmd.Stderr = s.Stderr

	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("cannot start worker command: %w", err)
	}
	go func() {
		cancel(s.cmd.Wait())
	}()
	defer func() {
		// make sure the process is reaped / killed if we disconnect
		if p := s.cmd.Process; p != nil {
			_ = p.Kill()
		}
	}()

	s.cmdStdinEncoder = json.NewEncoder(s.cmdStdin)
	s.cmdStdoutDecoder = json.NewDecoder(s.cmdStdout)

	// test the subcommand
	if err := s.cmdStdinEncoder.Encode(diragentapi.DirAgentRequest{
		Configure: &diragentapi.DirAgentConfigureRequest{},
	}); err != nil {
		return err
	}
	var resp diragentapi.DirAgentResponse
	if err := s.cmdStdoutDecoder.Decode(&resp); err != nil {
		return err
	}
	if resp.Error != nil {
		return fmt.Errorf("error: %s %s", resp.Error.Code, resp.Error.Message)
	}

	bo := backoff.Backoff{
		Min: time.Second,
		Max: time.Minute,
	}
	for {
		innerCtx, cancel := context.WithCancel(ctx)
		runStartTime := time.Now()
		err := s.runOnce(innerCtx)
		runDuration := time.Since(runStartTime)
		cancel()
		if err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				return nil
			}
			return ctx.Err()
		default:
			// ok
		}

		if runDuration > time.Minute {
			bo.Reset()
		}
		sleepTime := bo.Duration()
		log.Printf("ERROR: %s (will retry in %s)", err, sleepTime)
		time.Sleep(sleepTime)
	}
}

func (s *Service) runOnce(ctx context.Context) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	connectURL, err := url.Parse(s.Server)
	if err != nil {
		return fmt.Errorf("cannot parse server url %q: %w", s.Server, err)
	}
	connectURL.Path = "/api/diragent"
	connectURL.RawQuery = url.Values{
		"auth": {s.AuthToken},
	}.Encode()

	redactedConnectQuery := connectURL.Query()
	redactedConnectQuery.Set("auth", strings.Repeat("**", len(s.AuthToken)))
	redactedConnectURL := *connectURL
	redactedConnectURL.RawQuery = redactedConnectQuery.Encode()

	conn, wsResp, err := websocket.Dial(ctx, connectURL.String(), &websocket.DialOptions{HTTPClient: s.HTTPClient})
	if err != nil {
		return fmt.Errorf("cannot connect to server %q: %w", redactedConnectURL.String(), err)
	}
	if wsResp.StatusCode >= 400 {
		return fmt.Errorf("cannot connect to server %q: %d %s", redactedConnectURL.String(), wsResp.StatusCode, wsResp.Status)
	}
	log.Printf("connected to %s", redactedConnectURL.String())

	for {
		req := diragentapi.DirAgentRequest{}
		if err := wsjson.Read(ctx, conn, &req); err != nil {
			cancel(err)
			_ = conn.Close(websocket.StatusAbnormalClosure, err.Error())
			return err
		}
		switch {
		case req.Configure != nil:
			log.Printf("configure")
		case req.GetAccount != nil:
			log.Printf("get_account %s %s",
				lo.FromPtr(req.GetAccount.Ref.ImmutableID),
				lo.FromPtr(req.GetAccount.Ref.ID))
		case req.ListAccounts != nil:
			log.Printf("list_accounts")
		case req.ListGroups != nil:
			if req.ListGroups.NamePrefix != nil {
				log.Printf("list_groups starting with %s", *req.ListGroups.NamePrefix)
			} else {
				log.Printf("list_groups")
			}
		case req.PerformOperation != nil:
			log.Printf("perform_operation %s on %s%s",
				string(req.PerformOperation.Operation),
				req.PerformOperation.AccountImmutableID,
				lo.If(lo.FromPtr(req.PerformOperation.DryRun), " (dry run)").Else(""))
		case req.Ping != nil:
			log.Printf("ping")
		}

		if err := s.cmdStdinEncoder.Encode(req); err != nil {
			cancel(err)
			return err
		}

		var resp diragentapi.DirAgentResponse
		if err := s.cmdStdoutDecoder.Decode(&resp); err != nil {
			cancel(err)
			return err
		}

		// validate command output
		if resp.Error == nil {
			switch {
			case req.Configure != nil:
				if resp.Configure == nil {
					resp.Error = &diragentapi.DirAgentErrorResponse{
						Code:    diragentapi.InternalError,
						Message: "command must set 'configure' in response",
					}
				}
			case req.GetAccount != nil:
				if resp.GetAccount == nil {
					resp.Error = &diragentapi.DirAgentErrorResponse{
						Code:    diragentapi.InternalError,
						Message: "command must set 'get_account' in response",
					}
				}
			case req.ListAccounts != nil:
				if resp.ListAccounts == nil {
					resp.Error = &diragentapi.DirAgentErrorResponse{
						Code:    diragentapi.InternalError,
						Message: "command must set 'list_accounts' in response",
					}
				}
			case req.ListGroups != nil:
				if resp.ListGroups == nil {
					resp.Error = &diragentapi.DirAgentErrorResponse{
						Code:    diragentapi.InternalError,
						Message: "command must set 'list_groups' in response",
					}
				}
			case req.PerformOperation != nil:
				if resp.PerformOperation == nil {
					resp.Error = &diragentapi.DirAgentErrorResponse{
						Code:    diragentapi.InternalError,
						Message: "command must set 'perform_operation' in response",
					}
				}
			}
		}

		if resp.Error != nil {
			log.Printf("ERROR: %s %s", resp.Error.Code, resp.Error.Message)
		}

		if err := wsjson.Write(ctx, conn, resp); err != nil {
			cancel(err)
			return err
		}
	}

	// not reached.
}
