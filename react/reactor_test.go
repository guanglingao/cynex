package react

import (
	"cynex/log"
	"net/http"
	"testing"
)

type User struct {
}

func TestBindGet(t *testing.T) {

	BindGet("/index/{name}", new(User), "Index")
	Server.Start()

}

func (u *User) Index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("~~~~~~~~~~~~~~~~~~~~~~~~"))
	log.Info("Index is executing...")
}
