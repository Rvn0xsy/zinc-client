package main

import (
	"bufio"
	"context"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/schollz/progressbar/v3"
	"github.com/vjeantet/grok"
	client "github.com/zinclabs/sdk-go-zincsearch"
)

var (
	LogFile        = ""
	URL            = "http://localhost:4080"
	Username       = "admin"
	Password       = "admin"
	Index          = "index_name"
	LogFileLines   = 0
	Count          = 0
	ThreadNum      = 10
	DebugFileHanle *os.File
)

// DocumentStruct is a struct that represents a document to be indexed.
//
// It has two fields:
// - IndexName: a string representing the name of the index.
// - Document: a map[string]interface{} representing the document data.
type DocumentStruct struct {
	IndexName string                 `json:"indexName"` // The name of the index.
	Document  map[string]interface{} `json:"document"`  // The document data.
}

// init parses the command line arguments and initializes the necessary variables.
//
// This function sets the values of the LogFile, Username, Password, URL, Index, and ThreadNum variables based on the command line arguments passed.
// If no arguments are provided, default values are used.
// The function also creates a log file with a name derived from the Index and LogFile variables and opens it in append mode.
// If there is an error opening the log file, the program terminates.
// The function sets the output of the log package to a MultiWriter that writes to both the standard output and the log file.
// It sets the flags of the log package to include the date, time, and file name in the log output.
// Finally, it sets the output of the log package to the log file handle.
func init() {
	// 解析命令行
	flag.StringVar(&LogFile, "file", "access.log", "log filepath")
	flag.StringVar(&Username, "username", "admin", "zincsearch username")
	flag.StringVar(&Password, "password", "admin", "zincsearch password")
	flag.StringVar(&URL, "url", "http://localhost:4080", "zincsearch api host")
	flag.StringVar(&Index, "index", "index_name", "zincsearch index name")
	flag.IntVar(&ThreadNum, "thread", 4, "thread number")
	flag.Parse()
	logFileName := Index + "_" + filepath.Base(LogFile)
	DebugFileHanle, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
	multiWriter := io.MultiWriter(os.Stdout, DebugFileHanle)
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(DebugFileHanle)
	LogFileLines, err = CountLines(LogFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Log File Lines: ", LogFileLines, "\n")
}

// main is the entry point of the program.
//
// It starts by printing a start message to the log.
// The DebugFileHanle is closed at the end of the function.
//
// The program then initializes a WaitGroup and a pool of goroutines.
// The number of goroutines is determined by the ThreadNum variable.
// Each goroutine executes a function that creates a document using the
// IndexName and Document fields of a DocumentStruct.
//
// The program opens the LogFile and creates a bufio.Reader to read its contents.
// It then reads the file line by line and processes each line.
// Each line is parsed using a grok pattern to extract the username, password, and url.
// The extracted values are then used to create a document map.
// The document is then passed to the pool of goroutines for processing.
//
// The program waits for all goroutines to finish using the WaitGroup.
// Finally, it logs the total execution time.
func main() {
	log.Println("-------- Start -----------")
	defer DebugFileHanle.Close()
	startTime := time.Now()
	bar := progressbar.Default(int64(LogFileLines))
	// 设置并发
	var wg sync.WaitGroup
	grountinePool, _ := ants.NewPoolWithFunc(ThreadNum, func(i interface{}) {
		Doc := i.(DocumentStruct)
		createDocument(Doc.IndexName, Doc.Document)
		wg.Done()
	})

	defer grountinePool.Release()

	// 打开文件
	file, err := os.Open(LogFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 创建bufio.Reader
	reader := bufio.NewReader(file)

	defer grountinePool.ReleaseTimeout(5 * time.Second)
	// 逐行读取文件
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break // 读取到文件末尾或发生错误
		}
		// 处理每一行数据
		g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
		values, _ := g.Parse("(?P<url>\\S+) (?P<username>\\S+):(?P<password>\\S+)", line)

		document := map[string]interface{}{
			"username": values["username"],
			"password": values["password"],
			"url":      values["url"],
			"message":  strings.Trim(line, "\r\n"),
		}
		wg.Add(1)
		grountinePool.Invoke(DocumentStruct{Index, document})
		bar.Add(1)
	}
	wg.Wait()
	endTime := time.Now()
	log.Printf("[*] cost: %s \n", endTime.Sub(startTime))

}

// createDocument creates a document in the specified index using the provided
// document map. It takes in two parameters:
//
// - index: a string representing the index where the document will be created.
// - document: a map[string]interface{} representing the document to be created.
//
// The function returns nothing.
func createDocument(index string, document map[string]interface{}) {
	ctx := context.WithValue(context.Background(), client.ContextBasicAuth, client.BasicAuth{
		UserName: Username,
		Password: Password,
	})
	configuration := client.NewConfiguration()
	configuration.Servers = client.ServerConfigurations{
		client.ServerConfiguration{
			URL: URL,
		},
	}

	apiClient := client.NewAPIClient(configuration)
	resp, r, err := apiClient.Document.Index(ctx, index).Document(document).Execute()
	if err != nil {
		log.Fatalf("Error when calling `Document.Index`: %v\n %v \n", err, r)
	}
	Count = Count + 1
	// response from `Index`: MetaHTTPResponseID
	log.Printf("[*] Response from `Document.Index`: %v => %d \n", resp.GetId(), Count)
}
