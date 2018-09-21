// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

type Config struct {
	Period     time.Duration `config:"period"`       //The flush interval from statsd client, to elasticsearch
	UDPAddress string        `config:"statsdserver"` //udp listening
}

var DefaultConfig = Config{
	Period:     5 * time.Second,
	UDPAddress: ":8125",
}
