# Config-watcher

Small bash script for use as a sidecar container in a K8s environments. Watches
configmap-provisioned files, and reloads processes if their config file changed.

## Example usage

```
export RELOAD_SIGNAL=HUP
export TARGET_FILE=/etc/prometheus/prometheus.yml
export TARGET_PROCESS=prometheus

./config-watcher
```

## Configuration

Configuration is done via the following environment variables.

- `RELOAD_SIGNAL` (required): Signal to send the process to triger a reload.
  Must be valid value for `-s` parameter of `kill` command.
- `TARGET_FILE` (required): Path to file which to watch for changes.
- `TARGET_PROCESS`: Name of process which to send reload signal to. Either this
  or `TARGET_PID` must be set.
- `TARGET_PID`: PID of process which to send reload signal to. Either this or
  `TARGET_PROCESS` must be set.
- `VERBOSE` (default: ''): Set to non-empty value to enable verobse logging.
- `SLEEP` (default 1): Duration in seconds which to sleep between config file
  checks.
- `SLEEP_BEFORE_KILL` (default 1): Duration in seconds which to sleep before
  sending a reload signal when a change in the watched file was detectd.
  Ensures config map changes were populated to all containers.
