package beater

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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

	// 还需要维护一个文件采集时间信息, 并持久化到一个 json 文件当中
	// 先只考虑内存
	registrar map[string]time.Time
}

// New creates an instance of jobbeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	// 可以使用 1m 来表示分钟 1h 来表示小时
	// logp.Info("config period = %f", c.Period.Seconds())

	bt := &jobbeat{
		done:   make(chan struct{}),
		config: c,

		// 初始化 lastIndexTime
		lastIndexTime: time.Now(),

		// 加载 registrar
		registrar: loadRegistrar(),
	}
	return bt, nil
}

// Run starts jobbeat.
func (bt *jobbeat) Run(b *beat.Beat) error {
	logp.Info("jobbeat is running! Hit CTRL-C to stop it.")

	/* 维护文件注册信息
	 * {
	 *    "path": { "hash": "xxx", "time": "xxxx" }
	 * }
	 */

	// 加载 registrar 文件
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
	ext := ".job"

	now := time.Now()

	modified := false

	err := filepath.Walk(baseDir, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ext {
			modTime := info.ModTime()

			if last, ok := bt.registrar[path]; ok {
				// 采集过, 则比较上次采集时间和修改时间
				if last.Before(modTime) {
					// 上次采集时间在修改时间之前, 表示采集之后有修改, 需要补充采集
					bt.registrar[path] = now
					modified = true

					// ioutil.ReadFile(path)
					content, err := os.ReadFile(path)
					if err != nil {
						logp.Err("can not read file %s", path)
						return nil
					}

					// 采集
					event := beat.Event{
						Timestamp: now,
						Fields: common.MapStr{
							"type":    "job",
							"content": string(content),
						},
					}
					bt.client.Publish(event)

				}

			} else {
				// 没有采集过, 那么就采集它
				bt.registrar[path] = now
				modified = true

				content, err := os.ReadFile(path)
				if err != nil {
					logp.Err("can not read file %s", path)
					return nil
				}
				// 采集
				event := beat.Event{
					Timestamp: now,
					Fields: common.MapStr{
						"type":    "job",
						"content": string(content),
					},
				}
				bt.client.Publish(event)
			}

		}
		return nil
	})

	// todo 如果有修改, 则将 bt.registrar 写回文件
	if modified {
		// todo  写回文件
	}

	// 如果有错, 那么报错
	if err != nil {
		logp.Err(err.Error())
	}

	// 更新时间
	bt.lastIndexTime = now

	// event := beat.Event{
	// 	Timestamp: now,
	// 	Fields: common.MapStr{
	// 		"type": "job",
	// 	},
	// }
	// bt.client.Publish(event)
}

func loadRegistrar() map[string]time.Time {

	m := map[string]time.Time{}

	return m
}
