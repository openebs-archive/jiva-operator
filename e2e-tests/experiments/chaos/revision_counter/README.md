## Experiment Metadata

| Type  | Description                                                  | Storage | Applications | K8s Platform |
| ----- | ------------------------------------------------------------ | ------- | ------------ | ------------ |
| Chaos | Ensure that revision counter value does not change while dumping the data using jiva & ensure that all jiva replica consumes same amount of storage. | OpenEBS | Any          | Any          |

## Entry-Criteria

- Application should be created using jiva volume and should be up.
- Jiva volume should be created with three replica.

## Exit-Criteria

- Application and all three jiva replica should be up.
- All three jiva replica should consume same amount of storage.

## Procedure

- Obtain the nodes on which jiva replica pod is scheduled.

- Continuously delete jiva replica pod of one same node while dumping some data.

- Verify revision counter value in other jiva replica pod while deleting one jiva replica pod.

- After dumping the data verify all the three jiva replica is consuming same amount of storage.
  
## Environment Variables

| Parameters        | Description                                      |
| ----------------- | ------------------------------------------------ |
| APP_LABEL         | Namespace where OpenEBS components are deployed. |
| APP_NAMESPACE     | Namespace where application is deployed.         |
| BLOCK_COUNT       | Block Count to dump the data using dd.           |
| BLOCK_SIZE        | Block Size to dump the data using dd.            |
| FILE_NAME         | File Name to dump the data using dd.             |
| MOUNT_PATH        | Mount path where volume is mounted.              |