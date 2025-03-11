package metric_functions

import (
    "fmt"
    "os/exec"
    "strings"
    "runtime"
)

func GetMotherboardID() (string, error) {
    switch runtime.GOOS {
    case "linux":
        out, err := exec.Command("sudo", "dmidecode", "-s", "baseboard-serial-number").Output()
        if err != nil {
            return "", fmt.Errorf("error obteniendo motherboard ID: %w", err)
        }
        id := strings.TrimSpace(string(out))
        if id == "" {
            return "", fmt.Errorf("no se encontró motherboard ID")
        }
        return id, nil
    default:
        return "", fmt.Errorf("sistema operativo no soportado: %s", runtime.GOOS)
    }
}