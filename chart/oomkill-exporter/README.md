# oomkill-exporter Helm Chart

Helm chart for deploying oomkill-exporter as a DaemonSet in Kubernetes.

## Installation

```bash
helm install oomkill-exporter ./chart/oomkill-exporter
```

### With custom values

```bash
helm install oomkill-exporter ./chart/oomkill-exporter \
  --set image.tag=latest \
  --set config.logLevel=debug
```

## Configuration

See [values.yaml](values.yaml) for all available configuration options.

### Key Configuration Options

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.repository` | Image repository | `ghcr.io/joluc/oomkill-exporter` |
| `image.tag` | Image tag | Chart appVersion |
| `config.listenAddress` | Metrics listen address | `:9102` |
| `config.logLevel` | Log level (debug, info, warn, error) | `info` |
| `config.containerdSocket` | Path to containerd socket | `/run/containerd/containerd.sock` |
| `serviceMonitor.enabled` | Create ServiceMonitor for Prometheus Operator | `false` |
| `resources.limits.cpu` | CPU limit | `100m` |
| `resources.limits.memory` | Memory limit | `100Mi` |

## ServiceMonitor

To enable Prometheus Operator integration:

```bash
helm install oomkill-exporter ./chart/oomkill-exporter \
  --set serviceMonitor.enabled=true
```

## Uninstall

```bash
helm uninstall oomkill-exporter
```
