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
	"time"

	"github.com/samber/lo"
)

const adTimeFormat = "2006-01-02T15:04:05.000Z"

// UserGroup is info for user groups
type UserGroup struct {
	Name       string `json:"Name"`
	ObjectGUID string `json:"ObjectGUID"`
}

// Users is a list os User
type Users []*User

// User is info from AD on a certain user
type User struct {
	SamAccountName    string      `json:"SamAccountName"`
	DistinguishedName string      `json:"DistinguishedName"`
	Name              string      `json:"Name"`
	EmailAddress      string      `json:"EmailAddress"`
	EmployeeID        interface{} `json:"EmployeeID"`
	ObjectGUID        string      `json:"ObjectGUID"`
	MemberOf          []string    `json:"MemberOf"`
	LockedOut         bool        `json:"LockedOut"`
	WhenChanged       string      `json:"whenChanged"`
}

// GetADUserArgs is an requests args to user functions
type GetADUserArgs struct {
	Identity string
}

// GetADUser retrieved user information form AD
func GetADUser(s Client, args GetADUserArgs) (*Users, error) {
	cmdString := fmt.Sprintf("Get-ADUser -Identity %s -Properties * | ConvertTo-Json", args.Identity)
	stdout, err := s.Execute(cmdString)
	if err != nil {
		return nil, err
	}

	var rv Users

	// If we get one result, it'll be just be the object so check
	var singleUser User
	var singleUserErr error
	if singleUserErr = json.Unmarshal([]byte(stdout), &singleUser); singleUserErr == nil {
		rv = append(rv, &singleUser)
		return &rv, nil
	}

	var listUserError error
	if listUserError = json.Unmarshal([]byte(stdout), &rv); listUserError != nil {
		return nil, errors.Join(fmt.Errorf("could not unmarshall user list json: %w", listUserError), singleUserErr)
	}

	return &rv, nil
}

// ListADUsersArgs is request args to user functions
type ListADUsersArgs struct {
	UpdatedAfter *time.Time
	Cursor       *string
}

// ListADUsers lists ad users according to specified args
func ListADUsers(s Client, args ListADUsersArgs) (*Users, *string, error) {
	var filterString string
	if args.UpdatedAfter != nil {
		formatTimeCmdString := fmt.Sprintf("$listUsersTimeThreshold = (Get-Date \"%s\").ToUniversalTime().ToString(\"yyyyMMddHHmmss.0Z\")", args.UpdatedAfter.Format(adTimeFormat))
		_, err := s.Execute(formatTimeCmdString)
		if err != nil {
			return nil, nil, err
		}

		filterString = "-LDAPFilter \"(whenChanged>=$listUsersTimeThreshold)\""
	} else {
		filterString = "-Filter *"
	}

	cmdString := fmt.Sprintf("Get-ADUser %s -Properties * | Select-Object Name, SamAccountName, ObjectGUID, EmailAddress, @{Name='WhenChanged';Expression={$_.WhenChanged.ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ss.fffZ')}} | ConvertTo-Json", filterString)
	stdout, err := s.Execute(cmdString)
	if err != nil {
		return nil, nil, err
	}

	var rv Users
	if err := json.Unmarshal([]byte(stdout), &rv); err != nil {
		return nil, nil, err
	}

	sort.Slice(rv, func(i, j int) bool {
		return (rv)[i].ObjectGUID < (rv)[j].ObjectGUID
	})

	index := -1
	if args.Cursor != nil {
		_, index, _ = lo.FindIndexOf(rv, func(u *User) bool {
			return u.ObjectGUID == *args.Cursor
		})
	}

	end := index + 1 + defaultPageSize
	if end > len(rv) {
		end = len(rv)
	}
	response := rv[index+1 : end]
	var nextCursor *string
	if len(response) > 0 && end < len(rv) {
		nextCursor = &(rv)[end-1].ObjectGUID
	}

	return &response, nextCursor, nil
}

// GetADUserGroups returns AD user groups from the domain
func GetADUserGroups(s Client, user *User) (*[]UserGroup, error) {
	var userGroups []UserGroup
	for _, group := range user.MemberOf {
		cmdString := fmt.Sprintf("Get-ADGroup -Identity '%s' -Properties ObjectGUID | Select-Object Name,ObjectGUID | ConvertTo-Json", group)
		stdout, err := s.Execute(cmdString)
		if err != nil {
			return nil, err
		}

		// If we get one result, it'll be just be the object so check
		var singleGroup UserGroup
		var singleGroupErr error
		if singleGroupErr = json.Unmarshal([]byte(stdout), &singleGroup); singleGroupErr == nil {
			userGroups = append(userGroups, singleGroup)
			continue
		}

		var multipleGroups []UserGroup
		var multipleGroupsErr error
		if multipleGroupsErr = json.Unmarshal([]byte(stdout), &multipleGroups); multipleGroupsErr != nil {
			return nil, errors.Join(fmt.Errorf("could not unmarshall user list json: %w", multipleGroupsErr), multipleGroupsErr)
		}

		userGroups = append(userGroups, multipleGroups...)
	}

	return &userGroups, nil
}
