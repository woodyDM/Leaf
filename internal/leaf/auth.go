package leaf

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"math/rand"
	"net/http"
	"regexp"
)

var Counter = 0
var max = 10

type User struct {
	gorm.Model
	Name string `gorm:"unique_index"`
	Salt string
	Pass string
}

func NewUser(name string, rawPass string) (*User, error) {
	if name == "" {
		return nil, errors.New("name empty ")
	}
	if !nameMatch(name) {
		return nil, errors.New("only reg: \\w name supported")
	}
	if rawPass == "" {
		return nil, errors.New("pass empty ")
	}
	salt := RandomString(32)
	return &User{
		Name: name,
		Salt: salt,
		Pass: encPass(rawPass, salt),
	}, nil
}

//todo fix reg
func nameMatch(name string) bool {
	r := regexp.MustCompile(`(\w)+`)
	return r.MatchString(name)
}

func (u *User) passMatch(pass string) bool {
	if u.Salt == "" {
		panic("Invalid User ")
	}
	if u.Pass == "" {
		panic("Invalid user Pass")
	}
	if pass == "" {
		return false
	}
	enc := encPass(pass, u.Salt)
	return enc == u.Pass
}

func encPass(pass string, salt string) string {
	v := fmt.Sprintf("%s_%s", pass, salt)
	bs := sha256.Sum256([]byte(v))
	return fmt.Sprintf("%x", bs)
}

var CookieDomain = ""

const cookieName = "gsession"

var defaultLetters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
var authedMap = make(map[string]bool)

var cookieHandler gin.HandlerFunc = func(ctx *gin.Context) {
	c, err := ctx.Cookie(cookieName)
	if err == http.ErrNoCookie || c == "" {
		c = RandomString(32)
		ctx.SetCookie(cookieName, c, 3600, "/", CookieDomain, false, true)
	}
	ctx.Set(cookieName, c)
}

var authHandler gin.HandlerFunc = func(ctx *gin.Context) {
	c, _ := ctx.Cookie(cookieName)
	if c != "" {
		if _, exist := authedMap[c]; exist {
			// authed
			return
		}
	}
	//unAuthed
	ctx.AbortWithStatus(http.StatusUnauthorized)
}

// RandomString returns a random string with a fixed length
func RandomString(n int) string {
	b := make([]rune, n)
	len2 := len(defaultLetters)
	for i := range b {
		b[i] = defaultLetters[rand.Intn(len2)]
	}
	return string(b)
}

func FreshUser(name, pass string) error {
	if pass == "" {
		return errors.New("Pass empty ")
	}
	var us []User
	Db.Model(&User{}).
		Where("name = ? ", name).
		Find(&us)
	i := len(us)
	if i == 0 {
		user, err := NewUser(name, pass)
		if err != nil {
			return err
		}
		Db.Create(&user)
		return nil
	} else if i == 1 {
		u := us[0]
		newPass := encPass(pass, u.Salt)
		u.Pass = newPass
		Db.Updates(&u)
		return nil
	} else {
		return errors.New("User name not unique! ")
	}
}
