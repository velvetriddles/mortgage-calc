package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/velvetriddles/mortgage-calc/internal/cache"
	"github.com/velvetriddles/mortgage-calc/internal/config"
	"github.com/velvetriddles/mortgage-calc/internal/handler"
	"github.com/velvetriddles/mortgage-calc/internal/middleware"
	"github.com/velvetriddles/mortgage-calc/internal/service"
)

func main() {
	log.Println("Starting mortgage calculator...")

	cfg, err := config.LoadConfig("config.yml")
	if err != nil {
		log.Printf("Error loading configuration: %v, using default values", err)
		cfg = config.New()
	}

	mortCache := cache.NewMortCache()
	mux := http.NewServeMux()

	calculator := service.NewMortCalculator()
	mortHandler := handler.NewMortHandler(mortCache, calculator)

	mux.HandleFunc("/execute", mortHandler.Execute)
	mux.HandleFunc("/cache", mortHandler.GetCache)

	loggerMiddleware := middleware.Logger(mux)

	serverAddr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Server started on port %d", cfg.Port)
	log.Fatal(http.ListenAndServe(serverAddr, loggerMiddleware))
}
