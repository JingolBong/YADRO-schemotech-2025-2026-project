package plotting

import (
	"image/color"
	"os"

	"github.com/JingolBong/schemotech/internal/measurement"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func SaveBodePlot(points []measurement.Point, path string, title string) error {
	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = "Frequency (Hz)"
	p.Y.Label.Text = "Gain (dB)"
	p.X.Scale = plot.LogScale{}
	p.X.Tick.Marker = plot.LogTicks{}

	p.Add(plotter.NewGrid())

	expXY := make(plotter.XYs, len(points))
	theoXY := make(plotter.XYs, len(points))

	for i, pt := range points {
		expXY[i].X = pt.FreqHz
		expXY[i].Y = pt.GainDB
		theoXY[i].X = pt.FreqHz
		theoXY[i].Y = pt.TheoGain
	}

	expLine, err := plotter.NewLine(expXY)
	if err != nil {
		return err
	}
	expLine.Color = color.RGBA{B: 255, A: 255}
	expLine.Width = vg.Points(2)

	theoLine, err := plotter.NewLine(theoXY)
	if err != nil {
		return err
	}
	theoLine.Color = color.RGBA{R: 255, A: 255}
	theoLine.Dashes = []vg.Length{vg.Points(5), vg.Points(5)}

	p.Add(expLine, theoLine)
	p.Legend.Add("Experiment", expLine)
	p.Legend.Add("Theory", theoLine)

	if err := os.MkdirAll("output", 0o755); err != nil {
		return err
	}
	return p.Save(10*vg.Inch, 5*vg.Inch, path)
}

func SavePhasePlot(points []measurement.Point, path string, title string) error {
	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = "Frequency (Hz)"
	p.Y.Label.Text = "Phase (Degrees)"
	p.X.Scale = plot.LogScale{}
	p.X.Tick.Marker = plot.LogTicks{}

	p.Add(plotter.NewGrid())

	expXY := make(plotter.XYs, len(points))
	theoXY := make(plotter.XYs, len(points))

	for i, pt := range points {
		expXY[i].X = pt.FreqHz
		expXY[i].Y = pt.PhaseDeg
		theoXY[i].X = pt.FreqHz
		theoXY[i].Y = pt.TheoPhase
	}

	expLine, err := plotter.NewLine(expXY)
	if err != nil {
		return err
	}
	expLine.Color = color.RGBA{B: 255, A: 255}
	expLine.Width = vg.Points(2)

	theoLine, err := plotter.NewLine(theoXY)
	if err != nil {
		return err
	}
	theoLine.Color = color.RGBA{R: 255, A: 255}
	theoLine.Dashes = []vg.Length{vg.Points(5), vg.Points(5)}

	p.Add(expLine, theoLine)
	p.Legend.Add("Experiment", expLine)
	p.Legend.Add("Theory", theoLine)

	if err := os.MkdirAll("output", 0o755); err != nil {
		return err
	}
	return p.Save(10*vg.Inch, 5*vg.Inch, path)
}
