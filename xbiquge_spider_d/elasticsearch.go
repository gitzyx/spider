package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	ES_HOST       string = "http://172.26.80.40:9200"
	ES_BOOK_INDEX string = "xbiquge_index"
	ES_BOOK_TYPE  string = "xbiquge_type"
)

const (
	ES_UPDATE_RESULT string = "updated"
)

type ESBookRecord struct {
	BookId       string `json:"szBookId"`
	BookName     string `json:"szBookName"`
	Author       string `json:"szAuthor"`
	Introduction string `json:"szIntroduction"`
}

type ESBookUpdateRsp struct {
	Rresult string `json:"result"`
}

func UpdateBookInfo(objBook ESBookRecord) error {

	var anyErr error
	var szUrl string = fmt.Sprintf("%s/%s/%s/%s", ES_HOST,
		ES_BOOK_INDEX, ES_BOOK_TYPE, objBook.BookId)

	var byteArticle []byte
	byteArticle, anyErr = json.Marshal(&objBook)
	if anyErr != nil {
		_ptrLog.Warningf("json.Marshal error: %v,req: %v", anyErr, objBook)
		return anyErr
	}

	ptrHttpRequest, anyErr := http.NewRequest("PUT", szUrl, strings.NewReader(string(byteArticle)))
	if anyErr != nil {
		_ptrLog.Warningf("http.NewRequest error: %v,req: %v", anyErr, objBook)
		return anyErr
	}
	_ptrLog.Infof("http.NewRequest body: %v", objBook)

	ptrHttpRequest.Header.Set("Content-Type", "application/json")
	ptrHttpResp, anyErr := http.DefaultClient.Do(ptrHttpRequest)
	if anyErr != nil {
		_ptrLog.Warningf("http.DefaultClient.Do error: %v,req: %v", anyErr, objBook)
		return anyErr
	}
	byteRspBody, anyErr := ioutil.ReadAll(ptrHttpResp.Body)
	if anyErr != nil {
		_ptrLog.Warningf("ioutil.ReadAll error: %v,req: %v", anyErr, ptrHttpResp.Body)
		return anyErr
	}
	defer ptrHttpResp.Body.Close()

	_ptrLog.Infof("http PUT resp: %v", string(byteRspBody))
	var objESRsp ESBookUpdateRsp
	anyErr = json.Unmarshal(byteRspBody, &objESRsp)
	if anyErr != nil {
		_ptrLog.Warningf("json.Unmarshal error: %v,req: %v", anyErr, string(byteRspBody))
		return anyErr
	}

	return nil
}
