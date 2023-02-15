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
package dkg

import (
	"fmt"
	"io/ioutil"

	"github.com/getamis/alice/crypto/tss/ecdsa/cggmp/refresh"
	"github.com/getamis/alice/crypto/tss/ecdsa/cggmp/dkg"
	"github.com/getamis/alice/example/config"
	re "github.com/getamis/alice/example/refresh"
	"github.com/getamis/sirius/log"
	"gopkg.in/yaml.v2"
)

type DKGConfig struct {
	Port      int64   `yaml:"port"`
	Rank      uint32  `yaml:"rank"`
	Threshold uint32  `yaml:"threshold"`
	Peers     []int64 `yaml:"peers"`
}

func readDKGConfigFile(filaPath string) (*DKGConfig, error) {
	c := &DKGConfig{}
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

func writeDKGResult(id string, c *DKGConfig, result *dkg.Result) error {
	dkgResult := &re.RefreshConfig{
		Port: c.Port,
		Rank: c.Rank,
		Threshold: c.Threshold,
		Peers: c.Peers,
		Share: result.Share.String(),
		Pubkey: config.Pubkey{
			X: result.PublicKey.GetX().String(),
			Y: result.PublicKey.GetY().String(),
		},
		BKs: make(map[string]config.BK),
		PartialPubKey: make(map[string]config.PartialPubKey),
		SSid: result.SSid,
	}
	for peerID, bk := range result.Bks {
		dkgResult.BKs[peerID] = config.BK{
			X:    bk.GetX().String(),
			Rank: bk.GetRank(),
		}
	}
	for peerID, ppk := range result.PartialPubKey {
		dkgResult.PartialPubKey[peerID] = config.PartialPubKey{
			X: ppk.GetX().String(),
			Y: ppk.GetY().String(),
		}
	}
	err := config.WriteYamlFile(dkgResult, refreshInputPath(id))
	if err != nil {
		log.Error("Cannot write YAML file", "err", err)
		return err
	}
	return nil
}

func writeDKGRefreshResult(id string, c *DKGConfig, refreshInput *dkg.Result, result *refresh.Result) error {
	refreshResult := &re.RefreshResult{
		Port: c.Port,
		Rank: c.Rank,
		Threshold: c.Threshold,
		Peers: c.Peers,
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
	}
	for peerID, bk := range refreshInput.Bks {
		refreshResult.BKs[peerID] = config.BK{
			X:    bk.GetX().String(),
			Rank: bk.GetRank(),
		}
	}
	for peerID, ppk := range result.RefreshPartialPubKey {
		refreshResult.PartialPubKey[peerID] = config.PartialPubKey{
			X: ppk.GetX().String(),
			Y: ppk.GetY().String(),
		}
	}
	for peerID, ped := range result.PedParameter {
		refreshResult.Ped[peerID] = config.Ped{
			N: ped.Getn().String(),
			S: ped.Gets().String(),
			T: ped.Gett().String(),
		}
	}
	for peerID, y := range result.Y {
		refreshResult.AllY[peerID] = config.AllY{
			X: y.GetX().String(),
			Y: y.GetY().String(),
		}
	}
	err := config.WriteYamlFile(refreshResult, refreshOutputPath(id))
	if err != nil {
		log.Error("Cannot write YAML file", "err", err)
		return err
	}
	return nil
}

func refreshInputPath(id string) string {
	return fmt.Sprintf("refresh/%s-input.yaml", id)
}

// for DKGRefresh output
func refreshOutputPath(id string) string {
	return fmt.Sprintf("refresh/%s-output.yaml", id)
}
