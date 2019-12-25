package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/chanxuehong/wechat/mp/core"
	"github.com/chanxuehong/wechat/mp/material"
	"html/template"
	"strings"
	"tesou.io/platform/foot-parent/foot-api/common/base"
	"tesou.io/platform/foot-parent/foot-api/module/suggest/vo"
	"tesou.io/platform/foot-parent/foot-core/common/base/service/mysql"
	"tesou.io/platform/foot-parent/foot-core/common/utils"
	"tesou.io/platform/foot-parent/foot-core/module/analy/service"
	service2 "tesou.io/platform/foot-parent/foot-core/module/suggest/service"
	"time"
)

type SuggestWeekService struct {
	mysql.BaseService
	service.AnalyService
	service2.SuggestService
}

func (this *SuggestWeekService) Week(wcClient *core.Client) string {
	listData := this.AnalyService.ListDefaultData()
	articles := make([]material.Article, 1)
	first := material.Article{}
	first.Title = "最近七天战绩"
	first.Digest = "20191216-20191219"
	first.ThumbMediaId = "chP-LBQxy9SVbAFjwZ4QEo81I0bHaY3YDYRwVGmf7o8"

	var first_content string
	for _, e := range listData {
		bytes, _ := json.Marshal(e)
		first_content += string(bytes) + "<br/>"
	}
	first.Content = first_content
	articles[0] = first

	mediaId, err := material.AddNews(wcClient, &material.News{Articles: articles})
	if err != nil {
		base.Log.Error(err)
		return ""
	}
	return mediaId
}

func (this *SuggestWeekService) ModifyWeek(wcClient *core.Client) {
	param := new(vo.SuggestVO)
	now := time.Now()
	h2, _ := time.ParseDuration("-2h")
	endDate := now.Add(h2)
	param.EndDateStr = endDate.Format("2006-01-02 15:04:05")
	h168, _ := time.ParseDuration("-168h")
	beginDate := now.Add(h168)
	param.BeginDateStr = beginDate.Format("2006-01-02 15:04:05")
	param.IsDesc = true
	//今日推荐
	tempList := this.SuggestService.Query(param)
	//更新推送
	first := material.Article{}
	first.Title = "最近七天战绩"
	first.Digest = fmt.Sprintf("%v-%v", beginDate.Format("2006年01月02日"), now.Format("2006年01月02日"))
	first.ThumbMediaId = "chP-LBQxy9SVbAFjwZ4QEo81I0bHaY3YDYRwVGmf7o8"
	first.ShowCoverPic = 0
	first.Author = utils.GetVal("wechat","author")
	temp := vo.WeekVO{}
	temp.BeginDateStr = param.BeginDateStr
	temp.EndDateStr = param.EndDateStr
	temp.MatchCount = int64(len(tempList))

	var redCount, walkCount, blackCount, linkRedCount, tempLinkRedCount, linkBlackCount, tempLinkBlackCount int64
	temp.DataList = make([]vo.SuggestVO, len(tempList))
	for i, e := range tempList {
		if strings.EqualFold("正确", e.Result) {
			redCount++
			tempLinkRedCount++
			tempLinkBlackCount = 0
		} else if strings.EqualFold("错误", e.Result) {
			blackCount++
			tempLinkBlackCount++
			tempLinkRedCount = 0
		} else {
			walkCount++
		}

		if tempLinkRedCount > linkRedCount {
			linkRedCount = tempLinkRedCount
		}
		if tempLinkBlackCount > linkBlackCount {
			linkBlackCount = tempLinkBlackCount
		}

		e.MatchDateStr = e.MatchDate.Format("02号15:04")
		temp.DataList[i] = *e
	}
	temp.RedCount = redCount
	temp.WalkCount = walkCount
	temp.BlackCount = blackCount
	temp.LinkRedCount = linkRedCount
	temp.LinkBlackCount = linkBlackCount

	var buf bytes.Buffer
	tpl, err := template.New("week.html").Funcs(getFuncMap()).ParseFiles("assets/wechat/html/week.html")
	if err != nil {
		base.Log.Error(err)
	}
	if err := tpl.Execute(&buf, &temp); err != nil {
		base.Log.Fatal(err)
	}
	first.Content = buf.String()

	err = material.UpdateNews(wcClient, "chP-LBQxy9SVbAFjwZ4QEgcfOu5CZ67hiBgn5qnZ-Ac", 0, &first)
	if err != nil {
		base.Log.Error(err)
	}
}