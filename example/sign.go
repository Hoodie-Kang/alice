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
	"encoding/hex"
	"encoding/json"
	"net/http"
	"fmt"
	"os"
	"io"
	"strconv"

	signer "github.com/getamis/alice/crypto/tss/ecdsa/cggmp/sign"
	"github.com/getamis/alice/example/config"
	"github.com/getamis/alice/example/logger"
	"github.com/getamis/alice/example/node"
	"github.com/getamis/alice/example/utils"
	"github.com/libp2p/go-libp2p/core/network"
)

type SignConfig struct {
	Port          int64                     `json:"port"`
	Share         string                    `json:"share"`
	Pubkey        config.Pubkey             `json:"pubkey"`
	PartialPubKey map[string]config.ECPoint `json:"partialPubKey"`
	BKs           map[string]config.BK      `json:"bks"`
	Peers         []int64                   `json:"peers"`
	Threshold     uint32                    `json:"threshold"`
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

type User struct {
	Idx string `json:"idx"`
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


func ValidateToken(url string, token string, idx string) bool {
	client := &http.Client{}
	auth := fmt.Sprintf("%s/2.0/token", url)
	req, err := http.NewRequest("GET", auth, nil)
	if err != nil {
		fmt.Println("HTTP Request Creation Error", err)
		logger.Error("HTTP Request Creation Error", map[string]string{"err": err.Error()})
		return false
	}
	req.Header.Add("accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("HTTP Request Error:", err)
		logger.Error("HTTP Request Error", map[string]string{"err": err.Error()})
		return false
	}

	defer resp.Body.Close()
	fmt.Println("Response Code:", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Read Response Body Error:", err)
		logger.Error("Read Response Body Error", map[string]string{"err": err.Error()})
	}

	var user User
	err = json.Unmarshal(body, &user)
	if err != nil {
		fmt.Println("JSON Unmarshal Error:", err)
		logger.Error("JSON Unmarshal Error", map[string]string{"err": err.Error()})
	}

	fmt.Println(user)
	fmt.Println(user.Idx)

	if user.Idx != idx {
		fmt.Println("Invalid Wallet Error")
		return false
	}else {
		return true
	}
}

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
		logger.Error("Failed to add peers", map[string]string{"err": err.Error()})
	}

	signInput, err := utils.ConvertSignInput(config.Share, config.Pubkey, config.PartialPubKey, config.PaillierKey, config.Ped, config.BKs)
	if err != nil {
		logger.Error("Cannot get SignInput", map[string]string{"err": err.Error()})
	}
	message, _ := hex.DecodeString(msg)
	l := node.NewListener()
	service, err := signer.NewSign(config.Threshold, config.SSid, signInput.Share, signInput.PublicKey, signInput.PartialPubKey, signInput.PaillierKey, signInput.PedParameter, signInput.Bks, message, jwt, pm, l)
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
	// path := os.Getenv("file_path")
	// port := os.Getenv("port")
	// jwt := os.Getenv("jwt")
	// msg := os.Getenv("msg")
	// Sign(path, port, jwt, msg)
}
