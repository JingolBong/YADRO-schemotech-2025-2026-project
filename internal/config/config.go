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
		OffsetV      float64 `mapstructure:"offset_v"`
		Waveform     string  `mapstructure:"waveform"`
	} `mapstructure:"signal"`
	Sweep struct {
		StartHz float64 `mapstructure:"start_hz"`
		StopHz  float64 `mapstructure:"stop_hz"`
		Points  int     `mapstructure:"points"`
		Scale   string  `mapstructure:"scale"`
	} `mapstructure:"sweep"`
	Scope struct {
		ChIn     string `mapstructure:"ch_in"`
		ChOut    string `mapstructure:"ch_out"`
		Probe    string `mapstructure:"probe"`
		SettleMs int    `mapstructure:"settle_ms"`
	} `mapstructure:"scope"`
	Output struct {
		CSV       string `mapstructure:"csv"`
		GainPlot  string `mapstructure:"gain_plot"`
		VoutPlot  string `mapstructure:"vout_plot"`
		PhasePlot string `mapstructure:"phase_plot"`
	} `mapstructure:"output"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("ошибка чтения файла %s: %v", path, err)
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
