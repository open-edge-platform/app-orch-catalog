#!/bin/bash

# Copyright 2025 Intel Corp.
# SPDX-License-Identifier: Apache-2.0

#set -x
set -e

apt-get update

# Install Python and pip
python -m pip install --upgrade pip==26.0.1
python -m pip install openapi-spec-validator==0.8.4
python -m pip install -r requirements.txt

# Read .tool-versions file line by line
while IFS= read -r line; do
  # Skip empty lines and comments
  if [[ -z "$line" || "$line" =~ ^# ]]; then
    continue
  fi
  echo "Processing: $line"
  first_element=$(echo "$line" | awk '{print $1}')
  asdf plugin add "$first_element" || true
done < ".tool-versions"

asdf install

GOBIN=$(go env GOBIN)
export PATH=$PATH:"$GOBIN"
echo "export PATH=$PATH:$(go env GOBIN)" >> /etc/profile

#docker network create -d=bridge --subnet=172.19.0.0/24 kind

go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@v1.5.1
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
go install github.com/sudorandom/protoc-gen-connect-openapi@v0.25.4

asdf version
buf --version
docker --version
go version
kind version
kubectl version --client
k9s version -s
node --version
npm --version
protoc --version
protoc-gen-doc --version
pip --version
python --version
