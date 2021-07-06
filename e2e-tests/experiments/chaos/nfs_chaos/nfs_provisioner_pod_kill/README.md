## Experiment Metadata

| Type  | Description                  | Storage | Applications  | K8s Platform |
| ----- | ---------------------------- | ------- | ------------- | ------------ |
| Chaos | Fail the nfs application pod | OpenEBS | NFS           | Any          |

## Entry-Criteria

- NFS provisoner services are accessible & pods are healthy
- Application pods that are using NFS provisioner are healthy
- Application writes are successful 

## Exit-Criteria

- Applications that are using NFS provisioner are accessible & pods are healthy
- Data written prior to chaos is successfully retrieved/read

## Associated Utils 

- `chaoslib/pumba/pod_failure_by_sigkill.yaml` when the container runtime is pumba
- `chaoslib/containerd_chaos/crictl-chaos.yml` when the container runtime is containerd
- `chaoslib/crio_chaos/crio-crictl-chaos.yml` when the container runtime is cro 
   

## e2ebook Environment Variables

### Application

| Parameter         | Description                                                   |
| -------------     | ------------------------------------------------------------  |
| APP_NAMESPACE     | Namespace in which application pods are deployed              |
| APP_LABEL         | Unique Labels in `key=value` format of application deployment |
| CONTAINER_RUNTIME | container runtime to induce the chaos                         |
| NFS_NAMESPACE     | Namespace in which nfs provisioner is deployed                |
| NFS_LABEL         | Unique Labels in `key=value` format of NFS deployment         |

### Health Checks 

| Parameter              | Description                                                           |
| ---------------------- | --------------------------------------------------------------------- |
| LIVENESS_APP_NAMESPACE | Namespace in which external liveness pods are deployed, if any        |
| LIVENESS_APP_LABEL     | Unique Labels in `key=value` format for external liveness pod, if any |

## Procedure

This experiment kills the application container and verifies if the container is scheduled back and the data is intact. Based on CRI used, uses the relevant util to kill the application container.

After injecting the chaos into the component specified via environmental variable, e2e experiment observes the behaviour of corresponding Application.

