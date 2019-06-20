#!/usr/bin/env bash

export HUMAN_LOG=true

go build -o countX

./countX -dir="/Users/dave/Desktop/zebedee-data/content/zebedee/master" \
    -types="article, bulletin, compendium_landing_page, compendium_chapter, static_methodology"
