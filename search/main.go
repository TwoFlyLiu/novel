package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/twoflyliu/novel/engine"
	_ "github.com/twoflyliu/novel/extracter"
)

type SearchResult struct {
	novel *engine.Novel
	time  int64
}

func main() {
	var verbose bool
	var iconDirName, iconExt string
	var logDirName string
	flag.BoolVar(&verbose, "verbose", false, "enable debug information")
	flag.StringVar(&iconDirName, "id", "./icons", "icon directory name")
	flag.StringVar(&iconExt, "ie", ".img", "icon extension name")
	flag.StringVar(&logDirName, "ld", ".", "base dir name")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [-verbose] [-id 最终图标保存的目录名] [-ie 图标拓展名] [-ld 记录目录名称] 小说名称", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	if len(logDirName) > 1 && logDirName[len(logDirName)-1] == '/' {
		logDirName = logDirName[0 : len(logDirName)-1]
	}
	mgr := engine.NewDefaultEngine(verbose, "", "", iconDirName, iconExt, logDirName)
	urls := mgr.SearchSite(flag.Arg(0))
	log := mgr.GetLogger()

	log.Debug("Search Urls:", urls)

	ch := make(chan *engine.Novel, len(urls))

	for _, url := range urls {
		go func(url string) {
			novel, err := mgr.BaseInfoByURL(url)
			if err != nil {
				log.Info("URL:", url, "error:", err)
				ch <- nil
			} else {
				ch <- novel
			}
		}(url)
	}

	var novel *engine.Novel
	for i := 0; i < len(urls); i++ {
		novel = <-ch
		if novel != nil {
			break
		}
	}

	if novel == nil {
		fmt.Println("None")
		return
	}

	log.Info("MenuURL:", novel.MenuURL)
	log.Info("Name:", novel.Name)
	log.Info("Author:", novel.Author)
	log.Info("LastUpdateTime:", novel.LastUpdateTime)
	log.Info("NewestChapter:", novel.NewestLastChapterName)
	log.Info("Description:", novel.Description)
	log.Info("\n\n\n")

	// 下载小说对应的图标, 先写出图标
	mgr.DownloadAndSaveIcon(novel)

	// 然后输出搜索结果
	if len(iconExt) > 0 && iconExt[0] != '.' {
		iconExt = "." + iconExt
	}
	fmt.Printf("%s|%s|%s|%s|%s|%s\n", novel.MenuURL, novel.Name,
		novel.Author, novel.Description, novel.NewestLastChapterName,
		iconDirName+"/"+novel.Name+iconExt)
}
