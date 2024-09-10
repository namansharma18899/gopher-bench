package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
)

type HTTPBench struct {
	MethodName  
}

type Config struct {
	URL        string `json:"url"`
	Iterations int    `json:"iterations"`
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

func main() {
	config, err := loadConfig()
	if err != nil {
		panic(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"URL", "Latency", "Throughput(Mbps)"})

	for i := 0; i < config.Iterations; i++ {
		start := time.Now()
		resp, err := http.Get(config.URL)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		defer resp.Body.Close()

		elapsed := time.Since(start)
		latency := float64(elapsed.Milliseconds())

		body, _ := ioutil.ReadAll(resp.Body)
		size := len(body)
		throughput := float64(size) / latency / 125

		table.Append([]string{
			config.URL,
			fmt.Sprintf("%.2f ms", latency),
			fmt.Sprintf("%.2f Mbps", throughput),
		})
	}

	table.Render()
}
