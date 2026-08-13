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
date="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
platform="${DEVCTX_WAILS_PLATFORM:-}"
package_flag=()

if [[ "${DEVCTX_WAILS_NOPACKAGE:-0}" == "1" ]]; then
  package_flag=(-nopackage)
fi

ldflags="-X devctx/packages/core/version.Version=${version} -X devctx/packages/core/version.Commit=${commit} -X devctx/packages/core/version.Date=${date}"

args=(build -clean -trimpath -tags webkit2_41 -ldflags "${ldflags}" -o devctx)
if [[ -n "${platform}" ]]; then
  args+=(-platform "${platform}")
fi
args+=("${package_flag[@]}")

wails "${args[@]}"

artifact="build/bin/devctx"
if [[ "${platform}" == windows/* ]]; then
  artifact="build/bin/devctx.exe"
fi

if [[ -x "${artifact}" ]]; then
  "${artifact}" --version
else
  printf 'built artifacts in build/bin; verify the packaged executable with --version\n'
fi
