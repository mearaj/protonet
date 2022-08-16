package alog

import (
	log "github.com/sirupsen/logrus"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	//log.SetOutput(ioutil.Discard)
}
func Logger() *log.Entry {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		panic("Could not get context info for logger!")
	}
	index := strings.LastIndex(file, "/")
	index2 := strings.LastIndex(file[0:index], "/")
	filename := file[index2+1:] + ":" + strconv.Itoa(line)
	funcname := runtime.FuncForPC(pc).Name()
	fn := funcname[strings.LastIndex(funcname, ".")+1:]
	return log.WithField("file", filename).WithField("function", fn)
}
