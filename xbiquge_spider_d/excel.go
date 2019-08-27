package main

import (
	"fmt"
	"github.com/Luxurioust/excelize"
	"strconv"
	"time"
)

func WriteToExcel(mapBiqugeBookInfo map[string]BiqugeBookInfo) {

	if len(mapBiqugeBookInfo) <= 0 {
		return
	}

	var anyErr error
	var szSheet string = "xbiquge"
	var ptrExcel *excelize.File
	var szFileName string = fmt.Sprintf("./xbiquge_%s.xlsx",
		time.Now().Format("2006-01-02"))
	ptrExcel, anyErr = excelize.OpenFile(szFileName)
	if anyErr != nil {
		_ptrLog.Warningf("OpenFile error: %v, File: %v", anyErr, szFileName)
		ptrExcel = excelize.NewFile()
	}

	nIndex := ptrExcel.NewSheet(szSheet)
	ptrExcel.SetCellValue(szSheet, "A1", "书名")
	ptrExcel.SetCellValue(szSheet, "B1", "作者")
	ptrExcel.SetCellValue(szSheet, "C1", "分类")
	ptrExcel.SetCellValue(szSheet, "D1", "作品地址")
	ptrExcel.SetCellValue(szSheet, "E1", "简介")
	var i int64 = 2
	for _, objBookInfo := range mapBiqugeBookInfo {
		ptrExcel.SetCellValue(szSheet, "A"+strconv.FormatInt(i, 10), objBookInfo.BookName)
		ptrExcel.SetCellValue(szSheet, "B"+strconv.FormatInt(i, 10), objBookInfo.Author)
		ptrExcel.SetCellValue(szSheet, "C"+strconv.FormatInt(i, 10), objBookInfo.Category)
		ptrExcel.SetCellValue(szSheet, "D"+strconv.FormatInt(i, 10), objBookInfo.BookUrl)
		ptrExcel.SetCellValue(szSheet, "E"+strconv.FormatInt(i, 10), objBookInfo.Introduction)
		i++
	}

	ptrExcel.SetActiveSheet(nIndex)
	anyErr = ptrExcel.SaveAs(szFileName)
	if anyErr != nil {
		_ptrLog.Warningf("SaveAs error: %v", anyErr)
	}
}
