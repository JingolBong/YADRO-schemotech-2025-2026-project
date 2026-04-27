package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Mode   string `mapstructure:"mode"`
	Filter struct {
		ROhm   float64 `mapstructure:"r_ohm"`
		CFarad float64 `mapstructure:"c_farad"`
	} `mapstructure:"filter"`
	Signal struct {
		AmplitudeVPP float64 `mapstructure:"amplitude_vpp"`
		Waveform     string  `mapstructure:"waveform"`
	} `mapstructure:"signal"`
	Sweep struct {
		Points int `mapstructure:"points"`
	} `mapstructure:"sweep"`
	Stats struct {
		DeltaXMax  float64 `mapstructure:"delta_x_max"`
		MaxSamples int     `mapstructure:"max_samples"`
		MinFStep   float64 `mapstructure:"min_f_step"`
	} `mapstructure:"stats"`
	Output struct {
		GainPlot  string `mapstructure:"gain_plot"`
		PhasePlot string `mapstructure:"phase_plot"`
	} `mapstructure:"output"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if cfg.Mode != "lowpass" && cfg.Mode != "highpass" && cfg.Mode != "bandpass" && cfg.Mode != "wien" {
		return nil, fmt.Errorf("unsupported mode: %s", cfg.Mode)
	}

	return &cfg, nil
}
