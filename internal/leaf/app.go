package leaf

import (
	"bytes"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"os"
	"path"
	"strings"
	"time"
)

type Application struct {
	gorm.Model
	Name    string `json:"name"`
	Enable  bool   `json:"enable"`
	GitUrl  string `json:"gitUrl"`
	Command string `json:"command"`
}

type EnvShell struct {
	variableName string
	fileName     string
	content      string
}

var cloneShell = `
cd %s
git clone --progress %s
cd %s
%s
%s
`
var pullShell = `
cd %s
git pull
%s
%s
`
var noEnv = `echo "[Leaf]No envs to set"`

func queryApp(id string) (*Application, error) {
	var app Application
	if id == "" {
		return nil, errors.New("id not found")
	}
	Db.Find(&app, "id = ? ", id)
	if app.ID == 0 {
		return nil, errors.New(fmt.Sprintf("id %s not found", id))
	}
	return &app, nil
}

func parseGitFolder(url string) string {
	idx := strings.LastIndex(url, "/")
	if idx == -1 {
		return ""
	}
	idx2 := strings.LastIndex(url, ".git")
	if idx2 == -1 {
		return ""
	}
	return url[idx+1 : idx2]
}

func (app *Application) gitHome() string {
	return path.Join(GlobalConfig.Home, app.Name, gitHomePrefix)
}

func (app *Application) taskHome(seq int) string {
	return path.Join(GlobalConfig.Home, app.Name, fmt.Sprintf(taskPrefix, seq))
}

func (app *Application) home() string {
	return path.Join(GlobalConfig.Home, app.Name)
}

func (app *Application) run() (uint, error) {
	err := mkdir(app.gitHome())
	if err != nil {
		return 0, err
	}
	seq := taskSeq(app.ID)
	task, err := app.runCommand(seq)
	if err != nil {
		return 0, nil
	} else {
		return task.ID, nil
	}
}

func taskSeq(appId uint) int {
	var lastTask Task
	Db.Where(&Task{AppId: appId}).Last(&lastTask)
	if lastTask.ID == 0 {
		return 1
	} else {
		return lastTask.Seq + 1
	}
}

func (app *Application) runCommand(seq int) (*Task, error) {
	gitHome := app.gitHome()
	gitPath := path.Join(gitHome, parseGitFolder(app.GitUrl))
	if gitPath == "" {
		return nil, errors.New(fmt.Sprintf("invalid git url:%s", app.GitUrl))
	}
	dir, err := stat(gitHome)
	if err != nil {
		return nil, err
	}
	envShell := app.genEnvShell(seq)
	envCmd := envShell.command()
	var sh string
	if len(dir) == 0 {
		//need clone git rep
		sh = fmt.Sprintf(cloneShell, gitHome, app.GitUrl, gitPath, envCmd, app.Command)
	} else {
		sh = fmt.Sprintf(pullShell, gitPath, envCmd, app.Command)
	}
	//save pending task
	task := Task{
		AppId:   app.ID,
		Command: app.Command,
		Seq:     seq,
		Status:  Pending,
	}
	Db.Create(&task)
	//to start task in other goroutine
	cmd := createCmd(task.ID, app.ID, sh, envShell)
	CommonPool.Submit(cmd)
	return &task, nil
}

type EnvCommand struct {
	envs   []*EnvShell
	folder string
}

func (e *EnvCommand) command() string {
	s := e.envs
	if len(s) == 0 {
		return noEnv
	}
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("echo \"[Leaf]Export %d envs\"\n", len(s)))
	for _, it := range s {
		cmd := fmt.Sprintf("export %s=%s", it.variableName, it.fileName)
		buf.WriteString(fmt.Sprintf("%s\n", cmd))
	}
	return buf.String()
}

/**
//todo use db join
gen envShell
*/
func (app *Application) genEnvShell(seq int) *EnvCommand {
	envs := findByAppId(app.ID)
	allEnv := validEnvs()
	r := make([]*EnvShell, 0)
	stamp := time.Now().Unix()
	for i, it := range envs {
		if env, ok := allEnv[it.EnvId]; ok {
			fileName := fmt.Sprintf("temp_%d_%d.%s", stamp, i, "txt")
			file := path.Join(app.taskHome(seq), fileName)
			r = append(r, &EnvShell{
				variableName: it.Variable,
				fileName:     file,
				content:      env.Content,
			})
		}
	}
	return &EnvCommand{
		envs:   r,
		folder: app.taskHome(seq),
	}
}

func findAppByName(name string) (*Application, bool) {
	var app Application
	Db.Where("name", name).Find(&app)
	if app.ID == 0 {
		return nil, false
	} else {
		return &app, true
	}
}

func (app *Application) clear() error {
	dir := app.home()
	return os.RemoveAll(dir)
}

func (app *Application) delete() error {
	err := app.clear()
	if err != nil {
		return err
	}
	return Db.Delete(app).Error
}
