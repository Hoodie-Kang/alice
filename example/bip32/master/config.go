// Copyright © 2020 AMIS Technologies
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/getamis/alice/crypto/tss/ecdsa/cggmp/refresh"
	"github.com/getamis/alice/crypto/bip32/master"
	"github.com/getamis/alice/example/config"
	"github.com/getamis/sirius/log"
	"gopkg.in/yaml.v2"
)

type MasterConfig struct {
	Role     string   `yaml:"role"`
	Port     int64    `yaml:"port"`
	Rank     uint32   `yaml:"rank"`
	Peers    []int64  `yaml:"peers"`
}

type MasterResult struct {
	Role      string            				  `yaml:"role"`
	Port      int64              				  `yaml:"port"`
	Rank      uint32                              `yaml:"rank"`
	Threshold uint32                              `yaml:"threshold"`
	Peers     []int64            				  `yaml:"peers"`
	Share     string               				  `yaml:"share"`
	Pubkey    config.Pubkey       				  `yaml:"pubkey"`
	BKs       map[string]config.BK 				  `yaml:"bks"`	
	PartialPubKey map[string]config.PartialPubKey `yaml:"partialPubKey"`
	Ped map[string]config.Ped                     `yaml:"ped"`
	AllY map[string]config.AllY                   `yaml:"ally"`
	PaillierKey config.PaillierKey                `yaml:"paillierKey"`
	YSecret string 				                  `yaml:"ysecret"`
	SSid      []byte 							  `yaml:"ssid"`
	Seed      []byte							  `yaml:"seed"`
	ChainCode []byte							  `yaml:"chain-code"`
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

func writeMasterResult(con *MasterConfig, refreshInput *master.Result, result *refresh.Result) error {
	masterResult := &MasterResult{
		Role: con.Role,
		Port: con.Port,
		Rank: con.Rank,
		Threshold: 2,
		Peers: con.Peers,
		Share: result.RefreshShare.String(),
		Pubkey: config.Pubkey{
			X: refreshInput.PublicKey.GetX().String(),
			Y: refreshInput.PublicKey.GetY().String(),
		},
		BKs: make(map[string]config.BK),
		PartialPubKey: make(map[string]config.PartialPubKey),
		// for testing! private key p, q to make paillierkey
		// 실제 Refresh -> Sign 과정에서는 Refresh 의 결과로 *paillier.Paillier 를 넘겨서 활용할 수 있도록 해야함.
		PaillierKey: config.PaillierKey{
			P: result.Ped.GetP().String(),
			Q: result.Ped.GetQ().String(),
		},
		Ped: make(map[string]config.Ped),
		AllY: make(map[string]config.AllY),

		YSecret: result.YSecret.String(),	
		SSid: refreshInput.SSid,
		Seed: refreshInput.Seed,
		ChainCode: refreshInput.ChainCode,
	}
	for peerID, bk := range refreshInput.Bks {
		masterResult.BKs[peerID] = config.BK{
			X:    bk.GetX().String(),
			Rank: bk.GetRank(),
		}
	}
	for peerID, ppk := range result.RefreshPartialPubKey {
		masterResult.PartialPubKey[peerID] = config.PartialPubKey{
			X: ppk.GetX().String(),
			Y: ppk.GetY().String(),
		}
	}
	for peerID, ped := range result.PedParameter {
		masterResult.Ped[peerID] = config.Ped{
			N: ped.Getn().String(),
			S: ped.Gets().String(),
			T: ped.Gett().String(),
		}
	}
	for peerID, y := range result.Y {
		masterResult.AllY[peerID] = config.AllY{
			X: y.GetX().String(),
			Y: y.GetY().String(),
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
