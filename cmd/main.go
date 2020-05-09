package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	// "log"
	"os"
	"time"

	"github.com/pkg/profile"
)

func init() {
	initConfig()
	// cobra.OnInitialize()
	rootCmd.AddCommand(versionCmd)
}

func main() {
	// 增加profile
	var prof func(*profile.Profile)
	switch config.Prof {
	case "cpu":
		prof = profile.CPUProfile
	case "mem":
		prof = profile.MemProfile
	default:
		prof = nil
	}
	if prof != nil {
		defer profile.Start(prof, profile.ProfilePath("./prof")).Stop()
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use: "gencnf",
	Run: func(cmd *cobra.Command, args []string) {

		start := time.Now()
		gen := newGenerator()
		gen.GenConfig(config)
		gen.Close()
		dur := time.Since(start)
		gen.PrintErrors()
		fmt.Println("用时：", dur.String())

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
	fmt.Printf("loaded config from %s \n", viper.ConfigFileUsed())
}
