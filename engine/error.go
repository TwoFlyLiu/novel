package engine

import "fmt"

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

type NovelNotExistError struct {
	NovelName string
}

func (err *NovelNotExistError) Error() string {
	return fmt.Sprintf("Native %q novel does not exist!", err.NovelName)
}

func NewNovelNotExistError(novelName string) *NovelNotExistError {
	return &NovelNotExistError{novelName}
}
