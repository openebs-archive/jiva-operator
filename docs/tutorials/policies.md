## Jiva Volume Policies:

Jiva Volumes can be provisioned based on different policy configurations. JivaVolumePolicy has to be created prior to StorageClass 
and we have to mention the `JivaVolumePolicy` name in StorageClass parameters to provision jiva volume based on configured policy.

Following are list of policies that can be configured based on the requirements.

- [replicationFactor](#replication-factor)
- [Replica STS pod Anti Affinity](#replica-sts-pod-anti-affinity)
- [Target pod Affinity](#target-pod-affinity)
- [Resource Request and Limits](#resource-request-and-limits)
- [Priority Class](#priority-class)

Below StorageClass example contains `jivaVolumePolicy` parameter having `example-jivavolumepolicy` name set to configure the custom policy.

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-jiva-csi-sc
provisioner: jiva.csi.openebs.io
allowVolumeExpansion: true
parameters:
  cas-type: "jiva"
  jivaVolumePolicy: "example-jivavolumepolicy"
```

### Replication Factor:

Replication factor can be set based on the number of copies of the data required to be maintained.
If not provided, the replicationFactor is set to 3 by default.

```yaml
apiVersion: openebs.io/v1alpha1
kind: JivaVolumePolicy
metadata:
  name: example-jivavolumepolicy
  namespace: openebs
spec:
  target:
    replicationFactor: 3
```

### Replica STS Pod Anti Affinity:

The Stateful workloads access the OpenEBS storage volume by connecting to the Volume Target(Controller) Pod.
Replica Pod Anti Affinity policy can be used to distribute replica statefulset pod across the node to mitigate
high availability issues.

For example distributed applications like mongodb require the volumes to be spread across multiple nodes - just like its own replicas.
Cross scheduling them will cause performance and high availability issues. User will need to add the following podAntiAffinity
rule to volume Policy to enable replica anti affinity policies.

```yaml
apiVersion: openebs.io/v1alpha1
kind: JivaVolumePolicy
metadata:
  name: example-jivavolumepolicy
  namespace: openebs
spec:
  replicaSC: openebs-hostpath
  target:
    monitor: false
    replicationFactor: 1
  replica:
    affinity:
      podAntiAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
        - labelSelector:
            matchLabels:
              openebs.io/replica-anti-affinity: mongo-sts
          topologyKey: kubernetes.io/hostname

```

### Target Pod Affinity:

The Stateful workloads access the OpenEBS storage volume by connecting to the Volume Target(Controller) Pod.
Target Pod Affinity policy can be used to co-locate volume target pod on the same node as workload.
This feature makes use of the Kubernetes Pod Affinity feature that is dependent on the Pod labels.
User will need to add the following label to both Application and volume Policy.

Configured Policy having target-affinity label for example, using `kubernetes.io/hostname` as a topologyKey in JivaVolumePolicy:

```yaml
apiVersion: openebs.io/v1alpha1
kind: JivaVolumePolicy
metadata:
  name: example-jivavolumepolicy
  namespace: openebs
spec:
  target:
    affinity:
      podAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
        - labelSelector:
            matchExpressions:
            - key: openebs.io/target-affinity
              operator: In
              values:
              - fio-jiva                              // application-unique-label
          topologyKey: kubernetes.io/hostname
          namespaces: ["default"]                      // application namespace
```


Set the label configured in volume policy created above `openebs.io/target-affinity: fio-jiva` on the app pod which will be used to find pods by label, within the domain defined by topologyKey.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: fio-jiva
  namespace: default
  labels:
    name: fio-jiva
    openebs.io/target-affinity: fio-jiva
```

### Resource Request and Limits:

JivaVolumePolicy can be used to configure the volume Target/replica pod resource requests and
limits to ensure QOS. Below is the example to configure the target container resources
requests and limits, as well as auxResources configuration for the sidecar containers.

Learn more about (Resources configuration)[https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/]

```yaml
apiVersion: openebs.io/v1alpha1
kind: JivaVolumePolicy
metadata:
  name: example-jivavolumepolicy
  namespace: openebs
spec:
  target:
    resources:
      requests:
        memory: "64Mi"
        cpu: "250m"
      limits:
        memory: "128Mi"
        cpu: "500m"
    auxResources:
      requests:
        memory: "64Mi"
        cpu: "250m"
      limits:
        memory: "128Mi"
        cpu: "500m"
  replica:
    resources:
      requests:
        memory: "64Mi"
        cpu: "250m"
      limits:
        memory: "128Mi"
        cpu: "500m"
    auxResources:
      requests:
        memory: "64Mi"
        cpu: "250m"
      limits:
        memory: "128Mi"
        cpu: "500m"
```

### Target/Replica Pod Toleration:

This Kubernetes feature allows users to mark a node (taint the node) so that no pods can be scheduled to it, unless a pod explicitly tolerates the taint.
Using this Kubernetes feature we can label the nodes that are reserved (dedicated) for specific pods.

E.g. all the volume specific pods in order to operate flawlessly should be scheduled to nodes that are reserved for storage.

```yaml
apiVersion: openebs.io/v1alpha1
kind: JivaVolumePolicy
metadata:
  name: example-jivavolumepolicy
  namespace: openebs
spec:
  target:
    tolerations:
    - key: "key1"
      operator: "Equal"
      value: "value1"
      effect: "NoSchedule"
  replica:
    tolerations:
    - key: "key1"
      operator: "Equal"
      value: "value1"
      effect: "NoSchedule"
```


### Priority Class:

Priority classes can help you control the Kubernetes scheduler decisions to favor higher priority pods over lower priority pods.
The Kubernetes scheduler can even preempt (remove) lower priority pods that are running so that pending higher priority pods can be scheduled.
By setting pod priority, you can help prevent lower priority workloads from impacting critical workloads in your cluster, especially in cases where the cluster starts to reach its resource capacity.

Learn more about (PriorityClasses)[https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/#priorityclass]

*NOTE:* Priority class needs to be created before volume provisioning. In this case, `storage-critical` priority classes should exist.

```yaml
apiVersion: openebs.io/v1alpha1
kind: JivaVolumePolicy
metadata:
  name: example-jivavolumepolicy
  namespace: openebs
spec:
  target:
    priorityClassName: "storage-critical"
  replica:
    priorityClassName: "storage-critical"
```
