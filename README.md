# Go Dcp [![Go Reference](https://pkg.go.dev/badge/github.com/Trendyol/go-dcp.svg)](https://pkg.go.dev/github.com/Trendyol/go-dcp) [![Go Report Card](https://goreportcard.com/badge/github.com/Trendyol/go-dcp)](https://goreportcard.com/report/github.com/Trendyol/go-dcp)

This repository contains go implementation of a Couchbase Database Change Protocol (DCP) client.

### Contents

* [Why?](#why)
* [Usage](#usage)
* [Configuration](#configuration)
* [Examples](#examples)

### Why?

+ Our main goal is to build a dcp client for faster and stateful systems. We're already using this repository in below
  implementations:
    + [Elastic Connector](https://github.com/Trendyol/go-dcp-elasticsearch)
    + [Kafka Connector](https://github.com/Trendyol/go-dcp-kafka)
    + [Couchbase Connector](https://github.com/Trendyol/go-dcp-couchbase)

### Example

```go
package main

import (
  "github.com/Trendyol/go-dcp"
  "github.com/Trendyol/go-dcp/logger"
  "github.com/Trendyol/go-dcp/models"
)

func listener(ctx *models.ListenerContext) {
  switch event := ctx.Event.(type) {
  case models.DcpMutation:
    logger.Log.Info(
      "mutated(vb=%v,eventTime=%v) | id: %v, value: %v | isCreated: %v",
      event.VbID, event.EventTime, string(event.Key), string(event.Value), event.IsCreated(),
    )
  case models.DcpDeletion:
    logger.Log.Info(
      "deleted(vb=%v,eventTime=%v) | id: %v",
      event.VbID, event.EventTime, string(event.Key),
    )
  case models.DcpExpiration:
    logger.Log.Info(
      "expired(vb=%v,eventTime=%v) | id: %v",
      event.VbID, event.EventTime, string(event.Key),
    )
  }

  ctx.Ack()
}

func main() {
  connector, err := dcp.NewDcp("config.yml", listener)
  if err != nil {
    panic(err)
  }

  defer connector.Close()

  connector.Start()
}
```

### Usage

```
$ go get github.com/Trendyol/go-dcp

```

### Configuration

| Variable                                 |       Type        | Required |  Default   | Description                                                                                                             |
|------------------------------------------|:-----------------:|:--------:|:----------:|-------------------------------------------------------------------------------------------------------------------------|
| `hosts`                                  |     []string      |   yes    |     -      | Couchbase host like `localhost:8091`.                                                                                   |
| `username`                               |      string       |   yes    |     -      | Couchbase username.                                                                                                     |
| `password`                               |      string       |   yes    |     -      | Couchbase password.                                                                                                     |
| `bucketName`                             |      string       |   yes    |     -      | Couchbase DCP bucket.                                                                                                   |
| `dcp.group.name`                         |      string       |   yes    |            | DCP group name for vbuckets.                                                                                            |
| `scopeName`                              |      string       |    no    |  _default  | Couchbase scope name.                                                                                                   |
| `collectionNames`                        |     []string      |    no    |  _default  | Couchbase collection names.                                                                                             |
| `connectionBufferSize`                   |       uint        |    no    |  20971520  | [gocbcore](github.com/couchbase/gocbcore) library buffer size. `20mb` is default. Check this if you get OOM Killed.     |
| `connectionTimeout`                      |   time.Duration   |    no    |     5s     | Couchbase connection timeout.                                                                                           |
| `secureConnection`                       |       bool        |    no    |   false    | Enable TLS connection of Couchbase.                                                                                     |
| `rootCAPath`                             |      string       |    no    |  *not set  | if `secureConnection` set `true` this field is required.                                                                |
| `debug`                                  |       bool        |    no    |   false    | For debugging purpose.                                                                                                  |
| `dcp.bufferSize`                         |        int        |    no    |  16777216  | Go DCP listener pre-allocated buffer size. `16mb` is default. Check this if you get OOM Killed.                         |
| `dcp.connectionBufferSize`               |       uint        |    no    |  20971520  | [gocbcore](github.com/couchbase/gocbcore) library buffer size. `20mb` is default. Check this if you get OOM Killed.     |
| `dcp.connectionTimeout`                  |   time.Duration   |    no    |     5s     | DCP connection timeout.                                                                                                 |
| `dcp.listener.bufferSize`                |       uint        |    no    |    1000    | Go DCP listener buffered channel size.                                                                                  |
| `dcp.group.membership.type`              |      string       |    no    |            | DCP membership types. `couchbase`, `kubernetesHa`, `kubernetesStatefulSet` or `static`. Check examples for details.     |
| `dcp.group.membership.memberNumber`      |        int        |    no    |     1      | Set this if membership is `static`. Other methods will ignore this field.                                               |
| `dcp.group.membership.totalMembers`      |        int        |    no    |     1      | Set this if membership is `static` or `kubernetesStatefulSet`. Other methods will ignore this field.                    |
| `dcp.group.membership.rebalanceDelay`    |   time.Duration   |    no    |    20s     | Works for autonomous mode.                                                                                              |
| `leaderElection.enabled`                 |       bool        |    no    |   false    | Set this true for memberships  `kubernetesHa`.                                                                          |
| `leaderElection.type`                    |      string       |    no    | kubernetes | Leader Election types. `kubernetes`                                                                                     |
| `leaderElection.config`                  | map[string]string |    no    |  *not set  | Set lease key-values like `leaseLockName`,`leaseLockNamespace`.                                                         |
| `leaderElection.rpc.port`                |        int        |    no    |    8081    | This field is usable for `kubernetesStatefulSet` membership.                                                            |
| `checkpoint.type`                        |      string       |    no    |    auto    | Set checkpoint type `auto` or `manual`.                                                                                 |
| `checkpoint.autoReset`                   |      string       |    no    |  earliest  | Set checkpoint start point to `earliest` or `latest`.                                                                   |
| `checkpoint.interval`                    |   time.Duration   |    no    |    20s     | Checkpoint checking interval.                                                                                           |
| `checkpoint.timeout`                     |   time.Duration   |    no    |    60s     | Checkpoint checking timeout.                                                                                            |
| `healthCheck.disabled`                   |       bool        |    no    |   false    | Disable Couchbase connection health check.                                                                              |
| `healthCheck.interval`                   |   time.Duration   |    no    |    20s     | Couchbase connection health checking interval duration.                                                                 |
| `healthCheck.timeout`                    |   time.Duration   |    no    |     5s     | Couchbase connection health checking timeout duration.                                                                  |
| `rollbackMitigation.disabled`            |       bool        |    no    |   false    | Disable reprocessing for roll-backed Vbucket offsets.                                                                   |
| `rollbackMitigation.interval`            |   time.Duration   |    no    |   500ms    | Persisted sequence numbers polling interval.                                                                            |
| `rollbackMitigation.configWatchInterval` |   time.Duration   |    no    |     2s     | Cluster config changes listener interval.                                                                               |
| `metadata.type`                          |      string       |    no    | couchbase  | Metadata storing types.  `file` or `couchbase`.                                                                         |
| `metadata.readOnly`                      |       bool        |    no    |   false    | Set this for debugging state purposes.                                                                                  |
| `metadata.config`                        | map[string]string |    no    |  *not set  | Set key-values of config. `bucket`,`scope`,`collection`,`connectionBufferSize`,`connectionTimeout` for `couchbase` type |
| `api.disabled`                           |       bool        |    no    |   false    | Disable metric endpoints                                                                                                |
| `api.port`                               |        int        |    no    |    8080    | Set API port                                                                                                            |
| `metric.path`                            |      string       |    no    |  /metrics  | Set metric endpoint path.                                                                                               |
| `metric.averageWindowSec`                |      float64      |    no    |    10.0    | Set metric window range.                                                                                                |
| `logging.level`                          |      string       |    no    |    info    | Set logging level.                                                                                                      |

### Environment Variables

These environment variables will **overwrite** the corresponding configs.

| Variable                                    | Type |       Corresponding Config        |                         Description                          |
|---------------------------------------------|:----:|:---------------------------------:|:------------------------------------------------------------:|
| `GO_DCP__DCP_GROUP_MEMBERSHIP_MEMBERNUMBER` | int  | dcp.group.membership.memberNumber | To be able to prevent making deployment to scale up or down. 
| `GO_DCP__DCP_GROUP_MEMBERSHIP_TOTALMEMBERS` | int  | dcp.group.membership.totalMembers | To be able to prevent making deployment to scale up or down. 

### Monitoring

The client offers an API that handles different endpoints and expose several metrics.

### API

| Endpoint                | Description                                                                              | Debug Mode |
|-------------------------|------------------------------------------------------------------------------------------|------------|
| `GET /status`           | Returns a 200 OK status if the client is able to ping the couchbase server successfully. |            |
| `GET /rebalance`        | Triggers a rebalance operation for the vBuckets.                                         |            |
| `GET /states/offset`    | Returns the current offsets for each vBucket.                                            | x          | 
| `GET /states/followers` | Returns the list of follower clients if service discovery enabled                        | x          |
| `GET /debug/pprof/*`    | [Fiber Pprof](https://docs.gofiber.io/api/middleware/pprof/)                             | x          |

The Client collects relevant metrics and makes them available at /metrics endpoint.
In case you haven't configured a metric.path, the metrics will be exposed at the /metrics.

You can adjust the average window time for the metrics by specifying the value of metric.averageWindowSec.

### Exposed metrics

| Metric Name                          | Description                                                                           | Labels                  | Value Type |
|--------------------------------------|---------------------------------------------------------------------------------------|-------------------------|------------|
| cbgo_mutation_total                  | The total number of mutations on a specific vBucket                                   | vbId: ID of the vBucket | Counter    |
| cbgo_deletion_total                  | The total number of deletions on a specific vBucket                                   | vbId: ID of the vBucket | Counter    |
| cbgo_expiration_total                | The total number of expirations on a specific vBucket                                 | vbId: ID of the vBucket | Counter    |
| cbgo_seq_no_current                  | The current sequence number on a specific vBucket                                     | vbId: ID of the vBucket | Gauge      |
| cbgo_start_seq_no_current            | The starting sequence number on a specific vBucket                                    | vbId: ID of the vBucket | Gauge      |
| cbgo_end_seq_no_current              | The ending sequence number on a specific vBucket                                      | vbId: ID of the vBucket | Gauge      |
| cbgo_lag_current                     | The current lag on a specific vBucket                                                 | vbId: ID of the vBucket | Gauge      |
| cbgo_process_latency_ms_current      | The average process latency in milliseconds for the last metric.averageWindowSec      | N/A                     | Gauge      |
| cbgo_dcp_latency_ms_current          | The latest consumed dcp message latency in milliseconds                               | N/A                     | Counter    |
| cbgo_rebalance_current               | The number of total rebalance                                                         | N/A                     | Gauge      |
| cbgo_total_members_current           | The total number of members in the cluster                                            | N/A                     | Gauge      |
| cbgo_member_number_current           | The number of the current member                                                      | N/A                     | Gauge      |
| cbgo_membership_type_current         | The type of membership of the current member                                          | Membership type         | Gauge      |
| cbgo_offset_write_current            | The average number of the offset write for the last metric.averageWindowSec           | N/A                     | Gauge      |
| cbgo_offset_write_latency_ms_current | The average offset write latency in milliseconds for the last metric.averageWindowSec | N/A                     | Gauge      |

### Examples

- [example with couchbase membership](example/main.go)
- [couchbase membership config](example/config.yml) - thanks to [@onursak](https://github.com/onursak)
- [kubernetesStatefulSet membership config](example/config_k8s_stateful_set.yml)
- [kubernetesHa membership config](example/config_k8s_leader_election.yml)
- [static membership config](example/config_static.yml)
