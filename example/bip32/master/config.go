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
package master

import (
	"fmt"
	"io/ioutil"

	"github.com/getamis/alice/crypto/bip32/master"
	"github.com/getamis/alice/example/config"
	"github.com/getamis/sirius/log"
	"gopkg.in/yaml.v2"
)

type MasterConfig struct {
	Role     string   `yaml:"role"`
	Port     int64    `yaml:"port"`
	Rank     uint32   `yaml:"rank"`
	Peer     []int64  `yaml:"peer"`
}

type MasterResult struct {
	Role     string   `yaml:"role"`
	Port     int64    `yaml:"port"`
	Peer     []int64  `yaml:"peer"`
	Pubkey config.Pubkey        `yaml:"pubkey"`
	Share  string               `yaml:"share"`
	BKs    map[string]config.BK `yaml:"bks"`	
	Seed      []byte			`yaml:"seed"`
	ChainCode []byte			`yaml:"chain-code"`
}

func readMasterConfigFile(filaPath string) (*MasterConfig, error) {
	c := &MasterConfig{}
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

func writeMasterResult(con *MasterConfig, result *master.Result) error {
	masterResult := &MasterResult{
		Role: con.Role,
		Port: con.Port,
		Peer: con.Peer,
		Share: result.Share.String(),
		Pubkey: config.Pubkey{
			X: result.PublicKey.GetX().String(),
			Y: result.PublicKey.GetY().String(),
		},
		BKs: make(map[string]config.BK),
		Seed: result.Seed,
		ChainCode: result.ChainCode,
	}
	for peerID, bk := range result.Bks {
		masterResult.BKs[peerID] = config.BK{
			X:    bk.GetX().String(),
			Rank: bk.GetRank(),
		}
	}
	err := config.WriteYamlFile(masterResult, getFilePath(con.Role, con.Port))
	if err != nil {
		log.Error("Cannot write YAML file", "err", err)
		return err
	}
	return nil
}

func getFilePath(role string, id int64) string {
	return fmt.Sprintf("bip32/%s-%d-output.yaml", role, id)
}
