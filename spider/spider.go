package spider

import (
	"database/sql"

	"github.com/gocolly/colly"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// TODO: Context添加KV功能，能够结束请求链功能
// TODO: 思考出错, 中断后续爬虫的方法
func Run(task *Task) (<-chan struct{}, error) {
	var db *sql.DB
	var err error
	if task.OutputConfig.Type == OutputTypeMySQL {
		db, err = newDB(task.TaskConfig.OutputConfig.MySQLConf)
		if err != nil {
			logrus.Errorf("newDB failed! err:%#v", err)
			return nil, err
		}
	}

	nodesLen := len(task.Rule.Nodes)
	collectors := make([]*colly.Collector, 0, nodesLen)
	for i := 0; i < len(task.Rule.Nodes); i++ {
		c := newCollector(task.TaskConfig)
		if task.DisableCookies {
			c.DisableCookies()
		}
		collectors = append(collectors, c)
	}

	var ctx *Context
	for i := 0; i < nodesLen; i++ {
		if i != nodesLen-1 {
			ctx = newContext(task, collectors[i], collectors[i+1])
		} else {
			ctx = newContext(task, collectors[i], nil)
		}
		if task.OutputConfig.Type == OutputTypeMySQL {
			ctx.setOutputDB(db)
		}

		addCallback(ctx, task.Rule.Nodes[i])
	}

	c := newCollector(task.TaskConfig)
	headCtx := newContext(task, c, collectors[0])
	if err := task.Rule.Head(headCtx); err != nil {
		logrus.Errorf("exec rule head func err:%#v", err)
		return nil, errors.WithStack(err)
	}

	retCh := make(chan struct{}, 1)
	go func() {
		for i := 0; i < nodesLen; i++ {
			collectors[i].Wait()
		}
		retCh <- struct{}{}
		logrus.Infof("task:%s run completed...", task.Name)
	}()

	return retCh, nil
}

func addCallback(ctx *Context, node *Node) {
	if node.OnRequest != nil {
		ctx.c.OnRequest(func(req *colly.Request) {
			node.OnRequest(ctx, newRequest(req, ctx))
		})
	}

	if node.OnError != nil {
		ctx.c.OnError(func(res *colly.Response, e error) {
			node.OnError(ctx, newResponse(res, ctx), e)
		})
	}

	if node.OnResponse != nil {
		ctx.c.OnResponse(func(res *colly.Response) {
			node.OnResponse(ctx, newResponse(res, ctx))
		})
	}

	if node.OnHTML != nil {
		for selector, fn := range node.OnHTML {
			ctx.c.OnHTML(selector, func(el *colly.HTMLElement) {
				fn(ctx, newHTMLElement(el, ctx))
			})
		}
	}

	if node.OnXML != nil {
		for selector, fn := range node.OnXML {
			ctx.c.OnXML(selector, func(el *colly.XMLElement) {
				fn(ctx, newXMLElement(el, ctx))
			})
		}
	}

	if node.OnScraped != nil {
		ctx.c.OnScraped(func(res *colly.Response) {
			node.OnScraped(ctx, newResponse(res, ctx))
		})
	}

}