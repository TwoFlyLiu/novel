package common //包名称和目录名称未必要一样

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/twoflyliu/novel/engine"
)

type ExtracterPattern struct {
	NovelNamePattern                 string
	NovelAuhtorPattern               string
	NovelIconUrlPattern              string
	NovelLastUpdateTimePattern       string
	NovelNewestChapterNamePattern    string
	NovelDescriptionPattern          string
	MenuListPattern                  string
	MenuItemPattern                  string
	ChapterTitlePattern              string
	ChapterContentPattern            string
	BrElementPattern                 string
	EscapeElementPattern             string
	DivElementPattern                string
	ScriptElementPattern             string
	SearchFormPattern                string
	SearchFormMethodAttributePattern string
	SearchFormHiddenFieldPattern     string
	SearchFormShowFieldPattern       string
	SearchObjUrlPattern              string
}

type RegistrySearch struct {
	SearchUrlFmtStr string
	GBKEncoding     bool
	NeedEscape      bool
	Host            string
}

type RegistryExtracter struct {
	HostPattern  string
	ExtracterRef string
}

type SitesConfig struct {
	ExtracterMap          map[string]ExtracterPattern
	RegistrySearchList    []RegistrySearch
	RegistryExtracterList []RegistryExtracter
}

type ConfigExtracter struct {
	novelNamePatternSubMatch                  *regexp.Regexp
	novelAuthorPatternSubMatch                *regexp.Regexp
	novelIconUrlPatternSubMatch               *regexp.Regexp
	novelLastUpdateTimePatternSubMatch        *regexp.Regexp
	novelNewestLastChapterNamePatternSubMatch *regexp.Regexp
	novelDescriptionPatternSubMatch           *regexp.Regexp
	menuListPatternFind                       *regexp.Regexp
	menuItemPatternSubMatch                   *regexp.Regexp
	chapterTitlePatternSubMatch               *regexp.Regexp
	chapterContentPatternSubMatch             *regexp.Regexp
	brPatternReplaceNewLine                   *regexp.Regexp
	escapePatternRemove                       *regexp.Regexp
	divPatternRemove                          *regexp.Regexp
	scriptPatternRemove                       *regexp.Regexp
	searchFormFind                            *regexp.Regexp
	searchFormActionMethodSubmatch            *regexp.Regexp
	searchFormHiddenValueSubmatch             *regexp.Regexp
	searchFormNameFieldSubmatch               *regexp.Regexp
	searchObjUrlPattern                       string
}

func NewConfigExtracter(extracterName string,
	novelNamePattern string,
	novelAuthorPattern string,
	novelIconUrlPattern string,
	novelLastUpdateTimePattern string,
	novelNewestChapterNamePattern string,
	novelDescriptionPattern string,
	menuListPattern string,
	menuItemPattern string,
	chapterTitle string,
	chapterContentPattern string,
	brElementPattern string,
	escapeElementPattern string,
	divElementPattern string,
	scriptElementPattern string,
	searchFormPattern string,
	searchFormMethodAttributePattern string,
	searchFormHiddenFieldPattern string,
	searchFormShowFieldPattern string,
	searchObjUrlPattern string) *ConfigExtracter {
	var e ConfigExtracter

	e.novelNamePatternSubMatch = mustCompilePattern(extracterName, "NovelNamePattern", novelNamePattern)
	e.novelAuthorPatternSubMatch = mustCompilePattern(extracterName, "NovelAuthorPattern", novelAuthorPattern)
	e.novelIconUrlPatternSubMatch = mustCompilePattern(extracterName, "NovelIconUrlPattern", novelIconUrlPattern)
	e.novelLastUpdateTimePatternSubMatch = mustCompilePattern(extracterName, "NovelLastUpdateTimePattern", novelLastUpdateTimePattern)
	e.novelNewestLastChapterNamePatternSubMatch = mustCompilePattern(extracterName, "NovelNewestChapterNamePattern", novelNewestChapterNamePattern)
	e.novelDescriptionPatternSubMatch = mustCompilePattern(extracterName, "NovelDescriptionPattern", novelDescriptionPattern)
	e.menuListPatternFind = mustCompilePattern(extracterName, "MenuListPattern", menuListPattern)
	e.menuItemPatternSubMatch = mustCompilePattern(extracterName, "MenuItemPattern", menuItemPattern)
	e.chapterTitlePatternSubMatch = mustCompilePattern(extracterName, "ChapterTitlePattern", chapterTitle)
	e.chapterContentPatternSubMatch = mustCompilePattern(extracterName, "ChapterContentPattern", chapterContentPattern)
	e.brPatternReplaceNewLine = mustCompilePattern(extracterName, "BrElementPattern", brElementPattern)
	e.escapePatternRemove = mustCompilePattern(extracterName, "EscapeElementPattern", escapeElementPattern)
	e.divPatternRemove = mustCompilePattern(extracterName, "DivElementPattern", divElementPattern)
	e.scriptPatternRemove = mustCompilePattern(extracterName, "ScriptElementPattern", scriptElementPattern)
	e.searchFormFind = mustCompilePattern(extracterName, "SearchFormPattern", searchFormPattern)
	e.searchFormActionMethodSubmatch = mustCompilePattern(extracterName, "SearchFormMethodAttributePattern", searchFormMethodAttributePattern)
	e.searchFormHiddenValueSubmatch = mustCompilePattern(extracterName, "SearchFormHiddenFieldPattern", searchFormHiddenFieldPattern)
	e.searchFormNameFieldSubmatch = mustCompilePattern(extracterName, "SearchFormShowFieldPattern", searchFormShowFieldPattern)

	if searchObjUrlPattern == "" {
		fmt.Fprintf(os.Stderr, "%q.%q cannot be empty\n", extracterName, "SearchObjUrlPattern")
		os.Exit(1)
	} else {
		e.searchObjUrlPattern = searchObjUrlPattern
	}

	return &e
}

func mustCompilePattern(extracterName string, patternName string, patternStr string) *regexp.Regexp {
	pattern, err := regexp.Compile(patternStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Comple %q.%q=%q failed:%s\n", extracterName, patternName, patternStr, err.Error())
		os.Exit(1)
	}
	return pattern
}

const (
//CHINESE_SEC_STR                              = "："      //中文分号字符
//CHINESE_SEC_LEN                              = len("：") //中文分号长度
)

// 从fullPage中提取出小说名称
func (e *ConfigExtracter) ExtractNovelName(fullPage string) (name string) {
	matches := e.novelNamePatternSubMatch.FindStringSubmatch(fullPage)
	if len(matches) > 1 {
		name = matches[1]
	}
	return
}

// 从fullPage从提取出小说最后更新时间
func (e *ConfigExtracter) ExtractLastUpdateTime(fullPage string) (lastUpdateTime string) {
	matches := e.novelLastUpdateTimePatternSubMatch.FindStringSubmatch(fullPage)
	if len(matches) > 1 {
		lastUpdateTime = matches[1]
		if index := strings.Index(lastUpdateTime, CHINESE_SEC_STR); index != -1 { //去除不必要的前缀
			lastUpdateTime = lastUpdateTime[index+CHINESE_SEC_LEN:] //上面是中文，占用两个字节
		}
	}
	return
}

// 从fullPage从提取出小说作者
func (e *ConfigExtracter) ExtractNovelAuthor(fullPage string) (author string) {
	matches := e.novelAuthorPatternSubMatch.FindStringSubmatch(fullPage)
	if len(matches) > 1 {
		author = matches[1]
		if index := strings.Index(author, CHINESE_SEC_STR); index != -1 { ////去除不必要的前缀
			author = author[index+CHINESE_SEC_LEN:]
		}
	}
	return
}

// 从fullPage从提取出小说菜单列表
// 返回以[[url1, menu1], [ur2, menu2], ...]形式返回
func (e *ConfigExtracter) ExtractMenuList(fullPage string) (result [][]string) {
	result = make([][]string, 0)
	menuList := e.menuListPatternFind.FindString(fullPage)
	//fmt.Println("menuList len:", len(menuList))
	matches := e.menuItemPatternSubMatch.FindAllStringSubmatch(menuList, -1) //返回的是[][]string
	//fmt.Println("menuItem submatch len:", len(matches))
	for _, v := range matches {
		if len(v) > 2 {
			menu := make([]string, 0)
			menu = append(menu, v[1]) //添加的是url
			menu = append(menu, v[2]) //添加的是章节标题
			result = append(result, menu)
		}

	}
	return
}

// 从fullPage从提取出小说标题
func (e *ConfigExtracter) ExtractChapterTitle(fullPage string) (chapterTitle string) {
	matches := e.chapterTitlePatternSubMatch.FindStringSubmatch(fullPage)
	if len(matches) > 1 {
		chapterTitle = matches[1]
	}
	return
}

// 从fullPage从提取出小说内容
func (e *ConfigExtracter) ExtractChapterContent(fullPage string) (content string) {
	matches := e.chapterContentPatternSubMatch.FindStringSubmatch(fullPage)
	if len(matches) > 1 {
		content = matches[1]
		content = e.brPatternReplaceNewLine.ReplaceAllString(content, "\n") //替换<br/>为换行
		content = e.escapePatternRemove.ReplaceAllString(content, "")       //删除所有的形如&lt;&gt;的转义内容
		content = e.divPatternRemove.ReplaceAllString(content, "")          //删除所有的内容中的div元素标签
		content = e.scriptPatternRemove.ReplaceAllString(content, "")       //删除所有内容中的script元素标签
		content = strings.Replace(content, "nbsp;", " ", -1)                //替换掉转义字符
	}
	return
}

// 从url中提取出目录的url
func (extracter *ConfigExtracter) ExtractMenuURL(url string) (menuURL string) {
	if url[len(url)-4:] != "html" {
		menuURL = url
	} else {
		pos := strings.LastIndex(url, "/")
		menuURL = url[:pos+1]
	}
	return
}

func (e *ConfigExtracter) ExtractNewestLastChapterName(fullPage string) (newestLastChapterName string) {
	matches := e.novelNewestLastChapterNamePatternSubMatch.FindStringSubmatch(fullPage)
	if len(matches) > 1 {
		newestLastChapterName = matches[1]
	}
	return
}

func (e *ConfigExtracter) extractFormString(fullPage string) string {
	fmt.Print(fullPage)
	formString := e.searchFormFind.FindString(fullPage)
	return formString
}

// 从fullPage中提取出所有的搜索表单中的隐藏字段和值
func (e *ConfigExtracter) ExtractSearchFormHiddenValues(fullPage string) (values url.Values) {
	formString := e.extractFormString(fullPage)
	fmt.Println("len(formString):", len(formString))
	matches := e.searchFormHiddenValueSubmatch.FindAllStringSubmatch(formString, -1)
	for _, submatch := range matches {
		if len(submatch) > 2 {
			values.Set(submatch[1], submatch[2])
		}
	}
	return
}

// 从fullPage中提取出搜索表单的方法
func (e *ConfigExtracter) ExtractSearchFormMethodAndAction(fullPage string) (method string, actionUrl string) {
	matches := e.searchFormActionMethodSubmatch.FindStringSubmatch(fullPage)
	if len(matches) > 2 {
		actionUrl = matches[1]
		method = matches[2]
	}
	return
}

// 从fullPage中提取出搜索表单输入框的名称
func (e *ConfigExtracter) ExtractSearchFormSearchFieldName(fullPage string) (searchFieldName string) {
	formString := e.extractFormString(fullPage)
	submatch := e.searchFormNameFieldSubmatch.FindStringSubmatch(formString)
	if len(submatch) > 1 {
		searchFieldName = submatch[1]
	}
	return
}

func (e *ConfigExtracter) ExtractObjURL(name string, searchPage string) (string, bool) {
	fullSearchString := fmt.Sprintf(e.searchObjUrlPattern, name)
	searchObjUrlPattern := regexp.MustCompile(fullSearchString)
	matches := searchObjUrlPattern.FindStringSubmatch(searchPage)

	//fmt.Printf("pattern: %s", searchObjUrlPattern)
	//fmt.Printf("searchPage:%s", searchPage)

	if len(matches) > 1 {
		return matches[1], true
	}
	return "", false
}

func (e *ConfigExtracter) ExtractIconURL(menuPage string) (iconUrl string) {
	submatch := e.novelIconUrlPatternSubMatch.FindStringSubmatch(menuPage)
	if len(submatch) > 1 {
		iconUrl = submatch[1]
	}
	return
}

func (e *ConfigExtracter) ExtractNovelDescription(menuPage string) (description string) {
	submatch := e.novelDescriptionPatternSubMatch.FindStringSubmatch(menuPage)
	if len(submatch) <= 1 {
		return
	}

	description = submatch[1]

	description = strings.Replace(description, "<br>", "\n", -1)
	description = strings.Replace(description, "<br />", "\n", -1)
	description = strings.Replace(description, "<br/>", "\n", -1)
	description = strings.Replace(description, "&nbsp;", " ", -1)
	description = strings.TrimSpace(description)

	if left := strings.Index(description, "<p>"); left != -1 {
		if right := strings.Index(description, "</p>"); right != -1 {
			description = description[left+3 : right]
			description = strings.TrimSpace(description)
		}
	}

	return
}

// 使用一种自注册技术
func init() {
	// 加载本地的site.config
	var siteConfig SitesConfig

	fullpath := "./sites.json"
	file, err := os.Open(fullpath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "./sites.json not exist\n")
		os.Exit(2)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	err = json.Unmarshal(bytes, &siteConfig)

	if err != nil {
		fmt.Fprintf(os.Stderr, "sites.json corrupt!\n")
		os.Exit(2)
	}

	extracterMap := make(map[string]*ConfigExtracter)

	for n, e := range siteConfig.ExtracterMap {
		extracterMap[n] = NewConfigExtracter(
			n,
			e.NovelNamePattern,
			e.NovelAuhtorPattern,
			e.NovelIconUrlPattern,
			e.NovelLastUpdateTimePattern,
			e.NovelNewestChapterNamePattern,
			e.NovelDescriptionPattern,
			e.MenuListPattern,
			e.MenuItemPattern,
			e.ChapterTitlePattern,
			e.ChapterContentPattern,
			e.BrElementPattern,
			e.EscapeElementPattern,
			e.DivElementPattern,
			e.ScriptElementPattern,
			e.SearchFormPattern,
			e.SearchFormMethodAttributePattern,
			e.SearchFormHiddenFieldPattern,
			e.SearchFormShowFieldPattern,
			e.SearchObjUrlPattern,
		)
	}

	for _, s := range siteConfig.RegistrySearchList {
		engine.GlobalSiteSearcher.AddItem(s.SearchUrlFmtStr, s.NeedEscape, s.GBKEncoding, s.Host)
	}

	for _, e := range siteConfig.RegistryExtracterList {
		engine.RegisterExtracter(e.HostPattern, extracterMap[e.ExtracterRef])
	}
}
