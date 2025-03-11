package metric_functions

import (
    "fmt"
    "os/exec"
    "strings"
    "runtime"
    "io/ioutil"
)

func GetMotherboardID() (string, error) {
    switch runtime.GOOS {
    case "linux":
        out, err := exec.Command("sudo", "dmidecode", "-s", "baseboard-serial-number").Output()
        if err == nil {
            id := strings.TrimSpace(string(out))
            if id != "" && id != "Not Specified" && id != "Unknown" {
                return id, nil
            }
        }
        
        out, err = exec.Command("sudo", "dmidecode", "-s", "system-uuid").Output()
        if err == nil {
            id := strings.TrimSpace(string(out))
            if id != "" && id != "Not Specified" && id != "Unknown" {
                return id, nil
            }
        }
        
        data, err := ioutil.ReadFile("/sys/class/dmi/id/board_serial")
        if err == nil {
            id := strings.TrimSpace(string(data))
            if id != "" && id != "Not Specified" && id != "Unknown" {
                return id, nil
            }
        }
        
        data, err = ioutil.ReadFile("/sys/class/dmi/id/product_uuid")
        if err == nil {
            id := strings.TrimSpace(string(data))
            if id != "" {
                return id, nil
            }
        }
        
        return "", fmt.Errorf("no se encontró motherboard ID después de varios intentos")
    default:
        return "", fmt.Errorf("sistema operativo no soportado: %s", runtime.GOOS)
    }
}