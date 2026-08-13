#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  printf 'usage: %s <version>\n' "$0" >&2
  exit 2
fi

version="$1"

if [[ ! "${version}" =~ ^[0-9]+[.][0-9]+[.][0-9]+([-.][0-9A-Za-z.-]+)?$ ]]; then
  printf 'invalid version %q; expected semver such as 1.0.0\n' "${version}" >&2
  exit 2
fi

commit="$(git rev-parse --short=12 HEAD)"
build_date="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
platform="${DEVCTX_WAILS_PLATFORM:-}"

ldflags="-X devctx/packages/core/version.Version=${version} \
-X devctx/packages/core/version.Commit=${commit} \
-X devctx/packages/core/version.Date=${build_date}"

args=(
  build
  -clean
  -trimpath
  -ldflags "${ldflags}"
)

if [[ -n "${platform}" ]]; then
  args+=(-platform "${platform}")
fi

# Ubuntu 24.04 uses WebKitGTK 4.1.
# This build tag should only be used for Linux builds.
if [[ "${platform}" == linux/* ]]; then
  args+=(-tags webkit2_41)
fi

# Disable platform packaging when explicitly requested.
if [[ "${DEVCTX_WAILS_NOPACKAGE:-0}" == "1" ]]; then
  args+=(-nopackage)
fi

# Generate an NSIS installer on Windows.
if [[ "${DEVCTX_WAILS_NSIS:-0}" == "1" ]]; then
  args+=(-nsis)
fi

# Configure how WebView2 is handled on Windows.
if [[ -n "${DEVCTX_WAILS_WEBVIEW2:-}" ]]; then
  args+=(-webview2 "${DEVCTX_WAILS_WEBVIEW2}")
fi

printf 'Building Dev Context\n'
printf '  Version:  %s\n' "${version}"
printf '  Commit:   %s\n' "${commit}"
printf '  Date:     %s\n' "${build_date}"
printf '  Platform: %s\n\n' "${platform:-current}"

wails "${args[@]}"

printf '\nBuild completed successfully.\n'
printf 'Artifacts:\n'

find build/bin -maxdepth 2 -print