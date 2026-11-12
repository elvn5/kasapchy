#!/bin/sh
# Certbot deploy hook: applies renewed certificates without stopping existing connections.
set -eu
nginx -t
systemctl reload nginx
