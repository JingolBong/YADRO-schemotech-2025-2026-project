package measurement

import "math"

func GenerateSmartSweep(fc float64, points int, minStep float64) []float64 {
	var freqs []float64
	var lastF float64 = -1.0

	for i := 0; i < points; i++ {
		x := -1.0 + 2.0*float64(i)/float64(points-1)
		y := math.Pow(x, 3)
		f := fc * math.Pow(10, y)

		if lastF != -1.0 && (f-lastF) < minStep {
			continue
		}
		freqs = append(freqs, f)
		lastF = f
	}
	return freqs
}

func GetStudentT(n int) float64 {
	df := n - 1
	table := map[int]float64{
		1: 12.706, 2: 4.303, 3: 3.182, 4: 2.776,
		5: 2.571, 6: 2.447, 7: 2.365, 8: 2.306,
		9: 2.262, 10: 2.228,
	}
	if val, ok := table[df]; ok {
		return val
	}
	return 2.0
}

func CalculateStats(data []float64) (float64, float64) {
	n := float64(len(data))
	if n == 0 {
		return 0, 0
	}
	var mean float64
	for _, v := range data {
		mean += v
	}
	mean /= n

	var stdDev float64
	if n > 1 {
		var variance float64
		for _, v := range data {
			variance += math.Pow(v-mean, 2)
		}
		stdDev = math.Sqrt(variance / (n - 1))
	}
	return mean, stdDev
}
