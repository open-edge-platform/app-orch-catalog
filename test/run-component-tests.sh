#!/bin/bash

set -euo pipefail

# Set environment variables
export ORCH_DEFAULT_PASSWORD="ChangeMeOn1stLogin!"
export CODER_DIR="$(pwd)"
export GIT_USER="git"
export DOCKERHUB_TOKEN="your-dockerhub-token"
export DOCKERHUB_USERNAME="your-dockerhub-username"

# Checkout repositories
echo "Checking out edge-manageability-framework..."
git clone --branch "53e430f45faaee7671fa3c25f6acc3945aba2065" https://github.com/open-edge-platform/edge-manageability-framework.git emf
cd emf
git fetch --all
cd ..

echo "Checking out app-orch-catalog..."
git clone https://github.com/open-edge-platform/app-orch-catalog.git app-orch-catalog
cd app-orch-catalog
GIT_HASH_CHARTS=$(git rev-parse HEAD)
echo "GIT_HASH_CHARTS=${GIT_HASH_CHARTS}"
cd ..

# Deploy External Orchestrator
echo "Deploying External Orchestrator..."
cd emf
mage deploy:kindMinimal
echo "Orchestrator deployment done!"
kubectl -n dev get applications root-app -o yaml
cd ..

# Verify Kind Deployment
echo "Verifying Orchestrator deployment..."
cd emf
mage deploy:waitUntilComplete &
WAIT_PID=$!
while kill -0 $WAIT_PID 2>/dev/null; do
  echo "Waiting for Orchestrator deployment to complete..."
  kubectl get pods -A || true
  sleep 30
done
wait $WAIT_PID || true
mage router:stop router:start || true
echo "Router restarted"
cd ..

# Setup Test Environment
echo "Setting up test environment..."
sudo awk -i inplace '/BEGIN ORCH DEVELOPMENT HOSTS/,/END ORCH DEVELOPMENT HOSTS/ { next } 1' /etc/hosts
sudo awk -i inplace '/BEGIN ORCH SRE DEVELOPMENT HOST/,/END ORCH SRE DEVELOPMENT HOST/ { next } 1' /etc/hosts
cd emf
mage gen:hostfileTraefik | sudo tee -a /etc/hosts > /dev/null
mage gen:orchCa deploy:orchCa
cd ..

# Setup users and project/org
echo "Setting up users and project/org..."
cd emf
mage tenantUtils:createDefaultMtSetup
kubectl get projects.project -o json | jq -r ".items[0].status.projectStatus.uID"
cd ..

# Redeploy and Rebuild app-orch-catalog
echo "Redeploying and rebuilding app-orch-catalog..."
cd app-orch-catalog
make coder-redeploy
make coder-rebuild
cd ..

# Wait for app-orch-catalog pod to be Running
echo "Waiting for app-orch-catalog pod to be Running..."
while true; do
  POD_NAME=$(kubectl get pods -n orch-app -l app.kubernetes.io/instance=app-orch-catalog -o jsonpath='{.items[0].metadata.name}')
  POD_STATUS=$(kubectl get pod $POD_NAME -n orch-app -o jsonpath='{.status.phase}')
  if [ "$POD_STATUS" == "Running" ]; then
    echo "Pod $POD_NAME is in Running state."
    break
  else
    echo "Pod $POD_NAME is not in Running state. Current state: $POD_STATUS"
    echo "Waiting for 10 seconds before checking again..."
    sleep 10
  fi
done

# Run Catalog Component Tests
echo "Running Catalog Component Tests..."
cd app-orch-catalog/test
make component-tests
echo "Component tests done!"
cd ../..

# Get diagnostic information
echo "Collecting diagnostic information..."
kubectl get pods -o wide -A | tee pods-list.txt
kubectl describe pods -A | tee pods-describe.txt
cd emf
mage logutils:collectArgoDiags | tee ../argo-diag.txt
cd ..
kubectl get applications -o yaml -A | tee argocd-applications.yaml

echo "All steps completed successfully!"
