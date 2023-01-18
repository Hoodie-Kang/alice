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
package child

import (
	"fmt"
	"io/ioutil"

	"github.com/getamis/alice/crypto/bip32/child"
	"github.com/getamis/alice/example/config"
	"github.com/getamis/sirius/log"
	"gopkg.in/yaml.v2"
)

type ChildConfig struct {
	Role      string                `yaml:"role"`
	Port       int64                `yaml:"port"`
	Peer       []int64              `yaml:"peer"`
	Pubkey     config.Pubkey        `yaml:"pubkey"`
	Share      string               `yaml:"share"`
	BKs        map[string]config.BK `yaml:"bks"`	
	Seed       []byte			    `yaml:"seed"`
	Depth      byte                 `yaml:"depth"`
	// ChildIndex uint32	            `yaml:"childindex"`
	ChainCode  []byte			    `yaml:"chain-code"`
}

type ChildResult struct {
	Share      string           `yaml:"share"`
	Translate  string           `yaml:"translate"`
	Pubkey     config.Pubkey    `yaml:"pubkey"`
	ChainCode  []byte			`yaml:"chain-code"`
	Depth      byte			    `yaml:"depth"`
}

func readChildConfigFile(filaPath string) (*ChildConfig, error) {
	c := &ChildConfig{}
	yamlFile, err := ioutil.ReadFile(filaPath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func writeChildResult(con *ChildConfig, result *child.Result) error {
	childResult := &ChildResult{
		Share: result.Share.String(),
		Translate: result.Translate.String(),
		Pubkey: config.Pubkey{
			X: result.PublicKey.GetX().String(),
			Y: result.PublicKey.GetY().String(),
		},
		ChainCode: result.ChainCode,
		Depth: result.Depth,
	}

	err := config.WriteYamlFile(childResult, getFilePath(con.Role, con.Port))
	if err != nil {
		log.Error("Cannot write YAML file", "err", err)
		return err
	}
	return nil
}

func getFilePath(role string, id int64) string {
	return fmt.Sprintf("bip32/child/%s-%d-output.yaml", role, id)
}
