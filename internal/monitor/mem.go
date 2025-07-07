package monitor

import (
    "github.com/shirou/gopsutil/v3/mem"
    "github.com/shirou/gopsutil/v3/disk"
)

// Métricas relevantes del sistema
type SystemMetrics struct {
    MemPercent float64 // Porcentaje RAM usada
    SwapPercent int    // Porcentaje swap usada
    DiskFreeMB uint64  // Espacio libre en disco (MB)
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

    totalSwap := 0
    if swapStats.Total > 0 {
        totalSwap = int((swapStats.Used * 100) / swapStats.Total)
    }

	return &SystemMetrics{
		MemPercent:   v.UsedPercent,
		SwapPercent:  totalSwap,
		DiskFreeMB:   d.Free / (1024 * 1024),
	}, nil
}
