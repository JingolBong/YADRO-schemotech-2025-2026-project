package main

import (
	"context"
	"flag"
	"log"

	"github.com/JingolBong/schemotech/internal/config"
	"github.com/JingolBong/schemotech/internal/lab"
	"github.com/JingolBong/schemotech/internal/measurement"
	"github.com/JingolBong/schemotech/internal/plotting"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfgPath := flag.String("cfg", "configs/lpf_rc.yaml", "path to config file")
	flag.Parse()

	log.Printf("📂 Загружаем конфиг из: %s", *cfgPath)

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("❌ Ошибка загрузки конфига: %v", err)
	}

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ Ошибка подключения: %v", err)
	}
	defer conn.Close()

	client := lab.NewInstrumentControllerClient(conn)

	log.Printf("🚀 Запуск тестирования: %s", cfg.Mode)
	points, err := measurement.RunSweep(context.Background(), client, *cfg)
	if err != nil {
		log.Fatalf("❌ Ошибка измерений: %v", err)
	}

	err = plotting.SaveBodePlot(points, cfg.Output.GainPlot, "Bode Plot (Gain): "+cfg.Mode)
	if err != nil {
		log.Fatalf("❌ Ошибка сохранения АЧХ: %v", err)
	}
	log.Printf("✅ График АЧХ сохранен в %s", cfg.Output.GainPlot)

	err = plotting.SavePhasePlot(points, cfg.Output.PhasePlot, "Bode Plot (Phase): "+cfg.Mode)
	if err != nil {
		log.Fatalf("❌ Ошибка сохранения ФЧХ: %v", err)
	}
	log.Printf("✅ График ФЧХ сохранен в %s", cfg.Output.PhasePlot)
}
