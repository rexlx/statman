package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type StatsEndPoint struct {
	Run      bool   `json:"run"`
	Host     string `json:"host"`
	FileName string `json:"filename"`
	Port     int    `json:"port"`
}

type Stat struct {
	Value float64   `json:"value"`
	Time  time.Time `json:"time"`
	ID    string    `json:"id"`
}

func NewStatsEndPoint(host, filename string, port int) *StatsEndPoint {
	return &StatsEndPoint{
		Host:     host,
		FileName: filename,
		Port:     port,
	}
}

func (s *StatsEndPoint) PostStatHTTP(stat Stat) {
	url := fmt.Sprintf("http://%v:%v/%v", s.Host, s.Port, s.FileName)
	out, err := json.Marshal(stat)
	if err != nil {
		fmt.Println("(PostStatHTTP) Error marshalling stat:", err)
		return
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewReader(out))
	if err != nil {
		fmt.Println("(PostStatHTTP) Error creating request:", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("(PostStatHTTP) Error sending request:", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		fmt.Println("(PostStatHTTP) Error sending request, status code:", res.StatusCode)
	}

}
