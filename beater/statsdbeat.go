package beater

import (
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/sentient/statsdbeat/config"
)

type Statsdbeat struct {
	done       chan struct{}
	stopping   bool
	stopped    bool
	config     config.Config
	client     beat.Client
	address    *net.UDPAddr
	pipeline   beat.Pipeline // Interface to publish event.
	buffer     []beat.Event
	mux        sync.Mutex
	tmpStorage *tmpStorage
	log        *logp.Logger
}

//damn, why do I find out about the spool after my initial writing of the tmpStorage
var useSpool bool = true

// New Creates a statsdbeater
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
		//buffer:  make([]beat.Event), //I  don't think I can give a sensible pre-allocation here. So let's leave is empty
	}

	bt.address, err = net.ResolveUDPAddr("udp", c.UDPAddress)
	if err != nil {
		bt.log.Errorf("Failed to resolve udp address: %v, %v", c.UDPAddress, err)
		return nil, err
	}
	bt.log.Infof("Statsd server listening for UDP packages at '%v'", c.UDPAddress)

	bt.pipeline = b.Publisher

	if !useSpool {
		bt.tmpStorage, err = NewStorage(c.RetryStorageLocation, c.Period)
		if err != nil {
			return nil, fmt.Errorf("Failed to create tmp storage, %v", c.RetryStorageLocation)
		}
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

func (bt *Statsdbeat) Run(b *beat.Beat) error {
	bt.log.Info("statsdbeat is running! Hit CTRL-C to stop it.")
	var err error

	ce := ClientEventHandler{bt: bt}
	ph := PipelineACKHandler{bt: bt}

	bt.client, err = b.Publisher.ConnectWith(beat.ClientConfig{
		PublishMode: beat.GuaranteedSend,
		WaitClose:   10 * time.Second,
		//ACKCount:     ph.ACKCount,
		//ACKEvents:    ph.ACKEvents,
		ACKLastEvent: ph.ACKLastEvent,
		Events:       ce,
	})

	//defer bt.client.Close()

	if err != nil {
		return err
	}

	if !useSpool {
		//See if there were pending disk changes that still have to be send out.
		bt.tmpStorage.TrySync(bt.client)
	}

	//I was able to connect to ElasticSearch
	//now I'm will to received UDP packages
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
		}

		bt.sendStatsdBuffer()
	}
}

func (bt *Statsdbeat) sendStatsdBuffer() {
	bt.mux.Lock()
	if len(bt.buffer) > 0 {
		if useSpool {
			bt.client.PublishAll(bt.buffer)
		} else {
			bt.tmpStorage.SendBatch(bt.client, bt.buffer)
		}
		bt.buffer = nil
	}
	bt.mux.Unlock()
}

func (bt *Statsdbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}

type ClientEventHandler struct {
	bt *Statsdbeat
}

func (h ClientEventHandler) Closing() {
	h.bt.log.Info("statsdbeat CLOSING. Sending last buffer")
	h.bt.stopping = true
	h.bt.sendStatsdBuffer()

}

func (h ClientEventHandler) Closed() {
	h.bt.log.Info("statsdbeat CLOSED.")
	if !useSpool {
		h.bt.tmpStorage.Flush()
	}

}

func (h ClientEventHandler) Published() {
	h.bt.log.Debug("statsdbeat publishing. (it has not been confirmed yet)")
}

func (h ClientEventHandler) FilteredOut(ev beat.Event) {
	h.bt.log.Infof("filtered out %v", ev.Private)

}

func (h ClientEventHandler) DroppedOnPublish(ev beat.Event) {
	h.bt.log.Infof("dropped on publish %v", ev.Private)
}

type PipelineACKHandler struct {
	bt *Statsdbeat
}

// func (h PipelineACKHandler) ACKCount(n int) {
// 	h.bt.log.Infof("acknowledged %d", n)

// }

// func (h PipelineACKHandler) ACKEvents(privates []interface{}) {
// 	for _, e := range privates {
// 		h.bt.log.Infof("acknowledged %v", e)
// 	}
// }

//ACKLastEvent reports the last ACKed event out of a batch of ACKed events only.
func (h PipelineACKHandler) ACKLastEvent(private interface{}) {
	h.bt.log.Debug("last event from batch: %v", private)
	if !useSpool {
		h.bt.tmpStorage.RemoveBatch(private.(string))
	}

}
