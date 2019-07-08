Notes
=====


Collect batches of statsd events. Keep them in memory. Mark every batch with a unique ID.
When closing event is received, flush the pending batches to disk.
On startup; find any pending batches and try to publish them first.


# Naming your buckets

[Source](https://matt.aimonetti.net/posts/2013/06/26/practical-guide-to-graphite-monitoring/)

Properly naming your metrics is critical to avoid conflicts, confusing data and potentially wrong interpretation later on. I like to organize metrics using the following schema:

```
1 <namespace>.<instrumented section>.<target (noun)>.<action (past tense verb)>
```
Example:
```
1 accounts.authentication.password.attempted
2 accounts.authentication.password.succeeded
3 accounts.authentication.password.failed
```
I use nouns to define the target and past tense verbs to define the action. This becomes a useful convention when you need to nest metrics. In the above example, let’s say I want to monitor the reasons for the failed password authentications. Here is how I would organize the extra stats:
```
1 accounts.authentication.password.failure.no_email_found
2 accounts.authentication.password.failure.password_check_failed
3 accounts.authentication.password.failure.password_reset_required
```
As you can see, I used failure instead of failed in the stat name. The main reason is to avoid conflicting data. failed is an action and already has a data series allocated, if I were to add nested data using failed, the data would be collected but the result would be confusing. The other reason is because when we will graph the data, we will often want to use a wildcard * to collect all nested data in a series.

Graphite wild card usage example on counters:

1 accounts.authentication.password.failure.*
This should give us the same value as accounts.authentication.password.failed, so really, we should just collect the more detailed version and get rid of accounts.authentication.password.failed.

Following this naming convention should really help your data stay clean and easy to manage.



# Notes 

[Source](https://discuss.elastic.co/t/flushing-after-yourself/129204/4)

- producer publishes single events
- pipeline returns ordered ACK to producer or global ACK handler.
    - this means, your ACK handler will be invoked in the same order you have published events. Using ACKCount, the count returned is the first N events you have published. No out of order signaling
- pipeline always ACKs events. From producer point of view, there is no event drop.
  => event drops, infinite retry, is configured and handled in the publisher pipeline.

  
  
# StatsD Datagram
StatsD clients encode metrics into simple, text-based, UDP datagrams. Though your client takes care of forming these datagrams, by exploring the format we can learn important information about features that the StatsD protocol supports.

A StatsD datagram, which contains a single metric, has the following format:

```
<bucket>:<value>|<type>|@<sample rate>
```

## Bucket
The bucket is an identifier for the metric. Metric datagrams with the same bucket and the same type are considered occurrences of the same event by the server. In the example above, we used “login.invocations” and “login.time” as our buckets. Note that periods can be used in buckets to group related metrics. Buckets are not predefined; a client can send a metric with any bucket at any time, and the server will handle it appropriately.

## Value
The value is a number that is associated with the metric. Values have different meanings depending on the metric’s type.

## Sample Rate
The sample rate is used to indicate to the server that the metric was down-sampled. The sampling rate is intended to reduce the number of metric datagrams sent to the StatsD server, since the server’s aggregations can get expensive. The sample rate determines what percentage of the metric points a client should send to the server. The server accounts for this sampling by dividing the values it receives by the sample rate. For example, if a metric has a sampling rate of 0.1, only 10 percent of the metrics will be sent by the client to the server. The server will then divide the values for these metrics by 0.1 (or multiply by 10) to get an estimate of the true value in the case of additive metrics, such as the login invocation count we used in the example above.

## Type
The type determines what sort of event the metric represents. There are several metric types:

### COUNTERS
Counters count occurrences of an event. Counters are often used to determine the frequency at which an event is happening, as was done in the login example above. Counter metrics have “c” as their type in the datagram format. The value of a counter metric is the number of occurrences of the event that you wish to count, which may be a positive or negative whole number. Many clients implement “increment” and “decrement” functions, which are shorthand for counters with values of +1 or -1, respectively.

```
login.invocations:1|c        # increment login.invocations by 1
other_key:-100|c             # decrement other_key by 100
```

### TIMERS
Timers measure the amount of time an action took to complete, in milliseconds. Timers have “ms” as their metric type. The StatsD server will compute the mean, standard deviation, sum, and upper and lower bounds for a timer over one flush interval. The StatsD server can also be configured to compute histograms for these metrics (see this link for more information about histograms).

login.time:22|ms   # record a login.time event that took 22 ms

### GAUGES
Gauges are arbitrary, persistent values. Once a gauge is set to its value, the StatsD server will report the same value each flush period. After a gauge has been set, you can add a sign to a gauge’s value to indicate a change in value. Gauges have “g” as their type.

```
gas_tank:0.50|g   # set the gas tank metric to 50%
gas_tank:+0.50|g  # Add 50% to the gas tank. It now reads 100%
gas_tank:-0.75|g  # Subtract 75% from the gas tank. It now reads 25%
```

### SETS
Sets report the number of unique elements that are received in a flush period. The value of a set is a unique identifier for an element you wish to count. Sets have “s” as their type.

Assume the following metrics occur within one flush period:
```                       
# unique_users = 0

unique_users:foo|s     

# count an occurrence of user `foo`. unique_users = 1

unique_users:foo|s     

# we’ve already seen `foo`, so again unique_users = 1

unique_users:bar|s     
# unique_users = 2
```

After a flush, unique_users will reset to 0 until another metric is received.


# Kibana

statsd goes to the stasdbeat- index
you can use wildcards on the buckets

bucket:*password.fail*



========

See processor_test.go
```
{
			field: common.Field{
				Type: "object", ObjectType: "scaled_float",
				Name: "core.*.pct", ScalingFactor: 100, ObjectTypeMappingType: "float",
			},
			expected: common.MapStr{
				"core.*.pct": common.MapStr{
					"mapping": common.MapStr{
						"type":           "scaled_float",
						"scaling_factor": 100,
					},
					"match_mapping_type": "float",
					"path_match":         "core.*.pct",
				},
			},
		},
	}
```