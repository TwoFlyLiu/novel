package engine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

const (
	SEP = "/"
)

// NovelDao interface used to serialize Novel object
type Dao interface {
	// Save used to save novel to native file. Concret saving destination is determined by subclass.
	SaveNovel(novel *Novel, dirname string, suffix string) error

	// Load used to load novel from native file.
	LoadNovel(fullpath string) (*Novel, error)

	// 保存图标到本地
	SaveIcon(img []byte, dirname string, iconName string, suffix string) error
}

type ResourceDao struct{}

func makeDirIfNotExist(dirname string) (err error) {
	log.Debugf("makeDirIfNotExist [%s]", dirname)
	_, err = os.Stat(dirname)
	if err != nil {
		log.Debugf("create dir [%s]", dirname)
		err = os.MkdirAll(dirname, 0777)
		if err != nil {
			log.Debugf("Create dir %q fail! Error:%v", dirname, err)
			err = fmt.Errorf("Create dir %q fail! Error:%v", dirname, err)
		}
	}
	return
}

func (resourceDao *ResourceDao) SaveIcon(img []byte, dirname string, iconName string, suffix string) (err error) {
	err = makeDirIfNotExist(dirname)
	if err != nil {
		return
	}

	if len(suffix) > 0 && suffix[0] != '.' {
		suffix = "." + suffix
	}
	fullpath := fmt.Sprintf("%s%s%s%s", dirname, SEP, iconName, suffix)
	err = ioutil.WriteFile(fullpath, img, 0666)
	return
}

// JsonNovelDao a NovelDao's implementation, which is used to save novel to json file or load novel from json file.
// Recommended using NewJsonNovelDao function to create objects of this class.
type JsonNovelDao struct {
	ResourceDao
}

func (dao *JsonNovelDao) SaveNovel(novel *Novel, dirname string, suffix string) (err error) {
	err = makeDirIfNotExist(dirname)
	if err != nil {
		return
	}

	if len(suffix) > 0 && suffix[0] != '.' {
		suffix = "." + suffix
	}
	fullpath := fmt.Sprintf("%s/%s%s", dirname, novel.Name, suffix)
	log.Infof("Save nove to native %q", fullpath)
	file, err := os.OpenFile(fullpath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Debugf("Save novel to native fail, err:%v", err)
		return
	}
	defer file.Close()

	bytes, err := json.MarshalIndent(novel, "  ", "    ")
	if err != nil {
		return
	}

	_, err = file.Write(bytes)
	return
}

func (dao *JsonNovelDao) LoadNovel(fullpath string) (novel *Novel, err error) {
	file, err := os.Open(fullpath)
	if err != nil {
		err = NewNovelNotExistError(novel.Name)
		return
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	novel = new(Novel)
	err = json.Unmarshal(bytes, novel)

	if err != nil {
		novel = nil
		err = fmt.Errorf("novel data is corrupt!")
		return
	}

	return
}

// NewJsonNovelDao used to save novel to json file or load novel from json file
// args invalid count of args is 1, used to set saving destination. If not set, default is ./json
//
// Default saved distination is ./json/novel-name.json,
func NewJsonNovelDao() Dao {
	return &JsonNovelDao{}
}
