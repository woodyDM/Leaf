package leaf

import "testing"

func TestGitFolder(t *testing.T) {
	folder:= parseGitFolder("https://gitee.com/aaa/Java-test.git")
	if folder!="Java-test"{
		t.Error("error")
	}
}

