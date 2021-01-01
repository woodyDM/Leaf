package leaf

import "testing"

func TestExeCtx_Run(t *testing.T) {
	ctx := createCmd(1, "echo wd", nil)
	err:=ctx.Run()
	if err!=nil{
		t.Error("should no error")
	}
	if ctx.buf.String()!="wd\n"{
		t.Error("out put error")
	}
}

