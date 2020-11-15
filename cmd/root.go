package cmd

import (
	"github.com/govice/golinksd/pkg/daemon"
	"github.com/govice/golinksd/pkg/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Execute() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "golinksd",
	Short: "golinksd is a daemon for managing filesystem integrity over time",
	Run: func(cmd *cobra.Command, args []string) {
		d, err := daemon.New()
		if err != nil {
			log.Fatalln(err)
		}

		log.Logln("PORT: " + viper.GetString("port"))
		log.Logln("AUTH_SERVER: " + viper.GetString("auth_server"))

		if err := d.Execute(); err != nil {
			log.Fatalln(err)
		}

		if err := d.StopDaemon(); err != nil {
			log.Fatalln(err)
		}
	},
}
