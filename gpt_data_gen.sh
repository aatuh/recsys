#!/bin/bash

mkdir -p .trash/gpt/

./.weave api/ --ignore swagger --output .trash/gpt/api_source_code.txt
./.weave api/swagger --output .trash/gpt/swagger_docs.txt
cp README.md .trash/gpt/README.md