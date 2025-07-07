package cmd

import (
    "context"
    "log"
    "os"
    "os/signal"
    "time"

    "github.com/spf13/cobra"
    "Swaptimize/config"
    "Swaptimize/internal/monitor"
    "Swaptimize/internal/swap"
)

var runCmd = &cobra.Command{
    Use:   "run",
    Short: "Ejecuta el demonio Swaptimize",
    Long:  "Inicia el daemon que gestiona dinámicamente la swap del sistema según métricas de memoria.",
    Run: func(cmd *cobra.Command, args []string) {
        log.Println("🔄 Swaptimize iniciado")

        // Cargar configuración desde archivo .env
        settings, err := config.LoadSettings("/etc/swaptimize.env")
        if err != nil {
            log.Fatalf("❌ Error al cargar configuración: %v", err)
        }

        // Limpiar archivos swap residuales al arrancar
        swap.CleanUpSwapFilesOnStartup()

        // Preparar señal de interrupción
        ctx, cancel := context.WithCancel(context.Background())
        go listenForSignals(cancel)

        // Variables de intervalo
        defaultInterval := time.Duration(settings.SleepInterval) * time.Second
        dynamicInterval := defaultInterval

        // Estado de archivos swap
        swapIDCounter := 1

        // Existe swap inicial
        metricsAux, err := monitor.GetMetrics()
        hasSwap := metricsAux.TotalSwap

        counter := 1
        if hasSwap == 0 {
            counter = 2
        }

        for {
            select {
            case <-ctx.Done():
                log.Println("🧹 Swaptimize detenido correctamente.")
                return
            default:
                metrics, err := monitor.GetMetrics()
                if err != nil {
                    log.Printf("⚠️ Error al obtener métricas: %v", err)
                } else {
                    //log.Printf("📊 RAM: %.2f%% | Swap: %d%% | Disco libre: %dMB",
                    //    metrics.MemPercent, metrics.SwapPercent, metrics.DiskFreeMB)

                    // Crear nuevo swap si uso ≥ umbral alto y no se ha superado el máximo
                    if metrics.SwapPercent >= settings.ThresholdHigh || (hasSwap == 0 && swapIDCounter == 1) {
                        if swapIDCounter <= settings.MaxSwapFiles {
                            if err := swap.CreateSwapFile(swapIDCounter, settings.SwapSizeMB); err != nil {
                                log.Printf("❌ Error al crear swap: %v", err)
                            } else {
                                swapIDCounter++
                            }
                        } else {
                            log.Println("⛔ Máximo de archivos swap alcanzado.")
                        }
                    }
                    
                    // Eliminar swap si uso ≤ umbral bajo y hay al menos uno activado
                    if metrics.SwapPercent <= settings.ThresholdLow && swapIDCounter > counter {
                        swapIDCounter--
                        if err := swap.RemoveSwapFile(swapIDCounter); err != nil {
                            log.Printf("❌ Error al eliminar swap: %v", err)
                        }
                    }

                    // Ajustar intervalo dinámico si swap ≥ 90%
                    if metrics.SwapPercent >= 90 {
                        dynamicInterval = time.Duration(settings.SwapEmergencyInterval) * time.Second
                        log.Printf("⚡ Swap ≥ 90%%: ajustando intervalo a %ds",
                            settings.SwapEmergencyInterval)
                    } else {
                        dynamicInterval = defaultInterval
                    }
                }

                time.Sleep(dynamicInterval)
            }
        }
    },
}

func listenForSignals(cancel context.CancelFunc) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt)
    <-sigChan
    cancel()
}

func init() {
    rootCmd.AddCommand(runCmd)
}
