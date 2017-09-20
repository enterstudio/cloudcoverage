#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

pkgArgs=${*:-}

pkgs=${pkgArgs:-$(find cmd/* -type d)}

for dir in $pkgs
do
	filename=$(echo -n "$dir" | cut -d "/" -f 2)
	echo "Building directory $dir to $filename:"
	echo -n "	go build -o $filename \"$dir/main.go\" ... "
	GOARCH=amd64 go build -o "$filename" "$dir/main.go"
	GOARCH=arm GOARM=7 go build -o "$filename-arm" "$dir/main.go"
	echo "done!"
done

echo "Everything built!"
