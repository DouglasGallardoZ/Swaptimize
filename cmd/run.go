package cmd

import (
    "context"
    "log"
    "os"
    "os/signal"
    "os/exec"
    "time"
    "strings"

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

        // Estado inicial
        swapIDCounter := 1

        // Limpiar archivos swap residuales al arrancar
        hasSwap := true
        if IsSystemBootRecent() {
            swap.CleanUpSwapFilesOnStartup()
            initialMetrics, err := monitor.GetMetrics()
            if err != nil {
                log.Fatalf("❌ Error al obtener métricas iniciales: %v", err)
            }
            hasSwap = initialMetrics.TotalSwap > 0
        } else {
            log.Println("🔁 Reinicio del servicio detectado — preservando swap activa.")
            swapIDCounter, err = swap.CountActiveSwapFiles()
            
            if err != nil {
                log.Fatalf("❌ Error al contar archivos swap: %v", err)
            }
            
            swapIDCounter++
        }

        // Preparar señal de interrupción
        ctx, cancel := context.WithCancel(context.Background())
        go listenForSignals(cancel)

        // Intervalos de chequeo
        defaultInterval := time.Duration(settings.SleepInterval) * time.Second
        dynamicInterval := defaultInterval
        
        if err != nil {
            log.Fatalf("❌ Error al obtener métricas iniciales: %v", err)
        }

        // Mínimo swap activo permitido (protege la swap inicial)
        minSwapActive := 1
        if !hasSwap {
            minSwapActive = 2
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
                    time.Sleep(dynamicInterval)
                    continue
                }

                // Identificar arranque en frío sin swap activa
                isBootCold := !hasSwap && swapIDCounter == 1
                if isBootCold {
                    log.Println("⚠️ Sistema sin swap activa. Swaptimize iniciará con swap dinámica.")
                }

                // Crear swap si el uso ≥ umbral alto o es arranque en frío
                if metrics.SwapPercent >= settings.ThresholdHigh || isBootCold {
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

                // Eliminar swap si uso ≤ umbral bajo y hay más de los mínimos activos
                if metrics.SwapPercent <= settings.ThresholdLow && swapIDCounter > minSwapActive {
                    swapIDCounter--
                    if err := swap.RemoveSwapFile(swapIDCounter); err != nil {
                        log.Printf("❌ Error al eliminar swap: %v", err)
                    }
                }

                // Ajustar intervalo si swap ≥ 90%
                if metrics.SwapPercent >= 90 {
                    dynamicInterval = time.Duration(settings.SwapEmergencyInterval) * time.Second
                    log.Printf("⚡ Swap ≥ 90%%: ajustando intervalo a %ds", settings.SwapEmergencyInterval)
                } else {
                    dynamicInterval = defaultInterval
                }

                time.Sleep(dynamicInterval)
            }
        }
    },
}

func IsSystemBootRecent() bool {
    out, err := exec.Command("uptime", "-s").Output()
    if err != nil {
        return false
    }
    bootTimeStr := strings.TrimSpace(string(out))
    bootTime, err := time.Parse("2006-01-02 15:04:05", bootTimeStr)
    if err != nil {
        return false
    }

    return time.Since(bootTime) < 3*time.Minute
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
