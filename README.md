# Config-watcher

Small go application for use as a sidecar container in a K8s environments.
Watches configmap-provisioned files, and reloads processes if their config file
changed.

## Example usage

```
export RELOAD_SIGNAL=SIGHUP
export TARGET_FILES=/etc/prometheus/prometheus.yml,/etc/prometheus/rules.yml
export TARGET_PROCESS=prometheus

./config-watcher
```

## Configuration

Configuration is done via the following environment variables.

- `RELOAD_SIGNAL` (required): Signal to send the process to triger a reload.
  Must be the full signal name, eg `SIGHUP` or `SIGUSR1`.
- `TARGET_FILES` (required): Comma-separated list of paths or glob patterns to
  files which to watch for changes. Note that these globs might be more
  restrictive than what you are used to, with not support for eg `**/*` to
  match files in arbitarily deep directory hierarchies. See
  https://golang.org/pkg/path/filepath/#Match for the pattern syntax.
- `TARGET_PROCESS` (required): Name of process which to send reload signal to.
- `SLEEP_DURATION` (default 1): Duration in seconds which to sleep between
  config file checks.
- `SLEEP_BEFORE_RELOAD_DURATION` (default 1): Duration in seconds which to
  sleep before sending a reload signal when a change in the watched file was
  detectd.  Ensures config map changes were populated to all containers.
