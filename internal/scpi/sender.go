package scpi_sender

import (
	"fmt"
	"os/exec"
	"strings"
)

func SendSCPIGen(command string) string {
	out, err := exec.Command("python", ".\\internal\\python_script\\generator.py", command).CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Gen Error (%s): %v | %s\n", command, err, string(out))
	}
	return strings.TrimSpace(string(out))
}

func SendSCPIOwonOsci(command string) string {
	out, err := exec.Command("python", ".\\internal\\python_script\\osci.py", command).CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Osc Error (%s): %v | %s\n", command, err, string(out))
	}
	return strings.TrimSpace(string(out))
}
