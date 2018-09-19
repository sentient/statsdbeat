package beater

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
)

//regex might be getting too complicated? ->  ^(?P<bucket>(.*))\:(?P<value>(.*))\|(?P<type>(\-?\w))(\|(?P<rate>(.*)))?$
//var startd = regexp.MustCompile(`/^(:<ns>[^.]+)\.(:<grp>[^.]+)\.(:<tgt>[^.]+)(?:\.(:<act>[^.]+))?/`)
var bucket = regexp.MustCompile(`/^(:<ns>[^.]+)\.(:<grp>[^.]+)\.(:<tgt>[^.]+)(?:\.(:<act>[^.]+))?/`)

//msg has format <bucket>(,<k>=<v>)*:<value>|<type>|@<sample rate>
func ParseBeat(msg string) ([]beat.Event, error) {

	parts := strings.Split(msg, "|")
	if len(parts) < 2 || len(parts) > 3 {
		return nil, fmt.Errorf("Expecting 2 or 3 parts of | but was %d", len(parts))
	}

	//parts[0] has structure of  <bucket>(,<k>=<v>)*:<value>
	bucket, _, val, err := getBucketTagsValue(parts[0])
	if err != nil {
		return nil, err
	}

	_type := strings.TrimSpace(parts[1])
	e := &beat.Event{
		Timestamp: time.Now(),
	}

	switch _type {
	case "c":
		{
			e.Fields = common.MapStr{
				"val":    val,
				"bucket": bucket,
			}
		}
	default:
		{
			return nil, fmt.Errorf("Type %v not handled yet", _type)
		}
	}
	e.Fields.Put("testtags.a", "tag value a")
	e.Fields.Put("testtags.b", "tag value b")

	return []beat.Event{*e}, nil
}

func getBucketTagsValue(part string) (bucket string, tags map[string]string, val int, err error) {

	parts := strings.Split(part, ":")
	subParts := strings.Split(parts[0], ",")
	bucket = subParts[0]

	tags = make(map[string]string, len(subParts)-1)
	for i := 1; i < len(subParts); i++ {
		kv := strings.Split(subParts[i], "=")
		if len(kv) == 2 {
			tags[kv[0]] = kv[1]
		}
	}

	val, err = strconv.Atoi(parts[1])

	return bucket, tags, val, err
}
