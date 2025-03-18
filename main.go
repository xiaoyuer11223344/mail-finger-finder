package main

import (
	"bufio"
	"fmt"
	"mailfinger/query/pop3"
	"mailfinger/query/pop3s"
	"mailfinger/query/smtp"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const (
	workersNum = 5 // 并发工作协程数量
)

func main() {
	// 读取目标文件
	targets, err := readTargets("targets.txt")
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		return
	}

	// 创建工作任务通道
	jobs := make(chan string, len(targets))
	results := make(chan bool, len(targets))

	var wg sync.WaitGroup
	var completed int32 // 原子操作计数器

	// 启动工作协程池
	for i := 0; i < workersNum; i++ {
		wg.Add(1)
		go worker(jobs, results, &wg, &completed)
	}

	// 添加任务到通道
	go func() {
		for _, target := range targets {
			jobs <- target
		}
		close(jobs)
	}()

	// 启动进度监控
	go progressMonitor(&completed, len(targets))

	// 等待所有任务完成
	wg.Wait()
	close(results)
}

// 读取目标文件内容
func readTargets(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var targets []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		targets = append(targets, scanner.Text())
	}
	return targets, scanner.Err()
}

// 工作协程
func worker(jobs <-chan string, results chan<- bool, wg *sync.WaitGroup, completed *int32) {
	defer wg.Done()
	for target := range jobs {
		// 执行查询操作（示例）
		smtp.DoQuery(target)
		pop3.DoQuery(target)
		pop3s.DoQuery(target)
		// 原子操作更新计数器
		atomic.AddInt32(completed, 1)
		results <- true
	}
}

// 实时进度监控
func progressMonitor(completed *int32, total int) {
	for {
		current := atomic.LoadInt32(completed)
		progress := float64(current) / float64(total) * 100
		fmt.Printf("\r处理进度: %.2f%% (%d/%d)\n", progress, current, total)
		if current >= int32(total) {
			fmt.Println("\n所有任务已完成!")
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}
