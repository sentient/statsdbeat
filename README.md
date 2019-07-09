Statsdbeat
==========

Using the [beat framework](https://www.elastic.co/products/beats) to send [statsd formatted](https://github.com/b/statsd_spec) messages to ElasticSearch. 

# Download

[1.0 release](https://github.com/sentient/statsdbeat/releases)

# What we do 

We listen for UDP pacakges. And forwards them as `beat.Event` into Elastic Search at the index `statsdbeat-<agent-version>-<yyyy-mm-dd>`

Support
+ Following statsd types are supported
  
    | Type          | Example                                           |
    | ------------- | --------------------------------------------------|
    | Counters      | `platform-insights.test.counter.tick:1|c`         |
    | Gauge         | `platform-insights.test.gauge.num_goroutine:1|g`  |
    | Histogram     | `platform-insights.test.histogram.my_histo:17|h`  |
    | Timing        | `platform-insights.test.timing.ping:10|ms`        |

+ Tags (in InfluxDB notation format `counter,tagName=tagValue,anotherTag=withAnotherValue:1|c`)
+ Multi-Metric Packets


# Configuration

```
statsdbeat:
  statsdserver: ":8125"    # where should we listen for the UDP messages. Typically your localhost on port 8125
  period: 5s               # interval period the events (if any) are send to the output  
```

## Spooling

_Spooling to disk is currently a beta feature. Use with care._

You can configure in `statsbeat.yml` the spooling to disk

```
output.elasticsearch:
  hosts: ["https://vpc-<your-name>.<aws-region>.es.amazonaws.com:443"]

queue:
  spool:
    file:
      flush.timeout:1s 

```

# What we don't do (yet)

+ No pre-aggreation (roll-ups) of data before sending to Elastic Search

+ No Sets

+ No Sampling
  
+ No resend Gauge information ()

+ We don't compute percentile aggregations. Elastic Search has this already [build in](https://www.elastic.co/guide/en/elasticsearch/reference/current/search-aggregations-metrics-percentile-aggregation.html)

 
# Development

* [Getting Started](README-development.md)



# References:


+ [Etsy Statsd metric types](https://github.com/etsy/statsd/blob/master/docs/metric_types.md)
  - [Elastic Search backend](https://github.com/markkimsal/statsd-elasticsearch-backend), 
       (NPM module for Etsty Statsd to output to ElasticSearch )
