package tool

import (
	"net/url"

	iconv "github.com/djimenez/iconv-go"
)

func ConvertUTF8ToGBK(str string) (string, error) {
	return iconv.ConvertString(str, "utf-8", "gbk")
}

func ConvertGBKToUTF8(str string) (string, error) {
	return iconv.ConvertString(str, "gbk", "utf-8")
}

func ConvertToUTF8(str string, charset string) (string, error) {
	return iconv.ConvertString(str, charset, "utf-8")
}

// 将gbk中文转义为他们的以%ascii值形式
func EscapeString(gbk string) string {
	u, _ := url.Parse(gbk)
	return u.String()
}
