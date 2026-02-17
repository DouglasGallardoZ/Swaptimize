package monitor

import (
	"sync"

	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/disk"
)

// Métricas relevantes del sistema
type SystemMetrics struct {
	MemPercent        float64 // Porcentaje RAM usada
	MemUsageBytes     float64 // RAM usada en bytes
	MemAvailableBytes float64 // RAM disponible en bytes
	SwapPercent       int     // Porcentaje swap usada
	SwapUsageBytes    float64 // Swap usado en bytes
	SwapAvailBytes    float64 // Swap disponible en bytes
	DiskFreeMB        uint64  // Espacio libre en disco (MB)
	TotalSwap         int     // Capacidad total de swap en bytes. Si es 0, no hay swap activa.
	TotalSwapUsed     int     // Swap usado en bytes (Fase 1 addition)
	PSIValue          float64 // PSI "some" avg10 (presión de memoria) en porcentaje
}

// PSI Reader global (singleton)
var (
	psiReader *PSIReader
	psiMutex  sync.Once
)

// getPSIReader inicializa y retorna el PSIReader global
func getPSIReader() *PSIReader {
	psiMutex.Do(func() {
		psiReader = NewPSIReader()
	})
	return psiReader
}

// Extrae métricas actuales del sistema
func GetMetrics() (*SystemMetrics, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	d, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	swapStats, err := mem.SwapMemory()
	if err != nil {
		return nil, err
	}

	totalSwapPercent := 0
	if swapStats.Total > 0 {
		totalSwapPercent = int((swapStats.Used * 100) / swapStats.Total)
	}

	// Leer PSI value (presión de memoria)
	psiValue := 0.0
	reader := getPSIReader()
	if psiMetrics, err := reader.ReadMemoryPSI(); err == nil {
		psiValue = psiMetrics.SomeAvg10  // Usar "some" avg10 como indicador principal
	}

	return &SystemMetrics{
		MemPercent:        v.UsedPercent,
		MemUsageBytes:     float64(v.Used),
		MemAvailableBytes: float64(v.Available),
		SwapPercent:       totalSwapPercent,
		SwapUsageBytes:    float64(swapStats.Used),
		SwapAvailBytes:    float64(swapStats.Total - swapStats.Used),
		DiskFreeMB:        d.Free / (1024 * 1024),
		TotalSwap:         int(swapStats.Total),
		TotalSwapUsed:     int(swapStats.Used),
		PSIValue:          psiValue,
	}, nil
}
