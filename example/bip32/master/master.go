package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/getamis/alice/crypto/bip32/master"
	"github.com/getamis/alice/example/config"
	"github.com/getamis/alice/example/logger"
	"github.com/getamis/alice/example/node"
	"github.com/getamis/alice/example/utils"
	"github.com/libp2p/go-libp2p/core/network"
)

type MasterResult struct {
	Role     string   `json:"role"`
	Share     string               				  `json:"share"`
	Pubkey    config.Pubkey       				  `json:"pubkey"`
	BKs       map[string]config.BK 				  `json:"bks"`	
	PartialPubKey map[string]config.ECPoint       `json:"partialPubKey"`
	SSid      []byte 							  `json:"ssid"`
	Seed      []byte							  `json:"seed"`
	ChainCode []byte							  `json:"chain-code"`
}

type RefreshResult struct {
	Role      string            				  `json:"role"`
	Share     string               				  `json:"share"`
	Pubkey    config.Pubkey       				  `json:"pubkey"`
	BKs       map[string]config.BK 				  `json:"bks"`	
	PartialPubKey map[string]config.ECPoint       `json:"partialPubKey"`
	Ped map[string]config.Ped                     `json:"ped"`
	PaillierKey config.PaillierKey                `json:"paillierKey"`
	SSid      []byte 							  `json:"ssid"`
	Seed      []byte							  `json:"seed"`
	ChainCode []byte							  `json:"chain-code"`
}

func writeJsonFile(jsonData interface{}, filePath string) error {
	data, err := json.MarshalIndent(jsonData, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0600)
}

const masterProtocol = "/master/1.0.0"
const (
	circuitPath = "../../../crypto/circuit/bristolFashion/MPCSEED.txt"
)

func Master(role string, por string, per string) []byte {
	port, _ := strconv.ParseInt(por, 10, 64)
	peer, _ := strconv.ParseInt(per, 10, 64)
	peers := []int64{peer}

	host, err := node.MakeBasicHost(port)
	if err != nil {
		logger.Error("Failed to create a basic host", map[string]string{"err": err.Error()})
	}
	defer host.Close()

	pm := node.NewPeerManager(utils.GetPeerIDFromPort(port), host, masterProtocol)
	err = pm.AddPeers(peers)
	if err != nil {
		logger.Error("Failed to add peers", map[string]string{"err": err.Error()})
	}

	l := node.NewListener()
	sid := []byte("mastrsid")
	var service *master.Master
	if role == "Alice" {
		m, err := master.NewAlice(pm, sid, 0, circuitPath, l)
		if err != nil {
			logger.Error("Cannot create a new Alice", map[string]string{"err": err.Error()})
		}
		service = m
	} else if role == "Bob" {
		m, err := master.NewBob(pm, sid, 0, circuitPath, l)
		if err != nil {
			logger.Error("Cannot create a new Bob", map[string]string{"err": err.Error()})
		}
		service = m
	} else {
		logger.Error("Role must be Alice or Bob", map[string]string{"role": role, "err": err.Error()})
	}	
	
	// Create a new service
	node := node.New[*master.Message, *master.Result](service, l, pm)
	if err != nil {
		logger.Error("Failed to new service", map[string]string{"err": err.Error()})
	}

	host.SetStreamHandler(masterProtocol, func(s network.Stream) {
		node.Handle(s)
	})
	
	pm.EnsureAllConnected()

	result, err := node.Process()
	if err != nil {
		logger.Error("Failed to get result from Master", map[string]string{"err": err.Error()})
	}
	masterResult := &MasterResult{
		Role:  role,
		Share: result.Share.String(),
		Pubkey: config.Pubkey{
			X: result.PublicKey.GetX().String(),
			Y: result.PublicKey.GetY().String(),
		},
		BKs:           make(map[string]config.BK),
		PartialPubKey: make(map[string]config.ECPoint),
		SSid:          result.SSid,
		Seed:          result.Seed,
		ChainCode:     result.ChainCode,
	}
	for peerID, bk := range result.Bks {
		masterResult.BKs[peerID] = config.BK{
			X:    bk.GetX().String(),
			Rank: bk.GetRank(),
		}
	}
	for peerID, ppk := range result.PartialPubKey {
		masterResult.PartialPubKey[peerID] = config.ECPoint{
			X: ppk.GetX().String(),
			Y: ppk.GetY().String(),
		}
	}
	jsonData, err := json.Marshal(masterResult)
	if err != nil {
		logger.Error("json marshal error", map[string]string{"err": err.Error()})
	}
	path := fmt.Sprintf("./%s.json", role)
	writeJsonFile(masterResult, path)
	return jsonData
}

func main() {
	master := Master(os.Args[1], os.Args[2], os.Args[3])
	var data MasterResult
	err := json.Unmarshal(master, &data)
	if err != nil {
		logger.Error("json unmarshal error", map[string]string{"err": err.Error()})
	}
}
