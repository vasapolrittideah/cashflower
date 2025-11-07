#!/bin/bash

# Script to run golangci-lint on unique directories from staged Go files

# Get all staged .go files passed as arguments
staged_files=("$@")

# Extract unique directories from staged files
directories=()
for file in "${staged_files[@]}"; do
    dir=$(dirname "$file")

    if [[ ! " ${directories[@]} " =~ " ${dir} " ]]; then
        directories+=("$dir")
    fi
done

# Run golangci-lint fmt and run --fix for each unique directory
for dir in "${directories[@]}"; do
    echo "Running golangci-lint on directory: $dir"

    golangci-lint fmt "$dir"

    golangci-lint run --fix "$dir"
done
