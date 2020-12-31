package main

import (
	"fmt"
	"os"

	"github.com/go-kit/kit/log/level"
	"github.com/spf13/cobra"
	"github.com/zcong1993/x/pkg/log"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "log",
		Short: "Test cli",
	}

	log.Register(rootCmd.PersistentFlags())

	var testCmd = &cobra.Command{
		Use:   "test",
		Short: "Test sub command",
		Run: func(cmd *cobra.Command, args []string) {
			logger := log.MustNewLogger(cmd)
			level.Info(logger).Log("msg", "test")
		},
	}

	rootCmd.AddCommand(testCmd)

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
