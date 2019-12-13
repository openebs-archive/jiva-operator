# jiva-operator

[![Releases](https://img.shields.io/github/release/openebs/openebs/all.svg?style=flat-square)](https://github.com/openebs/openebs/releases)
[![Slack](https://img.shields.io/badge/chat!!!-slack-ff1493.svg?style=flat-square)]( https://openebs-community.slack.com)
[![Twitter](https://img.shields.io/twitter/follow/openebs.svg?style=social&label=Follow)](https://twitter.com/intent/follow?screen_name=openebs)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](https://github.com/openebs/openebs/blob/master/CONTRIBUTING.md)


https://openebs.org/

## Overview

Jiva operator is a custom kubernetes controller, which will continuously watch
for the JivaVolume CR and will do the bootstrapping of jiva components such as
creating service and deployment of jiva controller, deploy jiva replicas as
statefulsets using localpv for persisting the data.

Jiva Operator helps with managing the lifecycle and operations on Jiva Volumes.
Jiva Operator use the JivaVolume CR to perform the operations on Jiva Volume and it is implemented using Operator SDK.

Jiva Operator does the following:

- Launches Jiva Volume CRD into the cluster if not already present.
- Launches Jiva Volume components, when a new Jiva Volume CR is created and updates the Jiva Volume CR status attributes.
- Clears the Jiva Volume components when a Jiva Volume CR is deleted
- Performs update operations like Volume Expansion
- Handles scenarios like node failure and creation of new replicas.

A Jiva Volume comprises of the following components:

- Jiva Target Deployment and an associated Service
- Jiva Replica StatefulSet with Hostpath Local PVs for saving the data.

## Quick Start

### Prerequisite
- Kubernetes version should be > 1.14.

### Installation
Run following commands to proceed with the installation:
- Install openebs control plane components:
  ```
  kubectl apply -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/openebs-operator.yaml
  ```
- Install jiva-operator CRD:
  ```
  kubectl apply -f deploy/crds/openebs_v1alpha1_jivavolume_crd.yaml
  ```
- Install jiva-operator:
  ```
  kubectl create -f deploy/
  ```
- After the installation of control plane components and operator, it will look
  like below:
  ```
  jiva-operator-7765cbfffd-vt787                 1/1     Running   0          10s
  maya-apiserver-5c5d944d-fpkfj                  1/1     Running   2          2m5s
  openebs-admission-server-5959f9f9cd-vcwfw      1/1     Running   0          119s
  openebs-localpv-provisioner-57b44f4664-klsrw   1/1     Running   0          118s
  openebs-ndm-6dtjz                              1/1     Running   0          2m1s
  openebs-ndm-operator-f84848f77-j57vr           1/1     Running   1          2m
  openebs-ndm-qfrjf                              1/1     Running   0          2m1s
  openebs-ndm-tgpmk                              1/1     Running   0          2m1s
  openebs-provisioner-cd5759f96-jfcxb            1/1     Running   0          2m3s
  openebs-snapshot-operator-5f87bd54bf-mmtlh     2/2     Running   0          2m2s
  ```
