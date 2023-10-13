package beater

import (
	"fmt"
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"

	"github.com/Qiu-Weidong/jobbeat/config"
)

// jobbeat configuration.
type jobbeat struct {
	done   chan struct{}
	config config.Config
	client beat.Client

	lastIndexTime time.Time
}

// New creates an instance of jobbeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &jobbeat{
		done:   make(chan struct{}),
		config: c,

		// 初始化 lastIndexTime
		lastIndexTime: time.Now(),
	}
	return bt, nil
}

// Run starts jobbeat.
func (bt *jobbeat) Run(b *beat.Beat) error {
	logp.Info("jobbeat is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(bt.config.Period)
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		// event := beat.Event{
		// 	Timestamp: time.Now(),
		// 	Fields: common.MapStr{
		// 		"type":    b.Info.Name,
		// 		"counter": counter,
		// 	},
		// }
		// bt.client.Publish(event)
		bt.collectJobs(bt.config.Path, b)
		logp.Info("Event sent")
	}
}

// Stop stops jobbeat.
func (bt *jobbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}

func (bt *jobbeat) collectJobs(baseDir string, b *beat.Beat) {
	now := time.Now()
	// 更新时间
	bt.lastIndexTime = now

	event := beat.Event{
		Timestamp: now,
		Fields: common.MapStr{
			"type": "job",
		},
	}
	bt.client.Publish(event)
}
