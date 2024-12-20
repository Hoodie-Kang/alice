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
package main

import (
	"C"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/getamis/alice/crypto/tss/ecdsa/cggmp/dkg"
	"github.com/getamis/alice/crypto/tss/ecdsa/cggmp/refresh"
	"github.com/getamis/alice/example/config"
	"github.com/getamis/alice/example/logger"
	"github.com/getamis/alice/example/node"
	"github.com/getamis/alice/example/utils"
	"github.com/libp2p/go-libp2p/core/network"
)
import "os"

type DKGConfig struct {
	Port      int64   `json:"port"`
	Rank      uint32  `json:"rank"`
	Threshold uint32  `json:"threshold"`
	Peers     []int64 `json:"peers"`
}

type DKGResult struct {
	Port          int64                     `json:"port"`
	Peers         []int64                   `json:"peers"`
	Share         string                    `json:"share"`
	Pubkey        config.Pubkey             `json:"pubkey"`
	BKs           map[string]config.BK      `json:"bks"`
	PartialPubKey map[string]config.ECPoint `json:"partialPubKey"`
	SSid          []byte                    `json:"ssid"`
}

type RefreshConfig struct {
	Port          int64                     `json:"port"`
	Rank          uint32                    `json:"rank"`
	Threshold     uint32                    `json:"threshold"`
	Peers         []int64                   `json:"peers"`
	Share         string                    `json:"share"`
	Pubkey        config.Pubkey             `json:"pubkey"`
	BKs           map[string]config.BK      `json:"bks"`
	PartialPubKey map[string]config.ECPoint `json:"partialPubKey"`
	SSid          []byte                    `json:"ssid"`
}

type RefreshResult struct {
	Share         string                    `json:"share"`
	Pubkey        config.Pubkey             `json:"pubkey"`
	BKs           map[string]config.BK      `json:"bks"`
	PartialPubKey map[string]config.ECPoint `json:"partialPubKey"`
	Ped           map[string]config.Ped     `json:"ped"`
	PaillierKey   config.PaillierKey        `json:"paillierKey"`
	SSid          []byte                    `json:"ssid"`
}

const refreshProtocol = "/refresh/1.0.0"
const dkgProtocol = "/dkg/1.0.0"

func writeJsonFile(jsonData interface{}, filePath string) error {
	data, err := json.Marshal(jsonData)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0600)
}

func Dkg(por string, peer string) []byte {
	port, _ := strconv.ParseInt(por, 10, 64)
	peers, _ := strconv.ParseInt(peer, 10, 64)

	con := DKGConfig{
		Port:      port,
		Rank:      0,
		Threshold: 2,
		Peers:     []int64{peers},
	}
	// Make a host that listens on the given multiaddress.
	host, err := node.MakeBasicHost(con.Port)
	if err != nil {
		logger.Error("Failed to create a basic host", map[string]string{"err": err.Error()})
	}

	// Create a new peer manager.
	pm := node.NewPeerManager(utils.GetPeerIDFromPort(con.Port), host, dkgProtocol)
	err = pm.AddPeers(con.Peers)
	if err != nil {
		logger.Error("Failed to add peers", map[string]string{"err": err.Error()})
	}

	l := node.NewListener()

	// Create dkg
	service, err := dkg.NewDKG(utils.GetCurve(), pm, []byte("1"), con.Threshold, con.Rank, l)
	if err != nil {
		logger.Error("Cannot create a new dkg", map[string]string{"err": err.Error()})
	}

	// Create a new service.
	node := node.New[*dkg.Message, *dkg.Result](service, l, pm)
	if err != nil {
		logger.Error("Failed to new service", map[string]string{"err": err.Error()})
	}

	// Set a stream handler on the host.
	host.SetStreamHandler(dkgProtocol, func(s network.Stream) {
		node.Handle(s)
	})
	// Ensure all peers are connected before starting DKG process.
	pm.EnsureAllConnected()

	// Start DKG process.
	result, err := node.Process()
	if err != nil {
		logger.Error("Refresh Result error", map[string]string{"err": err.Error()})
	}
	dkgResult := &DKGResult{
		Port:  con.Port,
		Peers: con.Peers,
		Share: result.Share.String(),
		Pubkey: config.Pubkey{
			X: result.PublicKey.GetX().String(),
			Y: result.PublicKey.GetY().String(),
		},
		BKs:           make(map[string]config.BK),
		PartialPubKey: make(map[string]config.ECPoint),
		SSid:          result.SSid,
	}
	for peerID, bk := range result.Bks {
		if peerID == "id-10001" {
			peerID = "id-10003"
		} else {
			peerID = "id-10004"
		}
		dkgResult.BKs[peerID] = config.BK{
			X:    bk.GetX().String(),
			Rank: bk.GetRank(),
		}
	}
	for peerID, ppk := range result.PartialPubKey {
		if peerID == "id-10001" {
			peerID = "id-10003"
		} else {
			peerID = "id-10004"
		}
		dkgResult.PartialPubKey[peerID] = config.ECPoint{
			X: ppk.GetX().String(),
			Y: ppk.GetY().String(),
		}
	}
	jsonData, err := json.Marshal(dkgResult)
	if err != nil {
		logger.Error("json marshal error", map[string]string{"err": err.Error()})
	}
	return jsonData
}

func Refresh(con RefreshConfig, port string, peer string) {
	con.Port, _ = strconv.ParseInt(port, 10, 64)
	peers, _ := strconv.ParseInt(peer, 10, 64)
	con.Peers = []int64{peers}
	// Make a host that listens on the given multiaddress.
	host, err := node.MakeBasicHost(con.Port)
	if err != nil {
		logger.Error("Failed to create a basic host", map[string]string{"err": err.Error()})
	}
	defer host.Close()

	// Refresh needs results from DKG.
	dkgResult, err := utils.ConvertDKGResult(con.Pubkey, con.Share, con.BKs, con.PartialPubKey)
	if err != nil {
		logger.Error("Cannot get DKG result", map[string]string{"err": err.Error()})
	}

	// Create a new peer manager.
	pm := node.NewPeerManager(utils.GetPeerIDFromPort(con.Port), host, refreshProtocol)
	err = pm.AddPeers(con.Peers)
	if err != nil {
		logger.Error("Failed to add peers", map[string]string{"err": err.Error()})
	}

	l := node.NewListener()

	// Create a new service.
	service, err := refresh.NewRefresh(dkgResult.Share, dkgResult.PublicKey, pm, 2, dkgResult.PartialPubKey, dkgResult.Bks, 2048, con.SSid, l)
	if err != nil {
		logger.Error("Cannot create a new refresh", map[string]string{"err": err.Error()})
	}

	// Create a new node.
	node := node.New[*refresh.Message, *refresh.Result](service, l, pm)
	if err != nil {
		logger.Error("Failed to new service", map[string]string{"err": err.Error()})
	}

	// Set a stream handler on the host.
	host.SetStreamHandler(refreshProtocol, func(s network.Stream) {
		node.Handle(s)
	})

	// Ensure all peers are connected before starting refresh process.
	pm.EnsureAllConnected()

	result, err := node.Process()
	if err != nil {
		logger.Error("Refresh Result error", map[string]string{"err": err.Error()})
		return
	}

	p, q := result.PaillierKey.GetPQ()

	// func WriteRefreshResult(id string, input *RefreshConfig, result *refresh.Result, path string) error {
	refreshResult := &RefreshResult{
		Share: result.Share.String(),
		Pubkey: config.Pubkey{
			X: con.Pubkey.X,
			Y: con.Pubkey.Y,
		},
		BKs:           make(map[string]config.BK),
		PartialPubKey: make(map[string]config.ECPoint),
		PaillierKey: config.PaillierKey{
			P: p.String(),
			Q: q.String(),
		},
		Ped: make(map[string]config.Ped),

		SSid: con.SSid,
	}
	for peerID, bk := range con.BKs {
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

	dir, err := os.Getwd()
	if err != nil {
		logger.Error("get cwd error", map[string]string{"err": err.Error()})
	}
	path := fmt.Sprintf("%s/storage/emulated/0/Android/data/com.example.keygen_app/files/dkgresult.json", dir)
	err = writeJsonFile(refreshResult, path)
	if err != nil {
		fmt.Println(err)
	}
}

//export keygen
func keygen(a int64, b int64, c int64, d int64) {
	port_1 := strconv.FormatInt(a, 10)
	peer_1 := strconv.FormatInt(b, 10)
	port_2 := strconv.FormatInt(c, 10)
	peer_2 := strconv.FormatInt(d, 10)

	dkgresult := Dkg(port_1, peer_1)
	var data RefreshConfig
	err := json.Unmarshal(dkgresult, &data)
	if err != nil {
		logger.Error("json unmarshal error", map[string]string{"err": err.Error()})
	}
	Refresh(data, port_2, peer_2)
}

// func main() {
// }
