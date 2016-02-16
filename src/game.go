package main

import (

	"math/rand"
	"fmt"
	"time"

)

type Buscaminas struct {

	Board, StateBoard [][]int16
	R, C int16
	LastX, LastY int16
	Turn int16
	Players [2]string
	MinesLeft int16
	Score [2]int16
	HasBomb [2]bool
	Id int
}

type Response struct {

	X, Y, Val int16
}

func countMines(x, y int16, B *Buscaminas) (int16) {

	dX, dY := []int16{0, 1, 1, 1, 0, -1, -1, -1}, []int16{1, 1, 0, -1, -1, -1, 0, 1}

	var k int16 = 0
	for i := 0; i < 8; i++ {
		nx, ny := x+dX[i], y+dY[i]
		if nx < 0 || nx >= B.R || ny < 0 || ny >= B.C || B.Board[nx][ny] != -1 {
			continue
		} 
		k++
	}

	return k
}

func (B *Buscaminas) Init(R, C, mines int16, username string, id int) {
	
	seed := rand.NewSource(time.Now().UnixNano())
    rnd  := rand.New(seed)

	B.R, B.C = R, C
	B.LastX, B.LastY = -1, -1
	B.MinesLeft = mines
	B.Id = id
	B.Players = [2]string{username, ""}
	B.HasBomb = [2]bool{true, true}


	B.Board = make([][]int16, R)
	for i := range B.Board {
		B.Board[i] = make([]int16, C)
	}

	B.StateBoard = make([][]int16, R)
	for i := range B.StateBoard {
		B.StateBoard[i] = make([]int16, C)
	}

	for i := int16(0); i < mines; i++ {

		for {

			var x, y = rnd.Intn(int(R)), rnd.Intn(int(C))
			if B.Board[x][y] != -1 {
				
				B.Board[x][y] = -1
				break
			}
		}
	}

	for i := int16(0); i < R; i++ {

		for j := int16(0); j < C; j++ {

			if B.Board[i][j] != -1 {

				B.Board[i][j] = countMines(i, j, B)
			}
			B.StateBoard[i][j] = -1
		}
	}
}

func (B *Buscaminas) Move(coord [][2]int16, usedBomb bool, lastX, lastY int16) {

	if usedBomb {
		B.HasBomb[B.Turn] = false
	}
	B.LastX, B.LastY = lastX, lastY


	var keep bool = false
	for i := range coord {
		
		x, y := coord[i][0], coord[i][1]
		if B.Board[x][y] == -1 {
			
			B.Score[B.Turn]++
			B.MinesLeft--
			keep = true
		}

		B.StateBoard[x][y] = B.Turn
	}

	if !keep {

		if B.Turn == 0 {
			B.Turn = 1
		} else {
			B.Turn = 0
		}
	}
}

func (B *Buscaminas) PrintBoard() {
	for i := int16(0); i < B.R; i++ {
		for j := int16(0); j < B.C; j++ {
			if B.Board[i][j] == -1 {
				fmt.Printf("*")
			} else {
				fmt.Printf("%d", B.Board[i][j])
			}

		}
		fmt.Printf("\n");
	}
}

func (B *Buscaminas) PrintStateBoard() {
	for i := int16(0); i < B.R; i++ {
		for j := int16(0); j < B.C; j++ {
			fmt.Printf("%d", B.StateBoard[i][j])
		}
		fmt.Printf("\n");
	}
}