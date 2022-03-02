package service

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/nange/gospider/web/core"
	log "github.com/sirupsen/logrus"

	"github.com/nange/gospider/common"
	"github.com/nange/gospider/spider"
	"github.com/nange/gospider/web/model"
	"github.com/pkg/errors"
)

func GetSpiderTaskByModel(task *model.Task) (*spider.Task, error) {
	rule, err := spider.GetTaskRule(task.TaskRuleName)
	if err != nil {
		return nil, err
	}

	var optAllowedDomains []string
	if task.OptAllowedDomains != "" {
		optAllowedDomains = strings.Split(task.OptAllowedDomains, ",")
	}
	var urlFiltersReg []*regexp.Regexp
	if task.OptURLFilters != "" {
		urlFilters := strings.Split(task.OptURLFilters, ",")
		for _, v := range urlFilters {
			reg, err := regexp.Compile(v)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			urlFiltersReg = append(urlFiltersReg, reg)
		}
	}

	edb := model.ExportDB{}
	if task.OutputType == common.OutputTypeMySQL {
		query := model.NewExportDBQuerySet(core.GetGormDB())
		if err := query.IDEq(task.OutputExportDBID).One(&edb); err != nil {
			return nil, errors.Wrapf(err, "task.OutputExportDBID [%v]", task.OutputExportDBID)
		}
		if hasOutputConstraints(rule) && task.AutoMigrate {
			err = autoMigrate(&edb, rule)
			if err != nil {
				log.Errorf("autoMigrate err [%+v]", errors.WithStack(err))
			}
		}
	}

	config := spider.TaskConfig{
		CronSpec: task.CronSpec,
		Option: spider.Option{
			UserAgent:              task.OptUserAgent,
			MaxDepth:               task.OptMaxDepth,
			AllowedDomains:         optAllowedDomains,
			URLFilters:             urlFiltersReg,
			AllowURLRevisit:        rule.AllowURLRevisit,
			MaxBodySize:            task.OptMaxBodySize,
			IgnoreRobotsTxt:        rule.IgnoreRobotsTxt,
			InsecureSkipVerify:     rule.InsecureSkipVerify,
			ParseHTTPErrorResponse: rule.ParseHTTPErrorResponse,
			DisableCookies:         rule.DisableCookies,
		},
		Limit: spider.Limit{
			Enable:      task.LimitEnable,
			DomainGlob:  task.LimitDomainGlob,
			Delay:       time.Duration(task.LimitDelay) * time.Millisecond,
			RandomDelay: time.Duration(task.LimitRandomDelay) * time.Millisecond,
			Parallelism: task.LimitParallelism,
		},
		OutputConfig: spider.OutputConfig{
			Type: task.OutputType,
		},
	}
	if task.OutputType == common.OutputTypeMySQL {
		config.OutputConfig.MySQLConf = common.MySQLConf{
			Host:     edb.Host,
			Port:     edb.Port,
			User:     edb.User,
			Password: edb.Password,
			DBName:   edb.DBName,
		}
	} else if task.OutputType == common.OutputTypeCSV {
		config.OutputConfig.CSVConf = spider.CSVConf{
			CSVFilePath: "./csv_output",
		}
	}

	if task.OptRequestTimeout > 0 {
		config.Option.RequestTimeout = time.Duration(task.OptRequestTimeout) * time.Second
	}
	if urls := strings.TrimSpace(task.ProxyURLs); len(urls) > 0 {
		config.ProxyURLs = strings.Split(urls, ",")
	} else {
		config.ProxyURLs = getProxyList()
		config.ProxyURLs = []string{"http://1.180.156.226:65001/"}
	}

	return spider.NewTask(task.ID, *rule, config), nil
}

func hasOutputConstraints(rule *spider.TaskRule) (b bool) {
	if rule.OutputToMultipleNamespace {
		for k := range rule.MultipleNamespaceConf {
			if len(rule.MultipleNamespaceConf[k].OutputConstraints) > 0 {
				b = true
				return
			}
		}
	} else {
		b = len(rule.OutputConstraints) > 0
	}

	return
}

func autoMigrate(sdb *model.ExportDB, rule *spider.TaskRule) (err error) {
	db, err := common.NewGormDB(common.MySQLConf{
		Host:     sdb.Host,
		Port:     sdb.Port,
		User:     sdb.User,
		Password: sdb.Password,
		DBName:   sdb.DBName,
	})

	if err != nil {
		return
	}
	defer db.Close()

	err = spider.AutoMigrateHack(db, rule).Error
	return
}

func getProxyList() []string {
	var proxy_urls []string
	url := "http://127.0.0.1:9999/sql?query=SELECT%20*%20FROM%20proxy%20ORDER%20BY%20RAND()%20limit%2010"
	client := &http.Client{}
	res, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return proxy_urls
	}
	response, _ := client.Do(res)
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return proxy_urls
	}
	re_proxy := struct {
		Err string `json:"err"`
		Message []struct{
			Id int `json:"id"`
			Content string `json:"content"`
		} `json:"message"`
	}{}
	err = json.Unmarshal(body, &re_proxy)

	for _, item := range re_proxy.Message {
		proxy_urls = append(proxy_urls, "http://" + item.Content)
	}
	return proxy_urls
}
