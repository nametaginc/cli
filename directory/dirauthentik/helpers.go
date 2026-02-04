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
	"net/url"
	"strconv"
	"strings"
	"time"
)

func baseUserQuery(includeGroups bool) url.Values {
	query := url.Values{}
	query.Set("page_size", strconv.Itoa(defaultPageSize))
	query.Set("include_groups", strconv.FormatBool(includeGroups))
	query.Set("include_roles", "false")
	return query
}

func baseGroupQuery() url.Values {
	query := url.Values{}
	query.Set("page_size", strconv.Itoa(defaultPageSize))
	query.Set("include_users", "false")
	query.Set("include_parents", "false")
	query.Set("include_children", "false")
	return query
}

func parseAPITime(value string) *time.Time {
	if value == "" {
		return nil
	}
	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return &parsed
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return &parsed
	}
	return nil
}

func userImmutableID(user apiUser) string {
	if user.UUID != "" {
		return user.UUID
	}
	if user.PK != 0 {
		return strconv.Itoa(user.PK)
	}
	return ""
}

func userDisplayName(user apiUser) string {
	if user.Name != "" {
		return user.Name
	}
	if user.Username != "" {
		return user.Username
	}
	if user.Email != "" {
		return user.Email
	}
	return userImmutableID(user)
}

func appendUnique(values []string, entry string) []string {
	if entry == "" {
		return values
	}
	for _, existing := range values {
		if strings.EqualFold(existing, entry) {
			return values
		}
	}
	return append(values, entry)
}

func userExternalIDs(user apiUser) []string {
	ids := []string{}
	ids = appendUnique(ids, user.Email)
	ids = appendUnique(ids, user.Username)
	if user.UID != "" && !strings.EqualFold(user.UID, user.Email) && !strings.EqualFold(user.UID, user.Username) {
		ids = appendUnique(ids, user.UID)
	}
	if len(ids) == 0 {
		ids = appendUnique(ids, userImmutableID(user))
	}
	return ids
}

func isNumeric(value string) bool {
	if value == "" {
		return false
	}
	_, err := strconv.Atoi(value)
	return err == nil
}
