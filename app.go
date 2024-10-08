package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

type Config struct {
	URL        string `json:"url"`
	Iterations int    `json:"iterations"`
}

func init() {
	log.Formatter = new(logrus.JSONFormatter)
	log.Out = os.Stdout
	log.Level = logrus.InfoLevel
}

func loadConfig() (*Config, error) {
	file, err := os.Open("config.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &Config{}
	if err := json.NewDecoder(file).Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}

func benchApp(config Config, ch chan [2]float64, theadNum int) {
	start := time.Now()
	resp, err := http.Get(config.URL)
	if err != nil {
		log.WithFields(logrus.Fields{
			"threadNum": theadNum,
		}).Error("Error while hitting the Endpoint ", config.URL)
		fmt.Println("Error:", err)
	}
	log.WithFields(logrus.Fields{
		"threadNum": theadNum,
	}).Info("GET Request Successful ", config.URL)
	defer resp.Body.Close()
	elapsed := time.Since(start)
	latency := float64(elapsed.Milliseconds())

	body, _ := ioutil.ReadAll(resp.Body)
	size := len(body)
	throughput := float64(size) / latency / 125
	var arr [2]float64
	arr[0] = latency
	arr[1] = throughput
	ch <- arr
}

func main() {
	config, err := loadConfig()
	if err != nil {
		panic(err)
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"URL", "Latency", "Throughput(Mbps)"})
	threads := 10
	ch := make(chan [2]float64, threads)
	var wg sync.WaitGroup
	for i := 1; i <= threads; i++ {
		wg.Add(1)
		go func(config Config, ch chan [2]float64) {
			benchApp(config, ch, i)
			defer wg.Done()
		}(*config, ch)
	}
	go func() {
		wg.Wait()
		defer close(ch)
	}()

	for data := range ch {
		table.Append([]string{
			config.URL,
			fmt.Sprintf("%.2f ms", data[0]),
			fmt.Sprintf("%.2f Mbps", data[1]),
		})
	}
	table.Render()
}
