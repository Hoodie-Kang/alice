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
	"github.com/getamis/sirius/log"
)

type SignConfig struct {
	Port    int64                                 `yaml:"port"`
	Share   string                                `yaml:"share"`
	Pubkey  config.Pubkey                         `yaml:"pubkey"`
	PartialPubKey map[string]config.PartialPubKey `yaml:"partialPubKey"`
	BKs     map[string]config.BK                  `yaml:"bks"`
	Peers   []int64                               `yaml:"peers"`
	Threshold uint32			                  `yaml:"threshold"`
	SSid 	[]byte				                  `yaml:"ssid"`
	AllY    map[string]config.AllY                `yaml:"ally"`
	Ped     map[string]config.Ped                 `yaml:"ped"`
	PaillierKey config.PaillierKey                `yaml:"paillierKey"`
	Message string               
}

type SignResult struct {
	R string `yaml:"r"`
	S string `yaml:"s"`
	V uint   `yaml:"v"`
}

func ReadSignConfigFile(filaPath string) (*SignConfig, error) {
	c := &SignConfig{}
	yamlFile, err := os.ReadFile(filaPath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(yamlFile, c)
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
	
	signResult := &SignResult{
		R: R.String(),
		S: S.String(),
		V: V,
	}
	err := config.WriteJsonFile(signResult, getFilePath(id))
	if err != nil {
		log.Error("Cannot write YAML file", "err", err)
		return err
	}
	return nil
}

func getFilePath(id string) string {
	return fmt.Sprintf("./sign-%s-output.yaml", id)
}
