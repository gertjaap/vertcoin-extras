#!/bin/bash
cd react-ui
npm run build
set -e
rm -rf ../server/static
set +e
mv build ../server/static
cd ..
rm -rf releases
mkdir releases

GOOS=darwin GOARCH=amd64 packr build && zip ./releases/vertcoin-openassets-$1-mac-x64.zip vertcoin-openassets && rm vertcoin-openassets \
  && GOOS=linux GOARCH=amd64 packr build && zip ./releases/vertcoin-openassets-$1-linux-x64.zip vertcoin-openassets && rm vertcoin-openassets \
  && GOOS=linux GOARCH=386 packr build && zip ./releases/vertcoin-openassets-$1-linux-x86.zip vertcoin-openassets && rm vertcoin-openassets \
  && GOOS=windows GOARCH=amd64 packr build && zip ./releases/vertcoin-openassets-$1-windows-x64.zip vertcoin-openassets.exe && rm vertcoin-openassets.exe \
  && GOOS=windows GOARCH=386 packr build && zip ./releases/vertcoin-openassets-$1-windows-x86.zip vertcoin-openassets.exe && rm vertcoin-openassets.exe \
  && packr clean