#!/usr/bin/env bash

export HUMAN_LOG=true

go build -o fix

./fix -col="pdf_test_do_not_publish" \
    -limit=20 \
    -dir="/zebe-test" \
    -types="article, bulletin, compendium_landing_page, compendium_chapter, static_methodology"
