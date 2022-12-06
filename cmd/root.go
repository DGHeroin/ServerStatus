package cmd

import (
    "fmt"
    "github.com/DGHeroin/ServerStatus/cmd/agent"
    "github.com/DGHeroin/ServerStatus/cmd/server"
    "github.com/spf13/cobra"
    "os"
)

var (
    rootCmd = &cobra.Command{}
)

func Run() {
    rootCmd.AddCommand(server.Cmd, agent.Cmd)
    if err := rootCmd.Execute(); err != nil {
        _, _ = fmt.Fprintln(os.Stderr, err)
    }
}
