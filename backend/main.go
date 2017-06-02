package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/twoflyliu/novel/engine"
	_ "github.com/twoflyliu/novel/extracter"
)

func main() {
	var download, update, verbose, downloadIcon bool
	var downloadDir string
	var iconExt string
	var iconDir string
	var novelExt string
	var logDir string

	flag.BoolVar(&download, "g", false, "do download operator")
	flag.BoolVar(&downloadIcon, "gi", false, "if download icon")

	flag.BoolVar(&update, "u", false, "do update operator")
	flag.StringVar(&downloadDir, "d", "./json", "the directory of download object")
	flag.StringVar(&novelExt, "e", "", "the ext name of novel")
	flag.BoolVar(&verbose, "verbose", false, "enable debug information")
	flag.StringVar(&iconExt, "ie", "img", "icon ext name")
	flag.StringVar(&iconDir, "id", "icons", "icon native directory")
	flag.StringVar(&logDir, "ld", ".", "log dir name")
	flag.Parse()

	// 默认是下载操作
	if !download && !update {
		download = true
	}

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s -[g|u] [-d dirname] novel_name\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	if len(logDir) > 1 && logDir[len(logDir)-1] == '/' {
		logDir = logDir[0 : len(logDir)-1]
	}
	mgr := engine.NewDefaultEngine(verbose, downloadDir, novelExt, iconDir, iconExt, logDir)
	switch {
	case update:
		doUpdate(mgr, flag.Arg(0))
	case download:
		doDownload(mgr, flag.Arg(0), downloadDir, iconExt, downloadIcon)
	}
}

func doUpdate(mgr *engine.Engine, novelName string) {
	logger := mgr.GetLogger()
	logger.Debugf("Update novel %q", novelName)

	// 下面是从本地加载文件，但是如果本地没有对应的novel，他会自动下载的
	novel, err := mgr.NovelByName(novelName)
	CheckError(err)
	mgr.SyncNovel(novel) //手动更新
}

func doDownload(mgr *engine.Engine, url string, dirname string, iconExt string,
	downloadIcon bool) {
	logger := mgr.GetLogger()
	novel, err := mgr.NovelByURL(url)
	CheckError(err)
	logger.Debugf("Download novel to memory done!")
	err = mgr.SaveNovel(novel)
	CheckError(err)
	logger.Debugf("Save novel to native done!")

	if len(iconExt) > 0 && iconExt[0] != '.' {
		iconExt = "." + iconExt
	}

	if downloadIcon {
		mgr.DownloadAndSaveIcon(novel)
		logger.Debugf("Download and save icon to native done!")
	}
}

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
