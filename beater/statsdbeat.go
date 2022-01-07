package beater

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"

	"github.com/sentient/statsdbeat/config"
)

// Statsdbeat configuration.
type Statsdbeat struct {
	done   chan struct{}
	config config.Config
	client beat.Client
	//
	stopping bool
	stopped  bool
	address  *net.UDPAddr
	pipeline beat.Pipeline // Interface to publish event.
	buffer   []beat.Event
	mux      sync.Mutex
	log      *logp.Logger
	health   *HealthServer
}

// New creates an instance of statsdbeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	var err error
	if err = cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Statsdbeat{
		done:   make(chan struct{}),
		config: c,
		log:    logp.NewLogger("statsdbeat"),
	}

	bt.address, err = net.ResolveUDPAddr("udp", c.UDPAddress)
	if err != nil {
		bt.log.Errorf("Failed to resolve udp address: %v, %v", c.UDPAddress, err)
		return nil, err
	}
	bt.log.Infof("Statsd server listening for UDP packages at '%v'", c.UDPAddress)

	bt.pipeline = b.Publisher

	if len(c.TCPHealthAddress) > 0 {
		bt.log.Infof("Setup serving health checks at '%v'", c.TCPHealthAddress)
		bt.health = NewHealthCheck(c.TCPHealthAddress, bt.log)
	} else {
		bt.log.Info("No TCP health check configured. E.g. you could set statsdbeat.healthserver: \":8080\" to respond to TCP health checks")
	}

	return bt, nil
}

func (bt *Statsdbeat) listenAndBuffer(conn *net.UDPConn) {
	buf := make([]byte, 1024)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if bt.stopping || bt.stopped {
			return
		}
		statsdMsg := string(buf[0:n])
		if len(statsdMsg) > 0 {
			bt.log.Debug(fmt.Sprintf("Received %v from %v", statsdMsg, addr))

			events, err := ParseBeats(statsdMsg)
			if err != nil {
				bt.log.Error("Failed making a beat", zap.Error(err))
			} else {
				bt.mux.Lock()
				bt.buffer = append(bt.buffer, events...)
				bt.mux.Unlock()
			}
		}

		if err != nil {
			logp.Error(err)
		}
	}
}

// Run starts statsdbeat.
func (bt *Statsdbeat) Run(b *beat.Beat) error {
	bt.log.Info("statsdbeat is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.ConnectWith(beat.ClientConfig{
		PublishMode: beat.GuaranteedSend,
		WaitClose:   10 * time.Second,
	})

	if err != nil {
		return err
	}

	// I was able to connect to ElasticSearch
	// ready to receive UDP packages...
	conn, err := net.ListenUDP("udp", bt.address)
	if err != nil {
		return err
	}
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()

	go bt.listenAndBuffer(conn)

	if bt.health != nil {
		go bt.health.Serve()
	}

	ticker := time.NewTicker(bt.config.Period)

	for {
		select {
		case <-bt.done:
			if conn != nil {
				bt.stopped = true
				bt.log.Info("stop listening on UDP")
				conn.Close()
			}
			return nil
		case <-ticker.C:
			bt.sendStatsdBuffer()
		}
	}
}

func (bt *Statsdbeat) sendStatsdBuffer() {
	bt.mux.Lock()
	if len(bt.buffer) > 0 {
		bt.log.Info("Sending buffer " + strconv.Itoa(len(bt.buffer)))
		bt.client.PublishAll(bt.buffer)
		bt.buffer = nil
	}
	bt.mux.Unlock()
}

// Stop stops statsdbeat.
func (bt *Statsdbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
