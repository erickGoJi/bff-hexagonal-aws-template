package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppName                 string
	AppEnv                  string
	LogLevel                string
	DownstreamTimeout       time.Duration
	ServiceCatalogBaseURL   string
	ServicePricingBaseURL   string
	ServiceInventoryBaseURL string
	ServiceOrderBaseURL     string
	ServiceShipmentBaseURL  string
}

func Load() (Config, error) {
	timeoutMS := getEnv("DOWNSTREAM_TIMEOUT_MS", "1500")
	timeoutInt, err := strconv.Atoi(timeoutMS)
	if err != nil {
		return Config{}, fmt.Errorf("parse DOWNSTREAM_TIMEOUT_MS: %w", err)
	}

	cfg := Config{
		AppName:                 getEnv("APP_NAME", "hexagonal-bff"),
		AppEnv:                  getEnv("APP_ENV", "local"),
		LogLevel:                getEnv("LOG_LEVEL", "INFO"),
		DownstreamTimeout:       time.Duration(timeoutInt) * time.Millisecond,
		ServiceCatalogBaseURL:   getEnv("SERVICE_CATALOG_BASE_URL", "http://localhost:8081"),
		ServicePricingBaseURL:   getEnv("SERVICE_PRICING_BASE_URL", "http://localhost:8082"),
		ServiceInventoryBaseURL: getEnv("SERVICE_INVENTORY_BASE_URL", "http://localhost:8083"),
		ServiceOrderBaseURL:     getEnv("SERVICE_ORDER_BASE_URL", "http://localhost:8084"),
		ServiceShipmentBaseURL:  getEnv("SERVICE_SHIPMENT_BASE_URL", "http://localhost:8085"),
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback
	}

	return value
}
