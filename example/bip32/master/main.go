package master

import (
	"github.com/getamis/alice/example/peer"
	"github.com/getamis/alice/example/utils"
	"github.com/getamis/sirius/log"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const masterProtocol = "/master/1.0.0"
const refreshProtocol = "/master-refresh/1.0.0"
const (
	circuitPath = "../crypto/circuit/bristolFashion/MPCSEED.txt"
)

var configFile string

var Cmd = &cobra.Command{
	Use:   "master",
	Short: "2 party bip32 master process",
	Long:  `Make masters(Alice and Bob) to support 2 party Bip32`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := initService(cmd)
		if err != nil {
			log.Crit("Failed to init", "err", err)
		}

		config, err := readMasterConfigFile(configFile)
		if err != nil {
			log.Crit("Failed to read config file", "configFile", configFile, "err", err)
		}

		host, err := peer.MakeBasicHost(config.Port)
		if err != nil {
			log.Crit("Failed to create a basic host", "err", err)
		}

		pm := peer.NewPeerManager(utils.GetPeerIDFromPort(config.Port), host, masterProtocol)
		err = pm.AddPeers(config.Peers)
		if err != nil {
			log.Crit("Failed to add peers", "err", err)
		}

		service, err := NewService(config, pm)
		if err != nil {
			log.Crit("Failed to new service", "err", err)
		}
		
		host.SetStreamHandler(masterProtocol, func(s network.Stream) {
			service.Handle(s)
		})
		
		pm.EnsureAllConnected()
		service.Process()
		// For refresh //
		masterResult, err := service.master.GetResult()
		if err != nil {
			log.Warn("Failed to get result from Master", "err", err)
		}
		host, err = peer.MakeBasicHost(config.Port)
		if err != nil {
			log.Crit("Failed to create a basic host", "err", err)
		}
		// Create a new peer manager.
		pm2 := peer.NewPeerManager(utils.GetPeerIDFromPort(config.Port), host, refreshProtocol)
		err = pm2.AddPeers(config.Peers)
		if err != nil {
			log.Crit("Failed to add peers", "err", err)
		}
		// Create a new service.
		refreshService, err := NewRefreshService(config, masterResult, pm2)
		if err != nil {
			log.Crit("Failed to new service", "err", err)
		}
		// Set a stream handler on the host.
		host.SetStreamHandler(refreshProtocol, func(s network.Stream) {
			refreshService.Handle(s)
		})
		// Ensure all peers are connected before starting DKG process.
		pm2.EnsureAllConnected()

		refreshService.Process()
		return nil
	},
}

func init() {
	Cmd.Flags().String("config", "", "master config file path")
}

func initService(cmd *cobra.Command) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	configFile = viper.GetString("config")

	return nil
}