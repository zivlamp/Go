package main

import (
	"github.com/pwiecz/go-fltk"

	"fmt"
	"math/rand"
	"time"
)
// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ Main
const CELL = 30

type levelArg struct {
	size int
	qty int
}

type Pos struct {
	x int
	y int
}

func main() {
	levels := []levelArg{levelArg{8,8},levelArg{12,18},levelArg{16,32}}
	winPos := Pos{600, 300}
	index, cmdStart := 1, true
	for cmdStart == true {
		cmdStart, index, winPos = start(levels[index], winPos)
	}
}
// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ App Start
type dataTable struct {
	Size int
	Qty int
	step int
	flagsRemain int
	isMine [][]bool
	isOpen [][]bool
	isFlag [][]bool
	countMine [][]int
}

type widgetTable struct {
	cells [][]*fltk.Button
	btns [][]*fltk.Button
	remainShow *fltk.Box
}

func start(lv levelArg, winPos Pos) (bool, int, Pos) {
// ---------------------------------------------------------Data
	data := new(dataTable)
	data.dataInit(lv)
// ---------------------------------------------------------Window ->
	fltk.Lock()  //使fltk响应ticker事件，刷新界面
	win := fltk.NewWindow(CELL * data.Size, CELL * data.Size + 30)
	win.SetLabel("Mine Sweeper")
	win.SetPosition(winPos.x, winPos.y)
// ---------------------------------------------------------Timer
	timer := new(myTimer)
	timer.cmdRun = false
	timer.timerShow = fltk.NewBox(fltk.NO_BOX, CELL * data.Size / 2 - 100 / 2, 0, 100, 30, "")
// ---------------------------------------------------------Menu
	menu := fltk.NewMenuBar(0, 0, 60, 30, "")
	index, cmdStart := 0, false
	var menuItem = [3]string{"&Restart/Easy", "&Restart/Normal", "&Restart/Hard"}
	for i, txt := range menuItem {
		menu.Add(txt, func() func(){
			id := i
			return func() {
				winPos = Pos{win.X(), win.Y()}
				index, cmdStart = id, true
				win.Destroy()
			}
		}())
	}
// ---------------------------------------------------------Widget
	widget := new(widgetTable)
	widget.cells = make([][]*fltk.Button, data.Size)
	widget.btns = make([][]*fltk.Button, data.Size)
	for row := range widget.btns {
		widget.cells[row] = make([]*fltk.Button, data.Size)
		widget.btns[row] = make([]*fltk.Button, data.Size)
		for col := range widget.btns[row] {
			// Creat bottom cells
			widget.cells[row][col] = fltk.NewButton(col * CELL, 30 + row * CELL, CELL, CELL, "")
			widget.cells[row][col].SetBox(fltk.THIN_DOWN_BOX)
			widget.cells[row][col].ClearVisibleFocus()
			widget.cells[row][col].SetCallback(cellCall(row, col, widget, data, timer))
			// Creat top buttons
			widget.btns[row][col] = fltk.NewButton(col * CELL, 30 + row * CELL, CELL, CELL, "")
			widget.btns[row][col].ClearVisibleFocus()
			widget.btns[row][col].SetCallback(btnCall(row, col, widget, data, timer))
		}
	}
	widget.remainShow = fltk.NewBox(fltk.FLAT_BOX,  CELL * data.Size - 50, 0, 50, 30, fmt.Sprintf("%d", data.Qty))
// -------------------------------------------------------- <- Window
	win.End()
	win.Show()
	fltk.Run()
// -------------------------------------------------------- After window destroy
	timer.cmdRun = false
	return cmdStart, index, winPos

}

func (data *dataTable) dataInit(lv levelArg) {
	data.Size, data.Qty = lv.size, lv.qty
	data.step = data.Size * data.Size - data.Qty
	data.flagsRemain = data.Qty
	data.isMine = make([][]bool, data.Size)
	data.isOpen = make([][]bool, data.Size)
	data.isFlag = make([][]bool, data.Size)
	data.countMine = make([][]int, data.Size)
	for i := range data.isMine {
		data.isMine[i] = make([]bool, data.Size)
		data.isOpen[i] = make([]bool, data.Size)
		data.isFlag[i] = make([]bool, data.Size)
		data.countMine[i] = make([]int, data.Size)
	}
}

func (data *dataTable) genMines(x_0 int, y_0 int, widget *widgetTable) {
	rand.Seed(time.Now().Unix())
	genIndex := data.Qty
	for genIndex > 0 {
		x := rand.Intn(data.Size)
		y := rand.Intn(data.Size)
		if (x == x_0 && y == y_0) || data.isMine[x][y] == true {
			continue
		}
		data.isMine[x][y] = true
		row_begin,row_end,col_begin,col_end := getAround(x, y, data.Size)
		for r := row_begin; r < row_end + 1; r++ {
			for c:= col_begin; c < col_end + 1; c++ {
				data.countMine[r][c] += 1
			}
		}
		genIndex--
	}
	for r := range widget.cells {
		for c := range widget.cells[r] {
			if data.isMine[r][c] == true {
				widget.cells[r][c].SetColor(fltk.LIGHT1)
				widget.cells[r][c].SetLabel("\u058e")
			} else {
				widget.cells[r][c].SetColor(fltk.WHITE)
				cellLabel := fmt.Sprintf("%d", data.countMine[r][c])
				if cellLabel != "0" {widget.cells[r][c].SetLabel(cellLabel)}
			}
		}
	}
}

func getAround(row int, col int, size int) (int, int, int, int) {
	var row_begin, row_end, col_begin, col_end int
	if row > 0 {row_begin = row - 1} else {row_begin = row}
	if row < size - 1 {row_end = row + 1} else {row_end = row}
	if col > 0 {col_begin = col - 1} else {col_begin = col}
	if col < size - 1 {col_end = col + 1} else {col_end = col}
	return row_begin, row_end, col_begin, col_end
}

func btnCall(row int, col int, widget *widgetTable, data *dataTable, timer *myTimer) func() {
	return func() {
		if data.isOpen[row][col] == true {
			return
		}
		switch fltk.EventButton() {
		case fltk.RightMouse:
			if data.isFlag[row][col] == false {
				if data.flagsRemain == 0 {
					return
				}
				data.isFlag[row][col] = true
				widget.btns[row][col].SetColor(0xf0f08000)
				data.flagsRemain--
			} else {
				data.isFlag[row][col] = false
				widget.btns[row][col].SetColor(fltk.BACKGROUND_COLOR)
				data.flagsRemain++
			}
			remainShowLabel := fmt.Sprintf("%d", data.flagsRemain)
			widget.remainShow.SetLabel(remainShowLabel)
		case fltk.LeftMouse:
			if data.isFlag[row][col] == true {
				return
			}
			widget.btns[row][col].Hide()
			if data.step == data.Size * data.Size - data.Qty {
				data.genMines(row, col, widget)
				go timerSet(timer)
			}
			switch data.isMine[row][col] {
			case false:
				data.isOpen[row][col] = true
				data.step--
				if data.step == 0 {
					timer.cmdRun = false
					timer.timerShow.SetLabel(timer.durText)
					widget.doWin(data)
				}
				if data.countMine[row][col] == 0 {
					widget.openAround(row, col, data, timer)
				}
			case true:
				timer.cmdRun = false
				timer.timerShow.SetLabel("BOMB!")
				widget.cells[row][col].SetColor(fltk.RED)
				widget.doLost(data)
			}
		}
		
	}
}

func cellCall(row int, col int, widget *widgetTable, data *dataTable, timer *myTimer) func() {
	return func() {
		widget.openAround(row, col, data, timer)
	}
}

func (widget *widgetTable) openAround(row int, col int, data *dataTable, timer *myTimer) {
	row_begin,row_end,col_begin,col_end := getAround(row, col, data.Size)
	flag_around := 0
	for r := row_begin; r < row_end + 1; r++ {
		for c:= col_begin; c < col_end + 1; c++ {
			if data.isFlag[r][c] == true {
				flag_around += 1
			}
		}
	}
	if flag_around != data.countMine[row][col] {
		return
	}

	for r := row_begin; r < row_end + 1; r++ {
		for c:= col_begin; c < col_end + 1; c++ {
			btnCall(r, c, widget, data, timer)()
		}
	}	
}

func (widget *widgetTable) doLost(data *dataTable) {
	for r := range widget.cells {
		for c := range widget.cells[r] {
			widget.btns[r][c].SetCallback(func() {})
			widget.cells[r][c].SetCallback(func() {})
			if data.isOpen[r][c] == true {
				continue
			}
			widget.btns[r][c].Hide()
			if data.isFlag[r][c] == false {
				continue
			}
			if data.isMine[r][c] == true {
				widget.cells[r][c].SetColor(0x99aa7900)
			} else {
				widget.cells[r][c].SetLabel("X")
				widget.cells[r][c].SetLabelColor(fltk.RED)
			}
		}
	}
}

func (widget *widgetTable) doWin(data *dataTable) {
	for r := range widget.btns {
		for c := range widget.btns[r] {
			widget.btns[r][c].SetCallback(func() {})
			widget.cells[r][c].SetCallback(func() {})
			if data.isOpen[r][c] == true {
				continue
			} 
			widget.btns[r][c].SetColor(0x99aa7900)
			widget.btns[r][c].SetLabelSize(20)
			widget.btns[r][c].SetLabel("\u263a")
		}
	}
}
// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ Timer
type myTimer struct {
	cmdRun bool
	timerShow *fltk.Box
	durText string
}

// 为time.Duration类型实现Format输出的方法（借用time.Time类型）
type Timespan time.Duration

func (t Timespan) Format(format string) string {
	z := time.Unix(0, 0).UTC()
	return z.Add(time.Duration(t)).Format(format)
}

func timerSet(timer *myTimer) {
	timer.cmdRun = true
	beginTime := time.Now().Round(time.Second)
	myTicker := time.NewTicker(1 * time.Second)
	for {
		if timer.cmdRun == false {
			myTicker.Stop()
			return
		}
		select {
		case t := <-myTicker.C:
			fltk.Awake(func() {
				dur := t.Sub(beginTime)
				timer.durText = Timespan(dur).Format("15:04:05")
				timer.timerShow.SetLabel(timer.durText)
			})
		}
	}
}