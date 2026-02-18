#!/bin/bash
export SPANNER_EMULATOR_HOST=localhost:9010

# Create instance
gcloud spanner instances create test-instance \
    --config=emulator-config --description="Test Instance" --nodes=1

# Create database
gcloud spanner databases create catalog-db --instance=test-instance