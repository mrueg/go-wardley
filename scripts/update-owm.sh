#!/bin/bash

# This updates the dist folder with the latest build from owm

TMP_DIR=$(mktemp -d)

git clone git@github.com:damonsk/onlinewardleymaps.git --depth=1 "${TMP_DIR}"

cp next.config.js "${TMP_DIR}"/frontend 

docker run -it --rm -v "${TMP_DIR}":/tmp -w /tmp/frontend node:18-alpine3.18 sh -c "yarn install && yarn cache clean && yarn build"

rsync -avx --delete "${TMP_DIR}"/frontend/dist/ ../wardley/dist/
