package scpi_sender

import (
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	pyPath    = `C:\Users\boris\anaconda3\python.exe`
	genScript = filepath.Join("..", "internal", "python_script", "generator.py")
	oscScript = filepath.Join("..", "internal", "python_script", "osci.py")
)

func SendGen(cmd string) string {
	absScript, _ := filepath.Abs(genScript)
	out, _ := exec.Command(pyPath, absScript, cmd).CombinedOutput()
	return strings.TrimSpace(string(out))
}

func SendOsc(cmd string) string {
	absScript, _ := filepath.Abs(oscScript)
	out, _ := exec.Command(pyPath, absScript, cmd).CombinedOutput()
	return strings.TrimSpace(string(out))
}
