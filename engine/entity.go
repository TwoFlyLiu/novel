package engine

// Novel表示一个小说实体
type Novel struct {
	Name                  string     //小说名称
	LastUpdateTime        string     //小说最后更新时间
	Author                string     //小说作者
	MenuURL               string     //互联网上菜单页面所在的URL
	NewestLastChapterName string     //最新的最后章节的名称
	Menus                 []*Menu    //小说的目录
	Chapters              []*Chapter //小说的章节列表
}

// Menu表示小说的一个目录项
type Menu struct {
	Name string //目录名称
	URL  string //目录所指向的远程url内容
}

// Chapter表示小说的一个章节
type Chapter struct {
	Title   string //表示章节的标题
	Content string //表示章节的内容
}

func (novel *Novel) AddMenu(menu *Menu) {
	novel.Menus = append(novel.Menus, menu)
}

func (novel *Novel) AddChapter(chapter *Chapter) {
	novel.Chapters = append(novel.Chapters, chapter)
}

func NewMenu(name, url string) *Menu {
	return &Menu{name, url}
}

func NewChapter(title, content string) *Chapter {
	return &Chapter{title, content}
}
