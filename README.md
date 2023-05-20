# Openldap Operator

Openldap Operator for Kubernetes

## Installation

```
helm add repo openldap https://qwp0905.github.io/openldap-operator
helm repo update
```
```
helm upgrade --install openldap-operator openldap/openldap-operator
```

## Quick Start

```
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: openldap-admin
type: Opaque
data:
  password: # some of user secret
---
# cluster.yaml
apiVersion: openldap.kwonjin.click/v1
kind: OpenldapCluster
metadata:
  name: openldap
spec:
  template:
    image: qwp1216/openldap:2.6.4
    imagePullPolicy: IfNotPresent
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
  replicas: 3
  storage:
    volumeClaimTemplate:
      resources:
        requests:
          storage: 8Gi
      accessModes:
        - ReadWriteOnce
  openldapConfig:
    adminUsername: admin
    adminPassword:
      key: password
      name: openldap-admin
    root: dc=example,dc=com
```

