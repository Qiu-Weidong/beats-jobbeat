package beater

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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

	if !strings.HasSuffix(c.RegistrarPath, ".json") {
		// 需要添加文件名
		c.RegistrarPath = filepath.Join(c.RegistrarPath, "registrar.json")
	}
	// 可以使用 1m 来表示分钟 1h 来表示小时
	logp.Info("config period = %f, config registrar %s", c.Period.Seconds(), c.RegistrarPath)

	bt := &jobbeat{
		done:   make(chan struct{}),
		config: c,

		// 初始化 lastIndexTime
		lastIndexTime: time.Now(),

		// 加载 registrar
		registrar: loadRegistrar(c.RegistrarPath),
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

		for _, p := range bt.config.Path {
			bt.collectJobs(p, b)
		}

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
					bt.sendJob(path, b, modTime)
					modified = true
				}

			} else {
				// 没有采集过, 那么就采集它
				bt.sendJob(path, b, modTime)
				modified = true
			}

		}
		return nil
	})

	// todo 如果有修改, 则将 bt.registrar 写回文件
	if modified {
		// todo  写回文件
		saveRegistrar(bt.config.RegistrarPath, bt.registrar)
	} else {
		logp.Info("no file added at this period.")
	}

	// 如果有错, 那么报错
	if err != nil {
		logp.Err(err.Error())
	}

	// 更新时间
	bt.lastIndexTime = now
}

func (bt *jobbeat) sendJob(path string, b *beat.Beat, modtime time.Time) {
	now := time.Now()

	bt.registrar[path] = now
	content, err := os.ReadFile(path)
	if err != nil {
		logp.Err("can not read file %s", path)
		return
	}

	// 解析 xml
	var data job
	err = xml.Unmarshal(content, &data)
	if err != nil {
		logp.Err("parse job error: %v", err)
		return
	}

	// 采集
	event := beat.Event{
		Timestamp: now,
		Fields: common.MapStr{
			"type":     "job",
			"job":      data,
			"filename": filepath.Base(path),
			"path":     path,
			"modtime":  modtime,
		},
	}
	bt.client.Publish(event)
}

type item struct {
	Path          string    `json:"path"`
	CollectedTime time.Time `json:"collected_time"`
}

func loadRegistrar(registrarPath string) map[string]time.Time {

	m := map[string]time.Time{}

	// todo 加载 registrar
	content, err := os.ReadFile(registrarPath)
	if err != nil {
		return m
	}
	var items []item
	err = json.Unmarshal(content, &items)
	if err != nil {
		return m
	}

	for _, item := range items {
		// 按道理应该使用更新的时间
		m[item.Path] = item.CollectedTime
	}

	return m
}

func saveRegistrar(registrarPath string, m map[string]time.Time) {

	// 需要先判断目录是否存在
	dir := filepath.Dir(registrarPath)
	if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		// 首先创建目录
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			logp.Err("can not mkdir %s", dir)
			return
		}
	}
	file, err := os.Create(registrarPath)
	if err != nil {
		logp.Err("can not open file %s", registrarPath)
		return
	}
	defer file.Close()

	var items []item
	// encoder := json.NewEncoder(file)
	for key, value := range m {
		items = append(items, item{Path: key, CollectedTime: value})
	}
	encoder := json.NewEncoder(file)
	err = encoder.Encode(items)

	if err != nil {
		logp.Err("fail to write registrar.")
	}

}
