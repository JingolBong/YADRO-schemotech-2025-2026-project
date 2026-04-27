package measurement

import (
	"math"
	"testing"
)

func TestCalculateStats(t *testing.T) {
	data := []float64{4.0, 4.2, 3.9, 4.1}
	mean, std := CalculateStats(data)

	if math.Abs(mean-4.05) > 1e-5 {
		t.Errorf("Ожидалось среднее 4.05, получено %f", mean)
	}
	if math.Abs(std-0.129099) > 1e-4 {
		t.Errorf("Ожидалось СКО 0.129, получено %f", std)
	}
}

func TestGenerateSmartSweep(t *testing.T) {
	fc := 1000.0
	points := 21
	minStep := 10.0

	freqs := GenerateSmartSweep(fc, points, minStep)

	if len(freqs) == 0 {
		t.Fatal("Сетка частот пуста")
	}

	firstF := freqs[0]
	lastF := freqs[len(freqs)-1]

	if math.Abs(firstF-100.0) > 1e-1 {
		t.Errorf("Начальная частота должна быть 100 Гц, получено %f", firstF)
	}
	if math.Abs(lastF-10000.0) > 1e-1 {
		t.Errorf("Конечная частота должна быть 10000 Гц, получено %f", lastF)
	}

	for i := 1; i < len(freqs); i++ {
		diff := freqs[i] - freqs[i-1]
		if diff < minStep {
			t.Errorf("Нарушено условие различимости: шаг %f меньше минимального %f на частоте %f", diff, minStep, freqs[i])
		}
	}
}
