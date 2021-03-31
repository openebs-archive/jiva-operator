## How to Expand/Resize Jiva Volume

#### Prerequisites:

- StorageClass should have the following parameter set:
  `allowVolumeExpansion: true`
```yaml
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

#### Resizing volumes containing a file system:
It is internally a two step process for volumes containing a file system:
- Volume expansion
- FileSystem expansion

To resize a PV, edit the PVC definition and update the `spec.resources.requests.storage` to reflect the newly desired size, which must be greater than the original size.
- There are two scenarios when resizing a Jiva PV:
    - If the PV is attached to a pod, Jiva CSI driver expands the volume on the storage backend, rescans the device and resizes the filesystem.
    - When attempting to resize an unattached PV, Jiva CSI driver expands the volume on the storage backend. Once the PVC is bound to a pod, driver rescans the device and resizes the filesystem. 
- Kubernetes updates the PVC size after both the above mentioned steps are successfully completed.

For example, an application `busybox` pod is using the below PVC:
```sh
$ kubectl get pods
NAME            READY   STATUS    RESTARTS   AGE
busybox         1/1     Running   0          38m

$ kubectl get pvc
NAME                           STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS              AGE
jiva-pvc                       Bound    pvc-26dc4d24-1e2e-4727-9804-bcd7ce40364d   5Gi        RWO            openebs-jiva-csi-sc       1d

$ kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM               STORAGECLASS         REASON   AGE
pvc-26dc4d24-1e2e-4727-9804-bcd7ce40364d   5Gi        RWO            Delete           Bound    default/jiva-pvc    openebs-jiva-csi-sc           1d
```
To resize the PV from 5Gi to 10Gi, edit the PVC definition and update the `spec.resources.requests.storage` to 10Gi.
It may take few seconds to update the actual size in PVC resource, wait for the updated capacity to reflect in PVC status (pvc.status.capacity.storage).
```sh
$ kubectl edit pvc jiva-pvc
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  annotations:
    pv.kubernetes.io/bind-completed: "yes"
    pv.kubernetes.io/bound-by-controller: "yes"
    volume.beta.kubernetes.io/storage-provisioner: jiva.csi.openebs.io
  creationTimestamp: "2021-03-22T12:24:22Z"
  finalizers:
  - kubernetes.io/pvc-protection
  name: jiva-pvc
  namespace: default
  resourceVersion: "169312766"
  selfLink: /api/v1/namespaces/default/persistentvolumeclaims/jiva-pvc
  uid: 26dc4d24-1e2e-4727-9804-bcd7ce40364d
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
  storageClassName: openebs-jiva-csi-sc
  volumeMode: Filesystem
  volumeName: pvc-26dc4d24-1e2e-4727-9804-bcd7ce40364d
status:
  accessModes:
  - ReadWriteOnce
  capacity:
    storage: 5Gi
  phase: Bound
```

Now We can validate if the resize operation has worked correctly by checking the size of the PVC, PV, or describing the pvc to get all events.

```sh
$ kubectl describe pvc jiva-pvc

Name:          jiva-pvc
Namespace:     default
StorageClass:  openebs-jiva-csi-sc
Status:        Bound
Volume:        pvc-26dc4d24-1e2e-4727-9804-bcd7ce40364d
Labels:        <none>
Annotations:   pv.kubernetes.io/bind-completed: yes
               pv.kubernetes.io/bound-by-controller: yes
               volume.beta.kubernetes.io/storage-provisioner: jiva.csi.openebs.io
Finalizers:    [kubernetes.io/pvc-protection]
Capacity:      10Gi
Access Modes:  RWO
VolumeMode:    Filesystem
Mounted By:    busybox
Events:
  Type     Reason                      Age                From                                                                                      Message
  ----     ------                      ----               ----                                                                                      -------
  Warning  ExternalExpanding           109s  volume_expand                         Ignoring the PVC: didn't find a plugin capable of expanding the volume; waiting for an external controller to process this PVC.
  Normal   Resizing                    109s  external-resizer jiva.csi.openebs.io  External resizer is resizing volume pvc-26dc4d24-1e2e-4727-9804-bcd7ce40364d
  Normal   FileSystemResizeRequired    107s  external-resizer jiva.csi.openebs.io  Require file system resize of volume on node
  Normal   FileSystemResizeSuccessful  40s   kubelet                               MountVolume.NodeExpandVolume succeeded for volume "pvc-26dc4d24-1e2e-4727-9804-bcd7ce40364d"
```
```sh
$ kubectl get pvc
NAME                           STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS              AGE
jiva-pvc                       Bound    pvc-26dc4d24-1e2e-4727-9804-bcd7ce40364d   10Gi        RWO           openebs-jiva-csi-sc       1d

$ kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                   STORAGECLASS        REASON   AGE
pvc-849bd646-6d3f-4a87-909e-2416d4e00904   10Gi        RWO            Delete           Bound    default/jiva-csi       openebs-jiva-csi-sc          1d
```

#### Resizing volumes being used in BlockMode
Resizing volumes being used in block mode requires an additional step:
1. To resize a PV, edit the PVC definition and update the `spec.resources.requests.storage` to reflect the newly desired size, which must be greater than the original size. This step can be performed in a similar way as shown above.
2. Delete the application pod so that a new application pod is created and the volume with updated size is attached.
