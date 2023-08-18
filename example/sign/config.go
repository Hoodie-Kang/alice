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
package sign

import (
	"fmt"
	"os"
	"encoding/json"

	"github.com/getamis/alice/crypto/tss/ecdsa/cggmp/sign"
	"github.com/getamis/alice/example/config"
	// "github.com/getamis/sirius/log"
)

type SignConfig struct {
	Port    int64                                 `json:"port"`
	Share   string                                `json:"share"`
	Pubkey  config.Pubkey                         `json:"pubkey"`
	PartialPubKey map[string]config.PartialPubKey `json:"partialPubKey"`
	BKs     map[string]config.BK                  `json:"bks"`
	Peers   []int64                               `json:"peers"`
	Threshold uint32			                  `json:"threshold"`
	SSid 	[]byte				                  `json:"ssid"`
	AllY    map[string]config.AllY                `json:"ally"`
	Ped     map[string]config.Ped                 `json:"ped"`
	PaillierKey config.PaillierKey                `json:"paillierKey"`
	Message string               
}

type SignResult struct {
	R string `json:"r"`
	S string `json:"s"`
	V uint   `json:"v"`
}

func ReadSignConfigFile(filaPath string) (*SignConfig, error) {
	c := &SignConfig{}
	jsonFile, err := os.ReadFile(filaPath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jsonFile, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func WriteSignResult(id string, result *sign.Result) error {
	V := result.V
	R := result.R
	S := result.S
	fmt.Println("V:", V)
	fmt.Println("R:", R)
	fmt.Println("S:", S)
	
	// 결과 파일 작성
	// signResult := &SignResult{
	// 	R: R.String(),
	// 	S: S.String(),
	// 	V: V,
	// }
	// err := config.WriteJsonFile(signResult, getFilePath(id))
	// if err != nil {
	// 	log.Error("Cannot write JSON file", "err", err)
	// 	return err
	// }
	return nil
}

func getFilePath(id string) string {
	return fmt.Sprintf("./sign-%s-output.json", id)
}
