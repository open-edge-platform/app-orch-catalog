testing:

kubectl -n orch-app port-forward services/app-orch-catalog-rest-proxy 8081

curl -X "POST" http://localhost:8081/catalog.orchestrator.apis/v3/import -H "authorization: Bearer $api_token" -H "activeprojectid: <id>" -d "url=oci://ghcr.io/open-edge-platform/geti/helm/impt:2.9.0"
