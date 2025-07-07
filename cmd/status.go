package cmd

import (
    "fmt"
    "log"

    "github.com/spf13/cobra"
    "Swaptimize/internal/monitor"
    "Swaptimize/internal/swap"
)

var statusCmd = &cobra.Command{
    Use:   "status",
    Short: "Muestra las métricas actuales del sistema",
    Run: func(cmd *cobra.Command, args []string) {
        metrics, err := monitor.GetMetrics()
        if err != nil {
            log.Fatalf("❌ Error al obtener métricas: %v", err)
        }

        swapCount, err := swap.CountActiveSwapFiles()
        if err != nil {
            log.Printf("⚠️ Error al contar archivos swap: %v", err)
        }

        fmt.Println("📊 Estado del sistema:")
        fmt.Printf("  🔹 RAM:   %.2f%%\n", metrics.MemPercent)
        fmt.Printf("  🔹 Swap:  %d%%\n", metrics.SwapPercent)
        fmt.Printf("  🔹 Disco: %d MB libres\n", metrics.DiskFreeMB)
        if err == nil {
            fmt.Printf("  🔹 Swaps activos gestionados: %d\n", swapCount)
        }
    },
}

func init() {
    rootCmd.AddCommand(statusCmd)
}
