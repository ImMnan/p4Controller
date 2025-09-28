

## Arch - 1
1. Values added to Helm 
    - Values for replicas will be configured at install and added to manifest_in_queue configmap for the controller to pick.
    - The manifests will include the master server manifest and if deployed - replica server manifest

2. Helm install command
    a. the master server will be deployed first, will just miss the `SERVER_INIT` env var, should be true for commit server. 
    b. The controller will be deployed along side p4d
3. controller is designed to check the master spec
    - Update the manifest to manifest deployed cm
    - Switch the `SERVER_INIT` to true if it was false for some configuration like unpacking checkpoint
    - Update the spec to controller config
4. Done. 

5. In this setup the replicas will be deployed only through Helm chart upgrade or config map update to add in-queue manifests. Though controller is responsible for deploying the actual replica sts, it will not be doing it automatically, unless the helm chart has been upgraded to do so. 


## Arch - 2
1. Values added to helm - only for the master/commit server 
    - Values for replicas will be added as a templete for different types of servers
    - The minfests will include the master server manifest and if deployed - replica server manifest

2. Helm install command
    a. the master server will be deployed first, will just miss the `SERVER_INIT` env var, should be true for commit server. 
    b. The controller will be deployed along side p4d

3. controller is designed to check the master spec
    - Update the manifest to manifest deployed cm
    - Switch the `SERVER_INIT` to true if it was false for some configuration like unpacking checkpoint
    - Update the spec to controller config
4. Done.

5. In this setup, the replicas will be deployed through the controller, after checking the new entry in p4 servers list. 
6. P4 controller will check the p4 servers list every set interval, if it finds a new entry, it will be deployed, by editing the replicaset template in configmap. 



## Replica Initialisation process: 

### CUSTOMER STEPs: 

Firstly, the customer is expected to initiate a replica creation by: 

```sh
p4 -u super server fwd-1667
```
Updating the spec configuration: 

```sh
ServerID:    fwd-1667
Name:        fwd-1667
Type:        server
Services:    forwarding-replica
Address:     forward:1667
DistributedConfig:
    db.replication=readonly
    lbr.replication=readonly
    lbr.autocompress=1
    startup.1=pull -i 1
    startup.2=pull -u -i 1
    startup.3=pull -u -i 1
    P4TARGET=london:1666
    serviceUser=service
    monitor=1 # optional but required if using the 'p4 monitor show' command
    journalPrefix=/p4/journals/fw-replica # recommended
    P4TICKETS=/p4/.p4tickets # recommended
    P4LOG=/p4/logs/fw-replica.log # recommended
Description:
    Forwarding replica pointing to london:1666
```


Process that p4controller needs to manage on the **target/master server**: [Sourced from](https://help.perforce.com/helix-core/server-apps/p4sag/current/Content/P4SAG/replication-configure-forwarding-master.html)

1. Create the service user for the replication service. For example:
```sh
p4 -u super user -f service
```
The default user specification opens in your default editor. To make this user be of type service, add the following line: `Type: service`

2. Use the p4 group command to create a group for your service users and set the value of the timeout field. To avoid service users being logged out, consider using unlimited as the Timeout value. See Tickets and timeouts for service users.

3. Create a checkpoint of the target server. We may need to make sure the checkpoint, journal and version files are done on the same mount. 
```sh
p4 -u super admin checkpoint
```
> Do we need to create a service account for each type of replica? Or one is enough for all types of servers?

Process that p4controller needs to manage on the **replica server**: [Sourced from](https://help.perforce.com/helix-core/server-apps/p4sag/current/Content/P4SAG/replication-configure-forwarding-replica.html)

1. Restore from that checkpoint on the machine that the forwarding replica will run on. 
```sh
p4d -jr checkpoint_file
```
Considering that the checkpoint path is mounted on the replica server. 

2. Start the P4 server on the forwarding replica using the P4PORT value of forward:1667 and both the -n and -d options:
```sh
p4d -p forward:1667 -n -d
```

3. Set the serverid to fwd-1667 for the forwarding replica:
```sh
p4 -u super -p forward:1667 serverid fwd-1667
```

4. Confirm that the serverid is correctly set at the server address of forward:1667.
```sh
p4 -u super -p forward:1667 serverid
```
The output should be: `Server ID: fwd-1667`

5. Log the service user into the target server using the location of the tickets file specified in the Spec configuration section of Configure the target server for the forwarding replica: 
```sh
p4 -u super -p london:1666 -E P4TICKETS=/p4/.p4tickets  login service
```
> Note: london:1666 is the target server

6. On the forwarding replica, stop the server:
```sh
p4 -u super -p forward:1667 admin stop
```

7. Restart the server on the forwarding replica:
```sh
$ p4d -p forward:1667 -d
```

8. Confirm that the p4 pull commands specified in the fwd-1667 startup.N configurations are running:
```sh
p4 -u super -p forward:1667 monitor show -a
```

9. Confirm that the forwarding replica is replicating.
```sh
p4 -u super -p forward:1667 pull -l -j
```


## IDLE OPERATIONS


## CONTROLLER DESIGN


> [CH 1] LOOP START
1. Check the existing deployed resources (managed by p4controller) 
2. Verify the existing deployed with resources in the controller-cm [function configmap_manager]
3. If mismatch, for missing resources - 
    a. Deploy the missing resources  [function deployment_manager] (2 functions works separately based on init value - true or false)
    b. Update the deployed configmap [function configmap_manager]
4. If mismatch, for extra resources - 
    a. Remove them [function cleanup_manager - part of deployment_manager]
5. Read CH2 buffer for new servers to be deployed or existing servers to be removed 
    a. If new server found, update the configmap with new manifest [function configmap_manager]
    b. If server to be removed found, update the configmap to remove the manifest [function configmap_manager]
    c. free the channel. 
> [CH 1] LOOP END (loop again)



> [CH 2] LOOP START
1. Check the p4 servers list for new entries [function p4_client]
    a. If new entries are found, send the new server details to CH1 buffer
    b. Wait for channel to be free
> [CH 2] LOOP END (loop again)

