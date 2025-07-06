package main

import (
	"Swaptimize/internal/swap"
    "Swaptimize/internal/monitor"
    "log"
    "os"
    "os/signal"
    "time"
    "context"
    "strconv"

    "github.com/joho/godotenv"
    
)

func main() {
    // Cargar archivo .env si existe
    _ = godotenv.Load("/etc/manage_swap.env")

    // Leer configuración desde entorno o usar valores por defecto
    intervalSec := getEnvInt("SWAP_SLEEP_INTERVAL", 10)
    thresholdHigh := getEnvInt("SWAP_THRESHOLD_HIGH", 85)
    thresholdLow := getEnvInt("SWAP_THRESHOLD_LOW", 40)
    maxSwapFiles := getEnvInt("MAX_SWAP_FILES", 4)
    swapSize := getEnvInt("SWAP_SIZE", 4096)

    log.Println("🔄 Swaptimize iniciado")
    log.Printf("Intervalo: %ds | Umbral alto: %d%% | Umbral bajo: %d%% | Máx archivos swap: %d\n",
        intervalSec, thresholdHigh, thresholdLow, maxSwapFiles)

    // Canal para capturar señales y permitir cierre limpio
    ctx, cancel := context.WithCancel(context.Background())
    go listenForSignals(cancel)

    // Bucle principal
    ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
    defer ticker.Stop()

    swap.CleanUpSwapFilesOnStartup() // 🧹 Al iniciar, limpiar swaps previos

    for {
        select {
        case <-ctx.Done():
            log.Println("🧹 Swaptimize detenido correctamente.")
            return
        case <-ticker.C:
            runCheck(thresholdHigh, thresholdLow, swapSize, maxSwapFiles)
        }
    }
}

var swapIDCounter int = 1

func runCheck(thresholdHigh int, thresholdLow int, swapSize int, maxSwapFiles int) {
    metrics, err := monitor.GetMetrics()
    if err != nil {
        log.Printf("⚠️ Error al obtener métricas: %v", err)
        return
    }

    log.Printf("📊 RAM: %.2f%% | Swap: %d%% | Disco libre: %dMB",
    metrics.MemPercent, metrics.SwapPercent, metrics.DiskFreeMB)

    if metrics.SwapPercent >= thresholdHigh {
        log.Println("🚀 Umbral alto alcanzado — creando nuevo swap")

        if swapIDCounter <= maxSwapFiles { // Evita superar máximo
            if err := swap.CreateSwapFile(swapIDCounter, swapSize); err != nil {
                log.Printf("❌ Error al crear swap: %v", err)
            } else {
                swapIDCounter++
            }
        } else {
            log.Println("⛔ Máximo de archivos swap alcanzado")
        }

    } else if metrics.SwapPercent <= thresholdLow && swapIDCounter > 1 {
        swapIDCounter--
        log.Println("🧽 Umbral bajo alcanzado — eliminando swap")
        if err := swap.RemoveSwapFile(swapIDCounter); err != nil {
            log.Printf("❌ Error al eliminar swap: %v", err)
        }
    }
}


func getEnvInt(key string, defaultVal int) int {
    valStr := os.Getenv(key)
    if valStr == "" {
        return defaultVal
    }
    val, err := strconv.Atoi(valStr)
    if err != nil {
        return defaultVal
    }
    return val
}

func listenForSignals(cancel context.CancelFunc) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt)
    <-sigChan
    cancel()
}
