package model

import (
	"time"
)

type Game struct {
	Status    GameStatus
	Ready     map[UserId]struct{}
	Odais     []Odai
	Timeout   Timeout
	Timer     Timer
	DrawCount DrawCount
	ShowCount ShowCount
	ShowPhase GameShowPhase
	Canvas    Canvas
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
	Title     OdaiTitle
	SenderId  UserId
	AnswerId  UserId
	DrawerSeq []Drawer
	Img       []byte
}

type OdaiTitle string

type Drawer struct {
	UserId UserId
	Index  Index
}

type Index int

type Timeout int

type Timer time.Timer

type DrawCount int

type ShowCount int

type GameShowPhase int

const (
	GameShowPhaseOdai GameShowPhase = iota
	GameShowPhaseCanvas
	GameShowPhaseAnswer
)

// TODO: GAME_START で設定する
type Canvas struct {
	BoardName string // TODO: なんならenum 優先度低
	AllArea   int
}

func (g *Game) AddReady(uid UserId) {
	g.Ready[uid] = struct{}{}
}

func (g *Game) CancelReady(uid UserId) {
	delete(g.Ready, uid)
}

func (g *Game) AddOdai(uid UserId, title OdaiTitle) {
	g.Odais = append(g.Odais, Odai{
		Title:    title,
		SenderId: uid,
	})
}
