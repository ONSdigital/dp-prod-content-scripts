#!/usr/bin/env bash

export HUMAN_LOG=true

go build -o approval

./approval -dir="/zebe-test/collections" -col="pdf_test_do_not_publish"
