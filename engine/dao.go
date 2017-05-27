package engine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// NovelDao interface used to serialize Novel object
type NovelDao interface {
	// Save used to save novel to native file. Concret saving destination is determined by subclass.
	Save(novel *Novel) error

	// Load used to load novel from native file.
	Load(name string) (*Novel, error)
}

// JsonNovelDao a NovelDao's implementation, which is used to save novel to json file or load novel from json file.
// Recommended using NewJsonNovelDao function to create objects of this class.
type JsonNovelDao struct {
	dirname string // finally saved destination
}

func (dao *JsonNovelDao) Save(novel *Novel) (err error) {
	file, err := os.OpenFile(dao.dirname+"/"+novel.Name, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
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

func (dao *JsonNovelDao) Load(name string) (novel *Novel, err error) {
	file, err := os.Open(dao.dirname + "/" + name)
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
func NewJsonNovelDao(args ...string) NovelDao {
	var dirname = "./json"
	if len(args) > 0 {
		dirname = args[0]
	}
	return &JsonNovelDao{dirname}
}
