#!/bin/bash

echo 'Building compiler executable...'
go build -o the ./cmd/the
if [[ $? -eq 0 ]]; then
    echo 'Successfully built executable';
else
    echo 'Failed to build executable';
fi