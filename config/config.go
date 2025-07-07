package config

import (
    "os"
    "strconv"
    "github.com/joho/godotenv"
)

type Settings struct {
    SleepInterval  int
    ThresholdHigh  int
    ThresholdLow   int
    SwapSizeMB     int
    MaxSwapFiles   int
	SwapEmergencyInterval int
}

func LoadSettings(path string) (*Settings, error) {
    _ = godotenv.Load(path)

    return &Settings{
        SleepInterval:  getEnvInt("SWAP_SLEEP_INTERVAL", 30),
        ThresholdHigh:  getEnvInt("SWAP_THRESHOLD_HIGH", 85),
        ThresholdLow:   getEnvInt("SWAP_THRESHOLD_LOW", 40),
        SwapSizeMB:     getEnvInt("SWAP_SIZE", 4096),
        MaxSwapFiles:   getEnvInt("MAX_SWAP_FILES", 4),
		SwapEmergencyInterval: getEnvInt("SWAP_EMERGENCY_INTERVAL", 10),
    }, nil
}

func getEnvInt(key string, defaultVal int) int {
    valStr := os.Getenv(key)
    val, err := strconv.Atoi(valStr)
    if err != nil {
        return defaultVal
    }
    return val
}
