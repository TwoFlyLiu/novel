package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/twoflyliu/novel3/engine"
	_ "github.com/twoflyliu/novel3/extracter"
)

type SearchResult struct {
	novel   *engine.Novel
	time    int64
	iconURL string
}

func main() {
	var verbose bool
	flag.BoolVar(&verbose, "verbose", false, "enable debug information")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [-verbose] 小说名称", os.Args[0])
		os.Exit(1)
	}

	mgr := engine.NewDefaultEngine(verbose)
	urls := mgr.SearchSite(flag.Arg(0))
	log := mgr.GetLogger()

	var results []SearchResult

	for _, url := range urls {
		start := time.Now().Unix()
		novel, iconURL, err := mgr.BaseInfoByURL(url)
		end := time.Now().Unix()

		if err != nil {
			log.Info("URL:", url, "error:", err)
			continue
		}
		results = append(results, SearchResult{novel, end - start, iconURL})

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

	minPos := 0
	for i := 1; i < len(results); i++ {
		if results[minPos].time > results[minPos].time {
			minPos = i
		}
	}

	file, err := ioutil.TempFile("/tmp", "novel")
	engine.CheckError(err)
	defer file.Close()

	fmt.Printf("%s|%s|%s|%s|%s|%s\n", results[minPos].novel.MenuURL, results[minPos].novel.Name,
		results[minPos].novel.Author, results[minPos].novel.Description, results[minPos].novel.NewestLastChapterName,
		file.Name())

	iconContent, err := mgr.GetDownloader().Download(results[minPos].iconURL, 5)
	engine.CheckError(err)
	file.Write([]byte(iconContent))
}
