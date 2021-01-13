package leaf

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"syscall"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type ApplicationX struct {
	ID      uint
	Name    string     `json:"name"`
	Enable  bool       `json:"enable"`
	GitUrl  string     `json:"gitUrl"`
	Command string     `json:"command"`
	Envs    []*UsedEnv `json:"envs"`
}

func (a *ApplicationX) fill() {
	for _, it := range a.Envs {
		it.AppId = a.ID
	}
}

func ok(d interface{}) Response {
	return Response{
		Data: d,
	}
}

func fail(msg string) Response {
	return Response{
		Code: 500,
		Msg:  msg,
		Data: nil,
	}
}

var notFound = fail("Not found")
var paramError = fail("Param error")

type TaskPageQuery struct {
	AppId uint
	Page
}

func (query TaskPageQuery) defaultPage() {
	if query.PageNum == 0 {
		query.PageNum = 1
	}
	if query.PageSize == 0 {
		query.PageNum = 10
	}
}

func StartServer(port int) {
	// logger and recovery (crash-free) middleware
	router := gin.New()
	router.Use(gin.Recovery())
	logger := gin.Logger()

	router.GET("/api/health", health)
	router.GET("/api/env", env)

	router.Use(logger, cookieHandler).POST("/api/v1/login", loginC)

	v1 := router.Group("/api/v1", authHandler)
	{
		app := v1.Group("/app")
		{
			app.POST("", saveApplication)
			app.GET("", applicationDetail)
			app.GET("/list", appList)
			app.POST("/clear", clearApp)
			app.POST("/delete", deleteApp)
			app.POST("/run", runTask)
		}
		task := v1.Group("/task")
		{
			task.GET("/detail", taskQuery)
			task.GET("", taskPage)
			task.POST("/kill", taskKill)
		}
		env := v1.Group("/env")
		{
			env.GET("", getEnvC)
			env.POST("", saveOrUpdateEnv)
			env.GET("/list", getEnvListC)
			env.POST("/delete", deleteEnvC)
		}
	}

	e := router.Run(fmt.Sprintf(":%d", port))
	if e != nil {
		panic(fmt.Sprintf("Failed to start server %s.", e.Error()))
	}
}

func loginC(ctx *gin.Context) {
	var u User
	err := ctx.BindJSON(&u)
	if err != nil {
		ctx.JSON(200, paramError)
		return
	}
	if u.Name == "" {
		ctx.JSON(200, paramError)
		return
	}
	if u.Pass == "" {
		ctx.JSON(200, fail("No pass"))
		return
	}
	if Counter >= max {
		ctx.JSON(200, fail("Exceed max"))
		return
	}
	var us []User
	Db.Model(&User{}).
		Where("name = ?", u.Name).
		Find(&us)
	if len(us) == 1 {
		if us[0].passMatch(u.Pass) {
			Counter = 0
			if v, exists := ctx.Get(cookieName); exists {
				value, _ := v.(string)
				authedMap[value] = true
				ctx.JSON(200, ok("Success"))
				return
			}

		}
	}
	Counter++
	ctx.JSON(200, fail("登录失败"))
}

func deleteEnvC(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Query("id"))
	if err != nil {
		ctx.JSON(200, paramError)
		return
	}
	err = deleteEnv(uint(id))
	if err != nil {
		ctx.JSON(200, fail(err.Error()))
	} else {
		ctx.JSON(200, ok(id))
	}
}

func getEnvListC(ctx *gin.Context) {
	ctx.JSON(200, ok(listEnv()))
}

func saveOrUpdateEnv(ctx *gin.Context) {
	var env Env
	err := ctx.BindJSON(&env)
	if err != nil {
		ctx.JSON(200, paramError)
		return
	}
	if env.ID == 0 {
		Db.Create(&env)
		ctx.JSON(200, ok(env.ID))
	} else {
		dbEvn, exist := getEnv(env.ID)
		if !exist {
			ctx.JSON(200, fail("Not found"))
		} else {
			dbEvn.Name = env.Name
			dbEvn.Content = env.Content
			Db.Updates(&dbEvn)
			ctx.JSON(200, ok(dbEvn.ID))
		}
	}
}

func getEnvC(ctx *gin.Context) {
	id := ctx.Query("id")
	iid, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(200, paramError)
		return
	}
	env, exist := getEnv(uint(iid))
	if !exist {
		ctx.JSON(200, paramError)
	} else {
		ctx.JSON(200, ok(env))
	}
}

func clearApp(ctx *gin.Context) {
	id := ctx.Query("id")
	app, err := queryApp(id)
	if err != nil {
		ctx.JSON(200, paramError)
		return
	}
	err = app.clear()
	if err != nil {
		ctx.JSON(200, fail(err.Error()))
	} else {
		ctx.JSON(200, ok(app.ID))
	}
}

func deleteApp(ctx *gin.Context) {
	id := ctx.Query("id")
	app, err := queryApp(id)
	if err != nil {
		ctx.JSON(200, paramError)
		return
	}
	err = app.delete()
	if err != nil {
		ctx.JSON(200, fail(err.Error()))
	} else {
		ctx.JSON(200, ok(app.ID))
	}
}

func taskQuery(ctx *gin.Context) {
	if id, err := strconv.Atoi(ctx.Query("id")); err != nil {
		ctx.JSON(200, paramError)
		return
	} else {
		task, exist := taskDetail(uint(id))
		if !exist {
			ctx.JSON(200, paramError)
		} else {
			ctx.JSON(200, ok(task))
		}
	}
}

func taskKill(ctx *gin.Context) {
	if id, err := strconv.Atoi(ctx.Query("id")); err != nil {
		ctx.JSON(200, paramError)
	} else {
		run, exist := CommonPool.get(uint(id))
		if exist {
			ct := run.(*exeCtx)
			err := syscall.Kill(-ct.cmd.Process.Pid, syscall.SIGTERM)
			if err != nil {
				ctx.JSON(500, fail(fmt.Sprintf("停止失败:%s", err.Error())))
			} else {
				ctx.JSON(200, ok("OK"))
			}
		} else {
			ctx.JSON(500, fail("任务未在运行中"))
		}
	}
}

func taskPage(ctx *gin.Context) {
	var query TaskPageQuery
	err := ctx.BindQuery(&query)
	if err != nil {
		ctx.JSON(200, paramError)
	} else {
		_, err := queryApp(strconv.Itoa(int(query.AppId)))
		if err != nil {
			ctx.JSON(200, paramError)
		} else {
			query.defaultPage()
			ctx.JSON(200, ok(queryTasks(query)))
		}
	}
}

func runTask(ctx *gin.Context) {
	app, err := queryApp(ctx.Query("id"))
	if err != nil {
		ctx.JSON(200, notFound)
	} else {
		taskId, err := app.run()
		if err != nil {
			ctx.JSON(200, fail(err.Error()))
		} else {
			ctx.JSON(200, ok(taskId))
		}
	}
}

func env(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"home": GlobalConfig.Home,
		"OS":   GlobalConfig.OS,
	})
}

//todo remove invalid envs
func applicationDetail(ctx *gin.Context) {
	app, err := queryApp(ctx.Query("id"))
	if err != nil {
		ctx.JSON(200, notFound)
	} else {
		ctx.JSON(200, ok(ApplicationX{
			ID:      app.ID,
			Name:    app.Name,
			Enable:  app.Enable,
			GitUrl:  app.GitUrl,
			Command: app.Command,
			Envs:    findByAppId(app.ID),
		}))
	}
}

func appList(c *gin.Context) {
	var apps []Application
	Db.Order("id desc").Find(&apps)
	c.JSON(200, ok(apps))
}

func saveApplication(c *gin.Context) {
	var app ApplicationX
	err := c.BindJSON(&app)
	if err != nil {
		c.JSON(200, fail("Param error"))
		return
	}
	app.fill()
	folder := parseGitFolder(app.GitUrl)
	if folder == "" {
		c.JSON(200, fail("无效的gitUrl"))
		return
	}
	if app.Name == "" {
		c.JSON(200, fail("无效的名称"))
		return
	}
	dbApp, exist := findAppByName(app.Name)

	if app.ID == 0 {
		if exist {
			c.JSON(200, fail("应用名称已占用"))
			return
		}
		toSave := Application{
			Name:    app.Name,
			Enable:  app.Enable,
			GitUrl:  app.GitUrl,
			Command: app.Command,
		}
		Db.Create(&toSave)

		err = updateUsedEnv(toSave.ID, app.Envs)
		if err != nil {
			c.JSON(200, fail(err.Error()))
			return
		}
		c.JSON(200, ok(toSave.ID))

	} else {
		if !exist {
			c.JSON(200, fail("非法状态，应用不存在"))
			return
		}
		if dbApp.Name != app.Name {
			c.JSON(200, fail("非法状态，编辑时不能修改名称"))
			return
		}
		if dbApp.GitUrl != app.GitUrl {
			c.JSON(200, fail("非法状态，编辑时不能修改gitUrl"))
			return
		}

		err = updateUsedEnv(app.ID, app.Envs)
		if err != nil {
			c.JSON(200, fail(err.Error()))
			return
		}
		err = Db.Model(&dbApp).Updates(map[string]interface{}{
			"enable":  app.Enable,
			"command": app.Command,
		}).Error
		if err != nil {
			c.JSON(200, fail("更新失败"))
		} else {
			c.JSON(200, ok(app.ID))
		}
	}
}

func health(c *gin.Context) {
	c.String(200, "Service is on v1.1")
}
