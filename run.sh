#!/bin/sh
set -eux

MYSOCKS_PORT=58080 \
go run cmd/main.go
