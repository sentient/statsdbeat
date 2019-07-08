package beater

import (
	"reflect"
	"testing"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
)

func Test_getBucketTagsValue(t *testing.T) {
	type args struct {
		part string
	}
	tests := []struct {
		name       string
		args       args
		wantBucket string
		wantTags   map[string]interface{}
		wantVal    int
		wantErr    bool
	}{
		{"testSimple", args{"myCounter:1"},
			"myCounter",
			map[string]interface{}{},
			1,
			false,
		},
		{"testDotNotation", args{"my.counter.at:1"},
			"my.counter.at",
			map[string]interface{}{},
			1,
			false,
		},
		{"testNegative", args{"myCounter:-1"},
			"myCounter",
			map[string]interface{}{},
			-1,
			false,
		},
		{"testOneTags", args{"myCounter,myTag=a:1"},
			"myCounter",
			map[string]interface{}{
				"myTag": "a",
			},
			1,
			false,
		},
		{"testTwoTags", args{"myCounter,myTag=a,anotherTag=b:1"},
			"myCounter",
			map[string]interface{}{
				"myTag":      "a",
				"anotherTag": "b",
			},
			1,
			false,
		},
		{"testInvalidCount", args{"myCounter:ab"},
			"myCounter",
			map[string]interface{}{},
			0,
			true,
		},
		{"testValidFloat", args{"myCounter:1.0000"},
			"myCounter",
			map[string]interface{}{},
			1,
			false,
		},
		{"testValidFloat2", args{"myCounter:0000001.0000"},
			"myCounter",
			map[string]interface{}{},
			1,
			false,
		},
		{"testInvalidFloat", args{"myCounter:1.1"},
			"myCounter",
			map[string]interface{}{},
			0,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBucket, gotTags, gotVal, err := getBucketTagsValue(tt.args.part)
			if (err != nil) != tt.wantErr {
				t.Errorf("getBucketTagsValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotBucket != tt.wantBucket {
				t.Errorf("getBucketTagsValue() gotBucket = %v, want %v", gotBucket, tt.wantBucket)
			}
			if !reflect.DeepEqual(gotTags, tt.wantTags) {
				t.Errorf("getBucketTagsValue() gotTags = %v, want %v", gotTags, tt.wantTags)
			}
			if gotVal != tt.wantVal {
				t.Errorf("getBucketTagsValue() gotVal = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func Test_parseBeat(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name    string
		args    args
		want    []beat.Event
		wantErr bool
	}{
		{"testSimple",
			args{"myCounter:1|c"},
			[]beat.Event{
				beat.Event{
					Fields: common.MapStr{
						"statsd.bucket": "myCounter",
						"statsd.target": "myCounter",
						"statsd.type":   "counter",
						"statsd.value":  1,
					},
				},
			},
			false,
		},
		{"testWithTags",
			args{"myCounter,myTag=error,tagB=2:1|c"},
			[]beat.Event{
				beat.Event{
					Fields: common.MapStr{
						"statsd.bucket": "myCounter",
						"statsd.target": "myCounter",
						"statsd.type":   "counter",
						"statsd.value":  1,
						"statsd.ctx": map[string]interface{}{
							"myTag": "error",
							"tagB":  "2",
						},
					},
				},
			},
			false,
		},
		{"testGauge",
			args{"platform-insights.test.gauge.num_goroutine:4|g"},
			[]beat.Event{
				beat.Event{
					Fields: common.MapStr{
						"statsd.bucket":    "platform-insights.test.gauge.num_goroutine",
						"statsd.namespace": "platform-insights",
						"statsd.section":   "test",
						"statsd.target":    "gauge",
						"statsd.action":    "num_goroutine",
						"statsd.type":      "gauge",
						"statsd.value":     4,
					},
				},
			},
			false,
		},
		{"testHistogram",
			args{"platform-insights.test.histogram.not_sure:17|h"},
			[]beat.Event{
				beat.Event{
					Fields: common.MapStr{
						"statsd.bucket":    "platform-insights.test.histogram.not_sure",
						"statsd.namespace": "platform-insights",
						"statsd.section":   "test",
						"statsd.target":    "histogram",
						"statsd.action":    "not_sure",
						"statsd.type":      "histogram",
						"statsd.value":     17,
					},
				},
			},
			false,
		},
		{"testTiming",
			args{"platform-insights.test.timing.ping:100|ms"},
			[]beat.Event{
				beat.Event{
					Fields: common.MapStr{
						"statsd.bucket":    "platform-insights.test.timing.ping",
						"statsd.namespace": "platform-insights",
						"statsd.section":   "test",
						"statsd.target":    "timing",
						"statsd.action":    "ping",
						"statsd.type":      "timing",
						"statsd.value":     100,
					},
				},
			},
			false,
		},
		{"testInvalidParts",
			args{"myCounter:c"},
			[]beat.Event{},
			true,
		},
		{"testInvalidParts2",
			args{"myCounter|1|2|3:c"},
			[]beat.Event{},
			true,
		},
		{"testInvalidFloatValue",
			args{"myCounter:not_a_float|c"},
			[]beat.Event{},
			true,
		},
		{"testInvalidType",
			args{"myCounter:1|x"},
			[]beat.Event{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beats, err := ParseBeats(tt.args.msg)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBeat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(beats) != len(tt.want) {
				t.Errorf("ParseBeat array length = %v, want %v", len(beats), len(tt.want))
				return
			}

			//We want to skip the timestamp
			for i := 0; i < len(beats); i++ {
				//Got to flatten it!
				if !reflect.DeepEqual(beats[i].Fields.Flatten(), tt.want[i].Fields.Flatten()) {
					t.Errorf("ParseBeat() = \n%v, want \n%v", beats[i].Fields.Flatten(), tt.want[i].Fields.Flatten())
				}
			}

		})
	}
}

func TestParseBeats(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name    string
		args    args
		want    []beat.Event
		wantErr bool
	}{
		{"testSimple",
			args{"myCounter:1|c"},
			[]beat.Event{
				beat.Event{
					Fields: common.MapStr{
						"statsd.bucket": "myCounter",
						"statsd.target": "myCounter",
						"statsd.type":   "counter",
						"statsd.value":  1,
						"statsd": map[string]interface{}{
							"ctx": map[string]interface{}{},
						},
					},
				},
			},
			false,
		},
		{"testMultiline",
			args{"myCounter:1|c\ncnt2:2|c\ncnt3:3|c\ncnt4:4|c\ncnt5:5|c"},
			[]beat.Event{
				beat.Event{
					Fields: common.MapStr{
						"statsd.bucket": "myCounter",
						"statsd.target": "myCounter",
						"statsd.type":   "counter",
						"statsd.value":  1,
						"statsd": map[string]interface{}{
							"ctx": map[string]interface{}{},
						},
					},
				},
				beat.Event{
					Fields: common.MapStr{
						"statsd.bucket": "cnt2",
						"statsd.target": "cnt2",
						"statsd.type":   "counter",
						"statsd.value":  2,
						"statsd": map[string]interface{}{
							"ctx": map[string]interface{}{},
						},
					},
				},
				beat.Event{
					Fields: common.MapStr{
						"statsd.bucket": "cnt3",
						"statsd.target": "cnt3",
						"statsd.type":   "counter",
						"statsd.value":  3,
						"statsd": map[string]interface{}{
							"ctx": map[string]interface{}{},
						},
					},
				},
				beat.Event{
					Fields: common.MapStr{
						"statsd.bucket": "cnt4",
						"statsd.target": "cnt4",
						"statsd.type":   "counter",
						"statsd.value":  4,
						"statsd": map[string]interface{}{
							"ctx": map[string]interface{}{},
						},
					},
				},
				beat.Event{
					Fields: common.MapStr{
						"statsd.bucket": "cnt5",
						"statsd.target": "cnt5",
						"statsd.type":   "counter",
						"statsd.value":  5,
						"statsd": map[string]interface{}{
							"ctx": map[string]interface{}{},
						},
					},
				},
			},
			false,
		},
		{"testFloat",
			args{"myCounterAsFloat:1.0|c"},
			[]beat.Event{
				beat.Event{
					Fields: common.MapStr{
						"statsd.bucket": "myCounterAsFloat",
						"statsd.target": "myCounterAsFloat",
						"statsd.type":   "counter",
						"statsd.value":  1,
						"statsd": map[string]interface{}{
							"ctx": map[string]interface{}{},
						},
					},
				},
			},
			false,
		},
		{"testVault",
			args{"\nvault.test-phoenix-platform.runtime.num_goroutines:20.000000|g" +
				"\nvault.test-phoenix-platform.runtime.alloc_bytes:4523224.000000|g" +
				"\nvault.test-phoenix-platform.runtime.sys_bytes:72022264.000000|g" +
				"\nvault.test-phoenix-platform.runtime.malloc_count:96671.000000|g" +
				"\nvault.test-phoenix-platform.runtime.free_count:77885.000000|g" +
				"\nvault.test-phoenix-platform.runtime.heap_objects:18786.000000|g" +
				"\nvault.test-phoenix-platform.runtime.total_gc_pause_ns:3770211.000000|g" +
				"\nvault.test-phoenix-platform.runtime.total_gc_runs:5.000000|g"},
			[]beat.Event{
				beat.Event{
					Fields: common.MapStr{
						"statsd.bucket":    "vault.test-phoenix-platform.runtime.num_goroutines",
						"statsd.namespace": "vault",
						"statsd.section":   "test-phoenix-platform",
						"statsd.target":    "runtime",
						"statsd.action":    "num_goroutines",
						"statsd.type":      "gauge",
						"statsd.value":     20,
						"statsd": map[string]interface{}{
							"ctx": map[string]interface{}{},
						},
					},
				},
				beat.Event{
					Fields: common.MapStr{
						"statsd.bucket":    "vault.test-phoenix-platform.runtime.alloc_bytes",
						"statsd.namespace": "vault",
						"statsd.section":   "test-phoenix-platform",
						"statsd.target":    "runtime",
						"statsd.action":    "alloc_bytes",
						"statsd.type":      "gauge",
						"statsd.value":     4523224,
						"statsd": map[string]interface{}{
							"ctx": map[string]interface{}{},
						},
					},
				},
				beat.Event{
					Fields: common.MapStr{
						"statsd.bucket":    "vault.test-phoenix-platform.runtime.sys_bytes",
						"statsd.namespace": "vault",
						"statsd.section":   "test-phoenix-platform",
						"statsd.target":    "runtime",
						"statsd.action":    "sys_bytes",
						"statsd.type":      "gauge",
						"statsd.value":     72022264,
						"statsd": map[string]interface{}{
							"ctx": map[string]interface{}{},
						},
					},
				},
				beat.Event{
					Fields: common.MapStr{
						"statsd.bucket":    "vault.test-phoenix-platform.runtime.malloc_count",
						"statsd.namespace": "vault",
						"statsd.section":   "test-phoenix-platform",
						"statsd.target":    "runtime",
						"statsd.action":    "malloc_count",
						"statsd.type":      "gauge",
						"statsd.value":     96671,
						"statsd": map[string]interface{}{
							"ctx": map[string]interface{}{},
						},
					},
				},
				beat.Event{
					Fields: common.MapStr{
						"statsd.bucket":    "vault.test-phoenix-platform.runtime.free_count",
						"statsd.namespace": "vault",
						"statsd.section":   "test-phoenix-platform",
						"statsd.target":    "runtime",
						"statsd.action":    "free_count",
						"statsd.type":      "gauge",
						"statsd.value":     77885,
						"statsd": map[string]interface{}{
							"ctx": map[string]interface{}{},
						},
					},
				},
				beat.Event{
					Fields: common.MapStr{
						"statsd.bucket":    "vault.test-phoenix-platform.runtime.heap_objects",
						"statsd.namespace": "vault",
						"statsd.section":   "test-phoenix-platform",
						"statsd.target":    "runtime",
						"statsd.action":    "heap_objects",
						"statsd.type":      "gauge",
						"statsd.value":     18786,
						"statsd": map[string]interface{}{
							"ctx": map[string]interface{}{},
						},
					},
				},
				beat.Event{
					Fields: common.MapStr{
						"statsd.bucket":    "vault.test-phoenix-platform.runtime.total_gc_pause_ns",
						"statsd.namespace": "vault",
						"statsd.section":   "test-phoenix-platform",
						"statsd.target":    "runtime",
						"statsd.action":    "total_gc_pause_ns",
						"statsd.type":      "gauge",
						"statsd.value":     3770211,
						"statsd": map[string]interface{}{
							"ctx": map[string]interface{}{},
						},
					},
				},
				beat.Event{
					Fields: common.MapStr{
						"statsd.bucket":    "vault.test-phoenix-platform.runtime.total_gc_runs",
						"statsd.namespace": "vault",
						"statsd.section":   "test-phoenix-platform",
						"statsd.target":    "runtime",
						"statsd.action":    "total_gc_runs",
						"statsd.type":      "gauge",
						"statsd.value":     5,
						"statsd": map[string]interface{}{
							"ctx": map[string]interface{}{},
						},
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beats, err := ParseBeats(tt.args.msg)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBeat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(beats) != len(tt.want) {
				t.Errorf("ParseBeat array length = %v, want %v", len(beats), len(tt.want))
				return
			}

			//We want to skip the timestamp
			for i := 0; i < len(beats); i++ {

				//Got to flatten it!
				if !reflect.DeepEqual(beats[i].Fields.Flatten(), tt.want[i].Fields.Flatten()) {
					t.Errorf("ParseBeat() = \n%v, want \n%v", beats[i].Fields.Flatten(), tt.want[i].Fields.Flatten())
				}
			}

		})
	}
}

func Test_splitBucket(t *testing.T) {
	type args struct {
		bucket string
	}
	tests := []struct {
		name          string
		args          args
		wantNamespace string
		wantSection   string
		wantTarget    string
		wantAction    string
	}{
		{"simple", args{"counter"}, "", "", "counter", ""},
		{"simple", args{"account.login"}, "", "", "account", "login"},
		{"simple", args{"authentication.password.failure"}, "", "authentication", "password", "failure"},
		{"simple", args{"module.authentication.password.failure"}, "module", "authentication", "password", "failure"},
		{"simple", args{"accounts.authentication.password.failure.no_email_found"}, "accounts", "authentication", "password", "failure.no_email_found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNamespace, gotSection, gotTarget, gotAction := splitBucket(tt.args.bucket)
			if gotNamespace != tt.wantNamespace {
				t.Errorf("splitBucket() gotNamespace = %v, want %v", gotNamespace, tt.wantNamespace)
			}
			if gotSection != tt.wantSection {
				t.Errorf("splitBucket() gotSection = %v, want %v", gotSection, tt.wantSection)
			}
			if gotTarget != tt.wantTarget {
				t.Errorf("splitBucket() gotTarget = %v, want %v", gotTarget, tt.wantTarget)
			}
			if gotAction != tt.wantAction {
				t.Errorf("splitBucket() gotAction = %v, want %v", gotAction, tt.wantAction)
			}
		})
	}
}
