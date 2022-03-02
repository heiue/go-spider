package dianpindesigner

import (
	"fmt"
	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/nange/gospider/common"
	"github.com/nange/gospider/spider"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"strconv"
	"time"
)

func init() {
	spider.Register(rule)
}

var (
	// 北京、上海、广州、深圳、杭州、成都、重庆、武汉、南京、西安、长沙、郑州、苏州、天津、东莞、青岛、宁波、合肥、佛山、济南
	//city = []string{"北京", "上海", "广州", "深圳", "杭州", "成都", "重庆", "武汉", "南京", "西安", "长沙", "郑州", "苏州", "天津",}
	//"东莞", "青岛", "宁波", "合肥", "佛山1", "济南"}
	outputFields = []string{"cid", "城市", "公司名", "开店时间", "评分", "点评数", "均价", "设计方案数", "设计师数", "城区", "地址", "电话"}
	//outputFields=[]string{"city", "company_name", "dianping", "jujia", "fananshu", "shejishi", "chengqu", "bankuan", "dizhi", "dianhua"}
	constraints = spider.NewConstraints(outputFields,
		"VARCHAR(128) NOT NULL DEFAULT ''",
		"VARCHAR(128) NOT NULL DEFAULT ''",
		"VARCHAR(128) NOT NULL DEFAULT ''",
		"VARCHAR(128) NOT NULL DEFAULT ''",
		"VARCHAR(128) NOT NULL DEFAULT ''",
		"VARCHAR(128) NOT NULL DEFAULT ''",
		"VARCHAR(128) NOT NULL DEFAULT ''",
		"VARCHAR(128) NOT NULL DEFAULT ''",
		"VARCHAR(128) NOT NULL DEFAULT ''",
		"VARCHAR(128) NOT NULL DEFAULT ''",
		"VARCHAR(128) NOT NULL DEFAULT ''",
		"VARCHAR(128) NOT NULL DEFAULT ''",
	)
)

// NOTICE:
var rule = &spider.TaskRule{
	Name:              "大众点评装修",
	Description:       "抓取大众点评上全国各大城市所有装修",
	Namespace:         "dianping_designer",
	OutputFields:      outputFields,
	OutputConstraints: constraints,
	AllowURLRevisit:   true,
	Rule: &spider.Rule{
		Head: func(ctx *spider.Context) error { // 定义入口
			//a := pinyin.NewArgs()
			//r := pinyin.LazyPinyin(v_c, a)
			//c := strings.Join(r, "")
			//fmt.Printf("------------抓取当前城市: %v \n", c)
			cy := "tianjin"
			c_name := "天津"
			ctx.PutReqContextValue("city", c_name)
			/*i := 1
			for i <= 50 {
				err := ctx.VisitForNextWithContext("https://www.dianping.com/" + cy + "/ch90/g25475o2" + "p" + strconv.Itoa(i))
				//time.Sleep(time.Second * 70)
				if err != nil {
					log.Error("!!!!!!!!!!!!列表页需要验证")
					break
				}
				t := time.Tick(time.Second * 1)
				for countdown := 60; countdown > 0; countdown-- {
					fmt.Printf("\r%2d", countdown)
					<-t
				}
				i++
			}*/
			url := "https://www.dianping.com/" + cy + "/ch90/g25475o2"
			ctx.PutReqContextValue("url", url)
			ctx.VisitForNextWithContext(url)
			return nil
		},
		Nodes: map[int]*spider.Node{
			0: step1,
			1: step2,
			2: step3,
		},
	},
}
var step1 = &spider.Node{
	OnRequest: func(ctx *spider.Context, req *spider.Request) {
		req.Headers.Set("User-Agent", RandomString())
		//req.Headers.Set("Cookie", "fspop=test; _lxsdk_cuid=17f2463ad0dc8-0ce514567a0911-133f685c-13c680-17f2463ad0dc8; _lxsdk=17f2463ad0dc8-0ce514567a0911-133f685c-13c680-17f2463ad0dc8; _hc.v=05ed32bb-1a0b-fcf6-6d21-99d9f59fdcdf.1645582987; s_ViewType=10; Hm_lvt_602b80cf8079ae6591966cc70a3940e7=1645582989,1645777158; Hm_lvt_dbeeb675516927da776beeb1d9802bd4=1645583121,1645777161; cityid=2; default_ab=shopList%3AA%3A5; cy=2; cye=beijing; Hm_lpvt_dbeeb675516927da776beeb1d9802bd4=1646037048; Hm_lpvt_602b80cf8079ae6591966cc70a3940e7=1646037048; _lxsdk_s=17f3f72f402-6fb-766-9f0%7C%7C24")
		log.Infof("Visiting %s", req.URL.String())
	},
	OnError: func(ctx *spider.Context, res *spider.Response, err error) error {
		log.Errorf("Visiting failed! url:%s, err:%s", res.Request.URL.String(), err.Error())
		return Retry(ctx, 3)
	},
	OnHTML: map[string]func(ctx *spider.Context, el *spider.HTMLElement) error{
		`.pages a:nth-last-child(2)`: func(ctx *spider.Context, el *spider.HTMLElement) error {
			countTxt := el.Text
			if countTxt == "" {
				return nil
			}

			count64, err := strconv.ParseInt(countTxt, 10, 64)
			if err != nil {
				log.Errorf("pase page count err:%s", err.Error())
				return nil
			}
			fmt.Printf("countTotal : %v \n", count64)
			for i := 1; i <= int(count64); i++ {
				nextURL := fmt.Sprintf("%sp%d", el.Request.URL.String(), i)
				fmt.Printf("----------------------nextURL:%s \n", nextURL)
				ctx.VisitForNext(nextURL)
				//t := time.Tick(time.Second * 1)
				//for countdown := 10; countdown > 0; countdown-- {
				//	fmt.Printf("\r%2d", countdown)
				//	<-t
				//}
			}
			return nil
		},
	},
	//OnResponse: func(ctx *spider.Context, res *spider.Response) error {
	//	if ctx.GetReqContextValue("url") != res.Request.URL.String() {
	//		return nil
	//	}
	//	ctx.GetReqContextValue("city")
	//	for i := 1; i <= 2; i++ {
	//		fmt.Println("------------------------------------------第" + strconv.Itoa(i) + "页------------------------------------")
	//		ctx.VisitForNextWithContext(res.Request.URL.String() + "p" + strconv.Itoa(i))
	//		ctx.C().Wait()
	//		/*t := time.Tick(time.Second * 1)
	//		for countdown := 30; countdown > 0; countdown-- {
	//			fmt.Printf("\r%2d", countdown)
	//			<-t
	//		}*/
	//	}
	//	return nil
	//},
}

var step2 = &spider.Node{
	OnRequest: func(ctx *spider.Context, req *spider.Request) {
		req.Headers.Set("User-Agent", RandomString())
		//req.Headers.Set("Cookie", "fspop=test; _lxsdk_cuid=17f2463ad0dc8-0ce514567a0911-133f685c-13c680-17f2463ad0dc8; _lxsdk=17f2463ad0dc8-0ce514567a0911-133f685c-13c680-17f2463ad0dc8; _hc.v=05ed32bb-1a0b-fcf6-6d21-99d9f59fdcdf.1645582987; s_ViewType=10; Hm_lvt_602b80cf8079ae6591966cc70a3940e7=1645582989,1645777158; Hm_lvt_dbeeb675516927da776beeb1d9802bd4=1645583121,1645777161; cityid=2; default_ab=shopList%3AA%3A5; cy=2; cye=beijing; Hm_lpvt_dbeeb675516927da776beeb1d9802bd4=1646037048; Hm_lpvt_602b80cf8079ae6591966cc70a3940e7=1646037048; _lxsdk_s=17f3f72f402-6fb-766-9f0%7C%7C24")
		log.Infof("Visiting %s", req.URL.String())
	},
	OnError: func(ctx *spider.Context, res *spider.Response, err error) error {
		log.Errorf("Visiting failed! url:%s, err:%s", res.Request.URL.String(), err.Error())
		time.Sleep(time.Second * 1000)
		// 出错时重试三次
		return Retry(ctx, 3)
	},
	OnHTML: map[string]func(*spider.Context, *spider.HTMLElement) error{
		`.shop-list-item`: func(ctx *spider.Context, el *spider.HTMLElement) error {
			id := el.Attr("data-id")
			company_name := el.ChildText(".shop-title a")
			company_time := el.ChildText(".shop-title .shop-keys a")
			company_score := el.ChildText(".shop-info .dp-sml-score")
			dianping_count := el.ChildText(".shop-info-text-i>a")
			jujia := el.ChildText(".shop-info-text-i .ml-26")
			shejifangan_count := el.ChildText(".shop-team a:nth-of-type(1) i")
			shejishi_count := el.ChildText(".shop-team a:nth-of-type(2) i")
			chengqu := el.ChildText(".shop-location span")
			company_url := el.ChildAttr(".shop-title a", "href")

			ctx.PutReqContextValue("id", id)
			ctx.PutReqContextValue("company_name", company_name)
			ctx.PutReqContextValue("company_time", company_time)
			ctx.PutReqContextValue("company_score", company_score)
			ctx.PutReqContextValue("dianping_count", dianping_count)
			ctx.PutReqContextValue("jujia", jujia)
			ctx.PutReqContextValue("shejifangan_count", shejifangan_count)
			ctx.PutReqContextValue("shejishi_count", shejishi_count)
			ctx.PutReqContextValue("chengqu", chengqu)
			ctx.PutReqContextValue("company_url", company_url)

			return ctx.VisitForNextWithContext("http://m.dianping.com/shop/" + id)
		},
	},
}

var step3 = &spider.Node{
	OnRequest: func(ctx *spider.Context, req *spider.Request) {
		req.Headers.Set("User-Agent", RandomString())
		//req.Headers.Add("Cookie", "fspop=test; _lxsdk_cuid=17f2463ad0dc8-0ce514567a0911-133f685c-13c680-17f2463ad0dc8; _lxsdk=17f2463ad0dc8-0ce514567a0911-133f685c-13c680-17f2463ad0dc8; _hc.v=05ed32bb-1a0b-fcf6-6d21-99d9f59fdcdf.1645582987; Hm_lvt_602b80cf8079ae6591966cc70a3940e7=1645582989; s_ViewType=10; Hm_lvt_dbeeb675516927da776beeb1d9802bd4=1645583121; m_flash2=1; cityid=2; default_ab=shopList%3AC%3A5; pvhistory=6L+U5ZuePjo8L3N1Z2dlc3QvZ2V0SnNvbkRhdGE+OjwxNjQ1NjkxODQwODUzXV9b; cy=219; cye=dongguan; Hm_lpvt_602b80cf8079ae6591966cc70a3940e7=1645698737; Hm_lpvt_dbeeb675516927da776beeb1d9802bd4=1645698737; _lxsdk_s=17f2b360c6f-504-d2-ff6%7C%7C56")
		log.Infof("Visiting %s", req.URL.String())
	},
	OnError: func(ctx *spider.Context, res *spider.Response, err error) error {
		log.Errorf("Visiting failed! url:%s, err:%s", res.Request.URL.String(), err.Error())
		return Retry(ctx, 3)
	},
	OnHTML: map[string]func(*spider.Context, *spider.HTMLElement) error{
		`.shop-details`: func(ctx *spider.Context, el *spider.HTMLElement) error {
			//dizhi := el.ChildAttr(".address span", "title")
			//tel := el.ChildText(".telAndQQ strong")
			dizhi := el.ChildText(".info-details a span")
			tel := el.ChildAttr("#telphone", "href")

			id := ctx.GetReqContextValue("id")
			c := ctx.GetReqContextValue("city")
			company_name := ctx.GetReqContextValue("company_name")
			company_time := ctx.GetReqContextValue("company_time")
			company_score := ctx.GetReqContextValue("company_score")
			dianping_count := ctx.GetReqContextValue("dianping_count")
			jujia := ctx.GetReqContextValue("jujia")
			shejifangan_count := ctx.GetReqContextValue("shejifangan_count")
			shejishi_count := ctx.GetReqContextValue("shejishi_count")
			chengqu := ctx.GetReqContextValue("chengqu")

			return ctx.Output(map[int]interface{}{
				0:  id,
				1:  c,
				2:  company_name,
				3:  company_time,
				4:  company_score,
				5:  dianping_count,
				6:  jujia,
				7:  shejifangan_count,
				8:  shejishi_count,
				9:  chengqu,
				10: dizhi,
				11: tel,
			})
		},
	},
}

func Retry(ctx *spider.Context, count int) error {
	req := ctx.GetRequest()
	key := fmt.Sprintf("err_req_%s", req.URL.String())

	var et int
	if errCount := ctx.GetAnyReqContextValue(key); errCount != nil {
		et = errCount.(int)
		if et >= count {
			return fmt.Errorf("exceed %d counts", count)
		}
	}
	log.Infof("重试............errCount:%d, we wil retry url:%s, after 1 second", et+1, req.URL.String())
	time.Sleep(time.Second)
	ctx.PutReqContextValue(key, et+1)
	ctx.Retry()

	return nil
}

func RandomString() string {
	uA := browser.Random()
	fmt.Printf("Random User-Agent: %v \n", uA)
	return uA
}

func GetRandUserAgent() string {
	n := rand.Intn(80)
	fmt.Println("getRandUserAgent : " + common.User_agent[n])
	return common.User_agent[n]
}
