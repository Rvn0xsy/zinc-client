package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/vjeantet/grok"
	client "github.com/zinclabs/sdk-go-zincsearch"
)

var (
	LogFile  = ""
	URL      = "http://localhost:4080"
	Username = "admin"
	Password = "admin"
	Index    = "index_name"
	Count    = 0
)

func main() {
	flag.StringVar(&LogFile, "file", "access.log", "pass file")
	flag.StringVar(&Username, "username", "admin", "zincsearch username")
	flag.StringVar(&Password, "password", "admin", "zincsearch password")
	flag.StringVar(&URL, "url", "http://localhost:4080", "zincsearch api host")
	flag.StringVar(&Index, "index", "index_name", "zincsearch index name")
	flag.Parse()
	// 打开文件
	file, err := os.Open(LogFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 创建bufio.Reader
	reader := bufio.NewReader(file)

	// 逐行读取文件
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break // 读取到文件末尾或发生错误
		}
		// 处理每一行数据
		// fmt.Printl n(line)

		g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
		values, _ := g.Parse("(?P<url>\\S+) (?P<username>\\S+):(?P<password>\\S+)", line)

		// for k, v := range values {
		// 	fmt.Printf("%+15s: %s\n", k, v)
		// }
		document := map[string]interface{}{
			"username": values["username"],
			"password": values["password"],
			"url":      values["url"],
			"message":  line,
		}
		createDocument(Index, document)
	}
}

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
		fmt.Fprintf(os.Stderr, "Error when calling `Document.Index``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	Count = Count + 1
	// response from `Index`: MetaHTTPResponseID
	fmt.Fprintf(os.Stdout, "[*] Response from `Document.Index`: %v => %d \n", resp.GetId(), Count)
}
