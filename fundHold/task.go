package fundHold

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"spider/utils"
	"spider/utils/httpclient"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tidwall/gjson"
)

//基金信息
type fundInfo struct {
	Name    string
	Code    string
	Percent float64
}

var totalPercent float64

//股票信息
type gpInfo struct {
	Name        string
	Code        string
	Hangye      string
	FundCode    string
	FundName    string
	FundPercent float64
}

//债券信息
type zqInfo struct {
	Name        string
	Code        string
	FundCode    string
	FundPercent float64
}

//行业信息
type hangyeInfo struct {
	Name    string
	Percent float64
	GpList  []gpInfo
}

//基金组合
var fundList []fundInfo

//持仓股票
var gpList []gpInfo

//持仓债券
var zqList []zqInfo

//基金组合行业分布
var hyMap map[string]hangyeInfo = make(map[string]hangyeInfo)

func init() {
	// fmt.Println(utils.GetCurrentDir())
	content, err := ioutil.ReadFile(utils.GetCurrentDir() + "config.json")
	if err != nil {
		panic("初始化配置信息错误")
	}
	root := gjson.Parse(string(content[:]))
	root.Get("fundlist").ForEach(func(key, value gjson.Result) bool {
		f := fundInfo{Name: value.Get("name").String(), Code: value.Get("code").String(), Percent: value.Get("percent").Float()}
		totalPercent = totalPercent + value.Get("percent").Float()
		fundList = append(fundList, f)
		return true
	})
	fmt.Println(totalPercent)
	fmt.Println(fundList)

	log.Println("init success")
}

func makeGpList() {

	urlMode := `http://fundf10.eastmoney.com/FundArchivesDatas.aspx?type=jjcc&code=%s&topline=10&year=&month=`
	refererMode := `http://fundf10.eastmoney.com/ccmx_%s.html`

	for _, v := range fundList {
		url0 := fmt.Sprintf(urlMode, v.Code)
		fmt.Println(url0)
		referer0 := fmt.Sprintf(refererMode, v.Code)
		headers0 := httpclient.HGetHTML()
		headers0["Host"] = `fundf10.eastmoney.com`
		headers0["Referer"] = referer0
		result, err := httpclient.Get(url0, "", headers0, 20000)
		if err != nil {
			log.Print(err.Error())
			return
		}
		body := utils.BetweenString(result["body"], `content:"`, `",arryear`)

		dom, err := goquery.NewDocumentFromReader(strings.NewReader(body))
		if err != nil {
			log.Println(err.Error())
			return
		}

		dom.Find(".box").Eq(0).Find("tbody tr").Each(func(i int, s *goquery.Selection) {
			code := s.Find("td").Eq(1).Text()
			name := s.Find("td").Eq(2).Text()
			// fmt.Println(name)
			fundPercentStr := s.Find("td").Eq(6).Text()
			gCodeHref, _ := s.Find("td").Eq(1).Find("a").Eq(0).Attr("href")
			gCode := utils.BetweenString(gCodeHref, `quote.eastmoney.com/`, `.html`)

			fundPercentStr = strings.Replace(fundPercentStr, `%`, "", -1)
			fundPercent, err := strconv.ParseFloat(fundPercentStr, 64)
			if err != nil {
				log.Println(err.Error())
				return
			}
			fundPercent = v.Percent / totalPercent * fundPercent / 100 * 100
			fundPercent, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", fundPercent), 64)
			//行业
			hangye := ""
			//美股
			if strings.Index(gCodeHref, `/us/`) > 0 {
				gCode = utils.BetweenString(gCodeHref, `us/`, `.html`)
				url0 := `http://emweb.eastmoney.com/pc_usf10/CompanyInfo/index?color=web&code=` + gCode
				headers0["Host"] = `emweb.eastmoney.com`
				headers0["Referer"] = `http:` + gCodeHref
				result, err = httpclient.Get(url0, "", headers0, 20000)
				if err != nil {
					log.Print(err.Error())
					return
				}
				dom, err = goquery.NewDocumentFromReader(strings.NewReader(result["body"]))
				if err != nil {
					log.Println(err.Error())
					return
				}
				gCode2, _ := dom.Find("#global_fullcode").Eq(0).Attr("value")
				url := `http://emweb.eastmoney.com/pc_usf10/CompanyInfo/PageAjax?fullCode=` + gCode2
				headers := httpclient.HGetJSON()
				headers["Host"] = `emweb.eastmoney.com`
				headers["Referer"] = `http://emweb.eastmoney.com/pc_usf10/CompanyInfo/index?color=web&code=` + gCode
				result, err = httpclient.Get(url, "", headers0, 20000)
				if err != nil {
					log.Print(err.Error())
					return
				}
				root := gjson.Parse(result["body"])
				hangye = root.Get("data.gszl.0.INDUSTRY").String()
			} else if strings.Index(gCodeHref, `/hk/`) > 0 {
				//港股
				gCode = utils.BetweenString(gCodeHref, `hk/`, `.html`)
				url := `http://emweb.securities.eastmoney.com/PC_HKF10/CompanyProfile/PageAjax?code=` + gCode
				headers := httpclient.HGetJSON()
				headers["Host"] = `emweb.securities.eastmoney.com`
				headers["Referer"] = `http://emweb.securities.eastmoney.com/PC_HKF10/CompanyProfile/index?type=web&code=` + gCode
				result, err = httpclient.Get(url, "", headers0, 20000)
				if err != nil {
					log.Print(err.Error())
					return
				}
				root := gjson.Parse(result["body"])
				hangye = root.Get("gszl.sshy").String()
			} else {
				//A股
				url1 := `http://f10.eastmoney.com/f10_v2/CompanySurvey.aspx?code=` + gCode
				headers0["Host"] = `f10.eastmoney.com`
				headers0["Referer"] = `http:` + gCodeHref
				result, err = httpclient.Get(url1, "", headers0, 20000)
				if err != nil {
					log.Print(err.Error())
					return
				}
				dom, err = goquery.NewDocumentFromReader(strings.NewReader(result["body"]))
				if err != nil {
					log.Println(err.Error())
					return
				}
				gCode2, _ := dom.Find("#sCode123").Eq(0).Attr("value")

				url2 := `http://f10.eastmoney.com/CompanySurvey/CompanySurveyAjax?code=` + gCode2
				headers0["Referer"] = `http://f10.eastmoney.com/f10_v2/CompanySurvey.aspx?code=` + gCode
				result, err = httpclient.Get(url2, "", headers0, 20000)
				if err != nil {
					log.Print(err.Error())
					return
				}
				root := gjson.Parse(result["body"])
				hangye = root.Get("jbzl.sshy").String()
			}

			g := gpInfo{Name: name, Code: code, FundCode: v.Code, FundPercent: fundPercent, Hangye: hangye, FundName: v.Name}
			gpList = append(gpList, g)
		})
	}
	fmt.Println("重仓股=====================")
	// fmt.Println(gpList)
	gpListByte, _ := json.Marshal(gpList)
	fmt.Println(string(gpListByte))

	for _, v := range gpList {
		if hy, ok := hyMap[v.Hangye]; !ok {
			gps := make([]gpInfo, 1)
			gps[0] = v
			hyMap[v.Hangye] = hangyeInfo{Name: v.Hangye, Percent: v.FundPercent, GpList: gps}
		} else {
			percent := hy.Percent + v.FundPercent
			percent, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", percent), 64)
			gps := append(hy.GpList, v)
			hyMap[v.Hangye] = hangyeInfo{Name: v.Hangye, Percent: percent, GpList: gps}
		}
	}

	fmt.Println("行业透视====================")
	// fmt.Println(hyMap)
	hyMapByte, _ := json.Marshal(hyMap)
	fmt.Println(string(hyMapByte))

}

func makeZqList() {
	// url0 := `http://fundf10.eastmoney.com/FundArchivesDatas.aspx?type=zqcc&code=%s&year=`
	// referer0 := `http://fundf10.eastmoney.com/ccmx_%s.html`
}


func makePNG(){
	
}
