package main

import (
	"EasyBaseLine/core"
	"EasyBaseLine/utils"
	"fmt"
	"github.com/spf13/cobra"
)

var (
	// Flags for remote mode
	assetsFile     string
	remoteHost     string
	remoteUser     string
	remotePassword string
	remoteProto    string
	remotePort     int

	// Flag for server mode
	serverIP string

	// checkFile
	checkFile string
)

func init() {
	utils.Banner()
}

func main() {
	var rootCmd = &cobra.Command{Use: "EasyBaseLine"}
	var app = core.NewApplication()

	var localCmd = &cobra.Command{
		Use:   "local",
		Short: "Local mode (default)",
		Run: func(cmd *cobra.Command, args []string) {
			utils.Info(fmt.Sprintln("Running in local mode"))
			app.CheckFile = checkFile
			app.Run()
		},
	}

	localCmd.Flags().StringVar(&checkFile, "checkfile", "", "Specify a checkfile")

	var remoteCmd = &cobra.Command{
		Use:   "remote",
		Short: "Remote mode",
		Run: func(cmd *cobra.Command, args []string) {
			if assetsFile != "" {
				utils.Info(fmt.Sprintf("Running in remote mode with file: %s\n", assetsFile))
				app.File = assetsFile
				app.CheckFile = checkFile
				app.Remote = true
				app.Run()
			} else if remoteHost != "" && remoteUser != "" && remotePassword != "" {
				utils.Info(fmt.Sprintf("Running in remote mode with host: %s, user: %s, password: ******\n", remoteHost, remoteUser))
				app.Remote = true
				app.CheckFile = checkFile
				app.Username = remoteUser
				app.Password = remotePassword
				app.Host = remoteHost
				app.Port = remotePort
				app.Proto = remoteProto
				app.Run()
			} else {
				cmd.Help()
			}
		},
	}

	remoteCmd.Flags().StringVar(&assetsFile, "file", "", "Specify a assetfile")
	remoteCmd.Flags().StringVar(&checkFile, "checkfile", "", "Specify a checkfile")
	remoteCmd.Flags().StringVar(&remoteHost, "host", "", "Specify a host")
	remoteCmd.Flags().StringVar(&remoteUser, "user", "", "Specify a user")
	remoteCmd.Flags().StringVar(&remotePassword, "password", "", "Specify a password")
	remoteCmd.Flags().IntVarP(&remotePort, "port", "", 0, "Specify a port,such as 22,5985")
	remoteCmd.Flags().StringVar(&remoteProto, "proto", "", "Specify a proto,winrm or ssh")

	var serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Server mode",
		Run: func(cmd *cobra.Command, args []string) {
			foldersToCheck := []string{"config/json", "config/scripts", "checkItems"}
			for _, folder := range foldersToCheck {
				utils.CheckAndCreateFolder(folder)
			}
			utils.Info(fmt.Sprintln("Running in server mode"))
			app.Ip = serverIP
			app.ServerMode = true
			app.RunServer()
		},
	}

	serverCmd.Flags().StringVar(&serverIP, "ip", "", "Specify an IP address")

	rootCmd.AddCommand(localCmd, remoteCmd, serverCmd)

	if err := rootCmd.Execute(); err != nil {
		utils.Fatal(err.Error())
	}
}
