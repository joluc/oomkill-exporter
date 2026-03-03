# oomkill-exporter

Prometheus exporter that monitors OOM kills in Kubernetes by parsing kernel logs.

> Fork of [sapcc/kubernetes-oomkill-exporter](https://github.com/sapcc/kubernetes-oomkill-exporter)

## Overview

Watches `/dev/kmsg` for OOM kill events, retrieves pod metadata from containerd, and exposes metrics for Prometheus.

## Requirements

- Linux with `/dev/kmsg` access
- containerd runtime
- Kubernetes cluster

## Installation

### Helm Chart (Recommended)

```bash
helm install oomkill-exporter ./chart/oomkill-exporter
```

### Build from Source

```bash
make build
```

## Configuration

```bash
# Default settings
./oomkill-exporter

# Debug logging
./oomkill-exporter --log-level debug

# Custom listen address
./oomkill-exporter --listen-address :8080
```

### Flags

- `--listen-address` - Metrics HTTP endpoint (default: `:9102`)
- `--containerd-socket` - Containerd socket path (default: `/run/containerd/containerd.sock`)
- `--containerd-namespace` - Containerd namespace (default: `k8s.io`)
- `--regexp-pattern` - Custom regex for Pod UID and Container ID extraction
- `--log-level` - Log level: debug, info, warn, error (default: `info`)
- `--version` - Print version

## Metrics

**Endpoint:** `/metrics`

**Metric:** `klog_pod_oomkill` (counter)

**Labels:**
- `container_name`
- `namespace`
- `pod_name`
- `pod_uid`

### Example Prometheus Alert

```promql
sum by(namespace, pod_name) (changes(klog_pod_oomkill[30m])) > 2
```

## Development

```bash
make build          # Build binary
make test           # Run tests
make test-coverage  # Tests with coverage
make lint           # Run linter
make docker-build   # Build Docker image
make clean          # Clean artifacts
```

## License

Apache License 2.0
