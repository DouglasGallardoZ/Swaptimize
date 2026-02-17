package monitor

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// PSIMetrics contiene métricas de Pressure Stall Info
type PSIMetrics struct {
	// "some" = al menos una tarea está en stall
	SomeAvg10  float64 // Último 10 segundos
	SomeAvg60  float64 // Último 60 segundos
	SomeAvg300 float64 // Último 300 segundos
	SomeTotal  float64 // Total acumulado

	// "full" = todas las tareas están en stall
	FullAvg10  float64
	FullAvg60  float64
	FullAvg300 float64
	FullTotal  float64

	Timestamp time.Time
}

// PSIReader lee y parsea /proc/pressure/memory
type PSIReader struct {
	mu        sync.RWMutex
	lastRead  *PSIMetrics
	readCount int
	errors    int
}

// NewPSIReader crea un nuevo lector de PSI
func NewPSIReader() *PSIReader {
	return &PSIReader{
		lastRead: &PSIMetrics{},
	}
}

// ReadMemoryPSI lee y retorna las métricas actuales de memoria
func (pr *PSIReader) ReadMemoryPSI() (*PSIMetrics, error) {
	file, err := os.Open("/proc/pressure/memory")
	if err != nil {
		pr.mu.Lock()
		pr.errors++
		pr.mu.Unlock()
		return nil, fmt.Errorf("cannot open /proc/pressure/memory: %w", err)
	}
	defer file.Close()

	metrics := &PSIMetrics{
		Timestamp: time.Now(),
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "some") {
			parseSOMELine(line, metrics)
		} else if strings.HasPrefix(line, "full") {
			parseFULLLine(line, metrics)
		}
	}

	if err := scanner.Err(); err != nil {
		pr.mu.Lock()
		pr.errors++
		pr.mu.Unlock()
		return nil, err
	}

	pr.mu.Lock()
	pr.lastRead = metrics
	pr.readCount++
	pr.mu.Unlock()

	return metrics, nil
}

// parseSOMELine parsea línea "some"
func parseSOMELine(line string, metrics *PSIMetrics) {
	// Formato: some avg10=X.XX avg60=Y.YY avg300=Z.ZZ total=NNN
	fields := strings.Fields(line)
	for _, field := range fields {
		if strings.HasPrefix(field, "avg10=") {
			metrics.SomeAvg10, _ = strconv.ParseFloat(strings.TrimPrefix(field, "avg10="), 64)
		} else if strings.HasPrefix(field, "avg60=") {
			metrics.SomeAvg60, _ = strconv.ParseFloat(strings.TrimPrefix(field, "avg60="), 64)
		} else if strings.HasPrefix(field, "avg300=") {
			metrics.SomeAvg300, _ = strconv.ParseFloat(strings.TrimPrefix(field, "avg300="), 64)
		} else if strings.HasPrefix(field, "total=") {
			metrics.SomeTotal, _ = strconv.ParseFloat(strings.TrimPrefix(field, "total="), 64)
		}
	}
}

// parseFULLLine parsea línea "full"
func parseFULLLine(line string, metrics *PSIMetrics) {
	// Mismo formato que "some"
	fields := strings.Fields(line)
	for _, field := range fields {
		if strings.HasPrefix(field, "avg10=") {
			metrics.FullAvg10, _ = strconv.ParseFloat(strings.TrimPrefix(field, "avg10="), 64)
		} else if strings.HasPrefix(field, "avg60=") {
			metrics.FullAvg60, _ = strconv.ParseFloat(strings.TrimPrefix(field, "avg60="), 64)
		} else if strings.HasPrefix(field, "avg300=") {
			metrics.FullAvg300, _ = strconv.ParseFloat(strings.TrimPrefix(field, "avg300="), 64)
		} else if strings.HasPrefix(field, "total=") {
			metrics.FullTotal, _ = strconv.ParseFloat(strings.TrimPrefix(field, "total="), 64)
		}
	}
}

// GetLastMetrics retorna las últimas métricas leídas
func (pr *PSIReader) GetLastMetrics() *PSIMetrics {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	if pr.lastRead == nil {
		return &PSIMetrics{}
	}
	return pr.lastRead
}

// Stats retorna estadísticas del reader
func (pr *PSIReader) Stats() map[string]interface{} {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	return map[string]interface{}{
		"reads":   pr.readCount,
		"errors":  pr.errors,
		"last_ts": pr.lastRead.Timestamp,
	}
}

// LogMetrics registra las métricas en logs
func (pr *PSIReader) LogMetrics(metrics *PSIMetrics) {
	log.Printf("📊 PSI Memory Pressure:\n")
	log.Printf("  some: avg10=%.2f%% avg60=%.2f%% avg300=%.2f%%\n",
		metrics.SomeAvg10, metrics.SomeAvg60, metrics.SomeAvg300)
	log.Printf("  full: avg10=%.2f%% avg60=%.2f%% avg300=%.2f%%\n",
		metrics.FullAvg10, metrics.FullAvg60, metrics.FullAvg300)
}

// IsPressureHigh verifica si hay presión alta
func (metrics *PSIMetrics) IsPressureHigh(threshold float64) bool {
	// Usar "some" avg10 como indicador principal
	return metrics.SomeAvg10 > threshold
}

// GetAveragePressure retorna el promedio de presión
func (metrics *PSIMetrics) GetAveragePressure() float64 {
	return (metrics.SomeAvg10 + metrics.SomeAvg60 + metrics.SomeAvg300) / 3.0
}

// IsPSIAvailable verifica si PSI está disponible en el kernel
func IsPSIAvailable() bool {
	_, err := os.Stat("/proc/pressure/memory")
	return err == nil
}
