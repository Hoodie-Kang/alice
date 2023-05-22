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
package refresh

import (
	"fmt"
	"io/ioutil"

	"github.com/getamis/alice/crypto/tss/ecdsa/cggmp/refresh"
	"github.com/getamis/alice/example/config"
	"github.com/getamis/sirius/log"
	"gopkg.in/yaml.v2"
)

type RefreshConfig struct {
	Port      int64                               `yaml:"port"`
	Rank      uint32                              `yaml:"rank"`
	Threshold uint32                              `yaml:"threshold"`
	Peers     []int64 			                  `yaml:"peers"`
	Share     string                              `yaml:"share"`
	Pubkey    config.Pubkey                       `yaml:"pubkey"`
	BKs       map[string]config.BK                `yaml:"bks"`
	PartialPubKey map[string]config.PartialPubKey `yaml:"partialPubKey"`
	SSid      []byte			                  `yaml:"ssid"`
}

type RefreshResult struct {
	Port      int64                               `json:"port"`
	Rank      uint32                              `json:"rank"`
	Threshold uint32                              `json:"threshold"`
	Peers     []int64                             `json:"peers"`
	Share  string                                 `json:"share"`
	Pubkey config.Pubkey                          `json:"pubkey"`
	BKs    map[string]config.BK                   `json:"bks"`
	PartialPubKey map[string]config.PartialPubKey `json:"partialPubKey"`
	Ped map[string]config.Ped					  `json:"ped"`
	AllY map[string]config.AllY                   `json:"ally"`
	PaillierKey config.PaillierKey                `json:"paillierKey"`
	YSecret string 				                  `json:"ysecret"`
	SSid    []byte                                `json:"ssid"`
}

func ReadRefreshConfigFile(filaPath string) (*RefreshConfig, error) {
	c := &RefreshConfig{}
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

func WriteRefreshResult(id string, input *RefreshConfig, result *refresh.Result) error {
	refreshResult := &RefreshResult{
		Port: input.Port,
		Rank: input.Rank,
		Threshold: input.Threshold,
		Peers: input.Peers,
		Share: result.RefreshShare.String(),
		Pubkey: config.Pubkey{
			X: input.Pubkey.X,
			Y: input.Pubkey.Y,
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
		SSid: input.SSid,
	}
	for peerID, bk := range input.BKs {
		refreshResult.BKs[peerID] = config.BK{
			X:    bk.X,
			Rank: bk.Rank,
		}
	}
	for peerID, ppk := range result.RefreshPartialPubKey {
		refreshResult.PartialPubKey[peerID] = config.PartialPubKey{
			X: ppk.GetX().String(),
			Y: ppk.GetY().String(),
		}
	}
	// for peerID, ped := range result.PedParameter {
	// 	refreshResult.Ped[peerID] = config.Ped{
	// 		N: ped.Getn().String(),
	// 		S: ped.Gets().String(),
	// 		T: ped.Gett().String(),
	// 	}
	// }
	for peerID, y := range result.Y {
		refreshResult.AllY[peerID] = config.AllY{
			X: y.GetX().String(),
			Y: y.GetY().String(),
		}
	}
	err := config.WriteYamlFile(refreshResult, getFilePath(id))
	if err != nil {
		log.Error("Cannot write YAML file", "err", err)
		return err
	}
	return nil
}

func getFilePath(id string) string {
	return fmt.Sprintf("./refresh-%s-output.yaml", id)
}
