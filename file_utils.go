package main

import (
	"bufio"
	"os"
)

func CountLines(filePath string) (int, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// 创建bufio.Scanner来逐行读取文件，并设置最大token大小
	const maxTokenSize = 1 << 20 // 1MB
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	scanner.Buffer(nil, maxTokenSize) // 设置缓冲区大小

	lineCount := 0

	// 逐行扫描文件
	for scanner.Scan() {
		lineCount++
	}

	// 检查扫描过程中是否有错误发生
	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return lineCount, nil
}
