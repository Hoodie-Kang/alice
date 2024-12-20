package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/getamis/alice/crypto/bip32/child"
	"github.com/getamis/alice/example/config"
	"github.com/getamis/alice/example/logger"
	"github.com/getamis/alice/example/node"
	"github.com/getamis/alice/example/utils"
	"github.com/libp2p/go-libp2p/core/network"
)
type ChildConfig struct {
	Role       string               `json:"role"`
	Pubkey     config.Pubkey        `json:"pubkey"`
	Share      string               `json:"share"`
	BKs        map[string]config.BK `json:"bks"`	
	Seed       []byte			    `json:"seed"`
	Depth      byte                 `json:"depth"`
	ChildIndex uint32	            `json:"childindex"`
	ChainCode  []byte			    `json:"chain-code"`
}

type ChildResult struct {
	Role      string                              `json:"role"`
	Share     string               				  `json:"share"`
	Pubkey    config.Pubkey       				  `json:"pubkey"`
	BKs       map[string]config.BK 				  `json:"bks"`	
	PartialPubKey map[string]config.ECPoint       `json:"partialPubKey"`
	Translate  string           				  `json:"translate"`
	SSid      []byte 							  `json:"ssid"`
	Seed      []byte							  `json:"seed"`
	ChildIndex uint32	            			  `json:"childindex"`
	ChainCode []byte							  `json:"chain-code"`
	Depth     byte			 		              `json:"depth"`
}

func writeJsonFile(jsonData interface{}, filePath string) error {
	data, err := json.MarshalIndent(jsonData, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0600)
}

func readChildConfigFile(filaPath string) (*ChildConfig, error) {
	c := &ChildConfig{}
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

const childProtocol = "/child/1.0.0"
const (
	circuitPath = "../../../crypto/circuit/bristolFashion/MPCHMAC.txt"
)

func Child (filePath string, por string, per string) []byte {
	con, err := readChildConfigFile(filePath)
	port, _ := strconv.ParseInt(por, 10, 64)
	peer, _ := strconv.ParseInt(per, 10, 64)
	peers := []int64{peer}
	if err != nil {
		logger.Error("Failed to read config file", map[string]string{"path": filePath, "err": err.Error()})
	}
	host, err := node.MakeBasicHost(port)
	if err != nil {
		logger.Error("Failed to create a basic host", map[string]string{"err": err.Error()})
	}
	defer host.Close()

	pm := node.NewPeerManager(utils.GetPeerIDFromPort(port), host, childProtocol)
	err = pm.AddPeers(peers)
	if err != nil {
		logger.Error("Failed to add peers", map[string]string{"err": err.Error()})
	}

	l := node.NewListener()
	var service *child.Child
	sid := []byte("childsid")
	con.ChildIndex = uint32(2147483651)

	masterResult, err := utils.ConvertChildInput(con.Pubkey, con.Share, con.BKs, con.Seed, con.ChainCode)
	if err != nil {
		logger.Error("Cannot get Master result", map[string]string{"err": err.Error()})
	}

	if con.Role == "Alice" {
		c, err := child.NewAlice(pm, sid, masterResult.Share, masterResult.Bks, circuitPath, masterResult.ChainCode, con.Depth, con.ChildIndex, masterResult.PublicKey, l)
		if err != nil {
			logger.Error("Cannot create a new Alice", map[string]string{"err": err.Error()})
		}
		service = c
	} else if con.Role == "Bob" {
		c, err := child.NewBob(pm, sid, masterResult.Share, masterResult.Bks, circuitPath, masterResult.ChainCode, con.Depth, con.ChildIndex, masterResult.PublicKey, l)
		if err != nil {
			logger.Error("Cannot create a new Bob", map[string]string{"err": err.Error()})
		}
		service = c
	} else {
		logger.Error("Role must be Alice or Bob", map[string]string{"role": con.Role, "err": err.Error()})
	}

	node := node.New[*child.Message, *child.Result](service, l, pm)
	if err != nil {
		logger.Error("Failed to new service", map[string]string{"err": err.Error()})
	}

	host.SetStreamHandler(childProtocol, func(s network.Stream) {
		node.Handle(s)
	})
	
	pm.EnsureAllConnected()
	result, err := node.Process()
	if err != nil {
		logger.Error("Failed to get result from Child", map[string]string{"err": err.Error()})
	}
	childResult := &ChildResult{
		Role: con.Role,
		Share: result.Share.String(),
		Translate: result.Translate.String(),
		Pubkey: config.Pubkey{
			X: result.PublicKey.GetX().String(),
			Y: result.PublicKey.GetY().String(),
		},
		BKs:           make(map[string]config.BK),
		PartialPubKey: make(map[string]config.ECPoint),
		SSid:          result.SSid,
		Seed:          con.Seed,
		ChainCode:     result.ChainCode,
		ChildIndex:    con.ChildIndex,
	}
	for peerID, bk := range result.BKs {
		childResult.BKs[peerID] = config.BK{
			X:    bk.GetX().String(),
			Rank: bk.GetRank(),
		}
	}
	for peerID, ppk := range result.PartialPubKey {
		childResult.PartialPubKey[peerID] = config.ECPoint{
			X: ppk.GetX().String(),
			Y: ppk.GetY().String(),
		}
	}
	jsonData, err := json.Marshal(childResult)
	if err != nil {
		logger.Error("json marshal error", map[string]string{"err": err.Error()})
	}
	path := fmt.Sprintf("./child-%s-%d.json", con.Role, con.ChildIndex)
	writeJsonFile(childResult, path)
	return jsonData
}

func main() {
	child := Child(os.Args[1], os.Args[2], os.Args[3])
	var data ChildResult
	err := json.Unmarshal(child, &data)
	if err != nil {
		logger.Error("json unmarshal error", map[string]string{"err": err.Error()})
	}
}