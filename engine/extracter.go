package engine

import (
	"net/url"
	"regexp"
)

// 是一些通用的提取方法
const (
	CHARSET_PATTERN_SUBMATCH = `\<meta\s+http-equiv="Content-Type"\s+content="text/html;\s*charset\s*=\s*([\s\S]+?)"\s*/\>`
)

var (
	charsetPatternSubMatch *regexp.Regexp
)

// Extracter 可以从一个已有的页面中提取出必要的信息
// 应该和各个平台的小说网站相关联
type Extracter interface {
	// 从fullPage中提取出小说名称
	ExtractNovelName(fullPage string) string

	// 从fullPage从提取出小说最后更新时间
	ExtractLastUpdateTime(fullPage string) string

	// 从fullPage从提取出小说作者
	ExtractNovelAuthor(fullPage string) string

	// 从fullPage从提取出小说菜单列表
	// 返回以[[url1, menu1], [ur2, menu2], ...]形式返回
	ExtractMenuList(fullPage string) [][]string

	// 从menuPage中提取出小说图标的URL
	ExtractIconURL(menuPage string) string

	// 从fullPage从提取出小说标题
	ExtractChapterTitle(fullPage string) string

	// 从fullPage从提取出小说内容
	ExtractChapterContent(fullPage string) string

	// 从url中提取出目录的url
	ExtractMenuURL(url string) string

	// 从fullPage从提取出网上最新的章节的名称
	ExtractNewestLastChapterName(fullPage string) string

	// 从fullPage中提取出所有的搜索表单中的隐藏字段和值
	ExtractSearchFormHiddenValues(fullPage string) url.Values

	// 从fullPage中提取出搜索表单的方法
	ExtractSearchFormMethodAndAction(fullPage string) (string, string)

	// 从fullPage中提取出搜索表单输入框的名称
	ExtractSearchFormSearchFieldName(fullPage string) string

	// 从searchPage中，搜索出目标名为name的字符串
	ExtractObjURL(name string, searchPage string) (string, bool)
}

var globalExtracterManager map[string]Extracter

// 不允许重复注册
// regexpStr是主机名称正则表达式
func RegisterExtracter(regexpStr string, extracter Extracter) {
	if _, ok := globalExtracterManager[regexpStr]; !ok {
		globalExtracterManager[regexpStr] = extracter
	}
}

func AutoSelectExtracter(URL string) Extracter {
	u, _ := url.Parse(URL)
	for pat, ext := range globalExtracterManager {
		if m, _ := regexp.MatchString(pat, u.Host); m {
			return ext
		}
		if m, _ := regexp.MatchString(pat, URL); m {
			return ext
		}
	}
	return nil
}

func init() {
	var err error
	globalExtracterManager = make(map[string]Extracter, 0)
	charsetPatternSubMatch, err = regexp.Compile(CHARSET_PATTERN_SUBMATCH)
	CheckError(err)
}

func ExtractCharset(fullPage string) (charset string) {
	matches := charsetPatternSubMatch.FindStringSubmatch(fullPage)
	if len(matches) > 1 {
		charset = matches[1]
	}
	return
}
