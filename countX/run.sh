#!/usr/bin/env bash

export HUMAN_LOG=true

go build -o counter

./counter -dir="/zebe-test/master" \
    -types="article, bulletin, compendium_landing_page, compendium_chapter, static_methodology"
