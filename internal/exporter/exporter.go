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

package exporter

import (
	"context"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/node-problem-detector/pkg/systemlogmonitor/logwatchers/kmsg"
	"k8s.io/node-problem-detector/pkg/systemlogmonitor/logwatchers/types"
)

const defaultPattern = `^oom-kill.+,task_memcg=\/kubepods(?:\.slice)?\/.+\/(?:kubepods-burstable-)?pod(\w+[-_]\w+[-_]\w+[-_]\w+[-_]\w+)(?:\.slice)?\/(?:cri-containerd-)?([a-f0-9]+)`

var prometheusContainerLabels = map[string]string{
	"io.kubernetes.container.name": "container_name",
	"io.kubernetes.pod.namespace":  "namespace",
	"io.kubernetes.pod.uid":        "pod_uid",
	"io.kubernetes.pod.name":       "pod_name",
}

type Config struct {
	ListenAddress       string
	ContainerdSocket    string
	RegexpPattern       string
	ContainerdNamespace string
}

type Exporter struct {
	config               Config
	containerdClient     *containerd.Client
	kubernetesCounterVec *prometheus.CounterVec
	kmesgRE              *regexp.Regexp
	logger               *slog.Logger
}

func compilePattern(pattern string) (*regexp.Regexp, error) {
	if pattern == "" {
		pattern = defaultPattern
	}
	return regexp.Compile(pattern)
}

func New(cfg Config, logger *slog.Logger) (*Exporter, error) {
	kmesgRE, err := compilePattern(cfg.RegexpPattern)
	if err != nil {
		return nil, err
	}

	containerdClient, err := containerd.New(cfg.ContainerdSocket)
	if err != nil {
		return nil, err
	}

	var labels []string
	for _, label := range prometheusContainerLabels {
		labels = append(labels, strings.ReplaceAll(label, ".", "_"))
	}

	kubernetesCounterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "klog_pod_oomkill",
		Help: "Extract metrics for OOMKilled pods from kernel log",
	}, labels)

	prometheus.MustRegister(kubernetesCounterVec)

	return &Exporter{
		config:               cfg,
		containerdClient:     containerdClient,
		kubernetesCounterVec: kubernetesCounterVec,
		kmesgRE:              kmesgRE,
		logger:               logger,
	}, nil
}

func (e *Exporter) Run(ctx context.Context) error {
	defer e.containerdClient.Close()

	go e.startMetricsServer()

	kmsgWatcher := kmsg.NewKmsgWatcher(types.WatcherConfig{Plugin: "kmsg"})
	logCh, err := kmsgWatcher.Watch()
	if err != nil {
		return err
	}

	e.logger.Info("Started watching kernel log for OOM kills")

	for {
		select {
		case <-ctx.Done():
			e.logger.Info("Shutting down exporter")
			return ctx.Err()
		case log, ok := <-logCh:
			if !ok {
				e.logger.Warn("Log channel closed")
				return nil
			}
			e.processLogMessage(log.Message)
		}
	}
}

func (e *Exporter) startMetricsServer() {
	e.logger.Info("Starting prometheus metrics", "address", e.config.ListenAddress)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	server := &http.Server{
		Addr:              e.config.ListenAddress,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           mux,
	}

	if err := server.ListenAndServe(); err != nil {
		e.logger.Error("Metrics server failed", "error", err)
	}
}

func (e *Exporter) processLogMessage(message string) {
	podUID, containerID := e.getContainerIDFromLog(message)
	if containerID == "" {
		return
	}

	labels, err := e.getContainerLabels(containerID)
	if err != nil || labels == nil {
		e.logger.Warn("Could not get labels for container",
			"container_id", containerID,
			"pod_uid", podUID,
			"error", err)
		return
	}

	e.incrementPrometheusCounter(labels)
}

func (e *Exporter) getContainerIDFromLog(log string) (podUID, containerID string) {
	matches := e.kmesgRE.FindStringSubmatch(log)
	if matches == nil {
		return "", ""
	}
	return matches[1], matches[2]
}

func (e *Exporter) getContainerLabels(containerID string) (map[string]string, error) {
	ctx := namespaces.WithNamespace(context.Background(), e.config.ContainerdNamespace)
	container, err := e.containerdClient.ContainerService().Get(ctx, containerID)
	if err != nil {
		return nil, err
	}
	return container.Labels, nil
}

func (e *Exporter) incrementPrometheusCounter(containerLabels map[string]string) {
	labels := make(map[string]string)
	for key, label := range prometheusContainerLabels {
		labels[label] = containerLabels[key]
	}

	e.logger.Debug("Recording OOM kill", "labels", labels)

	counter, err := e.kubernetesCounterVec.GetMetricWith(labels)
	if err != nil {
		e.logger.Warn("Failed to get metric", "error", err)
		return
	}

	counter.Add(1)
}
