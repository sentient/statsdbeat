// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

type Config struct {
	Period               time.Duration `config:"period"`        //The flush interval from statsd client, to elasticsearch
	UDPAddress           string        `confing:"statsdserver"` //udp listening
	RetryStorageLocation string        `config:"storage"`       //flush to disk if we cannot connect.
}

var DefaultConfig = Config{
	Period:               5 * time.Second,
	UDPAddress:           ":8125",
	RetryStorageLocation: "./.storage/retry",
}
