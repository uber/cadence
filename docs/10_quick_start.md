# Quick Start

## Install Cadence Service Locally

### Install docker

Docker installation instructions: https://docs.docker.com/engine/installation/

### Run Cadence Server Using Docker Compose

Download Cadence docker-compose file:
```bash
quick_start: curl -O https://raw.githubusercontent.com/uber/cadence/master/docker/docker-compose.yml
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   675  100   675    0     0    958      0 --:--:-- --:--:-- --:--:--   958
quick_start: ls
docker-compose.yml
```
Start Cadence Service:
```bash
quick_start: docker-compose up
Creating network "quick_start_default" with the default driver
Pulling cadence (ubercadence/server:0.5.8)...
0.5.8: Pulling from ubercadence/server
db0035920883: Pull complete
82eed7f2d38e: Pull complete
f81e11a89e41: Pull complete
ae3538b1ae1c: Pull complete
23ddfb58e314: Pull complete
52a6bbeb81b5: Pull complete
a72c7949d8ac: Pull complete
1c3b1d477195: Pull complete
3312d4123248: Pull complete
5bbc95a38c5f: Pull complete
29176d1ce1ca: Pull complete
27ec3755f89c: Pull complete
0a5d2a29a5e5: Pull complete
Creating quick_start_statsd_1    ... done
Creating quick_start_cassandra_1 ... done
Creating quick_start_cadence_1   ... done
Creating quick_start_cadence-web_1 ... done
Attaching to quick_start_cassandra_1, quick_start_statsd_1, quick_start_cadence_1, quick_start_cadence-web_1
statsd_1       | *** Running /etc/my_init.d/00_regen_ssh_host_keys.sh...
statsd_1       | *** Running /etc/my_init.d/01_conf_init.sh...
cadence_1      | + CADENCE_HOME=/cadence
cadence_1      | + DB=cassandra
...
...
...
cadence_1      | {"level":"info","ts":"2019-06-06T15:26:38.199Z","msg":"Get dynamic config","name":"matching.longPollExpirationInterval","value":"1m0s","default-value":"1m0s","logging-call-at":"config.go:57"}
cadence_1      | {"level":"info","ts":"2019-06-06T15:26:38.199Z","msg":"Get dynamic config","name":"matching.updateAckInterval","value":"1m0s","default-value":"1m0s","logging-call-at":"config.go:57"}
cadence_1      | {"level":"info","ts":"2019-06-06T15:26:38.199Z","msg":"Get dynamic config","name":"matching.idleTasklistCheckInterval","value":"5m0s","default-value":"5m0s","logging-call-at":"config.go:57"}
cadence_1      | {"level":"info","ts":"2019-06-06T15:26:38.765Z","msg":"message is empty","service":"cadence-matching","component":"matching-engine","lifecycle":"Starting","wf-task-list-name":"cadence-archival-tl","wf-task-list-type":0,"logging-call-at":"matchingEngine.go:185"}
cadence_1      | {"level":"info","ts":"2019-06-06T15:26:38.775Z","msg":"message is empty","service":"cadence-matching","component":"matching-engine","lifecycle":"Started","wf-task-list-name":"cadence-archival-tl","wf-task-list-type":0,"logging-call-at":"matchingEngine.go:199"}
cadence_1      | {"level":"info","ts":"2019-06-06T15:26:38.891Z","msg":"message is empty","service":"cadence-matching","component":"matching-engine","lifecycle":"Starting","wf-task-list-name":"51f3b9fdfa7d:7feebe1f-95b2-44b8-8633-5ba7f4113508","wf-task-list-type":0,"logging-call-at":"matchingEngine.go:185"}
cadence_1      | {"level":"info","ts":"2019-06-06T15:26:38.900Z","msg":"message is empty","service":"cadence-matching","component":"matching-engine","lifecycle":"Started","wf-task-list-name":"51f3b9fdfa7d:7feebe1f-95b2-44b8-8633-5ba7f4113508","wf-task-list-type":0,"logging-call-at":"matchingEngine.go:199"}
cadence_1      | {"level":"info","ts":"2019-06-06T15:26:52.282Z","msg":"Get dynamic config","name":"history.shardUpdateMinInterval","value":"5m0s","default-value":"5m0s","logging-call-at":"config.go:57"}
cadence_1      | {"level":"info","ts":"2019-06-06T15:26:52.282Z","msg":"Get dynamic config","name":"history.emitShardDiffLog","value":"false","default-value":"false","logging-call-at":"config.go:57"}
cadence_1      | {"level":"info","ts":"2019-06-06T15:27:24.903Z","msg":"Get dynamic config","name":"history.transferProcessorCompleteTransferFailureRetryCount","value":"10","default-value":"10","logging-call-at":"config.go:57"}
cadence_1      | {"level":"info","ts":"2019-06-06T15:27:24.905Z","msg":"Get dynamic config","name":"history.timerProcessorCompleteTimerFailureRetryCount","value":"10","default-value":"10","logging-call-at":"config.go:57"}
```
### Register Domain Using CLI
From a different console window:
```bash
quick_start: docker run --network=host --rm ubercadence/cli:master --do test-domain domain register -rd 1
Unable to find image 'ubercadence/cli:master' locally
master: Pulling from ubercadence/cli
22dc81ace0ea: Pull complete
1a8b3c87dba3: Pull complete
91390a1c435a: Pull complete
07844b14977e: Pull complete
b78396653dae: Pull complete
5259e0c8568e: Pull complete
be8b5313e7cd: Pull complete
da2cfe74be81: Pull complete
5320bde81c0c: Pull complete
Digest: sha256:f5e5e708347909c8d3f74c47878b201d91606994394e94eaede9a80e3b9f077b
Status: Downloaded newer image for ubercadence/cli:master
Domain test-domain successfully registered.
quick_start:
```
Check that the domain is indeed registered:
```bash
quick_start: docker run --network=host --rm ubercadence/cli:master --do test-domain domain describe
Name: test-domain
Description:
OwnerEmail:
DomainData: map[]
Status: REGISTERED
RetentionInDays: 1
EmitMetrics: false
ActiveClusterName: active
Clusters: active
ArchivalStatus: DISABLED
Bad binaries to reset:
+-----------------+----------+------------+--------+
| BINARY CHECKSUM | OPERATOR | START TIME | REASON |
+-----------------+----------+------------+--------+
+-----------------+----------+------------+--------+
quick_start:
```
## Implement Hello World Java Workflow

### Include Cadence Java Client Dependency

Go to [Maven Repository Uber Cadence Java Client Page](https://mvnrepository.com/artifact/com.uber.cadence/cadence-client)
and find the latest version of the library. Include it as a dependency into your Java project. For example if you
are using Gradle the dependency looks like:
```
compile group: 'com.uber.cadence', name: 'cadence-client', version: '<latest_version>'
```
Make sure that the following file compiles