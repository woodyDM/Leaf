package leaf

import "time"

type Page struct {
	PageNum  int
	PageSize int
	Total    int
}

func (p Page) Offset() int {
	return (p.PageNum - 1) * p.PageSize
}

type TaskR struct {
	ID         uint
	CreatedAt  string     `json:"CreatedAt"`
	AppId      uint       `json:"appId"`
	Command    string     `json:"command"`
	Seq        int        `json:"seq"`
	Log        string     `json:"log"`
	Status     TaskStatus `json:"status"`
	StartTime  string     `json:"startTime"`
	FinishTime string     `json:"finishTime"`
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	} else {
		return t.Format("2006-01-02 15:04:05")
	}
}

func newTaskR(t Task) TaskR {
	return TaskR{
		ID:         t.ID,
		CreatedAt:  formatTime(&t.CreatedAt),
		AppId:      t.AppId,
		Command:    t.Command,
		Seq:        t.Seq,
		Log:        t.Log,
		Status:     t.Status,
		StartTime:  formatTime(t.StartTime),
		FinishTime: formatTime(t.FinishTime),
	}
}
