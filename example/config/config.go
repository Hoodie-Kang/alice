// Copyright © 2020 AMIS Technologies
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package config

import (
	"os"

	"encoding/json"
)

type Pubkey struct {
	X string `json:"x"`
	Y string `json:"y"`
}
// for testing
type PaillierKey struct {
	P string `json:"p"`
	Q string `json:"q"`
}

type BK struct {
	X    string `json:"x"`
	Rank uint32 `json:"rank"`
}

type PartialPubKey struct {
	X string `json:"x"`
	Y string `json:"y"`
}

type AllY struct {
	X string `json:"x"`
	Y string `json:"y"`
}

type Ped struct {
	N string `json:"n"`
	S string `json:"s"`
	T string `json:"t"`
}

func WriteJsonFile(jsonData interface{}, filePath string) error {
	data, err := json.Marshal(jsonData)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0600)
}
