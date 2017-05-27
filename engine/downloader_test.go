package engine

import "testing"

type DownloaderTestData struct {
	url      string
	expected bool
}

func TestDefaultDownloader(t *testing.T) {
	datas := []DownloaderTestData{
		DownloaderTestData{"http://www.qu.la/book/24868/", true},
		DownloaderTestData{"http://www.baidu.com/", true},
		DownloaderTestData{"www.dummy.com", false},
	}
	downloader := NewDefaultDownloader()

	for _, data := range datas {
		fullPage, _ := downloader.Download(data.url, 5)
		if (len(fullPage) > 0) != data.expected {
			t.Errorf("TestDefaultDownloader fail: Test [%s] expected:[%v], actual:[%v]", data.url, data.expected, !data.expected)
		}
	}
}
