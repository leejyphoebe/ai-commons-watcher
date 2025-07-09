package main

import (
	"fmt"
	"os"

	"ai-commons/cli"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ai-commons",
	Short: "AI Commons CLI",
	Long:  `A command line interface for submitting and managing AI jobs on the NSCC cluster.`,
}

func main() {
	godotenv.Load(".env")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(cli.InitCmd)
	rootCmd.AddCommand(cli.RunCmd)
	cli.InitCmd.PersistentFlags().String("config", "", "Path to ai-commons config file.")
	cli.RunCmd.PersistentFlags().String("config", "", "Path to ai-commons config file.")
}
