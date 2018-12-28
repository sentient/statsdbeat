package beater

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
)

/*ParseBeats takes a string constructs a  beat.Event.
  the msg has format <bucket>(,<k>=<v>)*:<value>|<type>|@<sample rate>
*/
func ParseBeats(msg string) ([]beat.Event, error) {
	parts := strings.Split(msg, "\n")
	result := []beat.Event{}
	for p := range parts {
		if len(strings.TrimSpace(parts[p])) == 0 {
			//skip empty lines
			continue
		}
		b, err := parseBeat(parts[p])
		if err != nil {
			return nil, err
		}
		result = append(result, b...)
	}
	return result, nil
}

func parseBeat(msg string) ([]beat.Event, error) {
	parts := strings.Split(msg, "|")
	if len(parts) < 2 || len(parts) > 3 {
		return nil, fmt.Errorf("Expecting 2 or 3 parts of | but was %d", len(parts))
	}

	//parts[0] has structure of  <bucket>(,<k>=<v>)*:<value>
	bucket, tags, val, err := getBucketTagsValue(parts[0])
	if err != nil {
		return nil, err
	}

	_type := strings.TrimSpace(parts[1])
	e := &beat.Event{
		Timestamp: time.Now(),
	}

	ns, sect, tgt, act := splitBucket(bucket)
	bucketMap := common.MapStr{
		"statsd.bucket": bucket,
		"statsd.target": tgt,
	}
	if len(act) > 0 {
		bucketMap.Put("statsd.action", act)
	}
	if len(sect) > 0 {
		bucketMap.Put("statsd.section", sect)
	}
	if len(ns) > 0 {
		bucketMap.Put("statsd.namespace", ns)
	}
	if len(tags) > 0 {
		bucketMap.Put("statsd.ctx", tags)
	}

	switch _type {
	case "c":
		{
			e.Fields = common.MapStr{
				"statsd.value": val,
				"statsd.type":  "counter",
			}
		}
	case "g":
		{
			e.Fields = common.MapStr{
				"statsd.value": val,
				"statsd.type":  "gauge",
			}
		}
	case "h":
		{
			e.Fields = common.MapStr{
				"statsd.value": val,
				"statsd.type":  "histogram",
			}
		}
	case "ms":
		{
			e.Fields = common.MapStr{
				"statsd.value": val,
				"statsd.type":  "timing",
			}
		}

	default:
		{
			return nil, fmt.Errorf("Type %v not handled yet", _type)
		}
	}

	e.Fields.Update(bucketMap)

	return []beat.Event{*e}, nil
}

func getBucketTagsValue(part string) (bucket string, tags map[string]interface{}, val int, err error) {

	parts := strings.Split(part, ":")
	subParts := strings.Split(parts[0], ",")
	bucket = subParts[0]

	tags = make(map[string]interface{}, len(subParts)-1)
	for i := 1; i < len(subParts); i++ {
		kv := strings.Split(subParts[i], "=")
		if len(kv) == 2 {
			tags[kv[0]] = kv[1]
		}
	}

	var fval float64
	if fval, err = strconv.ParseFloat(parts[1], 64); err == nil {
		if fval == float64(int(fval)) {
			val = int(fval)
		} else {
			return bucket, tags, 0, errors.New("failed to parse the value to an int " + parts[1])
		}
	}

	return bucket, tags, val, err
}

//accounts.authentication.password.failure.no_email_found
// We always have a target, then action, then section then namespace
func splitBucket(bucket string) (namespace string, section string, target string, action string) {
	parts := strings.Split(bucket, ".")
	l := len(parts)
	switch {
	case l == 1:
		{
			target = parts[0]
		}
	case l == 2:
		{
			target = parts[0]
			action = parts[1]
		}
	case l == 3:
		{
			section = parts[0]
			target = parts[1]
			action = parts[2]
		}
	case l == 4:
		{
			namespace = parts[0]
			section = parts[1]
			target = parts[2]
			action = parts[3]
		}
	case l > 4:
		{
			namespace = parts[0]
			section = parts[1]
			target = parts[2]
			action = strings.Join(parts[3:], ".")
		}
	}
	return namespace, section, target, action
}
