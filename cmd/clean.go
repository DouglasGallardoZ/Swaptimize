package cmd

import (
	"github.com/spf13/cobra"
	
    "Swaptimize/internal/swap"
)

var cleanCmd = &cobra.Command{
    Use:   "clean",
    Short: "Elimina todos los archivos swap activos",
    Run: func(cmd *cobra.Command, args []string) {
        swap.CleanUpSwapFilesOnStartup()
    },
}

func init() {
    rootCmd.AddCommand(cleanCmd)
}
