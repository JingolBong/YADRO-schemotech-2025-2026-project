package scpi_sender

import (
	"fmt"
	"os/exec"
	"strings"
)

func SendSCPIGen(command string) string {
	out, err := exec.Command("python", `.\internal\python_script\generator.py`, command).Output()
	result := strings.TrimSpace(string(out))
	if err != nil && result == "" {
		return fmt.Sprintf("ERROR: %v", err)
	}
	return result
}

func SendSCPIOwonOsci(command string) string {
	out, err := exec.Command("python", `.\internal\python_script\osci.py`, command).Output()
	result := strings.TrimSpace(string(out))
	if err != nil && result == "" {
		return fmt.Sprintf("ERROR: %v", err)
	}
	return result
}
