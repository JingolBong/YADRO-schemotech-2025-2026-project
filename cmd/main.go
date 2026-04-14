package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	scpi_sender "github.com/JingolBong/scheme/internal/scpi" // Поправь импорт под свой проект
)

// Парсер вольт (поддерживает научный формат, например 3.14e-02)
func parseVpp(resp string) float64 {
	if resp == "" || strings.Contains(resp, "ERROR") || strings.Contains(resp, "NAN") {
		return 0.0
	}
	re := regexp.MustCompile(`([-+]?[0-9]*\.?[0-9]+(?:[eE][-+]?[0-9]+)?)`)
	m := re.FindStringSubmatch(resp)
	if len(m) == 0 {
		return 0.0
	}
	val, _ := strconv.ParseFloat(m[1], 64)

	if strings.Contains(resp, "mV") {
		val /= 1000.0
	} else if strings.Contains(resp, "uV") {
		val /= 1e6
	}
	return val
}

// Динамические масштабы
func timebaseForFreq(f float64) string {
	switch {
	case f < 20:
		return "20ms"
	case f < 50:
		return "10ms"
	case f < 100:
		return "5ms"
	case f < 500:
		return "1ms"
	case f < 2000:
		return "200us"
	case f < 5000:
		return "100us"
	default:
		return "20us"
	}
}

func voltageScaleForFreq(f, fc float64) string {
	ratio := f / fc
	switch {
	case ratio < 0.5:
		return "1v"
	case ratio < 1.5:
		return "500mv"
	case ratio < 5.0:
		return "100mv"
	case ratio < 15.0:
		return "50mv"
	default:
		return "20mv"
	}
}

func logSpace(start, stop float64, num int) []float64 {
	res := make([]float64, num)
	step := (math.Log10(stop) - math.Log10(start)) / float64(num-1)
	for i := 0; i < num; i++ {
		res[i] = math.Pow(10, math.Log10(start)+float64(i)*step)
	}
	return res
}

func RunLowPassFilterTest(R, C float64) {
	fc := 1.0 / (2.0 * math.Pi * R * C)
	fmt.Printf("=====================================================\n")
	fmt.Printf("🔧 ЗАПУСК ТЕСТА ФНЧ (R = %.0f Ом, C = %.2e Ф)\n", R, C)
	fmt.Printf("📊 Расчетная частота среза fc ≈ %.2f Гц\n", fc)
	fmt.Printf("=====================================================\n")

	scpi_sender.SendSCPIGen(":CHANnel1:BASE:WAVe SINe")
	scpi_sender.SendSCPIGen(":CHANnel1:BASE:AMPLitude 4")
	scpi_sender.SendSCPIGen(":CHANnel1:BASE:OFFSet 0")
	scpi_sender.SendSCPIGen(":CHANnel1:OUTPut ON")

	scpi_sender.SendSCPIOwonOsci(":CH1:DISPlay ON")
	scpi_sender.SendSCPIOwonOsci(":CH2:DISPlay ON")
	scpi_sender.SendSCPIOwonOsci(":CH1:SCALe 1v")

	freqs := logSpace(10, 10000, 30)
	var fData, vinData, voutData, gainData []string

	fmt.Printf("%-10s | %-10s | %-10s | %-10s\n", "f, Гц", "Vin, В", "Vout, В", "Gain, dB")
	fmt.Println("-----------------------------------------------------")

	for _, f := range freqs {
		scpi_sender.SendSCPIGen(fmt.Sprintf(":CHANnel1:BASE:FREQuency %d", int(f)))
		scpi_sender.SendSCPIOwonOsci(fmt.Sprintf(":HORIzontal:SCALe %s", timebaseForFreq(f)))
		scpi_sender.SendSCPIOwonOsci(fmt.Sprintf(":CH2:SCALe %s", voltageScaleForFreq(f, fc)))

		// Ждем перестройки осциллографа
		time.Sleep(3 * time.Second)

		vinRaw := scpi_sender.SendSCPIOwonOsci(":MEASUrement:CH1:PKPK?")
		voutRaw := scpi_sender.SendSCPIOwonOsci(":MEASUrement:CH2:PKPK?")

		vin := parseVpp(vinRaw)
		vout := parseVpp(voutRaw)

		// Защита от просадок входа на осциллографе
		if vin < 0.1 {
			vin = 4.0
		}

		gainDb := -100.0
		if vin > 0.001 && vout > 0.0001 {
			gainDb = 20.0 * math.Log10(vout/vin)
		}

		fmt.Printf("%-10.1f | %-10.3f | %-10.3f | %-10.2f\n", f, vin, vout, gainDb)

		fData = append(fData, fmt.Sprintf("%.2f", f))
		vinData = append(vinData, fmt.Sprintf("%.3f", vin))
		voutData = append(voutData, fmt.Sprintf("%.3f", vout))
		gainData = append(gainData, fmt.Sprintf("%.2f", gainDb))
	}

	scpi_sender.SendSCPIGen(":CHANnel1:OUTPut OFF")

	// Генерация скрипта для графиков
	plotScript := fmt.Sprintf(`
import matplotlib.pyplot as plt
import numpy as np

f_data = [%s]
vin_data = [%s]
vout_data = [%s]
gain_data = [%s]
fc = %.2f

plt.figure(figsize=(14, 6))

plt.subplot(1, 2, 1)
plt.plot(f_data, vout_data, 'bo-', linewidth=2, label='Vout (Эксперимент)')
plt.axvline(x=fc, color='r', linestyle='--', label=f'Срез = {fc:.1f} Гц')
plt.xscale('log')
plt.xlabel('Частота (Гц)')
plt.ylabel('Напряжение (В)')
plt.title('Выходное напряжение Vout(f)')
plt.grid(True, which="both", ls="--", alpha=0.5)
plt.legend()

plt.subplot(1, 2, 2)
plt.plot(f_data, gain_data, 'go-', linewidth=2, label='Эксперимент (АЧХ)')
f_theory = np.logspace(1, 4, 100)
gain_theory = 20 * np.log10(1 / np.sqrt(1 + (f_theory/fc)**2))
plt.plot(f_theory, gain_theory, 'k--', label='Теория')
plt.axvline(x=fc, color='r', linestyle='--')
plt.xscale('log')
plt.xlabel('Частота (Гц)')
plt.ylabel('Коэффициент передачи (дБ)')
plt.title('АЧХ (Диаграмма Боде)')
plt.grid(True, which="both", ls="--", alpha=0.5)
plt.legend()

plt.tight_layout()
plt.savefig('rc_filter_plot.png', dpi=300)
plt.show()
`, strings.Join(fData, ","), strings.Join(vinData, ","), strings.Join(voutData, ","), strings.Join(gainData, ","), fc)

	os.WriteFile("plot_result.py", []byte(plotScript), 0644)
	fmt.Println("✅ Измерения завершены. Открываю графики...")

	// Запускаем отрисовку графиков асинхронно или синхронно
	cmd := exec.Command("python", "plot_result.py")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func main() {
	RunLowPassFilterTest(6800.0, 22e-9) // Твои параметры: 6.8 кОм и 22 нФ
}
