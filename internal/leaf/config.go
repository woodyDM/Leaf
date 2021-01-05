package leaf

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var GlobalConfig AppConfig
const gitHomePrefix = "workspace"
const taskPrefix = "task_%d"



type AppConfig struct {
	Home        string
	ShellEnable bool
	OS          string
}

func init() {
	GlobalConfig = AppConfig{
		Home:        getAppHome(),
		ShellEnable: IsLinux(),
		OS:          runtime.GOOS,
	}
	err := initHome()
	if err != nil {
		panic(fmt.Sprintf("Failed to create folder . %s", err))
	}
}

func initHome() error {
	home := GlobalConfig.Home
	return mkdir(home)
}

func stat(path string) ([]string, error) {
	r := make([]string, 0)

	dir, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, f := range dir {
		if f.IsDir() {
			r = append(r, f.Name())
		}
	}
	return r, nil
}

func mkdir(home string) error {

	if _, err := os.Stat(home); os.IsNotExist(err) {
		e := os.MkdirAll(home, 0770)
		if e != nil {
			return e
		}
		e = os.Chmod(home, 0770)
		if e == nil {
			log.Printf("Create folder %s success.\n", home)
		}
		return e
	}
	return nil
}

func IsLinux() bool {
	name := runtime.GOOS
	return name != "windows"
}

func getAppHome() string {
	return getHome() + "/.leaf"
}
func getHome() string {
	if IsLinux() {
		return getLinuxHome()
	} else {
		panic("Not support windows")
	}
}

func getLinuxHome() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	var holder bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &holder
	if err := cmd.Run(); err != nil {
		panic("Failed to get home")
	}
	result := strings.TrimSpace(holder.String())
	if result == "" {
		panic("Failed to parse home from buffer")
	}
	return result
}
