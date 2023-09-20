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
package dkg

import (
	"fmt"

	"github.com/getamis/alice/crypto/tss/ecdsa/cggmp/dkg"
	"github.com/getamis/alice/example/config"
	re "github.com/getamis/alice/example/refresh"
	"github.com/getamis/sirius/log"
)

type DKGConfig struct {
	Port      int64   `json:"port"`
	Rank      uint32  `json:"rank"`
	Threshold uint32  `json:"threshold"`
	Peers     []int64 `json:"peers"`
}

func writeDKGResult(id string, c *DKGConfig, refreshInput *dkg.Result, path string, info string) error {
	refreshResult := &re.RefreshResult{
		Port:      c.Port,
		Rank:      c.Rank,
		Threshold: c.Threshold,
		Peers:     c.Peers,
		Share:     refreshInput.Share.String(),
		Pubkey: config.Pubkey{
			X: refreshInput.PublicKey.GetX().String(),
			Y: refreshInput.PublicKey.GetY().String(),
		},
		BKs:           make(map[string]config.BK),
		PartialPubKey: make(map[string]config.ECPoint),
		SSid:          refreshInput.SSid,
	}
	for peerID, bk := range refreshInput.Bks {
		if peerID == "id-10001" {
			peerID = "id-10003"
		} else if peerID == "id-10002" {
			peerID = "id-10004"
		}
		refreshResult.BKs[peerID] = config.BK{
			X:    bk.GetX().String(),
			Rank: bk.GetRank(),
		}
	}
	for peerID, ppk := range refreshInput.PartialPubKey {
		if peerID == "id-10001" {
			peerID = "id-10003"
		} else if peerID == "id-10002" {
			peerID = "id-10004"
		}
		refreshResult.PartialPubKey[peerID] = config.ECPoint{
			X: ppk.GetX().String(),
			Y: ppk.GetY().String(),
		}
	}
	// ssid: []byte -> base64 encoded string in Json file
	err := config.WriteJsonFile(refreshResult, getFilePath(id, path, info))
	if err != nil {
		log.Error("Cannot write key file", "err", err)
		return err
	}
	return nil
}

// for DKGRefresh output
func getFilePath(id string, path string, info string) string {
	if id == "id-10001" {
		id = "Octet"
	} else {
		id = "User"
	}
	return fmt.Sprintf(path+"/%s_"+info, id)
}
