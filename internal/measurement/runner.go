package measurement

import (
	"context"
	"fmt"
	"math"

	"github.com/JingolBong/schemotech/internal/config"
	"github.com/JingolBong/schemotech/internal/lab"
)

func timebaseForFreq(f float64) string {
	if f < 20 {
		return "20ms"
	}
	if f < 100 {
		return "5ms"
	}
	if f < 500 {
		return "1ms"
	}
	if f < 2000 {
		return "200us"
	}
	return "50us"
}

func voltageScaleForFreq(f, fc float64, mode string) string {
	ratio := f / fc
	if mode == "highpass" {
		ratio = fc / f
	}

	switch {
	case ratio < 1.0:
		return "1v"
	case ratio < 3.0:
		return "500mv"
	case ratio < 10.0:
		return "200mv"
	default:
		return "100mv"
	}
}

func RunSweep(ctx context.Context, client lab.InstrumentControllerClient, cfg config.Config) ([]Point, error) {
	client.SetupInstruments(ctx, &lab.SetupRequest{
		AmplitudeVpp: float32(cfg.Signal.AmplitudeVPP),
		Waveform:     cfg.Signal.Waveform,
	})

	var results []Point
	fc := 1.0 / (2.0 * math.Pi * cfg.Filter.ROhm * cfg.Filter.CFarad)

	freqs := GenerateSmartSweep(fc, cfg.Sweep.Points, cfg.Stats.MinFStep)

	fmt.Printf("%-10s | %-6s | %-10s | %-10s | %-10s | %-10s\n", "f, Hz", "n", "Vin, V", "Vout, V", "Exp dB", "Phase Deg")

	for _, f := range freqs {
		var voutSamples []float64
		var vinSum, phaseSum float64
		n := 0

		for n < cfg.Stats.MaxSamples {
			resp, err := client.MeasurePoint(ctx, &lab.MeasureRequest{
				FrequencyHz: float32(f),
				Timebase:    timebaseForFreq(f),
				Vscale:      voltageScaleForFreq(f, fc, cfg.Mode),
			})
			if err != nil || resp.ErrorMsg != "" {
				continue
			}

			voutSamples = append(voutSamples, float64(resp.VoutVpp))
			vinSum += float64(resp.VinVpp)
			phaseSum += float64(resp.PhaseShiftDeg)
			n++

			if n >= 3 {
				_, s := CalculateStats(voutSamples)
				t := GetStudentT(n)

				nMin := math.Pow((t*s)/cfg.Stats.DeltaXMax, 2)

				if float64(n) >= nMin {
					break
				}
			}
		}

		if n == 0 {
			continue
		}

		voutMean, _ := CalculateStats(voutSamples)
		vinMean := vinSum / float64(n)
		phaseMean := phaseSum / float64(n)

		gainDb := -100.0
		if vinMean > 0.001 && voutMean > 0.0001 {
			gainDb = 20.0 * math.Log10(voutMean/vinMean)
		}

		theoDb, theoPhase := 0.0, 0.0
		switch cfg.Mode {
		case "lowpass":
			theoDb = TheoryLPF(f, cfg.Filter.ROhm, cfg.Filter.CFarad)
			theoPhase = TheoryPhaseLPF(f, cfg.Filter.ROhm, cfg.Filter.CFarad)
		case "highpass":
			theoDb = TheoryHPF(f, cfg.Filter.ROhm, cfg.Filter.CFarad)
			theoPhase = TheoryPhaseHPF(f, cfg.Filter.ROhm, cfg.Filter.CFarad)
		case "bandpass":
			theoDb = TheoryBPF(f, cfg.Filter.ROhm, cfg.Filter.CFarad)
			theoPhase = TheoryPhaseBPF(f, cfg.Filter.ROhm, cfg.Filter.CFarad)
		case "wien":
			theoDb = TheoryWien(f, cfg.Filter.ROhm, cfg.Filter.CFarad)
			theoPhase = TheoryPhaseWien(f, cfg.Filter.ROhm, cfg.Filter.CFarad)
		}

		fmt.Printf("%-10.1f | %-6d | %-10.3f | %-10.3f | %-10.2f | %-10.2f\n", f, n, vinMean, voutMean, gainDb, phaseMean)

		results = append(results, Point{
			FreqHz:    f,
			VinVpp:    vinMean,
			VoutVpp:   voutMean,
			GainDB:    gainDb,
			TheoGain:  theoDb,
			PhaseDeg:  phaseMean,
			TheoPhase: theoPhase,
		})
	}

	client.Shutdown(ctx, &lab.ShutdownRequest{})
	return results, nil
}
