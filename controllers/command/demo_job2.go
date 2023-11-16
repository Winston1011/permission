package command

import (
	"permission/helpers"
	"permission/models/demo"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"permission/pkg/golib/v2/zlog"
)

func DemoJob2(ctx *gin.Context, args ...string) error {
	p := &demo.NormalPage{
		No:   1,
		Size: 10,
	}
	o := &demo.FilterOption{
		IsNeedList: true,
	}
	demoList, _, err := demo.GetNormalList(ctx, o, p)
	if err != nil {
		return err
	}

	// cycle 任务中添加Notice
	zlog.AddField(ctx, zlog.String("demoCycle", "start"))

	/*
		在server.log 中打印出Notice信息
		用户自定义的 Notice 信息会在 access.log 中打印出来，理论上不需要用户主动打印
		但有些业务为了方便日志采集会有在server.log中打印的需求
	*/
	zlog.PrintFields(ctx)

	for _, d := range demoList {
		tmp := d
		helpers.Job.Run(ctx, func(newCtx *gin.Context) error {
			/*
				此处要规范使用ctx
				该函数是在Job中调用的，gin.Context 是新生成的，如果这里错误的使用了原有的ctx
				可能会导致多个协程操作同一个context的问题，引发 Context内部map并发不安全的问题
			*/
			zlog.Debugf(newCtx, "demo is ID=%d and name=%s", tmp.ID, tmp.Name)
			_, err := demo.GetDemoByID(newCtx, tmp.ID)
			if err != nil {
				return errors.WithMessagef(err, "once job run service GetDemoByName failed, input: %s， error: %s", d.Name, err.Error())
			}

			// job 中添加Notice
			zlog.AddField(newCtx, zlog.String("demoCycleJob", "success"))
			// server.log中打印Notice信息
			zlog.PrintFields(newCtx)
			return nil
		})
	}
	return nil
}
