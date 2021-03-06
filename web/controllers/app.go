package controllers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dalonghahaha/avenger/components/logger"
	"github.com/gin-gonic/gin"

	"Asgard/constants"
	"Asgard/models"
	"Asgard/providers"
	"Asgard/web/utils"
)

type AppController struct {
}

func NewAppController() *AppController {
	return &AppController{}
}

func (c *AppController) List(ctx *gin.Context) {
	groupID := utils.DefaultInt(ctx, "group_id", 0)
	agentID := utils.DefaultInt(ctx, "agent_id", 0)
	status := utils.DefaultInt(ctx, "status", -99)
	name := ctx.Query("name")
	page := utils.DefaultInt(ctx, "page", 1)
	user := utils.GetUser(ctx)
	where := map[string]interface{}{
		"status": status,
	}
	querys := []string{}
	if user.Role != constants.USER_ROLE_ADMIN {
		where["creator"] = user.ID
	}
	if groupID != 0 {
		where["group_id"] = groupID
		querys = append(querys, "group_id="+strconv.Itoa(groupID))
	}
	if agentID != 0 {
		where["agent_id"] = agentID
		querys = append(querys, "agent_id="+strconv.Itoa(agentID))
	}
	if status != -99 {
		querys = append(querys, "status="+strconv.Itoa(status))
	}
	if name != "" {
		where["name"] = name
		querys = append(querys, "name="+name)
	}
	appList, total := providers.AppService.GetAppPageList(where, page, constants.WEB_LIST_PAGE_SIZE)
	if appList == nil {
		utils.APIError(ctx, "获取应用列表失败")
	}
	list := []gin.H{}
	for _, app := range appList {
		list = append(list, utils.AppFormat(&app))
	}
	mpurl := "/app/list"
	if len(querys) > 0 {
		mpurl = "/app/list?" + strings.Join(querys, "&")
	}
	utils.Render(ctx, "app/list", gin.H{
		"Subtitle":   "应用列表",
		"List":       list,
		"Total":      total,
		"GroupList":  providers.GroupService.GetUsageGroup(),
		"AgentList":  providers.AgentService.GetUsageAgent(),
		"StatusList": constants.APP_STATUS,
		"GroupID":    groupID,
		"AgentID":    agentID,
		"Name":       name,
		"Status":     status,
		"Pagination": utils.PagerHtml(total, page, mpurl),
	})
}

func (c *AppController) Show(ctx *gin.Context) {
	app := utils.GetApp(ctx)
	utils.Render(ctx, "app/show", gin.H{
		"Subtitle": "查看应用",
		"App":      utils.AppFormat(app),
	})
}

func (c *AppController) Add(ctx *gin.Context) {
	utils.Render(ctx, "app/add", gin.H{
		"Subtitle":   "添加应用",
		"OutBaseDir": constants.WEB_OUT_DIR + "guard/",
		"GroupList":  providers.GroupService.GetUsageGroup(),
		"AgentList":  providers.AgentService.GetUsageAgent(),
	})
}

func (c *AppController) Create(ctx *gin.Context) {
	if utils.FormDefaultInt64(ctx, "agent_id", 0) == 0 {
		utils.APIError(ctx, "运行实例未选择")
		return
	}
	app := new(models.App)
	app.GroupID = utils.FormDefaultInt64(ctx, "group_id", 0)
	app.AgentID = utils.FormDefaultInt64(ctx, "agent_id", 0)
	app.Name = ctx.PostForm("name")
	app.Dir = ctx.PostForm("dir")
	app.Program = ctx.PostForm("program")
	app.Args = ctx.PostForm("args")
	app.StdOut = ctx.PostForm("std_out")
	app.StdErr = ctx.PostForm("std_err")
	app.Status = constants.APP_STATUS_PAUSE
	app.Creator = utils.GetUserID(ctx)
	if ctx.PostForm("auto_restart") != "" {
		app.AutoRestart = 1
	}
	if ctx.PostForm("is_monitor") != "" {
		app.IsMonitor = 1
	}
	ok := providers.AppService.CreateApp(app)
	if !ok {
		utils.APIError(ctx, "创建应用失败")
		return
	}
	utils.OpetationLog(utils.GetUserID(ctx), constants.TYPE_APP, app.ID, constants.ACTION_CREATE)
	utils.APIOK(ctx)
}

func (c *AppController) Edit(ctx *gin.Context) {
	app := utils.GetApp(ctx)
	utils.Render(ctx, "app/edit", gin.H{
		"Subtitle":  "编辑应用",
		"BackUrl":   utils.GetReferer(ctx),
		"Info":      utils.AppFormat(app),
		"GroupList": providers.GroupService.GetUsageGroup(),
		"AgentList": providers.AgentService.GetUsageAgent(),
	})
}

func (c *AppController) Update(ctx *gin.Context) {
	app := utils.GetApp(ctx)
	app.GroupID = utils.FormDefaultInt64(ctx, "group_id", 0)
	app.Name = ctx.PostForm("name")
	app.Dir = ctx.PostForm("dir")
	app.Program = ctx.PostForm("program")
	app.Args = ctx.PostForm("args")
	app.StdOut = ctx.PostForm("std_out")
	app.StdErr = ctx.PostForm("std_err")
	app.Updator = utils.GetUserID(ctx)
	if ctx.PostForm("auto_restart") != "" {
		app.AutoRestart = 1
	} else {
		app.AutoRestart = 0
	}
	if ctx.PostForm("is_monitor") != "" {
		app.IsMonitor = 1
	} else {
		app.IsMonitor = 0
	}
	if utils.FormDefaultInt64(ctx, "agent_id", 0) != 0 {
		app.AgentID = utils.FormDefaultInt64(ctx, "agent_id", 0)
	}
	ok := providers.AppService.UpdateApp(app)
	if !ok {
		utils.APIError(ctx, "更新应用失败")
		return
	}
	utils.OpetationLog(utils.GetUserID(ctx), constants.TYPE_APP, app.ID, constants.ACTION_UPDATE)
	utils.APIOK(ctx)
}

func (c *AppController) Copy(ctx *gin.Context) {
	app := utils.GetApp(ctx)
	_app := new(models.App)
	_app.GroupID = app.GroupID
	_app.Name = app.Name + "_copy"
	_app.AgentID = app.AgentID
	_app.Dir = app.Dir
	_app.Program = app.Program
	_app.Args = app.Args
	_app.StdOut = app.StdOut
	_app.StdErr = app.StdErr
	_app.AutoRestart = app.AutoRestart
	_app.IsMonitor = app.IsMonitor
	_app.Status = constants.APP_STATUS_PAUSE
	_app.Creator = utils.GetUserID(ctx)
	ok := providers.AppService.CreateApp(_app)
	if !ok {
		utils.APIError(ctx, "复制应用失败")
		return
	}
	utils.OpetationLog(utils.GetUserID(ctx), constants.TYPE_APP, app.ID, constants.ACTION_COPY)
	utils.APIOK(ctx)
}

func (c *AppController) Start(ctx *gin.Context) {
	app := utils.GetApp(ctx)
	agent := utils.GetAgent(ctx)
	if app.Status == constants.APP_STATUS_RUNNING {
		utils.APIError(ctx, "应用已经启动")
		return
	}
	client, err := providers.GetAgent(agent)
	if err != nil {
		utils.APIError(ctx, "初始化RPC客户端异常:\n"+err.Error())
		return
	}
	_app, err := client.GetApp(app.ID)
	if err != nil {
		utils.APIError(ctx, fmt.Sprintf("获取应用情况异常:%s", err.Error()))
		return
	}
	if _app == nil {
		err = client.AddApp(app)
		if err != nil {
			utils.APIError(ctx, fmt.Sprintf("添加应用异常:%s", err.Error()))
			return
		}
	}
	ok := providers.AppService.ChangeAPPStatus(app, constants.APP_STATUS_RUNNING, utils.GetUserID(ctx))
	if !ok {
		utils.APIError(ctx, "更新应用状态失败")
		return
	}
	utils.OpetationLog(utils.GetUserID(ctx), constants.TYPE_APP, app.ID, constants.ACTION_START)
	utils.APIOK(ctx)
}

func (c *AppController) ReStart(ctx *gin.Context) {
	app := utils.GetApp(ctx)
	agent := utils.GetAgent(ctx)
	client, err := providers.GetAgent(agent)
	if err != nil {
		utils.APIError(ctx, "初始化RPC客户端异常:\n"+err.Error())
		return
	}
	_app, err := client.GetApp(app.ID)
	if err != nil {
		utils.APIError(ctx, fmt.Sprintf("获取应用情况异常:%s", err.Error()))
		return
	}
	if _app == nil {
		err = client.AddApp(app)
		if err != nil {
			utils.APIError(ctx, fmt.Sprintf("重启应用异常:%s", err.Error()))
			return
		}
	} else {
		err = client.UpdateApp(app)
		if err != nil {
			utils.APIError(ctx, fmt.Sprintf("重启应用异常:%s", err.Error()))
			return
		}
	}
	utils.OpetationLog(utils.GetUserID(ctx), constants.TYPE_APP, app.ID, constants.ACTION_RESTART)
	utils.APIOK(ctx)
}

func (c *AppController) Pause(ctx *gin.Context) {
	app := utils.GetApp(ctx)
	agent := utils.GetAgent(ctx)
	client, err := providers.GetAgent(agent)
	if err != nil {
		utils.APIError(ctx, "初始化RPC客户端异常:\n"+err.Error())
		return
	}
	_app, err := client.GetApp(app.ID)
	if err != nil {
		utils.APIError(ctx, fmt.Sprintf("获取应用情况异常:%s", err.Error()))
		return
	}
	if _app != nil {
		err = client.RemoveApp(app.ID)
		if err != nil {
			utils.APIError(ctx, fmt.Sprintf("停止应用异常:%s", err.Error()))
			return
		}
	}
	ok := providers.AppService.ChangeAPPStatus(app, constants.APP_STATUS_PAUSE, utils.GetUserID(ctx))
	if !ok {
		utils.APIError(ctx, "更新应用状态失败")
		return
	}
	utils.OpetationLog(utils.GetUserID(ctx), constants.TYPE_APP, app.ID, constants.ACTION_PAUSE)
	utils.APIOK(ctx)
}

func (c *AppController) Delete(ctx *gin.Context) {
	app := utils.GetApp(ctx)
	agent := utils.GetAgent(ctx)
	if app.Status != constants.APP_STATUS_PAUSE {
		utils.APIError(ctx, "应用启动状态不能删除")
		return
	}
	client, err := providers.GetAgent(agent)
	if err != nil {
		utils.APIError(ctx, "初始化RPC客户端异常:\n"+err.Error())
		return
	}
	_app, err := client.GetApp(app.ID)
	if err != nil {
		utils.APIError(ctx, fmt.Sprintf("获取应用情况异常:%s", err.Error()))
		return
	}
	if _app != nil {
		err = client.RemoveApp(app.ID)
		if err != nil {
			utils.APIError(ctx, fmt.Sprintf("停止应用异常:%s", err.Error()))
			return
		}
	}
	ok := providers.AppService.ChangeAPPStatus(app, constants.APP_STATUS_DELETED, utils.GetUserID(ctx))
	if !ok {
		utils.APIError(ctx, "删除应用失败")
		return
	}
	utils.OpetationLog(utils.GetUserID(ctx), constants.TYPE_APP, app.ID, constants.ACTION_DELETE)
	utils.APIOK(ctx)
}

func (c *AppController) BatchStart(ctx *gin.Context) {
	appAgent := utils.GetAppAgent(ctx)
	for app, agent := range appAgent {
		if app.Status == constants.APP_STATUS_RUNNING {
			continue
		}
		client, err := providers.GetAgent(agent)
		if err != nil {
			logger.Errorf("App BatchStart GetAgent Error:[%d][%s]", app.ID, err.Error())
			continue
		}
		_app, err := client.GetApp(app.ID)
		if err != nil {
			logger.Errorf("App BatchStart GetApp Error:[%d][%s]", app.ID, err.Error())
			continue
		}
		if _app == nil {
			err = client.AddApp(app)
			if err != nil {
				logger.Errorf("App BatchStart AddApp Error:%s", err.Error())
			}
		}
		providers.AppService.ChangeAPPStatus(app, constants.APP_STATUS_RUNNING, utils.GetUserID(ctx))
		utils.OpetationLog(utils.GetUserID(ctx), constants.TYPE_APP, app.ID, constants.ACTION_START)
	}
	utils.APIOK(ctx)
}

func (c *AppController) BatchReStart(ctx *gin.Context) {
	appAgent := utils.GetAppAgent(ctx)
	for app, agent := range appAgent {
		client, err := providers.GetAgent(agent)
		if err != nil {
			logger.Errorf("App BatchReStart GetAgent Error:[%d][%s]", app.ID, err.Error())
			continue
		}
		_app, err := client.GetApp(app.ID)
		if err != nil {
			logger.Errorf("App BatchReStart GetAgentApp Error:[%d][%s]", app.ID, err.Error())
			continue
		}
		if _app == nil {
			err = client.AddApp(app)
			if err != nil {
				logger.Errorf("App BatchReStart AddAgentApp Error:[%d][%s]", app.ID, err.Error())
			}
		} else {
			err = client.UpdateApp(app)
			if err != nil {
				logger.Errorf("App BatchReStart UpdateAgentApp Error:[%d][%s]", app.ID, err.Error())
			}
		}
		utils.OpetationLog(utils.GetUserID(ctx), constants.TYPE_APP, app.ID, constants.ACTION_RESTART)
	}
	utils.APIOK(ctx)
}

func (c *AppController) BatchPause(ctx *gin.Context) {
	appAgent := utils.GetAppAgent(ctx)
	for app, agent := range appAgent {
		client, err := providers.GetAgent(agent)
		if err != nil {
			logger.Errorf("App BatchPause GetAgent Error:[%d][%s]", app.ID, err.Error())
			continue
		}
		_app, err := client.GetApp(app.ID)
		if err != nil {
			logger.Errorf("App BatchPause GetAgentApp Error:[%d][%s]", app.ID, err.Error())
			continue
		}
		if _app != nil {
			err = client.RemoveApp(app.ID)
			if err != nil {
				logger.Errorf("App BatchPause RemoveAgentApp Error:[%d][%s]", app.ID, err.Error())
				return
			}
		}
		providers.AppService.ChangeAPPStatus(app, constants.APP_STATUS_PAUSE, utils.GetUserID(ctx))
		utils.OpetationLog(utils.GetUserID(ctx), constants.TYPE_APP, app.ID, constants.ACTION_PAUSE)
	}
	utils.APIOK(ctx)
}

func (c *AppController) BatchDelete(ctx *gin.Context) {
	appAgent := utils.GetAppAgent(ctx)
	for app, agent := range appAgent {
		if app.Status == constants.APP_STATUS_RUNNING {
			continue
		}
		client, err := providers.GetAgent(agent)
		if err != nil {
			logger.Errorf("App BatchDelete GetAgent Error:[%d][%s]", app.ID, err.Error())
			continue
		}
		_app, err := client.GetApp(app.ID)
		if err != nil {
			logger.Errorf("App BatchDelete GetAgentApp Error:[%d][%s]", app.ID, err.Error())
			continue
		}
		if _app != nil {
			err = client.RemoveApp(app.ID)
			if err != nil {
				logger.Errorf("App BatchDelete RemoveAgentApp Error:[%d][%s]", app.ID, err.Error())
				return
			}
		}
		providers.AppService.ChangeAPPStatus(app, constants.APP_STATUS_DELETED, utils.GetUserID(ctx))
		utils.OpetationLog(utils.GetUserID(ctx), constants.TYPE_APP, app.ID, constants.ACTION_DELETE)
	}
	utils.APIOK(ctx)
}
