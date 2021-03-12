# Jiva Operator

[![Releases](https://img.shields.io/github/release/openebs/openebs/all.svg?style=flat-square)](https://github.com/openebs/openebs/releases)
[![Slack channel #openebs](https://img.shields.io/badge/slack-openebs-brightgreen.svg?logo=slack)](https://kubernetes.slack.com/messages/openebs)
[![Community Meetings](https://img.shields.io/badge/Community-Meetings-blue)](https://hackmd.io/hiRcXyDTRVO2_Zs9fp0CAg)
[![Twitter](https://img.shields.io/twitter/follow/openebs.svg?style=social&label=Follow)](https://twitter.com/intent/follow?screen_name=openebs)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](https://github.com/openebs/openebs/blob/master/CONTRIBUTING.md)
[![Go Report Card](https://goreportcard.com/badge/github.com/openebs/jiva-csi)](https://goreportcard.com/report/github.com/openebs/jiva-operator)


https://openebs.org/

### Overview

Jiva Operator helps with managing the lifecycle and operations on Jiva Volumes.

A Jiva Volume comprises of the following components:

- Jiva Controller Deployment and an associated Service.
- Jiva Replica StatefulSet with Hostpath Local PVs for saving the data.
- JivaVolume CR containing the status of the volume.

### Compatibility and Feature Matrix

| Operator Version | K8s Version | OpenEBS Version | Dynamic Provisioning | Resize (Expansion) | Snapshots | Raw Block | AccessModes | Status |
| ---------------- | --------------- | ------------------- | --------------- | --------------------------- | ----------- | --------- | --------- | ---------- |
| 2.6.0+ |   1.17+   |   2.6.0+   |   yes   |    yes    |   no   |   yes   |   RWO   | alpha |

### Installation
Install the following components before creating the volumes:
1. OpenEBS Control Plane:
  ```
  kubectl apply -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/openebs-operator.yaml
  ```
2. Jiva Operator:
  ```
  kubectl apply -f https://raw.githubusercontent.com/openebs/jiva-operator/master/deploy/operator.yaml
  ```
3. Jiva CSI Driver:
  ```
  kubectl apply -f https://raw.githubusercontent.com/openebs/jiva-operator/master/deploy/jiva-csi.yaml
  ```
Verify if all the above components are installed and running:
  ```
  $ kubectl get pods -n openebs
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
  openebs-jiva-csi-controller-0                  4/4     Running   0          6m14s
  openebs-jiva-csi-node-56t5g                    2/2     Running   0          6m13s
  openebs-jiva-csi-node-xtyhu                    2/2     Running   0          6m20s
  openebs-jiva-csi-node-h2unk                    2/2     Running   0          6m38s
  ```
### Steps to provision a Jiva Volume

1. Create Jiva volume policy to set various policies for creating
   jiva volume.
   A sample jiva volume policy CR looks like:
   ```
    apiVersion: openebs.io/v1alpha1
    kind: JivaVolumePolicy
    metadata:
      name: example-jivavolumepolicy
      namespace: openebs
    spec:
      replicaSC: openebs-hostpath
      enableBufio: false
      autoScaling: false
      target:
        replicationFactor: 1
        # monitor: false
        # auxResources:
        # tolerations:
        # resources:
        # affinity:
        # nodeSelector:
        # priorityClassName:
      # replica:
        # tolerations:
        # resources:
        # affinity:
        # nodeSelector:
        # priorityClassName:
    ```
2. Create a Storage Class to dynamically provision volumes by specifying above policy:
   ```
   apiVersion: storage.k8s.io/v1
   kind: StorageClass
   metadata:
     name: openebs-jiva-csi-sc
   provisioner: jiva.csi.openebs.io
   parameters:
     cas-type: "jiva"
     policy: "example-jivavolumepolicy"
   ```
3. Create PVC by specifying the above Storage Class in the PVC spec
   ```
   kind: PersistentVolumeClaim
   apiVersion: v1
   metadata:
     name: example-jiva-csi-pvc
   spec:
     storageClassName: openebs-jiva-csi-sc
     accessModes:
       - ReadWriteOnce
     resources:
       requests:
         storage: 4Gi
   ```
4. Deploy your application by specifying the PVC name:
   ```
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: fio
   spec:
     selector:
       matchLabels:
         name: fio
     replicas: 1
     strategy:
       type: Recreate
       rollingUpdate: null
     template:
       metadata:
         labels:
           name: fio
       spec:
         containers:
         - name: perfrunner
           image: openebs/tests-fio
           command: ["/bin/bash"]
           args: ["-c", "while true ;do sleep 50; done"]
           volumeMounts:
         - mountPath: /datadir
           name: fio-vol
         volumes:
         - name: fio-vol
           persistentVolumeClaim:
             claimName: example-jiva-csi-pvc
   ```
### Raising Issues And PRs

If you want to raise any issue for jiva-operator please do that at [openebs/openebs].

### Contributing

If you would like to contribute to code and are unsure about how to proceed,
please get in touch with the maintainers on Kubernetes Slack #openebs [channel].

Please read the contributing guidelines [here](./CONTRIBUTING.md).

### Code of conduct

Please read the community code of conduct [here](./CODE_OF_CONDUCT.md).

[Docker environment]: https://docs.docker.com/engine
[Go environment]: https://golang.org/doc/install
[openebs/openebs]: https://github.com/openebs/openebs
[channel]: https://kubernetes.slack.com/messages/openebs/
