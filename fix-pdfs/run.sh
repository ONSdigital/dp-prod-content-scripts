#!/usr/bin/env bash

export HUMAN_LOG=true

go build -o fix

./fix -col="pdfTestDoNotPublish" \
    -limit=20 \
    -dir="/zebe-test" \
    -types="article, bulletin, compendium_landing_page, compendium_chapter, static_methodology"
