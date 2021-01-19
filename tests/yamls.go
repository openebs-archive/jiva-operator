/*
Copyright © 2020 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package volume

var DeploymentName = "ubuntu"
var DeployYAML = `
apiVersion: apps/v1                                                                                                   
kind: Deployment                                                                                          
metadata:                                                                                                                                   
  name: ubuntu                                                                                                                              
  labels:                                                                                                                                   
    app.kubernetes.io/name: ubuntu
spec:
  selector:
    matchLabels:
      name: ubuntu
  replicas: 1
  strategy:
    type: Recreate
    rollingUpdate: null
  template:
    metadata:
      labels:
        name: ubuntu
    spec:
      containers:
      - name: ubuntu
        image: prateek14/ubuntu:18.04
        command: ["/usr/local/bin/pause"]
        livenessProbe:
          exec:
            command:
            - touch
            - /test1/file
          initialDelaySeconds: 5
          periodSeconds: 2
        volumeMounts:
        - mountPath: /test1
          name: my-volume
      volumes:
      - name: my-volume
        persistentVolumeClaim:
          claimName: jiva-pvc
`

var PVCName = "jiva-pvc"
var PVCYAML = `
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: jiva-pvc
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
  storageClassName: jiva-csi-sc
`

var ExpandedPVCYAML = `
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: jiva-pvc
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: jiva-csi-sc
`

var SCName = "jiva-sc"
var SCYAML = `
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: jiva-csi-sc
provisioner: jiva.csi.openebs.io
allowVolumeExpansion: true
parameters:
  cas-type: "jiva"
  policy: "example-jivavolumepolicy"
`

var NSName = "jiva-ns"
var NSYAML = `
kind: NameSpace
apiVersion: v1
metadata:
  name: jiva-ns
`

var policyYAML = `
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
    # monitor: false
    replicationFactor: 1
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
# status:
  # phase:
`
