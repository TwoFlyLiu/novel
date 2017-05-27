package engine

import (
	"os"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("logging")

//var format = logging.MustStringFormatter(`%{color:bold}%{time:2006-01-02 15:04:05.999Z-07:00} %{shortfunc}▶ [%{level:.4s}] %{id:03x}%{color:reset} %{message}`)
var format = logging.MustStringFormatter(`%{color:bold}%{time:2006-01-02 15:04:05} %{shortfunc} [%{level:.4s}] %{color:reset} %{message}`)

func configLog(enable bool) {
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	formatterBackend := logging.NewBackendFormatter(backend, format)
	leveledBackend := logging.AddModuleLevel(formatterBackend)

	if enable {
		leveledBackend.SetLevel(logging.DEBUG, "")
	} else {
		leveledBackend.SetLevel(logging.ERROR, "")
	}
	logging.SetBackend(leveledBackend)
}
