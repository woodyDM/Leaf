package leaf

import (
	"testing"
)

func TestAll(t *testing.T) {
	env := createEnv("test", "11")
	if env.ID == 0 {
		t.Error("error env")
	}
	e2, exist := getEnv(env.ID)
	if !exist {
		t.Fatal("should query")
	}
	if e2.Name!="test"{
		t.Error("Save error")
	}
	if e2.Content!="11"{
		t.Error("content error")
	}
	names:= listEnv()
	if names[0].Name!="test"{
		t.Error("list error")
	}
	if names[0].ID!= env.ID{
		t.Error("list error")
	}
	err := editEnv(env.ID, "test-m", "22","suf")
	if err != nil {
		t.Fatal(exist)
	}
	e3, exist := getEnv(env.ID)
	if !exist   {
		t.Fatal(exist)
	}
	if e3.Name!= "test-m" {
		t.Error("edit error")
	}
	if e3.Content!="22"{
		t.Error("edit content error")
	}
	deleteEnv(env.ID)
	_, exist = getEnv(env.ID)
	if exist  {
		t.Error("delete error")
	}
}

func TestCheckUnique(t *testing.T) {
	checkUnique([]*UsedEnv{
		{

			Variable: "",
			EnvId:    0,
			AppId:    0,
		},
	})
}

func TestEnvAll(t *testing.T) {
	all := listEnvAll()
	if all==nil{
		t.Error("nil envs")
	}
}


