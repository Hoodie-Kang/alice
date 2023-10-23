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
	"fmt"
	"os"
	"strconv"
	"time"

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

func OctetSign(path string, port string, jwt string, msg string) {
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

	for i := 0; i < 5; i++ {
		// Create sign
		time.Sleep(3 * time.Second)

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
}

func ClientSign(path string, port string, jwt string, msg string, opt string) {
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

	// TSS 시작 신호가 왔는지 매 1초 혹은 2초마다 뭐 확인하고 그런다. 서명 끝나고 여기로 돌아와서 다시 대기해야함
	for {
		// select {
		// case r, ok := <-resume:
		// 	if ok {
		// 		fmt.Println(r)
		// 	} else {
		// 		fmt.Println("channel closed!")
		// 	}
		// default:
		// fmt.Println("No value ready,,,,, retry")
		time.Sleep(3 * time.Second)
		l := node.NewListener()
		// Create sign
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
			fmt.Println("streamhandler")
		})

		// 연결 끊김 이벤트 핸들러 등록
		host.Network().Notify(&network.NotifyBundle{
			DisconnectedF: func(network.Network, network.Conn) {
				fmt.Println("Connection was closed, reconnect")
			},
		})
		
		pm.EnsureAllConnected()

		// Start sign process.
		result, err := node.Process()
		if err != nil {
			logger.Error("Sign Result error", map[string]string{"err": err.Error()})
		}

		// resume <- result.R.String()+"#"+result.S.String()+"#"+strconv.FormatUint(uint64(result.V), 10)
		fmt.Println(result)
		if opt == "true" {
			logger.Info("Docker started with only_once TRUE", map[string]string{"only_once": opt})
			return
		} else {
			logger.Info("Waiting for another sign", map[string]string{"only_once": opt})
			// 위에 대기하는 곳으로 돌아가는 코드
		}
		// }
	}
}

// var resume chan string

func main() {
	// path := os.Getenv("file_path")
	// port := os.Getenv("port")
	// jwt := os.Getenv("jwt")
	// msg := os.Getenv("msg")
	// only_once := os.Getenv("only_once")
	// Sign(path, port, jwt, msg, only_once)
	c1 := make(chan string)
	c2 := make(chan string)
 
	go func() {
	   for{
		  time.Sleep(5 * time.Second)
		  c1 <- "one"
	   }
	}()
	go func() {
	   for{
		  time.Sleep(10 * time.Second)
		  c2 <- "two"
	   }
	}()
 
	for{
	   fmt.Println("start select------------------")
	   select {
	   case msg1 := <-c1:
		  fmt.Println("received", msg1)
	   case msg2 := <-c2:
		  fmt.Println("received", msg2)
	   }
	   fmt.Println("end select-------------------\n\n")
	}

	// path := os.Args[1]
	// port := os.Args[2]
	// jwt := os.Args[3]
	// msg := os.Args[4]

	// // OctetSign(path,port,jwt,msg, "false")
	// ClientSign(path, port, jwt, msg, "false")
}
