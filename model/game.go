package model

import (
	"time"
)

type Game struct {
	Status    GameStatus
	Ready     map[UserId]struct{}
	Odais     []*Odai
	TimeLimit TimeLimit
	Timeout   Timeout
	Timer     Timer
	DrawCount DrawCount
	ShowCount ShowCount
	ShowPhase GameShowPhase
}

type GameStatus int

const (
	GameStatusRoom GameStatus = iota
	GameStatusOdai
	GameStatusDraw
	GameStatusAnswer
	GameStatusShow
)

type Odai struct {
	Title      OdaiTitle
	SenderId   UserId
	DrawerSeq  []Drawer
	Img        Img
	ImgUpdated bool
}

type OdaiTitle string

func (o OdaiTitle) String() string {
	return string(o)
}

type Img []byte

type Drawer struct {
	UserId UserId
	Index  Index // TODO: マージしてからAreaIdにする
}

type Index int

func (i Index) Int() int {
	return int(i)
}

func (g *Game) SetupDrawerSeq(members []UserId) {
	// TODO: DrawerSeqを埋める
}

type TimeLimit int

type Timeout int

type Timer time.Timer

type DrawCount int

func (d DrawCount) Int() int {
	return int(d)
}

type ShowCount int

type GameShowPhase int

const (
	GameShowPhaseOdai GameShowPhase = iota
	GameShowPhaseCanvas
	GameShowPhaseAnswer
)

func (g *Game) AddReady(uid UserId) {
	g.Ready[uid] = struct{}{}
}

func (g *Game) CancelReady(uid UserId) {
	delete(g.Ready, uid)
}

func (g *Game) AddOdai(uid UserId, title OdaiTitle) {
	g.Odais = append(g.Odais, &Odai{
		Title:    title,
		SenderId: uid,
		DrawerSeq: []Drawer{
			{
				UserId: uid,
				Index:  0, // TODO
			},
		},
	})
}

func (g *Game) AllDrawPhase() int {
	return len(g.Odais)
}

func (g *Game) ResetReady() {
	g.Ready = make(map[UserId]struct{})
}

func (g *Game) ResetImgUpdated() {
	for _, v := range g.Odais {
		v.ImgUpdated = false
	}
}
