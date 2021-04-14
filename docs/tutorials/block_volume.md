### Creating a new raw block PVC

`volumeMode: Block` need to be specified in the PVC spec in order to create and use a raw block volume.

1. Raw Block volume:

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: block-claim
spec:
  volumeMode: Block
  storageClassName: openebs-jiva-csi-sc
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi

```
### Using a raw Block PVC in POD

A devicePath needs to be specified for the block device inside the
Mysql container instead of the mountPath used for file system.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    run: mysql
  name: mysql
spec:
  replicas: 1
  selector:
    matchLabels:
      run: mysql
  strategy: {}
  template:
    metadata:
      labels:
        run: mysql
    spec:
      containers:
      - image: mysql
        imagePullPolicy: IfNotPresent
        name: mysql
        volumeDevices:
          - name: my-db-data
            devicePath: /dev/block
        env:
          - name: MYSQL_ROOT_PASSWORD
            value: test
      volumes:
        - name: my-db-data
          persistentVolumeClaim:
            claimName: block-claim
```
