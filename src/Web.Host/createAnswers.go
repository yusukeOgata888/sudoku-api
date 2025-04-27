package main

import (
	"math"
	"math/rand"
	"time"
)

// セルの構造体
type cell struct {
	Index  int
	Number  int
	CandidateNumberList  []int
	IsDirt bool
}

// レスポンスモデル
type answer struct {
	CellIndex int
	CellNumber int
}
// セルを9X9持った配列
var cellsArray []cell

// セルが属する行・列・3X3
type attribute struct {
	Row int
	Col int
	Square int
}

// セルを9X9持った配列を初期化する処理
func initCellsArray(index int, number int,candidateNumberList []int, isDirt bool){
	newCell := cell {
		Index: index,
		Number: number,
		CandidateNumberList: candidateNumberList,
	}
	cellsArray = append(cellsArray, newCell)
}

func createAnswer() []answer {
	var iterator int
	var candidateNumberList []int
	iterator = 1
	for i := 1 ; i<82 ; i++ {
		initCellsArray(i,0,nil,false)
	}

	for !isGoalNode(){
		candidateNumberList = getCandidateNumberList(iterator)
		// シードを更新
		rand.Seed(time.Now().UnixNano())
		// バックトラッキング中でない
		if cellsArray[iterator-1].IsDirt == false {
			if len(candidateNumberList) == 0 && len(cellsArray[iterator-1].CandidateNumberList) == 0{ 
				iterator = getIndexThatExistCandidateNumberList()
				setInitAfterIndexByParam(iterator)
			}
			if len(candidateNumberList) > 0 {
				i := rand.Intn(len(candidateNumberList))
				setNumberByIndex(iterator, candidateNumberList[i])
				setCandidateNumberListByIndex(iterator, remove(candidateNumberList, candidateNumberList[i]))
				iterator += 1
			}
		// バックトラッキング中
		} else {
			if len(cellsArray[iterator-1].CandidateNumberList) == 0 && len(candidateNumberList) > 0 {
				i := rand.Intn(len(candidateNumberList))
				setNumberByIndex(iterator, candidateNumberList[i])
				setCandidateNumberListByIndex(iterator, remove(candidateNumberList, candidateNumberList[i]))
				iterator += 1
			}
			if len(cellsArray[iterator-1].CandidateNumberList)>0 {
				candidateNumberList = cellsArray[iterator-1].CandidateNumberList
				i := rand.Intn(len(candidateNumberList))
				setNumberByIndex(iterator, candidateNumberList[i])
				setCandidateNumberListByIndex(iterator, remove(candidateNumberList, candidateNumberList[i]))
				iterator += 1
			}
			
		}
	}
	var result []answer;
	for cell := range cellsArray {
		responseCell := answer {
			CellIndex: cellsArray[cell].Index,
			CellNumber: cellsArray[cell].Number,
		}
		result = append(result, responseCell)
	}
	return result;
}

// 引数==インデックス以降のセルの各値を初期化する処理
func setInitAfterIndexByParam(index int){
	for i := range cellsArray {
		if cellsArray[i].Index == index {
			cellsArray[i].Number = 0
		}
		if cellsArray[i].Index > index {
			cellsArray[i].Number = 0
			cellsArray[i].IsDirt = false
			cellsArray[i].CandidateNumberList = nil
		}

	}
} 

// 第一引数の配列から第二引数の値を削除する
func remove(ints []int, search int) []int {
    result := []int{}
    for _, v := range ints {
        if v != search {
            result = append(result, v)
        }
    }
    return result
}

// 候補数列をもつセルの降順の先頭のインデックスを取得する処理
func getIndexThatExistCandidateNumberList() int{
	var result int
	for i := 80; i > 0; i-- {
		if len(cellsArray[i].CandidateNumberList) > 0{
			result = cellsArray[i].Index
			return result
		}
	}
	return result
}

// 第一引数をインデックスにもつセルに
// 第二引数をセルの解に設定する処理
func setNumberByIndex(index, number int){
	for i := range cellsArray {
		if cellsArray[i].Index == index {
			cellsArray[i].Number = number
			cellsArray[i].IsDirt = true
		}
	}
}

// 第一引数をインデックスにもつセルに
// 第二引数を候補数に設定する処理
func setCandidateNumberListByIndex(index int, candidateNumberList []int){
	for i := range cellsArray {
		if cellsArray[i].Index == index {
			cellsArray[i].CandidateNumberList = candidateNumberList
			cellsArray[i].IsDirt = true
		}
	}
}

//　第一引数と第二引数の差集合を返す。
func culcDifference(l1, l2 []int) []int {
	s := make(map[int]struct{}, len(l1))
 
	for _, data := range l2 {
		s[data] = struct{}{}
	}
 
	r := make([]int, 0, len(l2))
 
	for _, data := range l1 {
		if _, ok := s[data]; ok {
			continue
		}
 
		r = append(r, data)
	}
	return r
}


// 候補数列を取得する処理
func getCandidateNumberList(index int) []int{
	var Numbers = []int{1,2,3,4,5,6,7,8,9}
	var ownAttributes = getNumberByOwnAttribute(getOwnAttributes(index))
	return culcDifference(Numbers,ownAttributes)
}

// 引数である行・列・3X3属性が持つ値を取得する
func getNumberByOwnAttribute(value attribute) []int {
	var result []int
	for i := 1; i<82; i++ {
		if value.Row == getOwnAttributes(i).Row || value.Col == getOwnAttributes(i).Col || value.Square == getOwnAttributes(i).Square {
			result = append(result,cellsArray[i-1].Number)
		}
	}
	return result
}

// 引数をインデックスに持つセルが属する3X3を返す
func getThreeOnThree(row int, col int) int{
	var reslt int
	switch true {
	case 3 >= row:
		switch true {
		case  3 >= col:
			reslt = 1
		case  6 >= col:
			reslt = 2
		case  9 >= col:
			reslt = 3
		}
	case 4 <= row && 6 >= row:
		switch true{
		case  3 >= col:
			reslt = 4
		case  6 >= col:
			reslt = 5
		case  9 >= col:
			reslt = 6
		}
	case 7 <= row && 9 >= row:
		switch true{
		case  3 >= col:
			reslt = 7
		case  6 >= col:
			reslt = 8
		case  9 >= col:
			reslt = 9
		}
	}
	return reslt
}

// 引数==インデックスを保持しているセルの属する行・列・3X3を返す処理
func getOwnAttributes(value int) attribute {
	var index = float64(value)
	var row,col,square int
	if value % 9 == 0 {
    	row = value / 9
	} else {
		row = int(math.Floor(index/9+1))
	}
	col = int(math.Floor(index/9*10) - math.Floor(index/9)*10)
	if col == 0 {
		col = 9
	}
	square = getThreeOnThree(row,col)
	result := attribute{
		Row: row,
		Col: col,
		Square: square,
	}
	return result
}

// ボードが完成したかどうかを返す処理
func isGoalNode() bool {
	for cell := range cellsArray{
		if cellsArray[cell].Number == 0 {
			return false
		}
	}
	return true
}

