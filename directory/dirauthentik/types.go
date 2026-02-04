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

type pagination struct {
	Next       *int `json:"next"`
	Previous   *int `json:"previous"`
	Count      int  `json:"count"`
	Current    int  `json:"current"`
	TotalPages int  `json:"total_pages"`
	StartIndex int  `json:"start_index"`
	EndIndex   int  `json:"end_index"`
}

type apiGroup struct {
	PK   string `json:"pk"`
	Name string `json:"name"`
}

type apiUser struct {
	PK          int        `json:"pk"`
	UUID        string     `json:"uuid"`
	Username    string     `json:"username"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	UID         string     `json:"uid"`
	LastUpdated string     `json:"last_updated"`
	GroupsObj   []apiGroup `json:"groups_obj"`
}

type apiDevice struct {
	PK            string `json:"pk"`
	Type          string `json:"type"`
	MetaModelName string `json:"meta_model_name"`
}

type userListResponse struct {
	Pagination pagination `json:"pagination"`
	Results    []apiUser  `json:"results"`
}

type groupListResponse struct {
	Pagination pagination `json:"pagination"`
	Results    []apiGroup `json:"results"`
}

type linkResponse struct {
	Link string `json:"link"`
}
