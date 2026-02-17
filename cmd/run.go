package cmd

import (
    "context"
    "fmt"
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
    "Swaptimize/internal/metrics"
    "Swaptimize/internal/system"
)

// MetricsHistory mantiene un histórico circular de métricas para detectar picos
type MetricsHistory struct {
	Timestamps  [5]time.Time
	MemPercent  [5]float64
	SwapPercent [5]int
	PSIValue    [5]float64
	Index       int
	IsFull      bool
}

// Add registra nuevas métricas en el histórico circular
func (mh *MetricsHistory) Add(m *monitor.SystemMetrics) {
	mh.Timestamps[mh.Index] = time.Now()
	mh.MemPercent[mh.Index] = m.MemPercent
	mh.SwapPercent[mh.Index] = m.SwapPercent
	mh.PSIValue[mh.Index] = m.PSIValue

	mh.Index = (mh.Index + 1) % 5
	if mh.Index == 0 {
		mh.IsFull = true
	}
}

// GetDelta obtiene el máximo delta de MemPercent en los últimos 60s y el tiempo más antiguo
func (mh *MetricsHistory) GetDelta() (float64, time.Duration) {
	if !mh.IsFull && mh.Index < 2 {
		return 0, 0 // No hay suficientes datos
	}

	oldestIdx := mh.Index
	if mh.IsFull {
		// El índice actual apunta al siguiente a sobrescribir, que es el más antiguo
	} else if mh.Index > 0 {
		oldestIdx = 0 // Si no está lleno, empezar del principio
	} else {
		return 0, 0
	}

	newestIdx := (mh.Index - 1 + 5) % 5
	newestTime := mh.Timestamps[newestIdx]
	oldestTime := mh.Timestamps[oldestIdx]

	if newestTime.Before(oldestTime) || newestTime.Equal(oldestTime) {
		return 0, 0 // Datos no válidos
	}

	timeDiff := newestTime.Sub(oldestTime)
	memDelta := mh.MemPercent[newestIdx] - mh.MemPercent[oldestIdx]

	return memDelta, timeDiff
}

// GetSwapDelta obtiene el delta de SwapPercent en el histórico
func (mh *MetricsHistory) GetSwapDelta() (int, time.Duration) {
	if !mh.IsFull && mh.Index < 2 {
		return 0, 0 // No hay suficientes datos
	}

	oldestIdx := mh.Index
	if mh.IsFull {
		// El índice actual apunta al siguiente a sobrescribir, que es el más antiguo
	} else if mh.Index > 0 {
		oldestIdx = 0 // Si no está lleno, empezar del principio
	} else {
		return 0, 0
	}

	newestIdx := (mh.Index - 1 + 5) % 5
	newestTime := mh.Timestamps[newestIdx]
	oldestTime := mh.Timestamps[oldestIdx]

	if newestTime.Before(oldestTime) || newestTime.Equal(oldestTime) {
		return 0, 0 // Datos no válidos
	}

	timeDiff := newestTime.Sub(oldestTime)
	swapDelta := mh.SwapPercent[newestIdx] - mh.SwapPercent[oldestIdx]

	return swapDelta, timeDiff
}

// ShouldDetectPeak evalúa si hay un pico de MEMORIA que requiera respuesta inmediata
// NOTA: Solo detecta picos de memoria principal, no de swap (que es síntoma, no causa)
func (mh *MetricsHistory) ShouldDetectPeak(metrics *monitor.SystemMetrics, psiHighThreshold float64) bool {
	// Criterio 1: Delta de MemPercent > 10% en menos de 60 segundos (memoria principal saltó)
	memDelta, timeDelta := mh.GetDelta()
	if memDelta > 10.0 && timeDelta > 0 && timeDelta < 60*time.Second {
		return true
	}

	// Criterio 2: PSI "some" avg10 > umbral (presión extrema del kernel)
	if metrics.PSIValue > psiHighThreshold {
		return true
	}

	return false
}

// IsSwapFillingFast detecta si el swap se está llenando rápidamente
func (mh *MetricsHistory) IsSwapFillingFast(deltaThreshold int, window time.Duration) bool {
	swapDelta, timeDelta := mh.GetSwapDelta()
	if swapDelta > deltaThreshold && timeDelta > 0 && timeDelta < window {
		return true
	}
	return false
}

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

        // Inicializar Prometheus Metrics Server
        promMetrics := metrics.NewPrometheusMetrics(9100)
        if err := promMetrics.Start(); err != nil {
            log.Fatalf("❌ Error al iniciar servidor de métricas: %v", err)
        }

        // Detectar ambiente del sistema
        env, err := system.DetectEnvironment()
        if err != nil {
            log.Fatalf("❌ Error al detectar ambiente: %v", err)
        }

        // Crear gestor de perfiles
        profileMgr := system.NewProfileManager(env)
        profileMgr.LogConfiguration()

        // Usar valores del PERFIL, no los defaults de configuración
        profile := profileMgr.GetCurrentProfile()
        effectiveSwapSizeMB := profile.SwapSize
        effectiveMaxSwapFiles := profileMgr.AdaptiveMaxSwapFiles()
        
        // Sobrescribir settings con valores del perfil
        settings.SwapSizeMB = effectiveSwapSizeMB
        settings.MaxSwapFiles = effectiveMaxSwapFiles
        promMetrics.UpdateConfiguration(
            string(env.Type),
            env.Recommendation,
            profile.Name,
            profileMgr.GetModeDescription(),
            profileMgr.AdaptiveThresholdHigh(),
            profile.ThresholdLow,
            profileMgr.AdaptiveMaxSwapFiles(),
            profile.SwapSize,
            profileMgr.ShouldAllowCreation(),
        )

        // Estado inicial
        swapIDCounter := 1
        lastSwapActionTime := time.Now()
        lastSwapAction := "none" // "create" o "delete"
        metricsHistory := &MetricsHistory{}
        
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

                // Actualizar métricas en Prometheus
                promMetrics.UpdateGauges(
                    metrics.MemUsageBytes,
                    metrics.MemAvailableBytes,
                    metrics.MemPercent,
                    metrics.SwapUsageBytes,
                    metrics.SwapAvailBytes,
                    float64(metrics.SwapPercent),
                )

                // Agregar métricas al histórico para detección de picos
                metricsHistory.Add(metrics)

                // Identificar arranque en frío sin swap activa
                isBootCold := !hasSwap && swapIDCounter == 1
                if isBootCold {
                    log.Println("⚠️ Sistema sin swap activa. Swaptimize iniciará con swap dinámica.")
                }

                // Crear swap si el uso de MEMORIA ≥ umbral alto o si el SWAP se está llenando rápido
                // LÓGICA DUAL-PATH v2.1:
                // - Si hay PICO detectado: esperar solo 10s antes de crear (respuesta inmediata)
                // - Si es cambio gradual: mantener conservador con 300s de rate-limit
                
                timeSinceLastAction := time.Since(lastSwapActionTime)
                isPeak := metricsHistory.ShouldDetectPeak(metrics, settings.PSIHighThreshold)
                isSwapFillingFast := metricsHistory.IsSwapFillingFast(settings.SwapDeltaThreshold, 
                                                                       time.Duration(settings.SwapDeltaWindowSec)*time.Second)
                
                // Si hay pico, swap llenándose rápido O swap ya al 85%+ = respuesta inmediata (10s)
                // Si no = cambio gradual (300s)
                var minIntervalSeconds float64 = 300 // Cambio gradual
                if isPeak || isSwapFillingFast || metrics.SwapPercent >= 85 {
                    minIntervalSeconds = 10 // Respuesta inmediata
                }
                
                canCreate := (lastSwapAction != "delete") || (timeSinceLastAction > 120*time.Second)
                canCreate = canCreate && ((lastSwapAction != "create") || (timeSinceLastAction > time.Duration(minIntervalSeconds)*time.Second))
                
                // Evaluar si necesitamos un nuevo swap
                // CRITERIOS OBLIGATORIOS: 
                //   - MemPercent alto (>=85%) 
                //   - O SwapPercent ya al 85%+ (está siendo usado intensamente)
                //   - O arranque en frío
                shouldCreateNewSwap := (metrics.MemPercent >= float64(settings.ThresholdHigh) || 
                                       metrics.SwapPercent >= 85 ||
                                       isBootCold) && canCreate
                
                // Protección: Si hay swap activo, verificar que se está usando significativamente
                // EXCEPCIÓN: Si SwapPercent >= 85%, está claramente siendo usado, crear siempre
                if shouldCreateNewSwap && swapIDCounter > 1 && metrics.SwapPercent < 85 {
                    if metrics.SwapPercent < 60 {
                        // El swap actual no está siendo usado (< 60%), no crear uno nuevo
                        log.Printf("📊 Memoria alta (%.1f%%), pero swap actual solo al %d%% → no creando swap innecesario", 
                            metrics.MemPercent, metrics.SwapPercent)
                        shouldCreateNewSwap = false
                    }
                }
                
                if shouldCreateNewSwap {
                    if swapIDCounter <= settings.MaxSwapFiles {
                        if err := swap.CreateSwapFile(swapIDCounter, settings.SwapSizeMB); err != nil {
                            log.Printf("❌ Error al crear swap: %v", err)
                        } else {
                            logMsg := "🛠️ Swap creado"
                            if metrics.SwapPercent >= 85 {
                                logMsg += " [SWAP 85%+]"
                            } else if isPeak {
                                logMsg += " [PICO MEMORIA]"
                            } else if isSwapFillingFast {
                                logMsg += " [SWAP LLENÁNDOSE]"
                            } else {
                                logMsg += " [gradual]"
                            }
                            logMsg += fmt.Sprintf(" (MemPercent=%.1f%%, SwapPercent=%d%%, PSI=%.2f%%)", 
                                metrics.MemPercent, metrics.SwapPercent, metrics.PSIValue)
                            log.Println(logMsg)
                            swapIDCounter++
                            lastSwapActionTime = time.Now()
                            lastSwapAction = "create"
                        }
                    } else {
                        log.Println("⛔ Máximo de archivos swap alcanzado.")
                    }
                }

                // Eliminar swap si uso de MEMORIA ≤ umbral bajo y hay más de los mínimos activos
                // Con hysteresis: si acabamos de crear, esperar 30s antes de eliminar
                canDelete := (lastSwapAction != "create") || (timeSinceLastAction > 30*time.Second)
                
                if (metrics.MemPercent <= float64(settings.ThresholdLow) && swapIDCounter > minSwapActive) && canDelete {
                    swapIDCounter--
                    if err := swap.RemoveSwapFile(swapIDCounter); err != nil {
                        log.Printf("❌ Error al eliminar swap: %v", err)
                    } else {
                        log.Printf("📊 Swap eliminado. Esperando 120s antes de crear nuevamente.")
                        lastSwapActionTime = time.Now()
                        lastSwapAction = "delete"
                    }
                }

                // Ajustar intervalo si swap ≥ 90%
                if metrics.SwapPercent >= 90 {
                    dynamicInterval = time.Duration(settings.SwapEmergencyInterval) * time.Second
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
