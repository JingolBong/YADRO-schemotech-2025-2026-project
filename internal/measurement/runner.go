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
	case ratio < 0.5:
		return "1v"
	case ratio < 1.5:
		return "500mv"
	case ratio < 5.0:
		return "100mv"
	default:
		return "20mv"
	}
}

func RunSweep(ctx context.Context, client lab.InstrumentControllerClient, cfg config.Config) ([]Point, error) {
	client.SetupInstruments(ctx, &lab.SetupRequest{
		AmplitudeVpp: float32(cfg.Signal.AmplitudeVPP),
		Waveform:     cfg.Signal.Waveform,
	})

	var results []Point
	fc := 1.0 / (2.0 * math.Pi * cfg.Filter.ROhm * cfg.Filter.CFarad)
	step := (math.Log10(cfg.Sweep.StopHz) - math.Log10(cfg.Sweep.StartHz)) / float64(cfg.Sweep.Points-1)

	fmt.Printf("%-10s | %-10s | %-10s | %-10s | %-10s | %-10s\n", "f, Hz", "Vin, V", "Vout, V", "Exp dB", "Theo dB", "Phase Deg")
	fmt.Println("-------------------------------------------------------------------------")

	for i := 0; i < cfg.Sweep.Points; i++ {
		f := math.Pow(10, math.Log10(cfg.Sweep.StartHz)+float64(i)*step)

		resp, err := client.MeasurePoint(ctx, &lab.MeasureRequest{
			FrequencyHz: float32(f),
			Timebase:    timebaseForFreq(f),
			Vscale:      voltageScaleForFreq(f, fc, cfg.Mode),
		})

		if err != nil || resp.ErrorMsg != "" {
			continue
		}

		vin := float64(resp.VinVpp)
		vout := float64(resp.VoutVpp)

		gainDb := -100.0
		if vin > 0.001 && vout > 0.0001 {
			gainDb = 20.0 * math.Log10(vout/vin)
		}

		theoDb := 0.0
		theoPhase := 0.0

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

		fmt.Printf("%-10.1f | %-10.3f | %-10.3f | %-10.2f | %-10.2f | %-10.2f\n", f, vin, vout, gainDb, theoDb, resp.PhaseShiftDeg)

		results = append(results, Point{
			FreqHz:    f,
			VinVpp:    vin,
			VoutVpp:   vout,
			GainDB:    gainDb,
			TheoGain:  theoDb,
			PhaseDeg:  float64(resp.PhaseShiftDeg),
			TheoPhase: theoPhase,
		})
	}

	client.Shutdown(ctx, &lab.ShutdownRequest{})
	return results, nil
}
