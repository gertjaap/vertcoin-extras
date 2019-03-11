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

GOOS=darwin GOARCH=amd64 packr2 build && zip ./releases/vertcoin-extras-$1-mac-x64.zip vertcoin-extras && rm vertcoin-extras \
  && GOOS=linux GOARCH=amd64 packr2 build && zip ./releases/vertcoin-extras-$1-linux-x64.zip vertcoin-extras && rm vertcoin-extras \
  && GOOS=linux GOARCH=386 packr2 build && zip ./releases/vertcoin-extras-$1-linux-x86.zip vertcoin-extras && rm vertcoin-extras \
  && GOOS=windows GOARCH=amd64 packr2 build && zip ./releases/vertcoin-extras-$1-windows-x64.zip vertcoin-extras.exe && rm vertcoin-extras.exe \
  && GOOS=windows GOARCH=386 packr2 build && zip ./releases/vertcoin-extras-$1-windows-x86.zip vertcoin-extras.exe && rm vertcoin-extras.exe \
  && packr clean
