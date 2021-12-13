package random

import (
	"math/rand"
	"time"

	"github.com/21hack02win/nascalay-backend/model"
)

func init() {
	rand.Seed(int64(time.Now().Nanosecond()))
}

// AnswerId, DrawerSeqを埋める
func SetupMemberRoles(g *model.Game, members []model.User) {
	// n = メンバーの数 = お題の数
	n := len(g.Odais)
	// n * n の行列を作る
	rect := make([][]int, n)
	for i := 0; i < n; i++ {
		rect[i] = make([]int, g.Canvas.AllArea)
	}
	answers := RandIntArrayAllMove(n)
	for i := 0; i < n; i++ {
		g.Odais[i].AnswererId = g.Odais[answers[i]].SenderId
	}
	k := 0 // 横(描く回数)
	loop := true
	for loop {
		zurasu := RandIntArray(n - 1)
		for i := 0; i < n-1; i++ {
			if k == g.Canvas.AllArea {
				loop = false
				break
			}
			// 縦(お題の数)
			for j := 0; j < n; j++ {
				rect[j][k] = answers[(j+zurasu[i]+1)%n]
			}
			k++
		}
	}
	for i := 0; i < n; i++ {
		ars := RandIntArray(g.Canvas.AllArea)
		g.Odais[i].DrawerSeq = make([]model.Drawer, g.Canvas.AllArea)
		for j := 0; j < g.Canvas.AllArea; j++ {
			g.Odais[i].DrawerSeq[j] = model.Drawer{
				UserId: g.Odais[rect[i][j]].SenderId,
				AreaId: model.AreaId(ars[j]),
			}
		}
	}
}

func RandIntArray(n int) []int {
	arr := make([]int, n)
	for i := 0; i < n; i++ {
		arr[i] = i
	}
	rand.Shuffle(len(arr), func(i, j int) { arr[i], arr[j] = arr[j], arr[i] })
	return arr
}

func RandIntArrayAllMove(n int) []int {
	res := make([]int, n)
	inx := RandIntArray(n)
	for i := 0; i < n; i++ {
		res[inx[i]] = inx[(i+1)%n]
	}
	return res
}
