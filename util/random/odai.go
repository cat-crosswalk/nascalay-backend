package random

import (
	"math/rand"
	"time"
)

var (
	prefixData = [...]string{
		"ねこの", "異世界", "アイドル", "天然", "真っ黒な", "トラと", "空から", "伝説の", "昭和", "平成", "大盛り",
		"宇宙の", "古代", "伝統的な", "現代的な", "歪んだ", "空から", "庭で", "豆腐と",
	}
	suffixData = [...]string{
		"リンゴ", "おばけ", "なす", "トラ", "博士", "スイカ", "パイナップル", "メロン", "バナナ", "ピーマン", "ブドウ",
		"爆弾", "シンデレラ", "ひまわり", "ハロウィン", "宅配便", "みかん", "シャワー", "そば", "お茶", "お酒", "お菓子",
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

//nolint:unused,deadcode
func OdaiExample() string {
	prefix := prefixData[rand.Intn(len(prefixData))]
	suffix := suffixData[rand.Intn(len(suffixData))]
	return prefix + suffix
}
