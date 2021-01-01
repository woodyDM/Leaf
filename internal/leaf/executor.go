package leaf

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
)

type exeCtx struct {
	id      uint
	command string
	cmd     *exec.Cmd
	buf     *bytes.Buffer
	env     *EnvCommand
}


func (e *exeCtx) Info(msg string) {
	e.buf.WriteString(fmt.Sprintf("[Leaf] %s\n", msg))
}

func (e *exeCtx) Warning(msg string) {
	e.buf.WriteString("[Leaf] ============= WARNING ============= \n")
	e.buf.WriteString(fmt.Sprintf("[Leaf] %s\n", msg))
	e.buf.WriteString("[Leaf] =================================== \n")

}

var CommonPool = NewPool(4)

func (e *exeCtx) Run() error {
	return e.cmd.Run()
}

func createCmd(id uint, command string, shell *EnvCommand) *exeCtx {
	cmd := exec.Command("bash", "-c", command)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	return &exeCtx{
		id:      id,
		command: command,
		cmd:     cmd,
		buf:     &buf,
		env:     shell,
	}
}

type Pool struct {
	size      int
	ch        chan *exeCtx
	container map[uint]*exeCtx
	lock      sync.RWMutex
}

func (p *Pool) submit(ctx *exeCtx) {
	p.ch <- ctx
}
func (p *Pool) get(id uint) (*exeCtx, bool) {
	p.lock.RLock()
	ctx, ok := p.container[id]
	p.lock.RUnlock()
	return ctx, ok
}

func (p *Pool) start() {
	for i := 0; i < p.size; i++ {
		go func() {
			for it := range p.ch {
				log.Println("Start handle :", it.id)
				changeStatus(it)
				handleCtx(p, it)
			}
		}()
	}
}

func changeStatus(ctx *exeCtx) {
	var task Task
	Db.Find(&task, ctx.id)
	task.Status = Running
	Db.Updates(&task)
}

func handleCtx(p *Pool, ctx *exeCtx) {
	defer func() {
		p.lock.Lock()
		delete(p.container, ctx.id)
		p.lock.Unlock()
		if p := recover(); p != nil {
			log.Printf("Error when handle ctx: %v\n", p)
		}
	}()
	p.lock.Lock()
	p.container[ctx.id] = ctx
	p.lock.Unlock()
	err := createEvnFiles(ctx.env)
	if err != nil {
		ctx.Warning(err.Error())
		updateTaskStatus(ctx, Fail)
		return
	}
	ctx.Info("star to run shells ")
	err = ctx.Run()
	err2 := os.RemoveAll(ctx.env.folder)
	if err2 != nil {
		ctx.Warning(fmt.Sprintf("Unable to delete temp folder: %s", ctx.env.folder))
	}
	if err == nil {
		updateTaskStatus(ctx, Success)
	} else {
		updateTaskStatus(ctx, Fail)
	}
}

func updateTaskStatus(ctx *exeCtx, status TaskStatus) *Task {
	var task Task
	Db.Find(&task, ctx.id)
	task.Log = ctx.buf.String()
	task.Status = status
	Db.Updates(&task)
	return &task
}

func createEvnFiles(command *EnvCommand) error {
	if command == nil || len(command.envs) == 0 {
		return nil
	}
	err := mkdir(command.folder)
	if err != nil {
		return err
	}
	for _, it := range command.envs {
		e := writeToEvnFile(it)
		if e != nil {
			return e
		}
	}
	return nil
}

func writeToEvnFile(it *EnvShell) error {
	fileName := it.fileName
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0770)
	defer file.Close()
	if err != nil {
		return nil
	}
	_, err2 := file.WriteString(it.content)
	if err2 != nil {
		return errors.New(fmt.Sprintf("unable to write env file %s. ", it.fileName))
	}
	return nil
}

func NewPool(size int) *Pool {
	p := &Pool{
		size:      size,
		ch:        make(chan *exeCtx, 100000),
		container: make(map[uint]*exeCtx),
	}
	p.start()
	return p
}
