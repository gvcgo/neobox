package main

import (
	"fmt"
	"net/http"

	_ "github.com/moqsien/neobox/pkgs/conf"
	_ "github.com/moqsien/neobox/pkgs/proxy"
	"github.com/moqsien/neobox/pkgs/utils"
)

func Verify(httpClient *http.Client) bool {
	resp, err := httpClient.Head("https://www.google.com")
	if err != nil || resp == nil || resp.Body == nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func main() {
	// cnf := conf.GetDefaultNeoConf()
	// f := proxy.NewProxyFetcher(cnf)
	// f.Download()
	// f.DecryptAndLoad()
	client, _ := utils.GetHttpClient(2088, 10)
	r := Verify(client)
	fmt.Println(r)
}
