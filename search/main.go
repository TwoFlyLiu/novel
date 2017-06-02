package main

import (
	"flag"
	"fmt"
	"os"
	"time"

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

	var results []SearchResult

	for _, url := range urls {
		start := time.Now().UnixNano()
		novel, err := mgr.BaseInfoByURL(url)
		end := time.Now().UnixNano()

		if err != nil {
			log.Info("URL:", url, "error:", err)
			continue
		}
		results = append(results, SearchResult{novel, end - start})

		log.Info("MenuURL:", novel.MenuURL)
		log.Info("Name:", novel.Name)
		log.Info("Author:", novel.Author)
		log.Info("LastUpdateTime:", novel.LastUpdateTime)
		log.Info("NewestChapter:", novel.NewestLastChapterName)
		log.Info("Description:", novel.Description)
		log.Info("Time:", end-start)
		log.Info("\n\n\n")
	}

	if len(results) == 0 {
		fmt.Println("None") //表示没有结果
		return
	}

	// 找到下载速度最快的源，并且将接口通过命令行来返回
	minPos := 0
	for i := 1; i < len(results); i++ {
		if results[minPos].time > results[minPos].time {
			minPos = i
		}
	}

	// 下载小说对应的图标, 先写出图标
	mgr.DownloadAndSaveIcon(results[minPos].novel)

	// 然后输出搜索结果
	if len(iconExt) > 0 && iconExt[0] != '.' {
		iconExt = "." + iconExt
	}
	fmt.Printf("%s|%s|%s|%s|%s|%s\n", results[minPos].novel.MenuURL, results[minPos].novel.Name,
		results[minPos].novel.Author, results[minPos].novel.Description, results[minPos].novel.NewestLastChapterName,
		iconDirName+"/"+results[minPos].novel.Name+iconExt)
}
