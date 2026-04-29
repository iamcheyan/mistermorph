package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileStats 保存文件的统计信息
type FileStats struct {
	FileName string
	LineCount int
	CharCount int
	Error error
}

// worker 是一个工作协程，负责处理文件统计任务
func worker(ctx context.Context, taskChan <-chan string, resultChan chan<- FileStats, wg *sync.WaitGroup) {
	defer wg.Done()
	
	for {
		select {
		case <-ctx.Done():
			return
		case fileName, ok := <-taskChan:
			if !ok {
				return
			}
			
			stats, err := getFileStats(fileName)
			select {
			case <-ctx.Done():
				return
			case resultChan <- FileStats{
				FileName: fileName,
				LineCount: stats.lineCount,
				CharCount: stats.charCount,
				Error: err,
			}:
			}
		}
	}
}

// fileStatsInternal 保存内部文件统计数据
type fileStatsInternal struct {
	lineCount int
	charCount int
}

// getFileStats 统计单个文件的行数和字符数
func getFileStats(fileName string) (fileStatsInternal, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return fileStatsInternal{}, err
	}
	
	var lineCount, charCount int
	for _, b := range data {
		charCount++
		if b == '\n' {
			lineCount++
		}
	}
	
	// 处理没有换行符结尾的文件
	if charCount > 0 && data[charCount-1] != '\n' {
		lineCount++
	}
	
	return fileStatsInternal{
		lineCount: lineCount,
		charCount: charCount,
	}, nil
}

// scanTxtFiles 扫描当前目录下的所有 .txt 文件
func scanTxtFiles() ([]string, error) {
	var txtFiles []string
	
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && filepath.Ext(path) == ".txt" {
			txtFiles = append(txtFiles, path)
		}
		
		return nil
	})
	
	return txtFiles, err
}

func main() {
	// 创建可取消的 context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 扫描 .txt 文件
	txtFiles, err := scanTxtFiles()
	if err != nil {
		fmt.Printf("扫描文件失败: %v\n", err)
		os.Exit(1)
	}
	
	if len(txtFiles) == 0 {
		fmt.Println("未找到 .txt 文件")
		os.Exit(0)
	}
	
	// 创建通道
	const workerCount = 4
	taskChan := make(chan string, len(txtFiles))
	resultChan := make(chan FileStats, len(txtFiles))
	
	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(ctx, taskChan, resultChan, &wg)
	}
	
	// 发送任务到工作协程
	go func() {
		for _, fileName := range txtFiles {
			select {
			case <-ctx.Done():
				close(taskChan)
				return
			case taskChan <- fileName:
			}
		}
		close(taskChan)
	}()
	
	// 等待所有工作协程完成并关闭结果通道
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	// 收集并处理结果
	for result := range resultChan {
		if result.Error != nil {
			fmt.Printf("处理文件 %s 失败: %v\n", result.FileName, result.Error)
			cancel() // 取消所有任务
			
			// 等待所有协程退出以防止泄漏
			time.Sleep(100 * time.Millisecond)
			os.Exit(1)
		}
		
		fmt.Printf("文件 %s: 行数=%d, 字符数=%d\n", result.FileName, result.LineCount, result.CharCount)
	}
	
	fmt.Println("所有文件处理完成")
}
