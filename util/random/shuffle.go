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
// 計算量削れなかったので15人までしか実行できません
func SetupMemberRoles(g *model.Game, members []model.UserId) {
	// n = メンバーの数 = お題の数
	n := len(members)
	// n * n の行列を作る
	rect := make([][]int, n)
	for i := 0; i < n; i++ {
		rect[i] = make([]int, n-1)
		for j := 0; j < n-1; j++ {
			rect[i][j] = -1
		}
	}
	answers := RandIntArrayAllMove(n)
	// i 埋める数 j: 縦 k: 横
	for i := 0; i < n; i++ {
		// 埋まった行を記録
		filled := make([]bool, n)
		for a := 0; a < n; a++ {
			if answers[a] == i {
				filled[a] = true
				break
			}
		}
		k := 0
		cie := -1
		fail := 0
		for {
			if k == n-1 {
				break
			}
			// 空いているマスのindexを記録
			empty := make([]int, 0)
			for j := 0; j < n; j++ {
				if !filled[j] && rect[j][k] == -1 {
					empty = append(empty, j)
				}
			}
			if len(empty) == 0 {
				k--
				rect[cie][k] = -1
				filled[cie] = false
				fail++
				if fail > 10 {
					SetupMemberRoles(g, members)
					return
				}
				continue
			}
			cie = Choice(empty)
			rect[cie][k] = i
			filled[cie] = true
			k++
		}
	}
	for i := 0; i < n; i++ {
		ra := RandIntArray(g.Canvas.AllArea)
		k := 0
		f := false
		for {
			for j := 0; j < n-1; j++ {
				if k == g.Canvas.AllArea {
					g.Odais[i].AnswererId = members[answers[i]]
					f = true
					break
				}
				g.Odais[i].DrawerSeq = append(g.Odais[i].DrawerSeq, model.Drawer{
					UserId: members[rect[i][j]],
					Index:  model.Index(ra[k]),
				})
				k += 1
			}
			if f {
				break
			}
		}
	}
}

// 配列の中からランダムに1つ取り出す
func Choice(arr []int) int {
	if len(arr) == 0 {
		return -1
	}
	return arr[rand.Intn(len(arr))]
}

func RandIntArray(n int) []int {
	arr := make([]int, n)
	for i := 0; i < n; i++ {
		arr[i] = i
	}
	for i := 0; i < n; i++ {
		j := rand.Intn(n)
		arr[i], arr[j] = arr[j], arr[i]
	}
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
