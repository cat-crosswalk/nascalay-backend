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
	Timer     *Timer
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
	Title      OdaiTitle
	Answer     OdaiAnswer
	SenderId   UserId
	AnswererId UserId
	DrawerSeq  []Drawer
	Img        Img
	ImgUpdated bool
}

type OdaiTitle string

func (o OdaiTitle) String() string {
	return string(o)
}

type OdaiAnswer string

func (o OdaiAnswer) String() string {
	return string(o)
}

type Img []byte

type Drawer struct {
	UserId UserId
	AreaId AreaId
}

type AreaId int

func (i AreaId) Int() int {
	return int(i)
}

type TimeLimit int

type Timeout int

type Timer time.Timer

func NewTimer(duration time.Duration) *Timer {
	return (*Timer)(time.NewTimer(duration))
}

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
	g.Odais = append(g.Odais, &Odai{
		Title:     title,
		SenderId:  uid,
		DrawerSeq: []Drawer{},
	})
}

func (g *Game) AllDrawPhase() int {
	return len(g.Odais) - 1
}

func (g *Game) ResetReady() {
	g.Ready = make(map[UserId]struct{})
}

func (g *Game) ResetImgUpdated() {
	for _, v := range g.Odais {
		v.ImgUpdated = false
	}
}
