package engine

import "testing"

func TestExtractCharset(t *testing.T) {
	downloader := NewDefaultDownloader()
	fullPage, err := downloader.Download("http://www.37zw.net/", 5)
	if err != nil {
		t.Errorf("TestExtractCharset: %s!", "Download www.37w.net fail!")
	}
	expectedCharset := "gbk"
	charset := ExtractCharset(fullPage)
	if expectedCharset != charset {
		t.Errorf("TestExtractCharset: expected [%s], but got [%s]", expectedCharset, charset)
	}
}
