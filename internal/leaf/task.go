package leaf

import (
	"gorm.io/gorm"
	"time"
)

type TaskStatus int

const (
	Pending TaskStatus = iota
	Running
	Fail
	Success
)

type Task struct {
	gorm.Model
	AppId       uint
	Command     string
	Seq         int
	Log         string
	Status      TaskStatus
	StartTime   *time.Time
	FinishTime  *time.Time
	CostSeconds int64
}

type TaskPage struct {
	Page
	List []TaskR
}

func queryTasks(page TaskPageQuery) TaskPage {
	list := make([]Task, 0)
	Db.Model(&Task{}).
		Select("id, seq,app_id,status,created_at, start_time,finish_time ").
		Where("app_id = ? ", page.AppId).
		Order("id desc").
		Offset(page.Offset()).
		Limit(page.PageSize).
		Find(&list)
	var c int64
	Db.Model(&Task{}).
		Where("app_id = ? ", page.AppId).
		Count(&c)
	listR := make([]TaskR, 0)
	for _, t := range list {
		listR = append(listR, newTaskR(t))
	}
	return TaskPage{
		Page: Page{
			PageNum:  page.PageNum,
			PageSize: page.PageSize,
			Total:    int(c),
		},
		List: listR,
	}
}

func taskDetail(id uint) (*TaskR, bool) {
	var task Task
	Db.Find(&task, id)
	if task.Status == Running {
		ctx, exist := CommonPool.get(id)
		if exist {
			it, _ := ctx.(*exeCtx)
			task.Log = it.buf.String()
		} else {
			task.Status = Fail
			Db.Updates(&task)
		}
	}
	r := newTaskR(task)
	return &r, task.ID != 0
}
