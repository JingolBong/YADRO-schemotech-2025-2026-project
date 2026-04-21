package measurement

import "math"

type Point struct {
	FreqHz    float64
	VinVpp    float64
	VoutVpp   float64
	GainDB    float64
	TheoGain  float64
	PhaseDeg  float64
	TheoPhase float64
}

func TheoryLPF(f, R, C float64) float64 {
	fc := 1.0 / (2.0 * math.Pi * R * C)
	return 20 * math.Log10(1/math.Sqrt(1+math.Pow(f/fc, 2)))
}

func TheoryHPF(f, R, C float64) float64 {
	fc := 1.0 / (2.0 * math.Pi * R * C)
	return 20 * math.Log10((f/fc)/math.Sqrt(1+math.Pow(f/fc, 2)))
}

func TheoryBPF(f, R, C float64) float64 {
	w := 2 * math.Pi * f
	rc := R * C
	denominator := math.Sqrt(9 + math.Pow(w*rc-1/(w*rc), 2))
	return 20 * math.Log10(1/denominator)
}

func TheoryWien(f, R, C float64) float64 {
	w := 2 * math.Pi * f
	rc := R * C
	realPart := 1 - math.Pow(w*rc, 2)
	imagPart := 3 * w * rc
	magnitude := (w * rc) / math.Sqrt(realPart*realPart+imagPart*imagPart)
	return 20 * math.Log10(magnitude)
}

func TheoryPhaseLPF(f, R, C float64) float64 {
	fc := 1.0 / (2.0 * math.Pi * R * C)
	return -math.Atan(f/fc) * (180.0 / math.Pi)
}

func TheoryPhaseHPF(f, R, C float64) float64 {
	fc := 1.0 / (2.0 * math.Pi * R * C)
	return 90.0 - (math.Atan(f/fc) * (180.0 / math.Pi))
}

func TheoryPhaseBPF(f, R, C float64) float64 {
	w := 2 * math.Pi * f
	rc := R * C
	realPart := 1 - math.Pow(w*rc, 2)
	imagPart := 3 * w * rc
	return math.Atan2(realPart, imagPart) * (180.0 / math.Pi)
}

func TheoryPhaseWien(f, R, C float64) float64 {
	return TheoryPhaseBPF(f, R, C)
}
