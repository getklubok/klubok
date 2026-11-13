#!/bin/bash

BASE_DIR="./internal"

find $BASE_DIR -type d -name "mocks" | while read -r mock_dir; do
    find "$mock_dir" -type f -name "*.go" -exec rm -f {} \;
done
