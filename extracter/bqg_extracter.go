package common //包名称和目录名称未必要一样

import (
	//"fmt"

	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/twoflyliu/novel/engine"
)

type BiqugeExtracter struct {
	searchObjUrlPattern string
}

func NewExtracter(searchObjUrlPattern string) engine.Extracter {
	return &BiqugeExtracter{
		searchObjUrlPattern: searchObjUrlPattern,
	}
}

const (
	NOVEL_NAME_PATTERN_SUBMATCH                  = `\<div\s+id="info"[\s\S]+?\<h1\>([\s\S]+?)\</h1\>`
	NOVEL_AUTHOR_PATTERN_SUBMATCH                = `\<div\s+id="info"[\s\S]+?\<p\>([\s\S]+?)\</p\>`
	NOVEL_ICON_URL_PATTERN_SUBMATCH              = `\<div\s+id="fmimg"[\s\S]+?\<img.*?src="(.*?)"`
	NOVEL_LASTUPDATETIME_PATTERN_SUBMATCH        = `\<div\s+id="info"[\s\S]+?\<p\>[\s\S]+?\<p\>[\s\S]+?\<p\>([\s\S]+?)\</p\>`
	NOVEL_NEWESTLASTCHAPTERNAME_PATTERN_SUBMATCH = `\<div\s+id="info"[\s\S]+?\<p\>[\s\S]+?\<p\>[\s\S]+?\<p\>[\s\S]+?\<p\>[\s\S]+?\<a[\s\S]+?\>([\s\S]+?)\</a\>`
	NOVEL_DESCRIPTION_PATTERN_SUBMATCH           = `\<div\s+id="intro"\>([\s\S]+?)\</div\>`
	MENULIST_PATTERN_FIND                        = `\<div\s+id="list"[\s\S]+?\</div\>`
	MENUITEM_PATTERN_SUBMATCH                    = `\<a[\s\S]+?href="([\s\S]+?)"\s*\>([\s\S]+?)\</a\>`
	CHAPTERTITLE_PATTERN_SUBMATCH                = `\<div\s+class="bookname"[\s\S]+?\<h1\>([\s\S]+?)\</h1\>`
	CHAPTERCONTENT_PATTERN_SUBMATCH              = `\<div\s+id="content"\s*\>([\s\S]+?)\</div\>`
	BR_PATTERN_REPLACE_NEWLINE                   = `\<br\s*/\>`
	ESCAPE_PATTERN_REMOVE                        = `&[\s\S]+?;`
	DIV_PATTERN_REMOVE                           = `\<div[\s\S]+?\</div\>`
	SCRIPT_PATTERN_REMOVE                        = `\<script[\s\S]+?\</script\>`
	CHINESE_SEC_STR                              = "："      //中文分号字符
	CHINESE_SEC_LEN                              = len("：") //中文分号长度

	BQG_SEARCH_FORM_FIND                   = `\<form\s+id="bdcs-search-form"[\s\S]+?\</form\>`
	BQG_SEARCH_FORM_ACTION_METHOD_SUBMATCH = `\<form\s+id="bdcs-search-form"\s+action="([\s\S]+?)"\s+method="([\s\S]+?)"`
	BQG_SEARCH_FORM_HIDDEN_VALUE_SUBMATCH  = `\<input\s+name="(\w+)"\s+value="(\w+)"\s+type="hidden"\s*\>`
	BQG_SEARCH_FORM_NAME_FIELD_SUBMATCH    = `\<input[\s\S]+?name=(\w+)[\s\S]+?type="text"`
)

// 各个网站搜索URL
const (
	SEARCH_OBJ_URL_PATTERN_37ZW_STR = `\<a\s+href="([^"]+)"\s+target="_blank"\>\s*%s\s*\</a\>`                                                                                       //37中文网
	SEARCH_OBJ_URL_PATTERN_QU_STR   = `\<a\s+href="([^"]+)"\s+target="_blank"\>\s*%s\s*\</a\>`                                                                                       //笔趣阁
	SEARCH_OBJ_URL_PATTERN_XBQG_STR = `\<a\s+cpos="title"\s+href="([^"]+)" title="\s*%s\s*"\s+class="result-game-item-title-link"\s+target="_blank">[^<]*\<span\>\s*%[1]s\s*</span>` //新笔趣阁
)

var (
	novelNamePatternSubMatch                  *regexp.Regexp
	novelAuthorPatternSubMatch                *regexp.Regexp
	novelLastUpdateTimePatternSubMatch        *regexp.Regexp
	novelDescriptionPatternSubMatch           *regexp.Regexp
	novelIconUrlPatternSubMatch               *regexp.Regexp
	novelNewestLastChapterNamePatternSubMatch *regexp.Regexp
	menuListPatternFind                       *regexp.Regexp
	menuItemPatternSubMatch                   *regexp.Regexp
	chapterTitlePatternSubMatch               *regexp.Regexp
	chapterContentPatternSubMatch             *regexp.Regexp
	brPatternReplaceNewLine                   *regexp.Regexp
	escapePatternRemove                       *regexp.Regexp
	divPatternRemove                          *regexp.Regexp
	scriptPatternRemove                       *regexp.Regexp

	bqgSearchFormFind                 *regexp.Regexp
	bqgSearchFormActionMethodSubmatch *regexp.Regexp
	bqgSearchFormHiddenValueSubmatch  *regexp.Regexp
	bqgSearchFormNameFieldSubmatch    *regexp.Regexp
)

// 从fullPage中提取出小说名称
func (extracter *BiqugeExtracter) ExtractNovelName(fullPage string) (name string) {
	matches := novelNamePatternSubMatch.FindStringSubmatch(fullPage)
	if len(matches) > 1 {
		name = matches[1]
	}
	return
}

// 从fullPage从提取出小说最后更新时间
func (extracter *BiqugeExtracter) ExtractLastUpdateTime(fullPage string) (lastUpdateTime string) {
	matches := novelLastUpdateTimePatternSubMatch.FindStringSubmatch(fullPage)
	if len(matches) > 1 {
		lastUpdateTime = matches[1]
		if index := strings.Index(lastUpdateTime, CHINESE_SEC_STR); index != -1 { //去除不必要的前缀
			lastUpdateTime = lastUpdateTime[index+CHINESE_SEC_LEN:] //上面是中文，占用两个字节
		}
	}
	return
}

// 从fullPage从提取出小说作者
func (extracter *BiqugeExtracter) ExtractNovelAuthor(fullPage string) (author string) {
	matches := novelAuthorPatternSubMatch.FindStringSubmatch(fullPage)
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
func (extracter *BiqugeExtracter) ExtractMenuList(fullPage string) (result [][]string) {
	result = make([][]string, 0)
	menuList := menuListPatternFind.FindString(fullPage)
	//fmt.Println("menuList len:", len(menuList))
	matches := menuItemPatternSubMatch.FindAllStringSubmatch(menuList, -1) //返回的是[][]string
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
func (extracter *BiqugeExtracter) ExtractChapterTitle(fullPage string) (chapterTitle string) {
	matches := chapterTitlePatternSubMatch.FindStringSubmatch(fullPage)
	if len(matches) > 1 {
		chapterTitle = matches[1]
	}
	return
}

// 从fullPage从提取出小说内容
func (extracter *BiqugeExtracter) ExtractChapterContent(fullPage string) (content string) {
	matches := chapterContentPatternSubMatch.FindStringSubmatch(fullPage)
	if len(matches) > 1 {
		content = matches[1]
		content = brPatternReplaceNewLine.ReplaceAllString(content, "\n") //替换<br/>为换行
		content = escapePatternRemove.ReplaceAllString(content, "")       //删除所有的形如&lt;&gt;的转义内容
		content = divPatternRemove.ReplaceAllString(content, "")          //删除所有的内容中的div元素标签
		content = scriptPatternRemove.ReplaceAllString(content, "")       //删除所有内容中的script元素标签
		content = strings.Replace(content, "nbsp;", " ", -1)              //替换掉转义字符
	}
	return
}

// 从url中提取出目录的url
func (extracter *BiqugeExtracter) ExtractMenuURL(url string) (menuURL string) {
	if url[len(url)-4:] != "html" {
		menuURL = url
	} else {
		pos := strings.LastIndex(url, "/")
		menuURL = url[:pos+1]
	}
	return
}

func (extracter *BiqugeExtracter) ExtractNewestLastChapterName(fullPage string) (newestLastChapterName string) {
	matches := novelNewestLastChapterNamePatternSubMatch.FindStringSubmatch(fullPage)
	if len(matches) > 1 {
		newestLastChapterName = matches[1]
	}
	return
}

func (extracter *BiqugeExtracter) extractFormString(fullPage string) string {
	fmt.Print(fullPage)
	formString := bqgSearchFormFind.FindString(fullPage)
	return formString
}

// 从fullPage中提取出所有的搜索表单中的隐藏字段和值
func (extracter *BiqugeExtracter) ExtractSearchFormHiddenValues(fullPage string) (values url.Values) {
	formString := extracter.extractFormString(fullPage)
	fmt.Println("len(formString):", len(formString))
	matches := bqgSearchFormHiddenValueSubmatch.FindAllStringSubmatch(formString, -1)
	for _, submatch := range matches {
		if len(submatch) > 2 {
			values.Set(submatch[1], submatch[2])
		}
	}
	return
}

// 从fullPage中提取出搜索表单的方法
func (extracter *BiqugeExtracter) ExtractSearchFormMethodAndAction(fullPage string) (method string, actionUrl string) {
	matches := bqgSearchFormActionMethodSubmatch.FindStringSubmatch(fullPage)
	if len(matches) > 2 {
		actionUrl = matches[1]
		method = matches[2]
	}
	return
}

// 从fullPage中提取出搜索表单输入框的名称
func (extracter *BiqugeExtracter) ExtractSearchFormSearchFieldName(fullPage string) (searchFieldName string) {
	formString := extracter.extractFormString(fullPage)
	submatch := bqgSearchFormNameFieldSubmatch.FindStringSubmatch(formString)
	if len(submatch) > 1 {
		searchFieldName = submatch[1]
	}
	return
}

func (extracter *BiqugeExtracter) ExtractObjURL(name string, searchPage string) (string, bool) {
	fullSearchString := fmt.Sprintf(extracter.searchObjUrlPattern, name)
	searchObjUrlPattern := regexp.MustCompile(fullSearchString)
	matches := searchObjUrlPattern.FindStringSubmatch(searchPage)

	//fmt.Printf("pattern: %s", searchObjUrlPattern)
	//fmt.Printf("searchPage:%s", searchPage)

	if len(matches) > 1 {
		return matches[1], true
	}
	return "", false
}

func (extracter *BiqugeExtracter) ExtractIconURL(menuPage string) (iconUrl string) {
	submatch := novelIconUrlPatternSubMatch.FindStringSubmatch(menuPage)
	if len(submatch) > 1 {
		iconUrl = submatch[1]
	}
	return
}

func (extracter *BiqugeExtracter) ExtractNovelDescription(menuPage string) (description string) {
	submatch := novelDescriptionPatternSubMatch.FindStringSubmatch(menuPage)
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
	/*
		var err error
		novelNamePatternSubMatch, err = regexp.Compile(NOVEL_NAME_PATTERN_SUBMATCH)
		engine.CheckError(err)
		novelAuthorPatternSubMatch, err = regexp.Compile(NOVEL_AUTHOR_PATTERN_SUBMATCH)
		engine.CheckError(err)
		novelLastUpdateTimePatternSubMatch, err = regexp.Compile(NOVEL_LASTUPDATETIME_PATTERN_SUBMATCH)
		engine.CheckError(err)
		novelDescriptionPatternSubMatch, err = regexp.Compile(NOVEL_DESCRIPTION_PATTERN_SUBMATCH)
		engine.CheckError(err)
		novelNewestLastChapterNamePatternSubMatch, err = regexp.Compile(NOVEL_NEWESTLASTCHAPTERNAME_PATTERN_SUBMATCH)
		engine.CheckError(err)
		novelIconUrlPatternSubMatch, err = regexp.Compile(NOVEL_ICON_URL_PATTERN_SUBMATCH)
		engine.CheckError(err)
		menuListPatternFind, err = regexp.Compile(MENULIST_PATTERN_FIND)
		engine.CheckError(err)
		menuItemPatternSubMatch, err = regexp.Compile(MENUITEM_PATTERN_SUBMATCH)
		engine.CheckError(err)
		chapterTitlePatternSubMatch, err = regexp.Compile(CHAPTERTITLE_PATTERN_SUBMATCH)
		engine.CheckError(err)
		chapterContentPatternSubMatch, err = regexp.Compile(CHAPTERCONTENT_PATTERN_SUBMATCH)
		engine.CheckError(err)
		brPatternReplaceNewLine, err = regexp.Compile(BR_PATTERN_REPLACE_NEWLINE)
		engine.CheckError(err)
		escapePatternRemove, err = regexp.Compile(ESCAPE_PATTERN_REMOVE)
		engine.CheckError(err)
		divPatternRemove, err = regexp.Compile(DIV_PATTERN_REMOVE)
		engine.CheckError(err)
		scriptPatternRemove, err = regexp.Compile(SCRIPT_PATTERN_REMOVE)
		engine.CheckError(err)

		bqgSearchFormFind = regexp.MustCompile(BQG_SEARCH_FORM_FIND)
		bqgSearchFormActionMethodSubmatch = regexp.MustCompile(BQG_SEARCH_FORM_ACTION_METHOD_SUBMATCH)
		bqgSearchFormHiddenValueSubmatch = regexp.MustCompile(BQG_SEARCH_FORM_HIDDEN_VALUE_SUBMATCH)
		bqgSearchFormNameFieldSubmatch = regexp.MustCompile(BQG_SEARCH_FORM_NAME_FIELD_SUBMATCH)

		// 37中文网
		//注册提取器
		engine.RegisterExtracter("www.37zw.net", NewExtracter(SEARCH_OBJ_URL_PATTERN_37ZW_STR))

		////注册搜索器
		engine.GlobalSiteSearcher.AddItem("https://www.37zw.net/s/so.php?type=articlename&s=%s", true, true, "www.37zw.net")

		//// 笔趣阁
		//// 注册提取器
		engine.RegisterExtracter("www.qu.la", NewExtracter(SEARCH_OBJ_URL_PATTERN_QU_STR))

		//// 注册搜索器
		engine.GlobalSiteSearcher.AddItem("https://sou.xanbhx.com/search?siteid=qula&q=%s", false, false, "www.qu.la")

		// 新笔趣阁
		// 注册提取器
		engine.RegisterExtracter("www.xbiquge6.com", NewExtracter(SEARCH_OBJ_URL_PATTERN_XBQG_STR))

		// 注册搜索器
		engine.GlobalSiteSearcher.AddItem("https://www.xbiquge6.com/search.php?keyword=%s", true, false, "www.xbiquge6.com")
	*/
}
