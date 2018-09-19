package beater

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/logp"
)

var timeFormat = "20060102_150405"

//Note this is not really a file extension. That would have been .json, but we just want to filter it down further
var fileSuffix = "_statsd.json"

type tmpStorage struct {
	path         string
	eventsMap    map[string][]beat.Event
	truncateTime time.Duration
	log          *logp.Logger
}

/*
  NewStorage defines a file storage account at path 'path' in file blocks of 'duration' in seconds
  e.g. batch every 5 minute to \data\tmp  -> NewStorage('\data\tmp', 300)
*/
func NewStorage(path string, duration time.Duration) (*tmpStorage, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return nil, err
		}
	}

	t := tmpStorage{
		path:         path,
		eventsMap:    make(map[string][]beat.Event),
		truncateTime: time.Duration(duration),
		log:          logp.NewLogger("storage"),
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		t.log.Errorf("Failed to ReadDir %v", path)
		return nil, err
	}
	t.log.Infof("Using disk storage %v for resubmit", path)

	for _, f := range files {
		filename := f.Name()
		t.log.Debugf("Checking file %v", filename)
		if strings.HasSuffix(filename, fileSuffix) {
			var name = filename[0 : len(filename)-len(fileSuffix)]
			t.log.Debugf("Reading file %v", name)
			t.readFromDisk(name, filepath.Join(path, filename))
		} else {
			t.log.Debugf("Not matching file suffix %v (but was %v)", fileSuffix, filename)
		}
	}

	return &t, nil
}

func (t *tmpStorage) SendBatch(client beat.Client, events []beat.Event) {
	//Take timestamp of the last event and put that in the hourly record
	if len(events) > 0 {
		t.log.Debugf("Adding batch. size %d", len(events))
		lastEvent := events[len(events)-1]
		ts := lastEvent.Timestamp
		ti := ts.Truncate(t.truncateTime)
		key := ti.Format(timeFormat)
		//set the batch key to last timeslot
		for n := range events {
			events[n].Private = key
		}

		if _, exists := t.eventsMap[key]; exists {
			t.log.Warnf("How is this  possible? We already have %v", key)
		}
		t.eventsMap[key] = events

		client.PublishAll(events)

	} else {
		t.log.Debug("I got no statsd events to send")
	}

}

func (t *tmpStorage) TrySync(client beat.Client) {
	for _, batch := range t.eventsMap {
		client.PublishAll(batch)
	}
}

func (t *tmpStorage) RemoveBatch(batchID string) {
	t.log.Debugf("Remove batch %v from memory!", batchID)
	if _, exists := t.eventsMap[batchID]; exists {
		delete(t.eventsMap, batchID)
	} else {
		t.log.Warnf("What happened to batch %v ?", batchID)
	}

}

func (t *tmpStorage) Flush() {
	if len(t.eventsMap) > 0 {
		t.log.Warn("Persisting pending statsd data to disk!")
		for key, batch := range t.eventsMap {
			t.writeToDisk(key, batch)
		}
	}
}

func (t *tmpStorage) writeToDisk(key string, beats []beat.Event) {
	json, err := json.Marshal(beats)
	if err != nil {
		t.log.Error(err)
	}

	filename := key + "." + fileSuffix
	err = ioutil.WriteFile(path.Join(t.path, filename), json, 0644)
	if err != nil {
		t.log.Error(err)
	}
}

func (t *tmpStorage) readFromDisk(key string, filename string) {
	t.log.Warn("Reloading statsd data that has not been acknowledged.")
	blob, err := ioutil.ReadFile(filename)
	if err != nil {
		t.log.Error(err)
	}
	var events []beat.Event
	err = json.Unmarshal(blob, &events)
	if err != nil {
		t.log.Error(err)
	}
	//put the events back on the map
	t.eventsMap[key] = events
	err = os.Remove(filename)

	if err != nil {
		t.log.Error(err)
	}

}
