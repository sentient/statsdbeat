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
		want    *beat.Event
		wantErr bool
	}{
		{"testSimple",
			args{"myCounter:1|c"},
			&beat.Event{
				Fields: common.MapStr{
					"val":     1,
					"bucket":  "myCounter",
					"context": map[string]interface{}{},
				},
			},
			false,
		},
		{"testWithTags",
			args{"myCounter,myTag=error,tagB=2:1|c"},
			&beat.Event{
				Fields: common.MapStr{
					"val":    1,
					"bucket": "myCounter",
					"context": map[string]interface{}{
						"myTag": "error",
						"tagB":  "2",
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beats, err := ParseBeat(tt.args.msg)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBeat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(beats) != 1 {
				t.Errorf("Expected one beat, but was %d", len(beats))
			}
			got := beats[0]
			//We want to skip the timestamp

			if !reflect.DeepEqual(got.Fields, tt.want.Fields) {
				t.Errorf("ParseBeat() = %v, want %v", got.Fields, tt.want.Fields)
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
				// if beats[i].Fields.Flatten().StringToPrint() != tt.want[i].Fields.Flatten().StringToPrint() {
				// 	t.Errorf("ParseBeat failed\n%v,\n%v",
				// 		beats[i].Fields.Flatten().StringToPrint(),
				// 		tt.want[i].Fields.Flatten().StringToPrint())
				// }
				//Got to flatten it!
				if !reflect.DeepEqual(beats[i].Fields.Flatten(), tt.want[i].Fields.Flatten()) {
					t.Errorf("ParseBeat() = \n%v, want \n%v", beats[i].Fields, tt.want[i].Fields)
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
