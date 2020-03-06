# Config-watcher

Small bash script for use as a sidecar container in a K8s environments. Watches
configmap-provisioned files, and reloads processes if their config file changed.
