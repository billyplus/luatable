package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use: "gencnf",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var config Config

func initConfig() {
	viper.SetConfigFile("./config.json")
	viper.SetConfigType("json")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("err reading config:", err.Error())
		os.Exit(1)
	}
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Println("err unmarshal config:", err.Error())
		os.Exit(1)
	}
	fmt.Printf("%+v", config)
	fmt.Printf("loaded config from %s ", viper.ConfigFileUsed())
}
