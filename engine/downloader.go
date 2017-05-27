package engine

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/twoflyliu/novel2/tool"
)

// Downloader提供从互联网上下载制定url页面的的能力
type Downloader interface {
	// url - 要下载的地址
	// maxRetries - 当下载失败要进行重新下载的次数
	Download(url string, maxRetries int) (string, error)
}

type HttpDownloader struct{}

func (downloader *HttpDownloader) Download(url string, maxRetries int) (fullPage string, err error) {
	var response *http.Response
	// 如果下载不成功，则重复下载
	for maxRetries > 0 {
		maxRetries--
		response, err = http.Get(url)
		// log.Debugf("URL: %s, StatusLine: %s, Error: %v\n", url, response.Status, err)

		if err != nil {
			return
		}
		if response.Status == "200 OK" {
			break
		} else {
			err = fmt.Errorf("Status Line: %s", response.Status)
			response.Body.Close() //在循环体内部调用defer是不会自动执行的
		}
	}
	defer response.Body.Close()
	if maxRetries == 0 {
		return
	}

	buf, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return downloader.Download(url, maxRetries-1) //这儿当下载失败也是要养尝试重新下载
	}

	// 统一buf为utf编码
	fullPage = string(buf)
	charset := ExtractCharset(fullPage)

	charset = strings.TrimSpace(charset)
	charset = strings.ToLower(charset)

	// 这样写应该gbk, gb2312...
	if len(charset) != 0 && charset != "utf-8" && charset != "utf8" {
		fullPage, err = tool.ConvertToUTF8(fullPage, charset)
		if err != nil {
			err = fmt.Errorf("ConvertToUTF8 fail: %v, origin charset: %s, url: %s", err, charset, url)
		}
	}

	return
}

func NewDefaultDownloader() Downloader {
	return new(HttpDownloader)
}
