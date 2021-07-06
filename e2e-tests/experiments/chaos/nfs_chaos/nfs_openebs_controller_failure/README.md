## Experiment Metadata

| Type  | Description                                                  | Storage | K8s Platform |
| ----- | ------------------------------------------------------------ | ------- | ------------ |
| Chaos | Kill the Jiva controller container and check if gets created again | OpenEBS | Any          |

## Entry-Criteria

- NFS provisioner services are accessible.
- Application pods that are using NFS provisioner are healthy
- Application writes are successful 

## Exit-Criteria

- Application services are accessible & pods are healthy
- Data written prior to chaos is successfully retrieved/read
- Storage target pods are healthy

### Notes

- Typically used as a disruptive test, to cause loss of access to storage target by killing the containers.
- The container should be created again and it should be healthy.

## Associated Utils 

- `jiva_controller_pod_failure.yaml`

## e2e experiment Environment Variables

### Application

| Parameter        | Description                                                     |
| ---------------- | --------------------------------------------------------------- |
| APP_NAMESPACE    | Namespace in which application pods that are using nfs deployed |
| APP_LABEL        | Unique Labels in `key=value` format of application deployment   |
| NFS_NAMESPACE    | Namespace in which NFS provisioner pods are deployed            |
| NFS_LABEL        | Unique Labels in `key=value` format of NFS deployment           | 
| NFS_PVC          | Name of persistent volume claim used for NFS volume mounts      |
| TARGET_NAMESPACE | Namespace where OpenEBS is installed                            |

### Chaos 

| Parameter        | Description                                          |
| ---------------- | ---------------------------------------------------- |
| CHAOS_TYPE       | The type of chaos to be induced.                     |
| TARGET_CONTAINER | The container against which chaos has to be induced. |

### Procedure

This scenario validates the behaviour of application and OpenEBS persistent volumes in the amidst of chaos induced on OpenEBS data plane and control plane components.

After injecting the chaos into the component specified via environmental variable, e2e experiment observes the behaviour of corresponding OpenEBS PV and the application which consumes the volume.

