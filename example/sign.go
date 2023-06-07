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

	"github.com/getamis/alice/example/peer"
	"github.com/getamis/alice/example/sign"
	"github.com/getamis/alice/example/utils"
	"github.com/getamis/sirius/log"
	"github.com/libp2p/go-libp2p/core/network"
)

const signProtocol = "/sign/1.0.0"

//export Sign
func Sign(argv *C.char, msg *C.char) {
	arg := C.GoString(argv)
	config, err := sign.ReadSignConfigFile(arg)
	if err != nil {
		log.Crit("Failed to read config file", "configFile", arg, "err", err)
	}

	message := C.GoString(msg)
	fmt.Println("Message:", message)
	config.Message = message

	// Make a host that listens on the given multiaddress.
	host, err := peer.MakeBasicHost(config.Port)
	if err != nil {
		log.Crit("Failed to create a basic host", "err", err)
	}

	// Create a new peer manager.
	pm := peer.NewPeerManager(utils.GetPeerIDFromPort(config.Port), host, signProtocol)
	err = pm.AddPeers(config.Peers)
	if err != nil {
		log.Crit("Failed to add peers", "err", err)
	}

	// Create a new service.
	service, err := sign.NewService(config, pm)
	if err != nil {
		log.Crit("Failed to new service", "err", err)
	}
	// Set a stream handler on the host.
	host.SetStreamHandler(signProtocol, func(s network.Stream) {
		service.Handle(s)
	})

	// Ensure all peers are connected before starting sign process.
	err = pm.EnsureAllConnected()
	if err != nil {
		log.Crit("Connection Timeout", "err", err)
		return
	}
	// Start sign process.
	service.Process()
}

// func main() {}
