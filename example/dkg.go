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
)
import (
	"fmt"
	"strconv"
	"github.com/getamis/alice/crypto/tss/ecdsa/cggmp/dkg"
	"github.com/getamis/alice/example/config"
	"github.com/getamis/alice/example/node"
	"github.com/getamis/alice/example/utils"
	"github.com/libp2p/go-libp2p/core/network"
	logger "github.com/getamis/sirius/log"
)

type DKGConfig struct {
	Port      int64   `json:"port"`
	Rank      uint32  `json:"rank"`
	Threshold uint32  `json:"threshold"`
	Peers     []int64 `json:"peers"`
}

type DKGResult struct {
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

const dkgProtocol = "/dkg/1.0.0"

// for DKG output
func getFilePath(id string, path string, info string) string {
	if id == "id-10001" {
		id = "Octet"
	} else {
		id = "User"
	}
	return fmt.Sprintf(path+"/%s_"+info, id)
}

//export Dkg
func Dkg(argc *C.char, argv *C.char, arg *C.char, info *C.char) {
	port, _ := strconv.ParseInt(C.GoString(argc), 10, 64)
	peers, _ := strconv.ParseInt(C.GoString(argv), 10, 64)
	path := C.GoString(arg)
	information := C.GoString(info)

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
		Port:      con.Port,
		Rank:      con.Rank,
		Threshold: con.Threshold,
		Peers:     con.Peers,
		Share:     result.Share.String(),
		Pubkey:    config.Pubkey{
			X: result.PublicKey.GetX().String(),
			Y: result.PublicKey.GetY().String(),
		},
		BKs:           make(map[string]config.BK),
		PartialPubKey: make(map[string]config.ECPoint),
		SSid: result.SSid,
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
	for peerID, ppk := range result.PartialPubKey{
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

	err = config.WriteJsonFile(dkgResult, getFilePath(pm.SelfID(), path, information))
	if err != nil {
		logger.Error("Cannot write key file", "err", err)
	}
}

// func main() {}
