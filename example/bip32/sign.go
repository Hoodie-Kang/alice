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

	signer "github.com/getamis/alice/crypto/tss/ecdsa/cggmp/sign"
	"github.com/getamis/alice/crypto/tss/ecdsa/cggmp/refresh"
	"github.com/getamis/alice/example/config"
	"github.com/getamis/alice/example/logger"
	"github.com/getamis/alice/example/node"
	"github.com/getamis/alice/example/utils"
	"github.com/libp2p/go-libp2p/core/network"
)

type SignConfig struct {
	Share         string                    `json:"share"`
	Pubkey        config.Pubkey             `json:"pubkey"`
	PartialPubKey map[string]config.ECPoint `json:"partialPubKey"`
	BKs           map[string]config.BK      `json:"bks"`
	SSid          []byte                    `json:"ssid"`
	Ped           map[string]config.Ped     `json:"ped"`
	PaillierKey   config.PaillierKey        `json:"paillierKey"`
	Message       string
}

type SignResult struct {
	R string `json:"r"`
	S string `json:"s"`
	V uint   `json:"v"`
}

type RefreshConfig struct {
	Role      string                              `json:"role"`
	Share     string               				  `json:"share"`
	Pubkey    config.Pubkey       				  `json:"pubkey"`
	BKs       map[string]config.BK 				  `json:"bks"`	
	PartialPubKey map[string]config.ECPoint       `json:"partialPubKey"`
	SSid      []byte 							  `json:"ssid"`
}													

type RefreshResult struct {
	Share     string               				  `json:"share"`
	Pubkey    config.Pubkey       				  `json:"pubkey"`
	BKs       map[string]config.BK 				  `json:"bks"`	
	PartialPubKey map[string]config.ECPoint       `json:"partialPubKey"`
	Ped map[string]config.Ped                     `json:"ped"`
	PaillierKey config.PaillierKey                `json:"paillierKey"`
	SSid      []byte 							  `json:"ssid"`
}

func ReadRefreshConfigFile(filaPath string) (*RefreshConfig, error) {
	c := &RefreshConfig{}
	jsonFile, err := os.ReadFile(filaPath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jsonFile, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

const signProtocol = "/bip32sign/1.0.0"
const refreshProtocol = "/bip32refresh/1.0.0"

func Refresh(path string, por string, per string) []byte {
	con, _ := ReadRefreshConfigFile(path)
	port, _ := strconv.ParseInt(por, 10, 64)
	peer, _ := strconv.ParseInt(per, 10, 64)
	peers := []int64{peer}
	// Make a host that listens on the given multiaddress.
	host, err := node.MakeBasicHost(port)
	if err != nil {
		logger.Error("Failed to create a basic host", map[string]string{"err": err.Error()})
	}
	defer host.Close()

	// Refresh needs results from Child.
	childResult, err := utils.ConvertDKGResult(con.Pubkey, con.Share, con.BKs, con.PartialPubKey)
	if err != nil {
		logger.Error("Cannot get Child result", map[string]string{"err": err.Error()})
	}

	// Create a new peer manager.
	pm := node.NewPeerManager(utils.GetPeerIDFromPort(port), host, refreshProtocol)
	err = pm.AddPeers(peers)
	if err != nil {
		logger.Error("Failed to add peers", map[string]string{"err": err.Error()})
	}

	l := node.NewListener()

	// Create a new service.
	service, err := refresh.NewRefresh(childResult.Share, childResult.PublicKey, pm, 2, childResult.PartialPubKey, childResult.Bks, 2048, con.SSid, l)
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
	}

	p, q := result.PaillierKey.GetPQ()
	refreshResult := &RefreshResult{
		Share: result.Share.String(),
		Pubkey: config.Pubkey{
			X: con.Pubkey.X,
			Y: con.Pubkey.Y,
		},
		BKs: make(map[string]config.BK),
		PartialPubKey: make(map[string]config.ECPoint),
		PaillierKey: config.PaillierKey{
			P: p.String(),
			Q: q.String(),
		},
		Ped: make(map[string]config.Ped),
		SSid: con.SSid,

	}
	for peerID, bk := range con.BKs {
		if peerID == "id-10001" {
			peerID = "Alice"
		} else if peerID == "id-10002" {
			peerID = "Bob"
		}
		refreshResult.BKs[peerID] = config.BK{
			X:    bk.X,
			Rank: bk.Rank,
		}
	}
	for peerID, ppk := range result.PartialPubKey {
		if peerID == "id-10001" {
			peerID = "Alice"
		} else if peerID == "id-10002" {
			peerID = "Bob"
		}
		refreshResult.PartialPubKey[peerID] = config.ECPoint{
			X: ppk.GetX().String(),
			Y: ppk.GetY().String(),
		}
	}
	for peerID, ped := range result.PedParameter {
		if peerID == "id-10001" {
			peerID = "Alice"
		} else if peerID == "id-10002" {
			peerID = "Bob"
		}
		refreshResult.Ped[peerID] = config.Ped{
			N: ped.GetN().String(),
			S: ped.GetS().String(),
			T: ped.GetT().String(),
		}
	}
	jsonData, err := json.Marshal(refreshResult)
	if err != nil {
		logger.Error("json marshal error", map[string]string{"err": err.Error()})
	}
	return jsonData
}

func Sign(config RefreshResult, port int64, peer int64, message string) {
	// Make a host that listens on the given multiaddress.
	host, err := node.MakeBasicHost(port)
	if err != nil {
		logger.Error("Failed to create a basic host", map[string]string{"err": err.Error()})
	}
	var role, peerID string
	if port % 2 == 1 {
		role = "Alice"
		peerID = "Bob"
	} else {
		role = "Bob"
		peerID = "Alice"
	}
	// Create a new peer manager.
	pm := node.NewPeerManager(role, host, signProtocol)

	pm.AddPeer(peerID, peer)

	signInput, err := utils.ConvertSignInput(config.Share, config.Pubkey, config.PartialPubKey, config.PaillierKey, config.Ped, config.BKs)
	if err != nil {
		logger.Error("Cannot get SignInput", map[string]string{"err": err.Error()})
	}
	l := node.NewListener()
	msg := []byte(message)
	service, err := signer.NewSign(2, config.SSid, signInput.Share, signInput.PublicKey, signInput.PartialPubKey, signInput.PaillierKey, signInput.PedParameter, signInput.Bks, msg, "jwt", pm, l)
	if err != nil {
		logger.Error("Cannot create a new sign", map[string]string{"err": err.Error()})
	}

	// Create a new node.
	node := node.New[*signer.Message, *signer.Result](service, l, pm)
	if err != nil {
		logger.Error("Failed to new service", map[string]string{"err": err.Error()})
	}
	// Set a stream handler on the host.
	host.SetStreamHandler(signProtocol, func(s network.Stream) {
		node.Handle(s)
	})

	// 연결 끊김 이벤트 핸들러 등록
	host.Network().Notify(&network.NotifyBundle{
		DisconnectedF: func(network.Network, network.Conn) {
			fmt.Println("Connection was closed, reconnect")
			logger.Error("Connection was closed, reconnect", map[string]string{})
		},
	})

	// Ensure all peers are connected before starting sign process.
	pm.EnsureAllConnected()
	// Start sign process.
	result, err := node.Process()
	if err != nil {
		logger.Error("Sign Result error", map[string]string{"err": err.Error()})
	}
	fmt.Println(result.R.String() + "#" + result.S.String() + "#" + strconv.FormatUint(uint64(result.V), 10))
}

func main() {
	refresh := Refresh(os.Args[1], os.Args[2], os.Args[3])
	var data RefreshResult
	err := json.Unmarshal(refresh, &data)
	if err != nil {
		logger.Error("json unmarshal error", map[string]string{"err": err.Error()})
	}
	serverPort, _ := strconv.ParseInt(os.Args[2], 10, 64)
	clientPort, _ := strconv.ParseInt(os.Args[3], 10, 64)
	Sign(data, serverPort, clientPort, os.Args[4])
}
