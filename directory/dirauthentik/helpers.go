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
	"fmt"
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

func (p *Provider) parseAPITime(value string) *time.Time {
	return parseAPITime(value)
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

func (p *Provider) userImmutableID(user apiUser) string {
	return userImmutableID(user)
}

func (p *Provider) userDisplayName(user apiUser) string {
	if name, ok := userAttributeValue(user, p.NameAttribute); ok {
		return name
	}

	if user.Name != "" {
		return user.Name
	}
	if user.Username != "" {
		return user.Username
	}
	if user.Email != "" {
		return user.Email
	}
	return p.userImmutableID(user)
}

func (p *Provider) userBirthDate(user apiUser) *string {
	if birthDate, ok := userAttributeValue(user, p.BirthDateAttribute); ok {
		return &birthDate
	}
	return nil
}

func userAttributeValue(user apiUser, key string) (string, bool) {
	attributeKey := strings.TrimSpace(key)
	if attributeKey == "" || user.Attributes == nil {
		return "", false
	}
	value, ok := user.Attributes[attributeKey]
	if !ok {
		return "", false
	}
	stringValue, ok := value.(string)
	if !ok {
		return "", false
	}
	trimmed := strings.TrimSpace(stringValue)
	if trimmed == "" {
		return "", false
	}
	return trimmed, true
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

func (p *Provider) userExternalIDs(user apiUser) []string {
	return userExternalIDs(user)
}

func isNumeric(value string) bool {
	if value == "" {
		return false
	}
	_, err := strconv.Atoi(value)
	return err == nil
}

func (p *Provider) applyUserListFilters(query url.Values) error {
	path := strings.TrimSpace(p.Path)
	if path != "" {
		query.Set("path", path)
	}

	for _, group := range normalizedStringSet(p.GroupsByName) {
		query.Add("groups_by_name", group)
	}

	types, err := normalizedUserTypes(p.Types)
	if err != nil {
		return err
	}
	for _, userType := range types {
		query.Add("type", userType)
	}
	return nil
}

func normalizedStringSet(values []string) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		normalized = appendUnique(normalized, trimmed)
	}
	return normalized
}

func normalizedUserTypes(values []string) ([]string, error) {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.ToLower(strings.TrimSpace(value))
		if trimmed == "" {
			continue
		}
		if _, ok := allowedUserTypes[trimmed]; !ok {
			return nil, fmt.Errorf("invalid authentik user type filter %q", value)
		}
		normalized = appendUnique(normalized, trimmed)
	}
	return normalized, nil
}

var allowedUserTypes = map[string]struct{}{
	"external":                 {},
	"internal":                 {},
	"internal_service_account": {},
	"service_account":          {},
}
