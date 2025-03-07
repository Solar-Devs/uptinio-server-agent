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
        // dmidecode to get motherboard serial number in Linux
        cmd := exec.Command("sudo", "dmidecode", "-s", "baseboard-serial-number")
        output, err := cmd.Output()
        if err != nil {
            return "", err
        }
        return strings.TrimSpace(string(output)), nil
        
    case "windows":
        // wmic to get the motherboard serial number in Windows
        cmd := exec.Command("wmic", "baseboard", "get", "serialnumber")
        output, err := cmd.Output()
        if err != nil {
            return "", err
        }
        lines := strings.Split(string(output), "\n")
        for _, line := range lines {
            if line != "" && !strings.Contains(line, "SerialNumber") {
                return strings.TrimSpace(line), nil
            }
        }
        return "", fmt.Errorf("motherboard ID not found")
        
    default:
        return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
    }
}