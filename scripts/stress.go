package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type Response struct {
	URI          string
	StatusCode   int
	StartTime    time.Time
	ResponseTime time.Duration
}

func (r *Response) String() string {
	return fmt.Sprintf("%v %v", r.StatusCode, r.ResponseTime)
}

func (r *Response) toJSON() []byte {
	b, _ := json.Marshal(r)
	return b
}

func get(client *http.Client, uri string, ch chan *Response) {
	startTime := time.Now()
	req, _ := http.NewRequest("GET", uri, nil)
	//req.Header.Add("Content-Type", "Application/json")
	resp, err := client.Do(req)
	if err == nil {
		response := &Response{
			URI:          uri,
			StatusCode:   resp.StatusCode,
			ResponseTime: time.Since(startTime),
			StartTime:    startTime,
		}
		// 閉じないとfd足りなくなった
		defer resp.Body.Close()
		ch <- response
	} else {
		ch <- nil
	}
}

func dumpResponse(ctx context.Context, logFilePath string, ch chan *Response) {
	file, err := os.Create(logFilePath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer file.Close()
	var responseTimeTotal int64
	var successCount = 0
	var errorCount = 0
	totalStartTime := time.Now()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)

	resultFunc := func() {
		duration := time.Since(totalStartTime)
		fmt.Println("stopping...")
		fmt.Println(fmt.Sprintf("requested: %d", successCount+errorCount))
		rate := float64(successCount) / duration.Seconds()
		responseTime := float32(responseTimeTotal) / float32(1e6*successCount)
		fmt.Println(fmt.Sprintf("error: %d, rate %f req/sec, time: %f msec/req", errorCount, rate, responseTime))
	}

	for {
		select {
		case r := <-ch:
			log := r.toJSON()
			file.Write(log)
			file.WriteString("\n")

			if r == nil {
				errorCount++
				continue
			} else if r.StatusCode == 200 {
				successCount++
				responseTimeTotal += r.ResponseTime.Nanoseconds()
			} else {
				errorCount++
			}

			if (successCount+errorCount)%100 == 0 {
				fmt.Println(fmt.Sprintf("%d requested. error: %d", successCount+errorCount, errorCount))
			}
		case <-sigCh:
			resultFunc()
			return
		case <-ctx.Done():
			resultFunc()
			return
		}
	}
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	lines := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, nil
}

func main() {
	// 叩くurlをファイルから読み込む
	urls, err := readLines("./urls.txt")
	if err != nil {
		fmt.Println("urls.txt is required.")
		return
	}
	const log = "./response.log"
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	// 結果送信用のchan
	ch := make(chan *Response)
	concurrency, err := strconv.Atoi(os.Getenv("CONCURRENCY"))
	if err != nil {
		fmt.Println("CONCURRENCY must be integer")
		return
	}
	fmt.Println(fmt.Sprintf("stress start. concurrency: %d", concurrency))

	// init rand seed
	rand.Seed(time.Now().UnixNano())
	// timer
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	for i := 0; i < concurrency; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					n := rand.Intn(len(urls))
					get(client, urls[n], ch)
				}
			}
		}()
	}

	dumpResponse(ctx, log, ch)
}
