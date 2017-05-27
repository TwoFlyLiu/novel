// engine package is a backend or library package. Mainly used to extract novel on internet.
// And save extracted novel to native or load novel from native file .
// And sync native file to newest.
package engine

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
	"time"

	"github.com/op/go-logging"
)

const (
	MAX_RETRIES_COUNT = 5  //maximum numbers of downloads
	THREAD_COUNT      = 15 //the number of threads per download and extract task
)

//Engine is a entry of full engine package, which is actually a service class.
//The function of engine package is called primarily through this class.
//This class is not guaranteed to be thread-safe. It depends largely on the implementation of Download and NovelDao.
//Default Engine object is thread-safe, which is producted by NewDefaultEngine factory function.
type Engine struct {
	downloader Downloader //An object that implements Downloader interface. It is a aggregated object and mainly provide download function from  internet
	dao        NovelDao   //An object that implements NovelDao interface. It is a aggreated object and mainy provide serialize novel function
	threshold  int64      //The max time of extract baseinfo. If actually extracting time more thant threshold, search item will be removed
}

//NewEngine is a factory function used to create Engine object
//downloader - an implementation of Downloader
//dao - an implementation of NovelDao
//verbose - enable debug information
//threshold - threshold time
func NewEngine(downloader Downloader, dao NovelDao, verbose bool, threshold int64) *Engine {
	configLog(verbose) //配置日志
	GlobalSiteSearcher.loadIgnoredHosts()
	return &Engine{downloader: downloader, dao: dao, threshold: threshold}
}

//NewDefaultEngine is a handy factory function.It produces a thread-safe object, which uses the HttpDownloader object and
// the JsonDao object.
//
//verbose - enable debug information
func NewDefaultEngine(verbose bool) *Engine {
	return NewEngine(NewDefaultDownloader(),
		NewJsonNovelDao(), verbose, 1)
}

//Set threshold time of extract base info from url in second
func (engine *Engine) SetThreshold(threshold int64) {
	engine.threshold = threshold
}

//NovelByName - Use the novel name to download the content of novel from internet
//
//name - novel name
//return novel finally novel. err is to achieve error information if an error has occurred.
func (engine *Engine) NovelByName(name string) (novel *Novel, err error) {
	log.Debugf("Got novel by name %q", name)
	novel, err = engine.dao.Load(name) //先从本地获取

	// 当不存在，再从远程获取
	if err != nil {
		urls := engine.SearchSite(name)
		log.Debug("Got search result:", urls)
		for _, u := range urls {
			log.Debugf("Current use url %q", u)
			novel, err = engine.NovelByURL(u) //然后从可选的互联网上获取一个，当此互联网不可用或者出现问题的时候，则使用另一个网站
			if err == nil {
				log.Debug("Save novel to native")
				engine.dao.Save(novel)
				break //表明下载成功
			}
		}
	}
	return
}

//NovelByURL - download novel directly from internet
//return novel finally novel. err is to achieve error information if an error has occurred.
func (engine *Engine) NovelByURL(url string) (novel *Novel, err error) {
	extracter := AutoSelectExtracter(url)
	if extracter == nil {
		err = fmt.Errorf("%q extracter not implemented", url)
		return
	}

	novel = new(Novel)

	menuURL := extracter.ExtractMenuURL(url)
	fullPage, err := engine.downloader.Download(menuURL, MAX_RETRIES_COUNT)

	if err != nil {
		return
	}

	// 设置互联网上菜单所在的url
	novel.MenuURL = menuURL
	log.Debug("MenuURL:", menuURL)

	// 从fullPage中提取Name, Author, LastUpdateTime
	engine.constructNovelBase(fullPage, novel, extracter)

	// 从fullPage从提取出所有的菜单
	engine.constructNovelMenus(fullPage, novel, extracter)

	// 下载图标到本地
	engine.downloadIcon(fullPage, extracter, novel)

	// 下来所有的章节到novel.Chapaters中
	engine.constructNovelChapters(novel, extracter)

	// 保存内容到本地
	engine.dao.Save(novel)
	return
}

// BaseInfoByURL - downlaod the base information of novel directly from internet.
//
// url - the url of novel menu page, which is achieved by call SearchSite method
// novel - finally base information and the field of Chapters is invalid in novel.
// err - may contain error message
func (engine *Engine) BaseInfoByURL(netURL string) (novel *Novel, iconURL string, err error) {
	addr, err := url.Parse(netURL)
	CheckError(err)

	start := time.Now().Unix()
	extracter := AutoSelectExtracter(netURL)
	if extracter == nil {
		err = fmt.Errorf("%q extracter not implemented", netURL)
		return
	}

	novel = new(Novel)

	menuURL := extracter.ExtractMenuURL(netURL)
	fullPage, err := engine.downloader.Download(menuURL, MAX_RETRIES_COUNT)

	// 出错说明源有问题，那么就移除掉
	if err != nil {
		GlobalSiteSearcher.RemoveItem(addr.Host)
		return
	}

	path := extracter.ExtractIconURL(fullPage)
	pathURL, err := url.Parse(path)
	CheckError(err)

	hostURL, err := url.Parse(netURL)
	CheckError(err)
	iconURL = hostURL.ResolveReference(pathURL).String()

	// 设置互联网上菜单所在的url
	novel.MenuURL = menuURL
	// 从fullPage中提取Name, Author, LastUpdateTime
	engine.constructNovelBase(fullPage, novel, extracter)
	end := time.Now().Unix()

	if end-start > engine.threshold {
		GlobalSiteSearcher.RemoveItem(addr.Host)
	}
	return
}

func MustSelectSuitableExtracter(url string) (extracter Extracter) {
	extracter = AutoSelectExtracter(url)
	if extracter == nil {
		panic(fmt.Sprintf("Error: Cannot find suitable extract for %q", url))
	}
	return
}

// SyncNovel - update the content of novel to newest and save novel to native
func (engine *Engine) SyncNovel(novel *Novel) {
	lastMenuItem := novel.Menus[len(novel.Menus)-1]

	extracter := MustSelectSuitableExtracter(lastMenuItem.URL)
	menuPageURL := extracter.ExtractMenuURL(lastMenuItem.URL)
	menuPage, err := engine.downloader.Download(menuPageURL, MAX_RETRIES_COUNT)

	if err != nil {
		panic(fmt.Sprintf("Downlad page [%s] fail: %v", menuPageURL, err))
	}
	newestLastMenuName := extracter.ExtractNewestLastChapterName(menuPage)
	if strings.TrimSpace(lastMenuItem.Name) != strings.TrimSpace(newestLastMenuName) {
		engine.doUpdate(novel, menuPage, menuPageURL, extracter) //讲新的内容更新到内存和本地
		engine.Save(novel)                                       //保存到本地
	}
}

// Save - save novel to native, which mainly depends on the implementation of engine.dao
func (engine *Engine) Save(novel *Novel) {
	engine.dao.Save(novel)
}

func (engine *Engine) constructNovelBase(fullPage string, novel *Novel, extracter Extracter) {
	novel.Name = extracter.ExtractNovelName(fullPage)
	novel.Author = extracter.ExtractNovelAuthor(fullPage)
	novel.LastUpdateTime = extracter.ExtractLastUpdateTime(fullPage)
	novel.NewestLastChapterName = extracter.ExtractNewestLastChapterName(fullPage)
	novel.Description = extracter.ExtractNovelDescription(fullPage)

	log.Debug("Name", novel.Name)
	log.Debug("Author:", novel.Author)
	log.Debug("Last Update time:", novel.LastUpdateTime)
	log.Debug("Newest chapter:", novel.NewestLastChapterName)
}

func (engine *Engine) constructNovelMenus(fullPage string, novel *Novel, extracter Extracter) {
	menus := extracter.ExtractMenuList(fullPage)
	log.Debug("Menu count:", len(menus))
	for _, menu := range menus {
		m := new(Menu)
		m.URL = engine.joinMenuURLAndChapater(novel.MenuURL, menu[0])
		m.Name = menu[1]
		novel.AddMenu(m)
	}
}

// 使用多roroutine来下载章节内容
func (engine *Engine) constructNovelChapters(novel *Novel, extracter Extracter) {
	chapterCount := len(novel.Menus)
	novel.Chapters = make([]*Chapter, chapterCount) //预先设置好缓存

	msgChan := make(chan string, THREAD_COUNT)
	defer close(msgChan)

	perThreadJobCount := chapterCount / THREAD_COUNT

	for i := 0; i < THREAD_COUNT; i++ {
		//每个线程处理[i * len(novel.Menus) / THREAD_COUNT, (i + 1) * len(novel.Menus) / THREAD_COUNT)
		go engine.constructNovelChaptersT(novel.Chapters[i*perThreadJobCount:(i+1)*perThreadJobCount],
			novel.Menus[i*perThreadJobCount:(i+1)*perThreadJobCount], msgChan, i, extracter, novel.Name)
	}

	// 101, [0, 10), [10, 20), ...[90, 100)
	// 还剩下[100, 101)
	//这儿处理chapterCount / THREAD_COUNT，不能被整除的情况
	if (THREAD_COUNT * perThreadJobCount) != chapterCount {
		go engine.constructNovelChaptersT(novel.Chapters[THREAD_COUNT*perThreadJobCount:chapterCount],
			novel.Menus[THREAD_COUNT*perThreadJobCount:chapterCount], msgChan, THREAD_COUNT, extracter, novel.Name)
	}

	// 用来接受goroutine传过来的消息，并且还有个作用就是等待所有线程处理完毕
	for i := 0; i < chapterCount; i++ {
		fmt.Printf("<%%%.1f>", float64(i+1)/float64(chapterCount)*100)
		fmt.Println(<-msgChan)
	}

}

func (engine *Engine) doUpdate(novel *Novel, menuPage string, menuPageURL string, extracter Extracter) {
	menus := extracter.ExtractMenuList(menuPage)
	newMenuLen := len(menus)
	oldMenuLen := len(novel.Menus)

	// 更新新的菜单项
	for i := oldMenuLen; i < newMenuLen; i++ {
		menuItem := new(Menu)
		menuItem.URL = engine.joinMenuURLAndChapater(menuPageURL, menus[i][0])
		menuItem.Name = menus[i][1]
		novel.AddMenu(menuItem)
	}

	// 下面要从网上进行更新
	toUpdateLen := newMenuLen - oldMenuLen
	for i := 0; i < toUpdateLen; i++ {
		novel.AddChapter(new(Chapter)) //先提供空的，然后方便使用多线程来进行更新
	}

	// 把要更新的slice给去取出来
	updatedMenuSlice := novel.Menus[oldMenuLen:]
	toUpdateChapterSlice := novel.Chapters[oldMenuLen:]

	msgChan := make(chan string, toUpdateLen)
	defer close(msgChan)

	chapterCount := toUpdateLen
	perThreadJobCount := chapterCount / THREAD_COUNT
	tid := 0

	fmt.Printf("Begining update: %s\n", novel.Name)
	fmt.Printf("Updated menu list: %v\n", updatedMenuSlice)
	// chapterCount只要要大于THREAD_COUNT，下面的算法才成立
	if perThreadJobCount >= 1 {
		for i := 0; i < THREAD_COUNT; i++ {
			//每个线程处理[i * len(novel.Menus) / THREAD_COUNT, (i + 1) * len(novel.Menus) / THREAD_COUNT)
			go engine.constructNovelChaptersT(toUpdateChapterSlice[i*perThreadJobCount:(i+1)*perThreadJobCount],
				updatedMenuSlice[i*perThreadJobCount:(i+1)*perThreadJobCount], msgChan, tid, extracter, novel.Name)
			tid = tid + 1
		}
	}

	// 101, [0, 10), [10, 20), ...[90, 100)
	// 还剩下[100, 101)
	//这儿处理chapterCount / THREAD_COUNT，不能被整除的情况
	if (THREAD_COUNT * perThreadJobCount) != chapterCount {
		go engine.constructNovelChaptersT(toUpdateChapterSlice[THREAD_COUNT*perThreadJobCount:chapterCount],
			updatedMenuSlice[THREAD_COUNT*perThreadJobCount:chapterCount], msgChan, tid, extracter, novel.Name)
	}

	// 用来接受goroutine传过来的消息，并且还有个作用就是等待所有线程处理完毕
	for i := 0; i < chapterCount; i++ {
		fmt.Printf("<%%%g>", float32(i+1)/float32(chapterCount)*100)
		fmt.Println(<-msgChan)
	}
	fmt.Printf("Update %s done!\n", novel.Name)
}

func (engine *Engine) constructNovelChaptersT(chapters []*Chapter, menus []*Menu, msgChan chan string,
	tid int, extracter Extracter, name string) {
	jobCount := len(menus)
	for i := 0; i < jobCount; i++ {
		fullPage, err := engine.downloader.Download(menus[i].URL, MAX_RETRIES_COUNT)
		if err != nil {
			msgChan <- fmt.Sprintf("[%d] Dowloader.downloader(%s) fail[%s]: %v\n", tid, name, menus[i].URL, err)
			continue
		} else {
			msgChan <- fmt.Sprintf("[%d] Successfully download(%s) chapter %q", tid, name, menus[i].Name)
		}
		chapter := new(Chapter)
		chapter.Title = extracter.ExtractChapterTitle(fullPage)
		chapter.Content = extracter.ExtractChapterContent(fullPage)
		chapters[i] = chapter
	}
}

// 为了处理page是形如/book/4/2222.html形式
// 和2222.html形式
func (engine *Engine) joinMenuURLAndChapater(menuURL, page string) string {
	// 保证menuURL的最后一个元素必须是/
	if '/' != menuURL[len(menuURL)-1] {
		menuURL = menuURL + "/"
	}
	u, err := url.Parse(menuURL)
	if err != nil {
		panic(fmt.Sprintf("joinMenuURLAndChapater fail: %v", err))
	}

	root := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	page = strings.TrimSpace(page)
	if page[0] == '/' {
		return fmt.Sprintf("%s%s", root, page)
	} else {
		return fmt.Sprintf("%s%s", menuURL, page)
	}
}

// SearchSite - search the novel by name
//
// return - return multiple urls that novel's download page.
func (engine *Engine) SearchSite(name string) []string {
	return GlobalSiteSearcher.Search(name)
}

func (engine *Engine) downloadIcon(menuPage string, extracter Extracter, novel *Novel) {
	host, err := url.Parse(novel.MenuURL)
	CheckError(err)

	iconURL := extracter.ExtractIconURL(menuPage)
	path, err := url.Parse(iconURL)
	CheckError(err)

	// 后缀名是无所谓的
	filename := novel.Name + ".img"

	fullpath := host.ResolveReference(path)
	go func() {
		img, err := engine.downloader.Download(fullpath.String(), 5)
		CheckError(err)
		ioutil.WriteFile("./json/"+filename, []byte(img), 0666)
	}()
}

func (engine *Engine) GetLogger() *logging.Logger {
	return log
}

func (engine *Engine) GetDownloader() Downloader {
	return engine.downloader
}
