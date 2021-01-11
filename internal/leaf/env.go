package leaf

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
)

type UsedEnv struct {
	gorm.Model
	Variable string
	EnvId    uint
	AppId    uint
}

type SEnv struct {
	ID   uint
	Name string
}
type Env struct {
	gorm.Model
	Name    string `json:"name"`
	Content string `json:"content"`
}

func deleteEnv(id uint) error {
	var env Env
	e := Db.Find(&env, id).Error
	if e != nil {
		return e
	}
	if env.ID == 0 {
		return errors.New(fmt.Sprintf("Not found %d", id))
	}
	return Db.Delete(&env).Error
}

func getEnv(id uint) (*Env, bool) {
	var env Env
	e := Db.Find(&env, id).Error
	if e != nil {
		return nil, false
	} else {
		return &env, env.ID != 0
	}
}

func editEnv(id uint, name, content, suffix string) error {
	var env Env
	e := Db.Find(&env, id).Error
	if e != nil {
		return e
	}
	env.Name = name
	env.Content = content
	Db.Updates(&env)
	return nil
}

func createEnv(name, content string) *Env {
	env := Env{
		Name:    name,
		Content: content,
	}
	Db.Create(&env)
	return &env
}

func listEnvAll() []Env {
	l := make([]Env, 0)
	Db.Model(&Env{}).
		Order("id desc").
		Find(&l)
	return l
}

func listEnv() []SEnv {
	l := make([]Env, 0)
	Db.Model(&Env{}).Select("id, name").
		Order("id desc").
		Find(&l)
	r := make([]SEnv, 0)
	for _, e := range l {
		r = append(r, SEnv{
			ID:   e.ID,
			Name: e.Name,
		})
	}
	return r
}

func updateUsedEnv(appId uint, newEnvs []*UsedEnv) error {
	for _,it:=range newEnvs {
		it.AppId = appId
	}
	newMap, e := checkUnique(newEnvs)
	if e != nil {
		return e
	}
	dbs := findByAppId(appId)
	dbMap := make(map[uint]bool)
	for _, it := range dbs {
		dbMap[it.EnvId] = true
	}
	for _, it := range newEnvs {
		if _, ok := dbMap[it.EnvId]; !ok {
			//not exist in db ,should save
			it.ID = 0
			Db.Create(&it)
		} else {
			//exist in db, should update
			Db.Updates(&it)
		}
	}
	//clear unused dbEnv
	for _, it := range dbs {
		if _, ok := newMap[it.EnvId]; !ok {
			//no longer need db env, should delete
			Db.Delete(&it)
		}
	}
	return nil
}

func checkUnique(envs []*UsedEnv) (map[uint]bool, error) {
	m := make(map[uint]bool)
	for _, it := range envs {
		if _, ok := m[it.EnvId]; ok {
			return nil, errors.New("Env file not unique ")
		} else {
			m[it.EnvId] = true
		}
	}
	//todo check name unique
	return m, nil
}

/**
find app valid envs
*/
func findByAppId(id uint) []*UsedEnv {
	list := make([]*UsedEnv, 0)
	Db.Model(&UsedEnv{}).
		Where("app_id = ? ", id).
		Find(&list)
	m := validEnvIds()
	r := make([]*UsedEnv, 0)
	for _, it := range list {
		if _, ok := m[it.EnvId]; ok {
			r = append(r, it)
		}
	}
	return r
}

/**
current valid env ids
*/
func validEnvIds() map[uint]bool {
	validEnv := listEnv()
	m := make(map[uint]bool)
	for _, it := range validEnv {
		m[it.ID] = true
	}
	return m
}

func validEnvs() map[uint]Env {
	list := listEnvAll()
	m := make(map[uint]Env, 0)
	for _, it := range list {
		m[it.ID] = it
	}
	return m
}
