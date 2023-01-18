package child

import (
	"github.com/getamis/alice/example/peer"
	"github.com/getamis/alice/example/utils"
	"github.com/getamis/sirius/log"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const childProtocol = "/child/1.0.0"
const (
	circuitPath = "../crypto/circuit/bristolFashion/MPCHMAC.txt"
)

var configFile string

var Cmd = &cobra.Command{
	Use:   "child",
	Short: "2 party bip32 child process",
	Long:  `Make child under 2 party bip32 master(Alice and Bob)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := initService(cmd)
		if err != nil {
			log.Crit("Failed to init", "err", err)
		}

		config, err := readChildConfigFile(configFile)
		if err != nil {
			log.Crit("Failed to read config file", "configFile", configFile, "err", err)
		}

		host, err := peer.MakeBasicHost(config.Port)
		if err != nil {
			log.Crit("Failed to create a basic host", "err", err)
		}

		pm := peer.NewPeerManager(utils.GetPeerIDFromPort(config.Port), host, childProtocol)
		err = pm.AddPeers(config.Peer)
		if err != nil {
			log.Crit("Failed to add peers", "err", err)
		}

		service, err := NewService(config, pm)
		if err != nil {
			log.Crit("Failed to new service", "err", err)
		}
		
		host.SetStreamHandler(childProtocol, func(s network.Stream) {
			service.Handle(s)
		})
		
		pm.EnsureAllConnected()
		service.Process()
		return nil
	},
}

func init() {
	Cmd.Flags().String("config", "", "child config file path")
}

func initService(cmd *cobra.Command) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	configFile = viper.GetString("config")

	return nil
}