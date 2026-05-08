#!/usr/bin/env bash

set -euo pipefail

# This is the entrypoint script to the Docker image. It takes a number of
# environment variables and constructs the "go test" arguments to use.
#
#  KCP_KUBECONFIG
#    the path to a kubeconfig targetting the front-proxy
#
#  KCP_SHARDS
#    a comma-separated list of all shards, including the root shard,
#    for example "root,shard1,shard7".
#
#  KCP_SHARD_<uppercased name>_KUBECONFIG
#    for each shard in KCP_SHARDS, the path to a kubeconfig targetting
#    the shard's API server, for example
#    KCP_SHARD_ROOT_KUBECONFIG=/kubeconfigs/root.kubeconfig
#
#    Each of these kubeconfigs must have a "shard-base" context.

# Validate required environment variables
if [ -z "${KCP_KUBECONFIG:-}" ]; then
  echo "Error: KCP_KUBECONFIG environment variable is not set" >&2
  exit 1
fi

if [ -z "${KCP_SHARDS:-}" ]; then
  echo "Error: KCP_SHARDS environment variable is not set" >&2
  exit 1
fi

# Construct the --shard-kubeconfigs argument
shard_kubeconfigs=""
IFS=',' read -ra shards <<< "$KCP_SHARDS"

for shard in "${shards[@]}"; do
  # Convert shard name to uppercase for environment variable lookup
  shard_upper=$(echo "$shard" | tr '[:lower:]' '[:upper:]' | tr '-' '_')
  kubeconfig_var="KCP_SHARD_${shard_upper}_KUBECONFIG"

  # Get the kubeconfig path from the environment variable
  kubeconfig_path="${!kubeconfig_var:-}"

  if [ -z "$kubeconfig_path" ]; then
    echo "Error: $kubeconfig_var environment variable is not set for shard '$shard'" >&2
    exit 1
  fi

  # Append to the shard-kubeconfigs argument
  if [ -z "$shard_kubeconfigs" ]; then
    shard_kubeconfigs="${shard}=${kubeconfig_path}"
  else
    shard_kubeconfigs="${shard_kubeconfigs},${shard}=${kubeconfig_path}"
  fi
done

export NO_GORUN=1
export GOMAXPROCS=1

set -x
exec go test -parallel 1 -v -timeout 1h ./test/e2e/... -args \
  --kcp-kubeconfig="${KCP_KUBECONFIG}" \
  --shard-kubeconfigs="${shard_kubeconfigs}"
