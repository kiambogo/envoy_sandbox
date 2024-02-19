package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "gRPC Server and Client",
	Long:  "A simple gRPC server and client",
}

func main() {
	rootCmd.Execute()
}
