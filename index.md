# OpenEBS Jiva Helm Charts

<img width="300" align="right" alt="OpenEBS Logo" src="https://raw.githubusercontent.com/cncf/artwork/master/projects/openebs/stacked/color/openebs-stacked-color.png" xmlns="http://www.w3.org/1999/html">

[Helm3](https://helm.sh) must be installed to use the charts.
Please refer to Helm's [documentation](https://helm.sh/docs/) to get started.

Once Helm is set up properly, add OpenEBS Jiva repository to Helm repos:

```bash
helm repo add openebs-jiva https://openebs-archive.github.io/jiva-operator
```

You can then run `helm search repo openebs-jiva` to see the charts.


#### Update OpenEBS Jiva Repo

Once OpenEBS Jiva repository has been successfully fetched into the local system, it has to be updated to get the latest version. The OpenEBS Jiva charts repo can be updated using the following command.

```bash
helm repo update
```

#### Install using Helm 3

- Assign openebs namespace to the current context:
```bash
kubectl config set-context <current_context_name> --namespace=openebs
```

- If namespace is not created, run the following command
```bash
helm install <your-relase-name> openebs-jiva/jiva --create-namespace
```
- Else, if namespace is already created, run the following command
```bash
helm install <your-relase-name> openebs-jiva/jiva
```

_See [configuration](#configuration) below._

_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._


## Dependencies

By default this chart installs additional, dependent charts:

| Repository | Name | Version |
|------------|------|---------|
| https://openebs.github.io/dynamic-localpv-provisioner | localpv-provisioner | 3.5.0 |


To disable the dependency during installation, set `openebsLocalpv.enabled` to `false`.

```bash
helm install <your-relase-name> openebs-jiva/jiva --set openebsLocalpv.enabled=false
```

For more details on dependency see [Jiva chart readme](https://github.com/openebs/jiva-operator/blob/master/deploy/helm/charts/README.md).

_See [helm dependency](https://helm.sh/docs/helm/helm_dependency/) for command documentation._

## Uninstall Chart

```console
# Helm
$ helm uninstall [RELEASE_NAME]
```

This removes all the Kubernetes components associated with the chart and deletes the release.

_See [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) for command documentation._

## Upgrading Chart

```console
# Helm
$ helm upgrade [RELEASE_NAME] [CHART] --install
```

## Configuration

For more details and instructions see [Jiva chart readme](https://github.com/openebs/jiva-operator/blob/master/deploy/helm/charts/README.md).

The following table lists the configurable parameters of the OpenEBS Jiva chart and their default values.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| csiController.annotations | object | `{}` | CSI controller annotations |
| csiController.attacher.image.pullPolicy | string | `"IfNotPresent"` | CSI attacher image pull policy |
| csiController.attacher.image.registry | string | `"k8s.gcr.io/"` |  CSI attacher image registry |
| csiController.attacher.image.repository | string | `"k8scsi/csi-attacher"` |  CSI attacher image repo |
| csiController.attacher.image.tag | string | `"v3.1.0"` | CSI attacher image tag |
| csiController.attacher.name | string | `"csi-attacher"` |  CSI attacher container name|
| csiController.componentName | string | `"openebs-jiva-csi-controller"` | CSI controller component name |
| csiController.driverRegistrar.image.pullPolicy | string | `"IfNotPresent"` | CSI driver registrar image pull policy  |
| csiController.driverRegistrar.image.registry | string | `"k8s.gcr.io/"` | CSI driver registrar image registry |
| csiController.driverRegistrar.image.repository | string | `"k8scsi/csi-cluster-driver-registrar"` | CSI driver registrar image repo |
| csiController.driverRegistrar.image.tag | string | `"v1.0.1"` |  CSI driver registrar image tag|
| csiController.driverRegistrar.name | string | `"csi-cluster-driver-registrar"` | CSI driver registrar container name  |
| csiController.livenessprobe.image.pullPolicy | string | `"IfNotPresent"` | CSI livenessprobe image pull policy |
| csiController.livenessprobe.image.registry | string | `"k8s.gcr.io/"` |  CSI livenessprobe image registry |
| csiController.livenessprobe.image.repository | string | `"k8scsi/livenessprobe"` |  CSI livenessprobe image repo |
| csiController.livenessprobe.image.tag | string | `"v2.2.0"` | CSI livenessprobe image tag |
| csiController.livenessprobe.name | string | `"liveness-probe"` |  CSI livenessprobe container name|
| csiController.nodeSelector | object | `{}` |  CSI controller pod node selector |
| csiController.podAnnotations | object | `{}` | CSI controller pod annotations |
| csiController.provisioner.image.pullPolicy | string | `"IfNotPresent"` | CSI provisioner image pull policy |
| csiController.provisioner.image.registry | string | `"k8s.gcr.io/"` | CSI provisioner image pull registry |
| csiController.provisioner.image.repository | string | `"k8scsi/csi-provisioner"` | CSI provisioner image pull repository |
| csiController.provisioner.image.tag | string | `"v2.1.0"` | CSI provisioner image tag |
| csiController.provisioner.name | string | `"csi-provisioner"` | CSI provisioner container name |
| csiController.resizer.image.pullPolicy | string | `"IfNotPresent"` | CSI resizer image pull policy  |
| csiController.resizer.image.registry | string | `"k8s.gcr.io/"` | CSI resizer image registry |
| csiController.resizer.image.repository | string | `"k8scsi/csi-resizer"` |  CSI resizer image repository|
| csiController.resizer.image.tag | string | `"v1.1.0"` | CSI resizer image tag |
| csiController.resizer.name | string | `"csi-resizer"` | CSI resizer container name |
| csiController.resources | object | `{}` | CSI controller container resources |
| csiController.securityContext | object | `{}` | CSI controller security context |
| csiController.tolerations | list | `[]` | CSI controller pod tolerations |
| csiNode.annotations | object | `{}` | CSI Node annotations |
| csiNode.componentName | string | `"openebs-jiva-csi-node"` | CSI Node component name |
| csiNode.driverRegistrar.image.pullPolicy | string | `"IfNotPresent"` | CSI Node driver registrar image pull policy|
| csiNode.driverRegistrar.image.registry | string | `"k8s.gcr.io/"` | CSI Node driver registrar image registry |
| csiNode.driverRegistrar.image.repository | string | `"k8scsi/csi-node-driver-registrar"` | CSI Node driver registrar image repository |
| csiNode.driverRegistrar.image.tag | string | `"v2.0.1"` |  CSI Node driver registrar image tag|
| csiNode.driverRegistrar.name | string | `"csi-node-driver-registrar"` | CSI Node driver registrar container name |
| csiNode.kubeletDir | string | `"/var/lib/kubelet/"` | Kubelet root dir |
| csiNode.labels | object | `{}` | CSI Node pod labels |
| csiNode.nodeSelector | object | `{}` |   CSI Node pod nodeSelector |
| csiNode.podAnnotations | object | `{}` | CSI Node pod annotations |
| csiNode.resources | object | `{}` | CSI Node pod resources |
| csiNode.securityContext | object | `{}` | CSI Node pod security context |
| csiNode.tolerations | list | `[]` | CSI Node pod tolerations |
| csiNode.updateStrategy.type | string | `"RollingUpdate"` | CSI Node daemonset update strategy |
| csiNode.livenessprobe.image.pullPolicy | string | `"IfNotPresent"` | CSI livenessprobe image pull policy |
| csiNode.livenessprobe.image.registry | string | `"k8s.gcr.io/"` |  CSI livenessprobe image registry |
| csiNode.livenessprobe.image.repository | string | `"k8scsi/livenessprobe"` |  CSI livenessprobe image repo |
| csiNode.livenessprobe.image.tag | string | `"v2.2.0"` | CSI livenessprobe image tag |
| csiNode.livenessprobe.name | string | `"liveness-probe"` |  CSI livenessprobe container name|
| jivaOperator.annotations | object | `{}` | Jiva operator annotations |
| jivaOperator.componentName | string | `"jiva-operator"` | Jiva operator component name |
| jivaOperator.image.pullPolicy | string | `"IfNotPresent"` | Jiva operator image pull policy |
| jivaOperator.image.registry | string | `nil` | Jiva operator image registry |
| jivaOperator.image.repository | string | `"openebs/jiva-operator"` | Jiva operator image repository |
| jivaOperator.image.tag | string | `"2.6.0"` |  Jiva operator image tag |
| jivaOperator.nodeSelector | object | `{}` |  Jiva operator pod nodeSelector|
| jivaOperator.podAnnotations | object | `{}` | Jiva operator pod annotations |
| jivaOperator.resources | object | `{}` | Jiva operator pod resources |
| jivaOperator.securityContext | object | `{}` | Jiva operator security context |
| jivaOperator.tolerations | list | `[]` | Jiva operator pod tolerations |
| jivaCSIPlugin.image.pullPolicy | string | `"IfNotPresent"` | Jiva CSI driver image pull policy |
| jivaCSIPlugin.image.registry | string | `nil` | Jiva CSI driver image registry |
| jivaCSIPlugin.image.repository | string | `"openebs/jiva-csi"` |  Jiva CSI driver image repository |
| jivaCSIPlugin.image.tag | string | `"2.6.0"` | Jiva CSI driver image tag |
| jivaCSIPlugin.name | string | `"jiva-csi-plugin"` | Jiva CSI driver container name |
| jivaCSIPlugin.remount | string | `"true"` | Jiva CSI driver remount feature, enabled by default |


Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```bash
helm install  <your-relase-name> -f values.yaml  openebs-jiva/jiva
```

> **Tip**: You can use the default [values.yaml](values.yaml)
