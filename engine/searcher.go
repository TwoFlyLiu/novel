package engine

import "fmt"
import "os"

import "bufio"

import "github.com/twoflyliu/novel/tool"

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
	items       []*SearcherItem
	ignoredHost []string
}

var GlobalSiteSearcher *SiteSearcher

func (ss *SiteSearcher) AddItem(fmtSearchString string, escape bool, gbk bool, host string) {
	ss.items = append(ss.items, &SearcherItem{fmtSearchString, escape, gbk, host})
}

func (ss *SiteSearcher) RemoveItem(host string) {
	index := -1

	if len(ss.items) == 1 { //防止被删光了
		return
	}

	for i := 0; i < len(ss.items); i++ {
		if host == ss.items[i].host {
			index = i
			break
		}
	}

	if index != -1 {
		ss.items = append(ss.items[0:index], ss.items[index+1:]...) //slice 删除某个位置上的元素
	}

	// 将要移除的软件源保存到本地文件中
	if !ss.containsIgnoredHost(host) {
		ss.addIgnoredHost(host)
		ss.appendIgnoreHostToNative(host)
	}
}

func (ss *SiteSearcher) RemoveAllIgnoredHost() {
	ss.ignoredHost = []string{}
}

func (ss *SiteSearcher) containsIgnoredHost(host string) bool {
	for i := 0; i < len(ss.ignoredHost); i++ {
		if host == ss.ignoredHost[i] {
			return true
		}
	}
	return false
}

func (ss *SiteSearcher) appendIgnoreHostToNative(host string) {
	ignoredHostFilePath := config.BaseDirName() + "/" + config.IgnoredHostFileName()
	file, err := os.OpenFile(ignoredHostFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	CheckError(err)
	defer file.Close()
	file.Write([]byte(host))
	file.Write([]byte("\n"))
}

func (ss *SiteSearcher) loadIgnoredHosts() {
	log.Debug("Load Ignored hosts from native file")
	ignoredHostFilePath := config.BaseDirName() + "/" + config.IgnoredHostFileName()
	file, err := os.Open(ignoredHostFilePath)
	if err != nil {
		log.Debugf("%q not exist", ignoredHostFilePath)
		return
	}
	defer file.Close()

	buff := bufio.NewReader(file)
	for {
		host, err := buff.ReadString('\n')
		if err != nil {
			break
		}
		host = host[0 : len(host)-1]
		log.Debug("Ignored host:", host)
		ss.ignoredHost = append(ss.ignoredHost, host) //去除换行符
		ss.RemoveItem(host)                           //移除不好的搜索源
	}
}

func (ss *SiteSearcher) addIgnoredHost(host string) {
	ss.ignoredHost = append(ss.ignoredHost, host)
}

func (ss *SiteSearcher) Search(name string) []string {
	result := make([]string, 0)
	downloader := NewDefaultDownloader()

	ch := make(chan string, len(ss.items))

	for _, item := range ss.items {

		go func(item *SearcherItem, ch chan string) {
			extracter := AutoSelectExtracter(item.host)
			searchURL := ss.mkSearchURL(item, name)
			searchContent, err := downloader.Download(searchURL, MAX_RETRIES_COUNT)
			if err != nil {
				ch <- "none"
			}

			if err != nil {
				ch <- "none"
			} else if objURL, ok := extracter.ExtractObjURL(name, searchContent); ok {
				objURL = tool.FixUrl(objURL, searchURL)
				ch <- objURL
				log.Debugf("search url: %s -> %s", searchURL, objURL)
				return
			} else {
				ch <- "none"
			}

			log.Debugf("search url: %s -> none", searchURL)
		}(item, ch)
	}

	for i := 0; i < len(ss.items); i++ {
		objURL := <-ch
		if objURL != "none" {
			result = append(result, objURL)
			break //只要找到一个就可以
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
	GlobalSiteSearcher.ignoredHost = make([]string, 0)
}
