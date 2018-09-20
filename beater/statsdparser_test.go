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

func TestParseBeat(t *testing.T) {
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
