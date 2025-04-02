<!---
  SPDX-FileCopyrightText: (C) 2025 Intel Corporation
  SPDX-License-Identifier: Apache-2.0
-->
# Ingress KinD setup

In order to use k8s Ingress in kind, we need to setup cert-manager and nginx Ingress.  This allows us to more closely mimic demo server configuration in local development environment.  In the initial deployment, Ingress is accessing gRPC k8s service.  

Docker Desktop is adding `kubernetes.docker.internal` to hosts file so there is no need to do anything else.  

But if you are not using Docker Desktop, you will have to add this entry to `/etc/hosts`:
```
127.0.0.1 kubernetes.docker.internal
```

After installing kind cluster with `make kind`, you will need to wait for all pods in cert-manager and ingress-nginx namespaces to be up and `Running`:

```
kubectl get pods -A
NAMESPACE            NAME                                         READY   STATUS      RESTARTS   AGE
cert-manager         cert-manager-99bb69456-knmwc                 1/1     Running     0          4m14s
cert-manager         cert-manager-cainjector-ffb4747bb-q5cfd      1/1     Running     0          4m14s
cert-manager         cert-manager-webhook-545bd5d7d8-bghgb        1/1     Running     0          4m14s
ingress-nginx        ingress-nginx-admission-create-2sj8j         0/1     Completed   0          4m14s
ingress-nginx        ingress-nginx-admission-patch-6nv7c          0/1     Completed   0          4m14s
ingress-nginx        ingress-nginx-controller-58c49c4db-n4src     1/1     Running     0          4m14s
kube-system          coredns-565d847f94-b7cbc                     1/1     Running     0          4m14s
kube-system          coredns-565d847f94-d4tsl                     1/1     Running     0          4m14s
kube-system          etcd-kind-control-plane                      1/1     Running     0          4m28s
kube-system          kindnet-22d62                                1/1     Running     0          4m14s
kube-system          kube-apiserver-kind-control-plane            1/1     Running     0          4m28s
kube-system          kube-controller-manager-kind-control-plane   1/1     Running     0          4m28s
kube-system          kube-proxy-hxvw5                             1/1     Running     0          4m14s
kube-system          kube-scheduler-kind-control-plane            1/1     Running     0          4m28s
local-path-storage   local-path-provisioner-684f458cdd-6tpvh      1/1     Running     0          4m14s
```

Then you can deploy application catalog helm chart:
```
make chart-install-kind
```

