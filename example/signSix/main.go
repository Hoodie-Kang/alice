// Copyright © 2020 AMIS Technologies
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package signSix

import (
	"github.com/getamis/alice/example/peer"
	"github.com/getamis/alice/example/utils"
	"github.com/getamis/sirius/log"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const signSixProtocol = "/signSix/1.0.0"

var configFile string

var Cmd = &cobra.Command{
	Use:   "signSix",
	Short: "SignSix process",
	Long:  `Signing(6 Round) for using the secret shares to generate a signature.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := initService(cmd)
		if err != nil {
			log.Crit("Failed to init", "err", err)
		}

		c, err := readSignSixConfigFile(configFile)
		if err != nil {
			log.Crit("Failed to read config file", "configFile", configFile, "err", err)
		}

		// Make a host that listens on the given multiaddress.
		host, err := peer.MakeBasicHost(c.Port)
		if err != nil {
			log.Crit("Failed to create a basic host", "err", err)
		}

		// Create a new peer manager.
		pm := peer.NewPeerManager(utils.GetPeerIDFromPort(c.Port), host, signSixProtocol)
		err = pm.AddPeers(c.Peers)
		if err != nil {
			log.Crit("Failed to add peers", "err", err)
		}

		// Create a new service.
		service, err := NewService(c, pm)
		if err != nil {
			log.Crit("Failed to new service", "err", err)
		}
		// Set a stream handler on the host.
		host.SetStreamHandler(signSixProtocol, func(s network.Stream) {
			service.Handle(s)
		})

		// Ensure all peers are connected before starting signSix process.
		pm.EnsureAllConnected()

		// Start signSix process.
		service.Process()

		return nil
	},
}

func init() {
	Cmd.Flags().String("config", "", "signSix config file path")
}

func initService(cmd *cobra.Command) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	configFile = viper.GetString("config")

	return nil
}
