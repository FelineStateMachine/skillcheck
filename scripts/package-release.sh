#!/bin/sh
set -eu

version=${VERSION:-dev}
out=${OUT_DIR:-dist}
mkdir -p "$out"
for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64; do
  os=${target%/*}; arch=${target#*/}; ext=""; test "$os" = windows && ext=.exe
  name="skilltrace_${version}_${os}_${arch}"
  GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=$version" -o "$out/$name/skilltrace$ext" ./cmd/skilltrace
  cp README.md "$out/$name/README.md"
  (cd "$out" && tar -czf "$name.tar.gz" "$name")
  rm -rf "$out/$name"
done
