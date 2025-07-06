package main

import (
    "fmt"
    "log"
    "Swaptimize/internal/monitor"
)

func main() {
    metrics, err := monitor.GetMetrics()
    if err != nil {
        log.Fatalf("Error al obtener métricas: %v", err)
    }

    fmt.Println("📊 Estado del sistema:")
    fmt.Printf("  🔹 RAM usada:      %.2f%%\n", metrics.MemPercent)
    fmt.Printf("  🔹 Swap usada:     %d%%\n", metrics.SwapPercent)
    fmt.Printf("  🔹 Disco libre:    %d MB\n", metrics.DiskFreeMB)
}
