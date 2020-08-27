package backend

import (
	"github.com/go-xorm/xorm"
	"github.com/xiusin/pine"
	"github.com/xiusin/pine/cache"
	"github.com/xiusin/pinecms/src/application/models"
	"github.com/xiusin/pinecms/src/application/models/tables"
	"github.com/xiusin/pinecms/src/common/helper"
	"html/template"
	"strconv"
	"strings"
)

type SystemController struct {
	pine.Controller
}

func (c *SystemController) RegisterRoute(b pine.IRouterWrapper) {
	b.GET("/system/menulist", "MenuList")
	b.POST("/system/menu-edit", "MenuEdit")
	b.POST("/system/menu-quick-save", "MenuQuickSave")
	b.POST("/system/menu-add", "MenuAdd")
	b.POST("/system/menu-delete", "MenuDelete")
	b.POST("/system/menu-order", "MenuOrder")
	b.ANY("/system/menu-tree", "MenuSelectTree")
	b.POST("/system/menu-check", "MenuCheck")
	b.ANY("/system/loglist", "LogList")
	b.ANY("/system/tail", "TailList")
	b.ANY("/system/log-delete", "LogDelete")
	b.ANY("/system/ws-connection", "WsConnection")
}

func (c *SystemController) MenuList() {
	helper.Ajax(models.NewMenuModel().GetTree(models.NewMenuModel().GetAll(), 0), 0, c.Ctx())
}

func (c *SystemController) MenuAdd(iCache cache.AbstractCache) {
	menu := &tables.Menu{}
	if err := c.Ctx().BindForm(menu); err != nil {
		helper.Ajax(err, 1, c.Ctx())
		return
	}
	if id, _ := c.Ctx().Value("orm").(*xorm.Engine).InsertOne(menu); id > 0 {
		clearMenuCache(iCache, c.Ctx().Value("orm").(*xorm.Engine))
		helper.Ajax("新增菜单成功", 0, c.Ctx())
	} else {
		helper.Ajax("新增菜单失败", 1, c.Ctx())
	}
	return
}

func (c *SystemController) MenuDelete(iCache cache.AbstractCache, orm *xorm.Engine) {
	id, _ := c.Ctx().PostInt64("id", 0)
	if id < 1 {
		helper.Ajax("参数失败", 1, c.Ctx())
		return
	}
	if exists, _ := orm.Where("parentid = ?", id).Count(&tables.Menu{}); exists > 0 {
		helper.Ajax("删除菜单失败,有下级菜单", 1, c.Ctx())
		return
	}
	if row, _ := c.Ctx().Value("orm").(*xorm.Engine).Id(id).Delete(&tables.Menu{}); row > 0 {
		clearMenuCache(iCache, c.Ctx().Value("orm").(*xorm.Engine))
		helper.Ajax("删除菜单成功", 0, c.Ctx())
	} else {
		helper.Ajax("删除菜单失败", 1, c.Ctx())
	}
}

func (c *SystemController) MenuOrder(iCache cache.AbstractCache) {
	posts := c.Ctx().PostData()
	data := map[int64]int64{}
	for k, v := range posts {
		k = strings.Replace(k, "order[", "", 1)
		k = strings.Replace(k, "]", "", 1)
		s, e := strconv.Atoi(k)
		if e != nil {
			continue
		}
		sort, e := strconv.Atoi(v[0])
		if e != nil {
			continue
		}
		data[int64(s)] = int64(sort)
	}
	var flag int64
	for id, val := range data {
		menu := new(tables.Menu)
		menu.Listorder = val
		affected, _ := c.Ctx().Value("orm").(*xorm.Engine).Id(id).Update(menu)
		if affected > 0 {
			flag++
		}
	}
	if flag > 0 {
		clearMenuCache(iCache, c.Ctx().Value("orm").(*xorm.Engine))
		helper.Ajax("排序更新成功", 0, c.Ctx())
	} else {
		helper.Ajax("排序规则没有发生任何改变", 1, c.Ctx())
	}
}

func (c *SystemController) MenuEdit(iCache cache.AbstractCache, orm *xorm.Engine) {
	menu := &tables.Menu{}
	var err error
	if string(c.Ctx().Request.Header.ContentType()) == "application/json" {
		err = c.Ctx().BindJSON(menu)
	} else {
		err = c.Ctx().BindForm(menu)
	}
	if err != nil {
		helper.Ajax(err.Error(), 1, c.Ctx())
		return
	}
	if menu.Id == 0 {
		helper.Ajax("请选择要操作的数据", 1, c.Ctx())
		return
	}
	if _, exist := models.NewMenuModel().GetInfo(menu.Id); !exist {
		helper.Ajax("没有此菜单", 1, c.Ctx())
		return
	}
	if result, _ := orm.Id(menu.Id).Update(menu); result > 0 {
		clearMenuCache(iCache, orm)
		helper.Ajax("修改菜单成功", 0, c.Ctx())
	} else {
		helper.Ajax("修改菜单失败", 1, c.Ctx())
	}
}

func (c *SystemController) MenuQuickSave(iCache cache.AbstractCache, orm *xorm.Engine) {
	c.MenuEdit(iCache, orm)
}

func (c *SystemController) MenuSelectTree() {
	var tree []map[string]interface{}
	tree = append(tree, map[string]interface{}{
		"value":    0,
		"label":    "作为一级菜单",
		"children": models.NewMenuModel().GetSelectTree(models.NewMenuModel().GetAll(), 0),
	})
	helper.Ajax(tree, 0, c.Ctx())
}

func (c *SystemController) MenuCheck() {
	name := c.Ctx().FormValue("name")
	if name == "" {
		helper.Ajax("用户名为空", 1, c.Ctx())
		return
	}
	if models.NewMenuModel().CheckName(name) {
		helper.Ajax("用户名已存在", 1, c.Ctx())
		return
	}

	helper.Ajax("正常", 0, c.Ctx())
}

func (c *SystemController) WsConnection() {
	//if conn, err := upgrader.Upgrade(c.Ctx().Writer(), c.Ctx().Request(), nil); err != nil {
	//	//报错了，直接返回底层的websocket链接就会终断掉
	//	pine.Logger().Print("连接ws错误", err)
	//	return
	//} else {
	//	conf := pine.Make("pinecms.config").(*config.Config)
	//	logPath := filepath.Join(conf.LogPath, "pinecms.log")
	//	fileInfo, err := os.Stat(logPath)
	//	if err != nil {
	//		return
	//	}
	//	size := fileInfo.Size()
	//	ta, err := tail.TailFile(logPath, tail.Config{
	//		ReOpen:      true,
	//		MustExist:   true,
	//		Poll:        false,
	//		Follow:      true,
	//		MaxLineSize: 0,
	//	})
	//	if err != nil {
	//		return
	//	}
	//	for text := range ta.Lines {
	//		size -= int64(len(text.Text) + 1) // 添加一个换行符字节
	//		if size <= 2048 {
	//			if err = conn.WriteMessage(websocket.TextMessage, []byte(text.Text)); err != nil {
	//				pine.Logger().Error("tail log failed", err)
	//				_ = ta.Stop()
	//				ta.Cleanup()
	//				ta = nil
	//				return
	//			}
	//		}
	//	}
	//}
}

func (c *SystemController) TailList() {
	// 扫描日志文件可选日志内容 设置最大行号
	c.Ctx().Render().HTML("backend/system_tail.html")
}

func (c *SystemController) LogList() {
	page, _ := c.Ctx().GetInt64("page")
	rows, _ := c.Ctx().GetInt64("rows")

	if page > 0 {
		list, total := models.NewLogModel().GetList(page, rows)
		c.Ctx().Render().JSON(map[string]interface{}{"rows": list, "total": total})
		return
	}

	menuid, _ := c.Ctx().GetInt64("menuid")
	table := helper.Datagrid("system_menu_logList", "/b/system/loglist", helper.EasyuiOptions{
		"title":   models.NewMenuModel().CurrentPos(menuid),
		"toolbar": "system_loglist_datagrid_toolbar",
	}, helper.EasyuiGridfields{
		"用户名": {"field": "Username", "width": "20", "index": "0"},
		"模块":  {"field": "Controller", "width": "15", "index": "1"},
		"方法":  {"field": "Action", "width": "15", "index": "2"},
		"URL": {"field": "Querystring", "width": "100", "formatter": "systemLogViewFormatter", "index": "3"},
		"时间":  {"field": "Time", "width": "30", "index": "4"},
		"IP":  {"field": "Ip", "width": "25", "index": "5"},
	})
	c.Ctx().Render().ViewData("dataGrid", template.HTML(table))
	c.Ctx().Render().HTML("backend/system_loglist.html")

}

func (c *SystemController) LogDelete() {
	if models.NewLogModel().DeleteAll() {
		helper.Ajax("清空日志成功", 0, c.Ctx())
	} else {
		helper.Ajax("清空日志失败", 1, c.Ctx())
	}
}
