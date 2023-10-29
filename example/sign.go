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
	"os"
	"fmt"
	"encoding/hex"
	"encoding/json"
	"strconv"

	"github.com/getamis/alice/example/logger"
	"github.com/getamis/alice/example/node"
	"github.com/getamis/alice/example/config"
	"github.com/getamis/alice/example/utils"
	signer "github.com/getamis/alice/crypto/tss/ecdsa/cggmp/sign"
	"github.com/libp2p/go-libp2p/core/network"
)

type SignConfig struct {
	Port    int64                                 `json:"port"`
	Share   string                                `json:"share"`
	Pubkey  config.Pubkey                         `json:"pubkey"`
	PartialPubKey map[string]config.ECPoint       `json:"partialPubKey"`
	BKs     map[string]config.BK                  `json:"bks"`
	Peers   []int64                               `json:"peers"`
	Threshold uint32			                  `json:"threshold"`
	SSid 	[]byte				                  `json:"ssid"`
	Ped     map[string]config.Ped                 `json:"ped"`
	PaillierKey config.PaillierKey                `json:"paillierKey"`
	Message string               
}

type SignResult struct {
	R string `json:"r"`
	S string `json:"s"`
	V uint   `json:"v"`
}

func ReadSignConfigFile(filaPath string) (*SignConfig, error) {
	c := &SignConfig{}
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

const signProtocol = "/sign/1.0.0"

func Sign(path string, port string, jwt string, msg string) {
	logger.Info("Sign started", map[string]string{})
	config, err := ReadSignConfigFile(path)
	if err != nil {
		logger.Panic("Failed to read key file", map[string]string{"err": err.Error(), "path": path})
	}

	config.Port, _ = strconv.ParseInt(port, 10, 64)
	config.Message = msg

	// Make a host that listens on the given multiaddress.
	host, err := node.MakeBasicHost(config.Port)
	if err != nil {
		logger.Error("Failed to create a basic host", map[string]string{"err": err.Error()})
	}

	// Create a new peer manager.
	pm := node.NewPeerManager(utils.GetPeerIDFromPort(config.Port), host, signProtocol)
	err = pm.AddPeers(config.Peers)
	if err != nil {
 		logger.Error("Failed to add peers",  map[string]string{"err": err.Error()})
	}

	l := node.NewListener()

	signInput, err := utils.ConvertSignInput(config.Share, config.Pubkey, config.PartialPubKey, config.PaillierKey, config.Ped, config.BKs)
	if err != nil {
		logger.Error("Cannot get SignInput", map[string]string{"err": err.Error()})
	}
	message, _ := hex.DecodeString(msg)
	
	// Create sign
	service, err := signer.NewSign(config.Threshold, config.SSid, signInput.Share, signInput.PublicKey, signInput.PartialPubKey, signInput.PaillierKey, signInput.PedParameter, signInput.Bks, message, jwt, pm, l)
	if err != nil {
		logger.Error("Cannot create a new sign", map[string]string{"err": err.Error()})
	}
	
	// Create a new node.
	node := node.New[*signer.Message, *signer.Result] (service, l, pm)
	if err != nil {
		logger.Error("Failed to new service",  map[string]string{"err": err.Error()})
	}
	// Set a stream handler on the host.
	host.SetStreamHandler(signProtocol, func(s network.Stream) {
		node.Handle(s)
	})

	// Ensure all peers are connected before starting sign process.
	pm.EnsureAllConnected()
	
	// Start sign process.                 
	result, err := node.Process()
	if err != nil {
		logger.Error("Sign Result error", map[string]string{"err": err.Error()})
	}
	// return result.R.String()+"#"+result.S.String()+"#"+strconv.FormatUint(uint64(result.V), 10) // R+S+V
	fmt.Println("result:", result)
}

func main() {
	path := os.Getenv("FILE_PATH")
	port := os.Getenv("PORT")
	jwt := os.Getenv("JWT")
	msg := os.Getenv("msg")
	// path := os.Args[1]
	// port := os.Args[2]
	// jwt := os.Args[3]
	// msg := os.Args[4]

	Sign(path, port, jwt, msg)
}
