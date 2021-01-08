package leaf

import "gorm.io/gorm"

type TaskStatus int

const (
	Pending TaskStatus = iota
	Running
	Fail
	Success
)

type Task struct {
	gorm.Model
	AppId   uint       `json:"appId"`
	Command string     `json:"command"`
	Seq     int        `json:"seq"`
	Log     string     `json:"log"`
	Status  TaskStatus `json:"status"`
}

type TaskPage struct {
	Page
	List []Task
}

func queryTasks(page TaskPageQuery) TaskPage {
	list := make([]Task, 0)
	Db.Model(&Task{}).
		Select("id, seq,app_id,status,created_at").
		Where("app_id = ? ", page.AppId).
		Order("id desc").
		Offset(page.Offset()).
		Limit(page.PageSize).
		Find(&list)
	var c int64
	Db.Model(&Task{}).
		Where("app_id = ? ", page.AppId).
		Count(&c)
	return TaskPage{
		Page: Page{
			PageNum:  page.PageNum,
			PageSize: page.PageSize,
			Total:    int(c),
		},
		List: list,
	}
}

func taskDetail(id uint) (*Task, bool) {
	var task Task
	Db.Find(&task, id)
	if task.Status==Running{
		ctx, exist := CommonPool.get(id)
		if exist    {
			it,_:=ctx.(*exeCtx)
			task.Log = it.buf.String()
		}
	}
	return &task, task.ID != 0
}
