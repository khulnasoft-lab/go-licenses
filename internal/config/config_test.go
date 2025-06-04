package config

import (
	"github.com/spf13/viper"
	"os"
	"testing"
)

func TestLoadConfigFromFile_FileNotFound(t *testing.T) {
	_, err := LoadConfigFromFile(viper.New(), "nonexistent.yaml")
	if err == nil {
		t.Error("expected error for missing config file, got nil")
	}
}

func TestLoadConfigFromFile_InvalidFormat(t *testing.T) {
	f, err := os.CreateTemp("", "bad_config_*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	_, err = LoadConfigFromFile(viper.New(), f.Name())
	if err == nil {
		t.Error("expected error for invalid config format, got nil")
	}
}
