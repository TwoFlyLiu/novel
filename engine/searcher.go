package engine

import "fmt"

import "github.com/twoflyliu/novel3/tool"

type Searcher interface {
	Search(name string) []string
}

type SearcherItem struct {
	fmtSearchString string //格式化搜索字符串，可以使用这个字符串合成合法的站内搜索url
	escape          bool   //url路径中的中文是否进行转义
	gbk             bool   //表示站内的网站是否只支持gbk

	//所以写他来从extracter的管理者中来进行获取
	host string //该网站的域名，写他是种折中的设计方案，最好应该是extracter，但是extracter应该是单例的对象，
}

// 表示站内搜索
type SiteSearcher struct {
	items []*SearcherItem
}

var GlobalSiteSearcher *SiteSearcher

func (ss *SiteSearcher) AddItem(fmtSearchString string, escape bool, gbk bool, host string) {
	ss.items = append(ss.items, &SearcherItem{fmtSearchString, escape, gbk, host})
}

func (ss *SiteSearcher) Search(name string) []string {
	result := make([]string, 0)
	for _, item := range ss.items {
		extracter := AutoSelectExtracter(item.host)
		searchURL := ss.mkSearchURL(item, name)
		searchContent, err := NewDefaultDownloader().Download(searchURL, MAX_RETRIES_COUNT)
		CheckError(err)

		if objURL, ok := extracter.ExtractObjURL(name, searchContent); ok {
			result = append(result, objURL)
		}
	}
	return result
}

func (ss *SiteSearcher) mkSearchURL(item *SearcherItem, name string) string {
	var err error
	if item.gbk {
		name, err = tool.ConvertUTF8ToGBK(name)
		CheckError(err)
	}
	if item.escape {
		name = tool.EscapeString(name)
	}
	return fmt.Sprintf(item.fmtSearchString, name)
}

type NativeSearcher struct{}

func (ss *NativeSearcher) Search(name string) (result []string) {
	return
}

func init() {
	GlobalSiteSearcher = new(SiteSearcher)
	GlobalSiteSearcher.items = make([]*SearcherItem, 0)
}
