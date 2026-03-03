// Copyright 2024 SAP SE
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joluc/oomkill-exporter/internal/exporter"
	"github.com/sapcc/go-api-declarations/bininfo"
)

func main() {
	var (
		listenAddress       string
		containerdSocket    string
		containerdNamespace string
		regexpPattern       string
		versionFlag         bool
		logLevel            string
	)

	flag.StringVar(&listenAddress, "listen-address", ":9102", "The address to listen on for HTTP requests")
	flag.StringVar(&containerdSocket, "containerd-socket", "/run/containerd/containerd.sock", "Path to containerd socket")
	flag.StringVar(&containerdNamespace, "containerd-namespace", "k8s.io", "Containerd namespace to use")
	flag.StringVar(&regexpPattern, "regexp-pattern", "", "Custom regexp pattern to match and extract Pod UID and Container ID")
	flag.BoolVar(&versionFlag, "version", false, "Print version info")
	flag.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.Parse()

	if versionFlag {
		fmt.Printf("Version: %s\n", bininfo.Version())
		os.Exit(0)
	}

	// Setup structured logging
	var level slog.Level
	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)

	// Create exporter configuration
	cfg := exporter.Config{
		ListenAddress:       listenAddress,
		ContainerdSocket:    containerdSocket,
		ContainerdNamespace: containerdNamespace,
		RegexpPattern:       regexpPattern,
	}

	// Initialize exporter
	exp, err := exporter.New(cfg, logger)
	if err != nil {
		logger.Error("Failed to create exporter", "error", err)
		os.Exit(1)
	}

	// Setup signal handling
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("Starting OOM Kill Exporter",
		"version", bininfo.Version(),
		"listen_address", listenAddress,
		"containerd_socket", containerdSocket,
		"containerd_namespace", containerdNamespace,
	)

	// Run exporter
	if err := exp.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("Exporter failed", "error", err)
		os.Exit(1)
	}

	logger.Info("Exporter stopped gracefully")
}
