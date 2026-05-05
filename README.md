# prom-trim

> Utility to prune stale Prometheus recording rules based on query usage metrics

---

## Installation

```bash
go install github.com/yourorg/prom-trim@latest
```

Or build from source:

```bash
git clone https://github.com/yourorg/prom-trim.git && cd prom-trim && go build -o prom-trim .
```

---

## Usage

Point `prom-trim` at your Prometheus instance and a rules file. It will identify recording rules that have not been queried within the specified window and output a pruned rules file.

```bash
prom-trim \
  --prometheus http://localhost:9090 \
  --rules /etc/prometheus/rules.yml \
  --stale-window 30d \
  --output pruned-rules.yml
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--prometheus` | `http://localhost:9090` | Prometheus base URL |
| `--rules` | `rules.yml` | Path to input recording rules file |
| `--stale-window` | `30d` | Age threshold for considering a rule stale |
| `--output` | `stdout` | Path for pruned rules output |
| `--dry-run` | `false` | Print stale rules without writing output |

### Example — dry run

```bash
prom-trim --prometheus http://prom.internal:9090 --rules rules.yml --stale-window 14d --dry-run
```

```
[stale] job:http_requests:rate5m
[stale] instance:node_cpu:avg
2 stale rules found (dry-run, no changes written)
```

---

## Requirements

- Go 1.21+
- Prometheus 2.x with query logging or `query_range` API access

---

## License

MIT © yourorg