#!/bin/sh
set -e
rm -rf manpages
mkdir manpages
go run ./cli man | gzip -c -9 >manpages/aws-taggy.1.gz
