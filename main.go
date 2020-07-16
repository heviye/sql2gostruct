package main

import (
	"log"

	"github.com/heviye/sql2gostruct/db"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "",
	Short: "",
	Long:  "将SQL的表结构转换成Golang结构",
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(db.OracleCmd)
}

func main() {
	err := Execute()
	if err != nil {
		log.Fatalf("cmd.Execute err: %v", err)
	}
}
