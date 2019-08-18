package tool

import (
	"bytes"
	"io/ioutil"
	"net/url"

	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func ConvertUTF8ToGBK(str string) (string, error) {
	data, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader([]byte(str)),
		simplifiedchinese.GB18030.NewEncoder()))
	return string(data), err
}

func ConvertGBKToUTF8(str string) (string, error) {
	data, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader([]byte(str)),
		simplifiedchinese.GB18030.NewDecoder()))
	return string(data), err
}

func ConvertToUTF8(str string, charset string) (string, error) {
	charset = strings.ToUpper(charset)
	if len(charset) >= 2 && charset[0:2] == "GB" {
		return ConvertGBKToUTF8(str)
	}
	return "", nil
}

// 将gbk中文转义为他们的以%ascii值形式
func EscapeString(gbk string) string {
	u, _ := url.Parse(gbk)
	return u.String()
}
