# Quickstart

## Prerequisites

1. Kubernetes version 1.17 or higher.

2. iSCSI initiator utils installed on all the worker nodes(If you are using rancher based cluster perform steps mentioned [here](troubleshooting/rancher_prerequisite.md)).


| OPERATING SYSTEM | iSCSI PACKAGE         | Commands to install iSCSI                                | Verify iSCSI Status         |
| ---------------- | --------------------- | -------------------------------------------------------- | --------------------------- |
| RHEL/CentOS      | iscsi-initiator-utils | <ul><li>sudo yum install iscsi-initiator-utils -y</li><li>sudo systemctl enable --now iscsid</li><li>modprobe iscsi_tcp</li><li>echo iscsi_tcp >/etc/modules-load.d/iscsi-tcp.conf</li></ul> | sudo systemctl status iscsid.service |
| Ununtu/ Debian   | open-iscsi            |  <ul><li>sudo apt install open-iscsi</li><li>sudo systemctl enable --now iscsid</li><li>modprobe iscsi_tcp</li><li>echo iscsi_tcp >/etc/modules-load.d/iscsi-tcp.conf</li></ui>| sudo systemctl status iscsid.service |
| RancherOS        | open-iscsi            |  <ul><li>sudo ros s enable open-iscsi</li><li>sudo ros s up open-iscsi</li></ui>| ros service list iscsi |


3. Access to install RBAC components into kube-system namespace.

4. OpenEBS version 2.6.0 or higher.
```
kubectl apply -f https://raw.githubusercontent.com/openebs/openebs/master/k8s/openebs-operator.yaml
```

## Install

### Using Helm Charts:
 
Install Jiva components using [jiva helm charts](https://github.com/openebs/jiva-operator/tree/master/deploy/helm/charts).

### Using Operator:

Install the [latest release](https://github.com/openebs/jiva-operator/releases) using Jiva Operator yamls.

```
kubectl apply -f https://openebs.github.io/charts/jiva-operator.yaml
```
Once installed using any of the above methods, verify that Jiva Operator and jiva csi pods are running. 

```bash
$ kubectl get pod -n openebs

NAME                                           READY   STATUS    RESTARTS   AGE
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

1. Create Jiva volume policy to set various policies for creating a jiva volume.
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
    By default, volume data is stored at `/var/openebs/<pvc-*>` at the worker nodes,
    to change this behavior, a new replicaSC needs to be created.
    This tutorial can be referred to create replicaSC and make use of various policies.

2. Create a Storage Class to dynamically provision volumes by specifying above policy:
   ```
   apiVersion: storage.k8s.io/v1
   kind: StorageClass
   metadata:
     name: openebs-jiva-csi-sc
   provisioner: jiva.csi.openebs.io
   allowVolumeExpansion: true
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
   ```
   $ kubectl get pvc
   NAME                   STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS           AGE
   example-jiva-csi-pvc   Bound    pvc-ffc1e885-0122-4b5b-9d36-ae131717a77b   4Gi        RWO            openebs-jiva-csi-sc    1m
   ```

   Verify volume is ready to serve IOs.
   ```
   $ kubectl get jivavolume pvc-ffc1e885-0122-4b5b-9d36-ae131717a77b -n openebs
   NAME                                       REPLICACOUNT   PHASE   STATUS
   pvc-ffc1e885-0122-4b5b-9d36-ae131717a77b   1              Ready   RW
   ```

4. Deploy an application using the above PVC:
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
   ```
   $ kubectl get po
   NAME                   READY   STATUS    RESTARTS   AGE
   fio-68c4c5b545-vg2rc   1/1     Running   0          2m
   ```
