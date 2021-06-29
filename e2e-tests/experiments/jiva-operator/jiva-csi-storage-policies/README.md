# e2eBook to deploy the OpenEBS Jiva CSI Driver.

## Description
   - This e2eBook is capable of setting up OpenEBS Jiva CSI Driver and create a storageclass.

   - This test constitutes the below files. 

### run_e2e_test.yml
   - This includes the e2e job which triggers the test execution. The pod includes several environmental variables such as 
        - PROVIDER_STORAGE_CLASS : The name of storageclass to create using jiva csi provisioner
        - REPLICA_COUNT : The number of volume replicas to be created.
        - REPLICA_SC : The name of storage class used to create volume replicas.

### jiva-csi-sc.j2
   - The storage class template which has to be populated with the given variables

### test_vars.yml
   - This test_vars file has the list of test specific variables used in e2eBook

### test.yml
   - test.yml is the playbook where the test logic is built to deploy OpenEBS Jiva CSI Driver and create stoarge class.
