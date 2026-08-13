#!/usr/bin/env bash
set -euo pipefail

count="${DEVCTX_BENCH_COUNT:-5}"

go test \
  ./packages/core/cli \
  ./packages/application \
  -run '^$' \
  -bench 'Benchmark(DirectLaunchPreparation|InteractiveLaunchStateLoading)$' \
  -benchmem \
  -count "${count}"
