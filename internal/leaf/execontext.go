package leaf

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

type exeCtx struct {
	id      uint
	command string
	cmd     *exec.Cmd
	buf     *bytes.Buffer
	env     *EnvCommand
}

func (ctx *exeCtx) whenError(e error) {
	ctx.Warning(e.Error())
	updateTaskStatus(ctx, Fail)
}

func (ctx *exeCtx) runnerId() uint {
	return ctx.id
}

func (ctx *exeCtx) run() {
	taskToRunning(ctx)
	err := createEvnFiles(ctx.env)
	if err != nil {
		ctx.Warning(err.Error())
		updateTaskStatus(ctx, Fail)
		return
	}
	ctx.Info("star to run shells ")
	err = ctx.cmd.Start()
	if err != nil {
		updateTaskStatus(ctx, Fail)
		return
	}
	go ctx.cmd.Wait()
	for a:=0;a<200;a++{
		log.Println(ctx.cmd.ProcessState)
		time.Sleep(time.Second)

	}

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

func (ctx *exeCtx) shutdown() {
	//todo shutdown graceful
}

func (ctx *exeCtx) Info(msg string) {
	ctx.buf.WriteString(fmt.Sprintf("[Leaf] %s\n", msg))
}

func (ctx *exeCtx) Warning(msg string) {
	ctx.buf.WriteString("============= WARNING ============= \n")
	ctx.buf.WriteString(fmt.Sprintf("[Leaf] %s\n", msg))
	ctx.buf.WriteString("=================================== \n")

}

func createCmd(id uint, command string, shell *EnvCommand) *exeCtx {
	cmd := exec.Command("/bin/bash", "-cxe", command)
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

func taskToRunning(ctx *exeCtx) {
	var task Task
	Db.Find(&task, ctx.id)
	task.Status = Running
	Db.Updates(&task)
}

func updateTaskStatus(ctx *exeCtx, status TaskStatus) *Task {
	var task Task
	Db.Find(&task, ctx.id)
	task.Log = ctx.buf.String()
	task.Status = status
	Db.Updates(&task)
	return &task
}
