<!--
 Copyright (c) 2019 Intel Corporation

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
-->

# ![Swan diagram](/images/swan-logo-48.png) Swan
---
## KRICO Experiment

This experiment uses KRICO (Komponent Rekomendacji dla Inteligentnych Chmur Obliczeniowych)
for workload prediction and classification.

It consists of three sub experiments:
- Metric Gathering which provides data for KRICO neural network.
- Classification Experiment which runs workloads, gather metrics from them and in the end do classification of these workloads.
- Prediction Experiment which for passed parameters do prediction.

More information about KRICO (in Polish) [krico.gda.pl](http://krico.gda.pl/)

---
## Required software and services

To run the experiment you need following services:

- [OpenStack Rocky](https://www.openstack.org/software/rocky/)
- [Snap Telemetry](https://github.com/intelsdi-x/snap) installed on OpenStack hypervisor node
- [Cassandra Database](http://cassandra.apache.org/)
- KRICO Service

### KRICO installation
##### Requirements:
* Python 2.7
* [Pipenv](https://github.com/pypa/pipenv)

##### Installation and running
Edit existing config file (```experiments/krico/config.yml```), put your information about database and api.

```bash
database:
    host: your host
    port: your port
    keyspace: krico
    replication_factor: 1

api:
    host: your ip
    port: 5000
```

Go to experiment folder. Install virtual environment and run KRICO service.

```bash
cd experiments/krico/
pipenv install
pipenv run python main.py -c config.yml
```
---
## Running the experiments


##### Metric gathering
```bash
krico-metric-gathering
```
Parameters available:

```
-aggressor_address string                                
        IP address of aggressor node. (default "127.0.0.0")

-ssh_key string                              
        SSH key path (default "~/.ssh/id_rsa")

-remote_ssh_key_path string                                         
        Key for user in from flag "remote_ssh_user" used for connecting to remote nodes.
        Default value is '$HOME/.ssh/id_rsa' (default "/home/vagrant/.ssh/id_rsa")      

-username string                  
        Username (default "cirros")

-snapteld_address http://%s:%s                                      
        Snapteld address in http://%s:%s format (default "http://127.0.0.1:8181")

-cassandra_address string                                                                     
        Address of Cassandra DB endpoint for Metadata and Snap Publishers. (default "127.0.0.1")

-remote_ssh_user string                                 
        Login used for connecting to remote nodes.        
        Default value is current user. (default "vagrant")

-experiment_load_duration duration           
        Load duration on HP task. (default 15s)

-experiment_peak_load 0                                                                                          
        Maximum load that will be generated on HP workload. If value is 0, then maximum possible load will be found by Swan.

-os_keypair_name string
        Openstack Keypair Name (default "swan")

-cassandra_keyspace_name string                        
        Keyspace used to store metadata. (default "swan")

-host_aggregate_id int                                                         
        ID of host aggregate which VM must be running in (default -1)

-memcached_listening_address string                 
        IP address of interface that Memcached will be listening on.
        It must be actual device address, not '0.0.0.0'. (default "127.0.0.1")

-mutilate_records int
        Number of memcached records to use (-r). (default 5000000)

-krico_api_address string                                        
        Ip address of KRICO API service. (default "localhost:5000")
```

##### Classification
```bash
krico-classification
```
Parameters available:

```
-aggressor_address string                                
        IP address of aggressor node. (default "127.0.0.0")

-ssh_key string                              
        SSH key path (default "~/.ssh/id_rsa")

-remote_ssh_key_path string                                         
        Key for user in from flag "remote_ssh_user" used for connecting to remote nodes.
        Default value is '$HOME/.ssh/id_rsa' (default "/home/vagrant/.ssh/id_rsa")      

-username string                  
        Username (default "cirros")

-snapteld_address http://%s:%s                                      
        Snapteld address in http://%s:%s format (default "http://127.0.0.1:8181")

-cassandra_address string                                                                     
        Address of Cassandra DB endpoint for Metadata and Snap Publishers. (default "127.0.0.1")

-remote_ssh_user string                                 
        Login used for connecting to remote nodes.        
        Default value is current user. (default "vagrant")

-experiment_load_duration duration           
        Load duration on HP task. (default 15s)

-experiment_peak_load 0                                                                                          
        Maximum load that will be generated on HP workload. If value is 0, then maximum possible load will be found by Swan.

-os_keypair_name string
        Openstack Keypair Name (default "swan")

-cassandra_keyspace_name string                        
        Keyspace used to store metadata. (default "swan")

-host_aggregate_id int                                                         
        ID of host aggregate which VM must be running in (default -1)

-ycsb_path string                                                   
        Path to YCSB binary file. (default "ycsb")    

-redis_sudo                   
        Run Redis server in sudo

-redis_isolate                             
        Run Redis server in new namespace pid

-krico_api_address string                                        
        Ip address of KRICO API service. (default "localhost:5000")
```


##### Prediction
```bash
krico-prediction
```
Parameters available:

```
-experiment_load_duration duration           
        Load duration on HP task. (default 15s)

-cassandra_keyspace_name string                        
        Keyspace used to store metadata. (default "swan")

-krico_api_address string                                        
        Ip address of KRICO API service. (default "localhost:5000")

-krico_prediction_category string
        Workload category                      

-krico_prediction_clients string                                                             
        Clients (default "0.0")                                                                

-krico_prediction_disk string                                                                
        Disk (default "0.0")                                                                   

-krico_prediction_image string                                                               
        Workload image (default "default")                                                     

-krico_prediction_memory string                                                              
        Memory  (default "0.0")                                                                

-krico_prediction_ratio string                                                               
        Ratio (default "0.0")                                                                  
```
