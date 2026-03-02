/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "reatemodule",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,

	Run: func(cmd *cobra.Command, args []string) {
		moduleName, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Fatalln(err)
		}

		if moduleName == "" {
			log.Fatalln("Module name is required")
		}

		path, err := cmd.Flags().GetString("path")
		if err != nil {
			os.Exit(1)
		}

		if path == "" {
			log.Fatalln("Path is required")
		}

		err = createModules(&moduleName, &path)
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println("Module created successfully")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

	rootCmd.Flags().StringP("name", "n", "", "Name of the module")
	rootCmd.Flags().StringP("path", "p", "", "Path of the module")
}
