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
package refresh

import (
	"os"
	"encoding/json"

	"github.com/getamis/alice/crypto/tss/ecdsa/cggmp/refresh"
	"github.com/getamis/alice/example/config"
	"github.com/getamis/sirius/log"
)

type RefreshConfig struct {
	Port          int64                           `json:"port"`
	Rank          uint32                          `json:"rank"`
	Threshold     uint32                          `json:"threshold"`
	Peers         []int64                         `json:"peers"`
	Share         string                          `json:"share"`
	Pubkey        config.Pubkey                   `json:"pubkey"`
	BKs           map[string]config.BK            `json:"bks"`
	PartialPubKey map[string]config.ECPoint       `json:"partialPubKey"`
	SSid          []byte                          `json:"ssid"`
}

type RefreshResult struct {
	Port          int64                           `json:"port"`
	Rank          uint32                          `json:"rank"`
	Threshold     uint32                          `json:"threshold"`
	Peers         []int64                         `json:"peers"`
	Share         string                          `json:"share"`
	Pubkey        config.Pubkey                   `json:"pubkey"`
	BKs           map[string]config.BK            `json:"bks"`
	PartialPubKey map[string]config.ECPoint       `json:"partialPubKey"`
	Ped           map[string]config.Ped           `json:"ped"`
	Y             map[string]config.ECPoint       `json:"y"`
	PaillierKey   config.PaillierKey              `json:"paillierKey"`
	YSecret       string                          `json:"ysecret"`
	SSid          []byte                          `json:"ssid"`
}

func ReadRefreshConfigFile(filaPath string) (*RefreshConfig, error) {
	c := &RefreshConfig{}
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

func WriteRefreshResult(id string, input *RefreshConfig, result *refresh.Result, path string) error {
	p, q := result.PaillierKey.GetPQ()
	refreshResult := &RefreshResult{
		Port:      input.Port,
		Rank:      input.Rank,
		Threshold: input.Threshold,
		Peers:     input.Peers,
		Share:     result.Share.String(),
		Pubkey: config.Pubkey{
			X: input.Pubkey.X,
			Y: input.Pubkey.Y,
		},
		BKs:           make(map[string]config.BK),
		PartialPubKey: make(map[string]config.ECPoint),
		PaillierKey: config.PaillierKey{
			P: p.String(),
			Q: q.String(),
		},
		Ped:  make(map[string]config.Ped),
		Y: make(map[string]config.ECPoint),

		YSecret: result.YSecret.String(),
		SSid:    input.SSid,
	}
	for peerID, bk := range input.BKs {
		if peerID == "id-10003" {
			peerID = "Octet"
		} else if peerID == "id-10004" {
			peerID = "User"
		}
		refreshResult.BKs[peerID] = config.BK{
			X:    bk.X,
			Rank: bk.Rank,
		}
	}
	for peerID, ppk := range result.PartialPubKey {
		if peerID == "id-10003" {
			peerID = "Octet"
		} else if peerID == "id-10004" {
			peerID = "User"
		}
		refreshResult.PartialPubKey[peerID] = config.ECPoint{
			X: ppk.GetX().String(),
			Y: ppk.GetY().String(),
		}
	}
	for peerID, ped := range result.PedParameter {
		if peerID == "id-10003" {
			peerID = "Octet"
		} else if peerID == "id-10004" {
			peerID = "User"
		}
		refreshResult.Ped[peerID] = config.Ped{
			N: ped.GetN().String(),
			S: ped.GetS().String(),
			T: ped.GetT().String(),
		}
	}

	for peerID, y := range result.Y {
		if peerID == "id-10003" {
			peerID = "Octet"
		} else if peerID == "id-10004" {
			peerID = "User"
		}
		refreshResult.Y[peerID] = config.ECPoint{
			X: y.GetX().String(),
			Y: y.GetY().String(),
		}
	}
	err := config.WriteJsonFile(refreshResult, path)
	if err != nil {
		log.Error("Cannot write key file", "err", err)
		return err
	}
	return nil
}
