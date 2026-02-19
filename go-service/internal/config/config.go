package config

import (
	"flag"
	"os"
	"strconv"
)

// Config holds all service configuration
type Config struct {
	// Server ports
	HTTPPort  int
	GRPCPort  int
	PProfPort int

	// Observability
	EnableDebugLog bool
	EnablePProf    bool
}

// Load reads configuration from flags and environment variables
// Flags take precedence over environment variables
func Load() *Config {
	cfg := &Config{}

	// Define flags
	flag.IntVar(&cfg.HTTPPort, "http-port", getEnvInt("HTTP_PORT", 8080), "HTTP server port")
	flag.IntVar(&cfg.GRPCPort, "grpc-port", getEnvInt("GRPC_PORT", 9090), "gRPC server port")
	flag.IntVar(&cfg.PProfPort, "pprof-port", getEnvInt("PPROF_PORT", 6060), "pprof server port")
	flag.BoolVar(&cfg.EnableDebugLog, "debug", getEnvBool("DEBUG", false), "Enable debug logging")
	flag.BoolVar(&cfg.EnablePProf, "enable-pprof", getEnvBool("ENABLE_PPROF", true), "Enable pprof endpoint")

	flag.Parse()

	return cfg
}

// getEnvInt retrieves an integer from environment with a default fallback
func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

// getEnvBool retrieves a boolean from environment with a default fallback
func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			return boolVal
		}
	}
	return defaultVal
}
