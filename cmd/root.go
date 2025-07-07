package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "swaptimize",
    Short: "Gestor dinámico de memoria swap para Linux",
    Long:  "Swaptimize permite monitorear y gestionar swap en estaciones de trabajo y laptops de forma eficiente y segura.",
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println("Error:", err)
    }
}
