package cmd

import (
    "github.com/spf13/cobra"

    "fmt"
    "log"
    "Swaptimize/internal/monitor"
)

var statusCmd = &cobra.Command{
    Use:   "status",
    Short: "Muestra las métricas actuales del sistema",
    Run: func(cmd *cobra.Command, args []string) {
        metrics, err := monitor.GetMetrics()
        if err != nil {
            log.Fatalf("Error: %v", err)
        }
        fmt.Println("📊 Estado del sistema:")
        fmt.Printf("  🔹 RAM:   %.2f%%\n", metrics.MemPercent)
        fmt.Printf("  🔹 Swap:  %d%%\n", metrics.SwapPercent)
        fmt.Printf("  🔹 Disco: %d MB libres\n", metrics.DiskFreeMB)
    },
}

func init() {
    rootCmd.AddCommand(statusCmd)
}
