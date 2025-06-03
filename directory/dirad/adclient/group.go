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
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/samber/lo"
)

// ADGroup represents ADGroup membership info
type ADGroup struct {
	Name       string `json:"Name"`
	ObjectGUID string `json:"ObjectGUID"`
}

// ADGroups is a list of ADGroup
type ADGroups []*ADGroup

// GetADGroupArgs is a struct of request args
type GetADGroupArgs struct {
	NamePrefix *string
	MaxCount   *int64
	Cursor     *string
}

// GetADGroups retrieves groups from AD.
func GetADGroups(s Client, args GetADGroupArgs) (*ADGroups, *string, error) {
	prefix := ""
	if args.NamePrefix != nil {
		prefix = *args.NamePrefix
	}

	escapedPrefix := strings.ReplaceAll(prefix, "'", "''")
	cmdString := fmt.Sprintf("Get-ADGroup -Filter \"Name -like '%s*'\" | Select-Object Name, ObjectGUID | ConvertTo-Json", escapedPrefix)
	stdout, err := s.Execute(cmdString)
	if err != nil {
		return nil, nil, err
	}

	var rv ADGroups
	// If we get one result, it'll be just be the object so check
	var singleGroup ADGroup
	var singleGroupErr error
	if singleGroupErr = json.Unmarshal([]byte(stdout), &singleGroup); singleGroupErr == nil {
		rv = append(rv, &singleGroup)
		return &rv, nil, nil
	}

	var listGroupError error
	if listGroupError = json.Unmarshal([]byte(stdout), &rv); listGroupError != nil {
		return nil, nil, errors.Join(fmt.Errorf("could not unmarshall user list json: %w", listGroupError), singleGroupErr)
	}

	sort.Slice(rv, func(i, j int) bool {
		return (rv)[i].ObjectGUID < (rv)[j].ObjectGUID
	})

	index := -1
	if args.Cursor != nil {
		_, index, _ = lo.FindIndexOf(rv, func(u *ADGroup) bool {
			return u.ObjectGUID == *args.Cursor
		})
	}

	pageSize := defaultPageSize
	if args.MaxCount != nil {
		pageSize = int(*args.MaxCount)
	}

	end := index + 1 + pageSize
	if end > len(rv) {
		end = len(rv)
	}
	response := rv[index+1 : end]
	var nextCursor *string
	if len(response) > 0 && end < len(rv) {
		nextCursor = &(rv)[end-1].ObjectGUID
	}

	return &rv, nextCursor, nil
}
