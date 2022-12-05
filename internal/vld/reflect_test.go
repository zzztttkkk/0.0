package vld

import (
	"fmt"
	"github.com/zzztttkkk/0.0/internal/utils"
	"net/http"
	"reflect"
	"testing"
	"time"
)

type User struct {
	Name      string    `vld:"name"`
	Email     string    `vld:"email"`
	Age       int       `vld:"age"`
	CreatedAt time.Time `vld:"created_at"`
	Nums      []int     `vld:"nums"`
}

type AuthInfo struct {
	Token string `vld:"token"`
}

func (authinfo *AuthInfo) FromRequest(req *http.Request) (any, error) {
	return &AuthInfo{
		Token: req.Header.Get("Authorization"),
	}, nil
}

func TestGetRules(t *testing.T) {
	rules := GetRules(reflect.TypeOf(User{}))
	req := utils.Must(http.NewRequest("Post", "/oops", nil))
	req.PostForm = map[string][]string{}
	req.PostForm["name"] = []string{"ztk<Spk>"}
	req.PostForm["email"] = []string{"ztk@local.dev"}
	req.PostForm["age"] = []string{"123"}
	req.PostForm["created_at"] = []string{"189123000"}
	req.PostForm["nums"] = []string{"1", "2", "3"}

	fmt.Println(rules.BindAndValidate(req))
}
