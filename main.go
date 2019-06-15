package main

import (
	"cynex/log"
	"cynex/react"
	"net/http"
)

func init() {
	react.BindGet("/index/{name}", new(User), "Index")
	react.PipelineBefore(new(User), "Index", new(User), "Before")
	react.PipelineBefore(new(User), "Before", new(User), "BeforeBefore")
	react.PipelineAfter(new(User), "Index", new(User), "After")
	react.PipelineAfter(new(User), "After", new(User), "AfterAfter")
	react.BindDownload("/download", "readme.md")
	react.BindStatic("/static", "/")
}

func main() {
	react.Server.Start()
}

type User struct {
}

func (u *User) Index(w http.ResponseWriter, r *http.Request) {
	log.Debug("Index 方法已经执行")
	strOut := ""
	if v := r.FormValue("name"); v != "" {
		for i := 0; i < 49; i++ {
			strOut += v
		}
		log.Debug("WriteOut:" + strOut)
		w.Write([]byte(strOut))
	}
}

func (u *User) Before(w http.ResponseWriter, r *http.Request) {
	log.Debug("我在前面执行了")
}

func (u *User) BeforeBefore(w http.ResponseWriter, r *http.Request) {
	log.Debug("我在最前面执行了")
}
func (u *User) After(w http.ResponseWriter, r *http.Request) {
	log.Debug("我在后面执行了")
}
func (u *User) AfterAfter(w http.ResponseWriter, r *http.Request) {
	log.Debug("我在最后面执行了")
}
