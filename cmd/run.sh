#!/bin/sh
set -eux

exec ssh -vvv -o "ProxyCommand $(pwd)/cmd connect --sni $1" aidan@"$1"
