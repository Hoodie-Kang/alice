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
	"strconv"

	"github.com/getamis/alice/example/logger"
	"github.com/getamis/alice/example/peer"
	"github.com/getamis/alice/example/sign"
	"github.com/getamis/alice/example/utils"
	"github.com/libp2p/go-libp2p/core/network"
)

const signProtocol = "/sign/1.0.0"

func Sign(path string, port string, jwt string, msg string) {
	logger.Info("Sign started", map[string]string{})
	config, err := sign.ReadSignConfigFile(path)
	if err != nil {
		// log.Crit("Failed to read config file", "configFile", path, "err", err)
		logger.Panic("Failed to read key file", map[string]string{"err": err.Error(), "path": path})
	}

	config.Port, _ = strconv.ParseInt(port, 10, 64)
	config.Message = msg

	// Make a host that listens on the given multiaddress.
	host, err := peer.MakeBasicHost(config.Port)
	if err != nil {
		logger.Error("Failed to create a basic host", map[string]string{"err": err.Error()})
	}

	// Create a new peer manager.
	pm := peer.NewPeerManager(utils.GetPeerIDFromPort(config.Port), host, signProtocol)
	err = pm.AddPeers(config.Peers)
	if err != nil {
		logger.Error("Failed to add peers",  map[string]string{"err": err.Error()})
	}

	// Create a new service.
	service, err := sign.NewService(config, jwt, pm)
	if err != nil {
		logger.Error("Failed to new service",  map[string]string{"err": err.Error()})
	}
	// Set a stream handler on the host.
	host.SetStreamHandler(signProtocol, func(s network.Stream) {
		service.Handle(s)
	})

	// Ensure all peers are connected before starting sign process.
	err = pm.EnsureAllConnected()
	if err != nil {
		logger.Error("Connection Timeout",  map[string]string{"err": err.Error()})
	}
	// Start sign process.
	service.Process()
	// result, err := service.Sign.GetResult()
	// if err != nil {
	// 	logger.Error("Sign Result error",  map[string]string{"err": err.Error()})
	// }
	// return result.R.String()+"#"+result.S.String()+"#"+strconv.FormatUint(uint64(result.V), 10) // R+S+V
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
