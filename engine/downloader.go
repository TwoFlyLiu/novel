package engine

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/twoflyliu/novel/tool"
)

type Downloader interface {
	Download(url string, retries int) (string, error)
}

type TimeoutDownloader struct {
	timeout time.Duration
}

func NewTimeoutDownloader(timeout time.Duration) *TimeoutDownloader {
	return &TimeoutDownloader{timeout: timeout}
}

func (downloader *TimeoutDownloader) Download(url string, maxRetries int) (result string, err error) {
	return downloader.doDownloadAndRetryIfFail(url, maxRetries)
}

// maxRetries 表示下载失败， 重新尝试的次数
// 如果maxRetries = 0，那么就下载一次，如果等于1，那么如果下载失败，就会重新再下载一次
func (downloader *TimeoutDownloader) doDownloadAndRetryIfFail(url string, maxRetries int) (content string, err error) {
	// always download util success
	if maxRetries < 0 {
		for {
			content, err = downloader.doDownload(url)
			if err == nil {
				log.Debugf("Must downloader: url: [%v], err: [%v]", url, err)
				return
			} else {
				err = fmt.Errorf("ErrorType: %T, Error: %v", err, err)
				log.Debugf("Must downloader: url: [%v], err: [%v]", url, err)
			}
		}
	}

	// do retry download util maxRetries
	for maxRetries >= 0 {
		log.Debugf("Retry downloader: retry: [%v], url: [%v], err: [%v]", maxRetries, url, err)
		content, err = downloader.doDownload(url)
		if err == nil {
			break
		}
		maxRetries--
	}
	return
}

func (downloader *TimeoutDownloader) doDownload(url string) (page string, err error) {
	defer func() {
		if e := recover(); e != nil {
			var ok bool
			if err, ok = e.(error); !ok {
				err = fmt.Errorf("Error:%v", e)
			}
		}
	}()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Timeout:   downloader.timeout,
		Transport: tr,
	} //设置超时时间

	request, _ := http.NewRequest(http.MethodGet, url, nil)
	request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.80 Safari/537.36")

	resp, err := client.Do(request)
	CheckError(err)
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	CheckError(err)

	page, err = downloader.convertToUTF8(bytes)
	CheckError(err)
	return
}

func (downloader *TimeoutDownloader) convertToUTF8(bytes []byte) (fullPage string, err error) {
	// 统一buf为utf编码
	fullPage = string(bytes)
	charset := ExtractCharset(fullPage)

	charset = strings.TrimSpace(charset)
	charset = strings.ToLower(charset)

	// 这样写应该gbk, gb2312...
	if len(charset) != 0 && charset != "utf-8" && charset != "utf8" {
		fullPage, err = tool.ConvertToUTF8(fullPage, charset)
		if err != nil {
			err = fmt.Errorf("ConvertToUTF8 fail: %v, origin charset: %s", err, charset)
		}
	}
	return
}

func NewDefaultDownloader() Downloader {
	return NewTimeoutDownloader(5 * time.Second)
}
