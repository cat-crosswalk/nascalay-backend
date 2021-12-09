package random

import (
	"math/rand"
	"time"
)

var (
	prefixData = [...]string{
		"ねこの", "異世界", "アイドル", "天然", "真っ黒な", "トラと", "空から", "伝説の", "昭和", "平成", "大盛り",
		"宇宙の", "古代",
	}
	suffixData = [...]string{
		"リンゴ", "おばけ", "なす", "トラ", "博士", "スイカ", "パイナップル", "メロン", "バナナ", "ピーマン", "ブドウ",
		"爆弾",
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

//nolint:unused,deadcode
func odaiExample() string {
	prefix := prefixData[rand.Intn(len(prefixData))]
	suffix := suffixData[rand.Intn(len(suffixData))]
	return prefix + suffix
}
