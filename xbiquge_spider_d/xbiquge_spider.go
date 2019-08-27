package main

import (
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
	"sync"
	"time"
)

func GetAllBook() (map[string]BiqugeBookInfo, error) {

	var mapBiqugeBookInfo map[string]BiqugeBookInfo = make(map[string]BiqugeBookInfo)
	var anyErr error
	var szUrl string = "http://www.xbiquge.la/xiaoshuodaquan/"

	ptrRequest, anyErr := http.NewRequest("GET", szUrl, nil)
	if anyErr != nil {
		_ptrLog.Warningf("http.NewRequest error: %v, url: %v", anyErr, szUrl)
		return mapBiqugeBookInfo, anyErr
	}

	objClient := http.Client{Timeout: time.Second * 30}
	ptrResp, anyErr := objClient.Do(ptrRequest)
	if anyErr != nil {
		_ptrLog.Warningf("http.Do error: %v, url: %v", anyErr, szUrl)
		return mapBiqugeBookInfo, anyErr
	}
	if ptrResp.StatusCode != 200 {
		_ptrLog.Warningf("http.Do error: %v, url: %v", ptrResp.StatusCode, szUrl)
		return mapBiqugeBookInfo, anyErr
	}

	defer ptrResp.Body.Close()
	ptrDocument, anyErr := goquery.NewDocumentFromReader(ptrResp.Body)
	if anyErr != nil {
		_ptrLog.Warningf("goquery.NewDocumentFromReader error: %v, url: %v", anyErr, szUrl)
		return mapBiqugeBookInfo, anyErr
	}

	//<li><a href="http://www.xbiquge.la/21/21223/">召唤梦魇</a></li>
	ptrSelection := ptrDocument.Find("[href^='http://www.xbiquge.la/']")
	ptrSelection.Each(func(i int, selection *goquery.Selection) {
		var objBiqugeBookInfo BiqugeBookInfo
		objBiqugeBookInfo.BookName = selection.Text()
		objBiqugeBookInfo.BookUrl, _ = selection.Attr("href")

		aryBookId := strings.Split(objBiqugeBookInfo.BookUrl, "/")
		if len(aryBookId) > 2 {
			objBiqugeBookInfo.BookId = aryBookId[len(aryBookId)-2]
		}

		if len(objBiqugeBookInfo.BookName) > 0 {
			_ptrLog.Infof("book: %v", objBiqugeBookInfo)
			mapBiqugeBookInfo[objBiqugeBookInfo.BookName] = objBiqugeBookInfo
		}
	})

	return mapBiqugeBookInfo, nil
}

func GetInfoFromBookUrl(objBiqugeBookInfo BiqugeBookInfo) (BiqugeBookInfo, error) {

	var anyErr error
	var szUrl string = objBiqugeBookInfo.BookUrl

	ptrRequest, anyErr := http.NewRequest("GET", szUrl, nil)
	if anyErr != nil {
		_ptrLog.Warningf("http.NewRequest error: %v, url: %v", anyErr, szUrl)
		return objBiqugeBookInfo, anyErr
	}

	objClient := http.Client{Timeout: time.Second * 30}
	ptrResp, anyErr := objClient.Do(ptrRequest)
	if anyErr != nil {
		_ptrLog.Warningf("http.Do error: %v, url: %v", anyErr, szUrl)
		return objBiqugeBookInfo, anyErr
	}
	if ptrResp.StatusCode != 200 {
		_ptrLog.Warningf("http.Do error: %v, url: %v", ptrResp.StatusCode, szUrl)
		return objBiqugeBookInfo, anyErr
	}

	defer ptrResp.Body.Close()
	ptrDocument, anyErr := goquery.NewDocumentFromReader(ptrResp.Body)
	if anyErr != nil {
		_ptrLog.Warningf("goquery.NewDocumentFromReader error: %v, url: %v", anyErr, szUrl)
		return objBiqugeBookInfo, anyErr
	}

	//<meta property="og:description" content="    大墟的祖训说，天黑，别出门。
	// 大墟残老村的老弱病残们从江边捡到了一个婴儿，取名秦牧，含辛茹苦将他养大。
	// 这一天夜幕降临，黑暗笼罩大墟，秦牧走出了家门……做个春风中荡漾的反派吧！
	// 瞎子对他说。秦牧的反派之路，正在崛起！"/>
	//<meta property="og:novel:author" content="宅猪"/>
	//<meta property="og:novel:category" content="玄幻小说"/>
	ptrSelection := ptrDocument.Find("[property^='og:description']")
	szIntroduction, bExist := ptrSelection.Attr("content")
	if bExist {
		objBiqugeBookInfo.Introduction = szIntroduction
	}
	ptrSelection = ptrDocument.Find("[property^='og:novel:author']")
	szAuthor, bExist := ptrSelection.Attr("content")
	if bExist {
		objBiqugeBookInfo.Author = szAuthor
	}

	ptrSelection = ptrDocument.Find("[property^='og:novel:category']")
	szCategory, bExist := ptrSelection.Attr("content")
	if bExist {
		objBiqugeBookInfo.Category = szCategory
	}

	return objBiqugeBookInfo, nil
}

func DoSpider() {

	var anyErr error
	var mapBiqugeBookInfo map[string]BiqugeBookInfo = make(map[string]BiqugeBookInfo)
	mapBiqugeBookInfo, anyErr = GetAllBook()
	if anyErr != nil {
		return
	}

	mapBiqugeBookInfo = BatchFillBookInfo(mapBiqugeBookInfo)

	SendToElasticSearch(mapBiqugeBookInfo)
	WriteToExcel(mapBiqugeBookInfo)
}

func BatchFillBookInfo(mapBiqugeBookInfo map[string]BiqugeBookInfo) map[string]BiqugeBookInfo {

	var objMapMutex sync.Mutex
	var mapResult map[string]BiqugeBookInfo = make(map[string]BiqugeBookInfo)
	mapResult = mapBiqugeBookInfo
	var anyErr error

	var aryBiqugeBookInfo []BiqugeBookInfo
	for _, objBookInfo := range mapResult {
		aryBiqugeBookInfo = append(aryBiqugeBookInfo, objBookInfo)
	}

	const nSize int = 100
	var nBatchSize int = nSize
	var nIndex int = 0
	var nResLen int = len(aryBiqugeBookInfo)
	for {
		nBatchSize = nSize
		if nResLen <= nBatchSize {
			nBatchSize = nResLen
		}

		objWaitGroup := sync.WaitGroup{}
		objWaitGroup.Add(nBatchSize)
		for i := nIndex; i < (nIndex + nBatchSize); i++ {
			objBiqugeBookInfo := aryBiqugeBookInfo[i]
			go func(i int, objBookInfo BiqugeBookInfo) {

				objBookInfo, anyErr = GetInfoFromBookUrl(objBookInfo)
				if anyErr == nil {
					objMapMutex.Lock()
					mapResult[objBookInfo.BookName] = objBookInfo
					objMapMutex.Unlock()
				}

				objWaitGroup.Done()
			}(i, objBiqugeBookInfo)
		}
		objWaitGroup.Wait()

		nIndex += nBatchSize
		nResLen -= nBatchSize
		if nResLen <= 0 {
			break
		}
	}

	return mapResult
}

func SendToElasticSearch(mapBiqugeBookInfo map[string]BiqugeBookInfo) {

	if len(mapBiqugeBookInfo) <= 0 {
		return
	}

	for _, objBook := range mapBiqugeBookInfo {
		var objESBook ESBookRecord
		objESBook.BookId = objBook.BookId
		objESBook.BookName = objBook.BookName
		objESBook.Author = objBook.Author
		objESBook.Introduction = objBook.Introduction
		UpdateBookInfo(objESBook)
	}
}
