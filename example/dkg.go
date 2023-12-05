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
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/getamis/alice/crypto/tss/ecdsa/cggmp/dkg"
	"github.com/getamis/alice/crypto/tss/ecdsa/cggmp/refresh"

	"github.com/getamis/alice/example/config"
	"github.com/getamis/alice/example/node"
	"github.com/getamis/alice/example/utils"
	logger "github.com/getamis/sirius/log"
	"github.com/libp2p/go-libp2p/core/network"
)

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

const refreshProtocol = "/refresh/1.0.0"
const dkgProtocol = "/dkg/1.0.0"

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
		logger.Error("Failed to create a basic host", "err", err)
	}

	// Create a new peer manager.
	pm := node.NewPeerManager(utils.GetPeerIDFromPort(con.Port), host, dkgProtocol)
	err = pm.AddPeers(con.Peers)
	if err != nil {
		logger.Error("Failed to add peers", "err", err)
	}

	l := node.NewListener()

	// Create dkg
	service, err := dkg.NewDKG(utils.GetCurve(), pm, []byte("1"), con.Threshold, con.Rank, l)
	if err != nil {
		logger.Error("Cannot create a new dkg", "err", err)
	}

	// Create a new service.
	node := node.New[*dkg.Message, *dkg.Result](service, l, pm)
	if err != nil {
		logger.Error("Failed to new service", "err", err)
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
		logger.Error("Refresh Result error", "err", err)
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
		logger.Error("json marshal error", err)
	}
	return jsonData
}

func Refresh(con RefreshConfig, por string) {
	port, _ := strconv.ParseInt(por, 10, 64)
	con.Port = port
	if con.Peers[0] == 10002 {
		con.Peers[0] = 10004
	} else {
		con.Peers[0] = 10003
	}
	// Make a host that listens on the given multiaddress.
	host, err := node.MakeBasicHost(con.Port)
	if err != nil {
		logger.Error("Failed to create a basic host", "err", err)
	}
	defer host.Close()

	// Refresh needs results from DKG.
	dkgResult, err := utils.ConvertDKGResult(con.Pubkey, con.Share, con.BKs, con.PartialPubKey)
	if err != nil {
		logger.Error("Cannot get DKG result", "err", err)
	}

	// Create a new peer manager.
	pm := node.NewPeerManager(utils.GetPeerIDFromPort(con.Port), host, refreshProtocol)
	err = pm.AddPeers(con.Peers)
	if err != nil {
		logger.Error("Failed to add peers", "err", err)
	}

	l := node.NewListener()

	// Create a new service.
	service, err := refresh.NewRefresh(dkgResult.Share, dkgResult.PublicKey, pm, 2, dkgResult.PartialPubKey, dkgResult.Bks, 2048, con.SSid, l)
	if err != nil {
		logger.Error("Cannot create a new refresh", "err", err)
	}

	// Create a new node.
	node := node.New[*refresh.Message, *refresh.Result](service, l, pm)
	if err != nil {
		logger.Error("Failed to new service", "err", err)
	}

	// Set a stream handler on the host.
	host.SetStreamHandler(refreshProtocol, func(s network.Stream) {
		node.Handle(s)
	})

	// Ensure all peers are connected before starting refresh process.
	pm.EnsureAllConnected()

	result, err := node.Process()
	if err != nil {
		logger.Error("Refresh Result error", "err", err)
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
	jsonData, err := json.Marshal(refreshResult)
	if err != nil {
		logger.Error("json marshal error", err)
	}
	fmt.Println(jsonData)
}

func main() {
	dkgresult := Dkg(os.Args[1], os.Args[2])
	var data RefreshConfig
	err := json.Unmarshal(dkgresult, &data)
	if err != nil {
		logger.Error("json unmarshal error", err)
	}
	Refresh(data, os.Args[3])
}
