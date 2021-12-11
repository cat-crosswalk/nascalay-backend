package model

import (
	"time"
)

type Game struct {
	Status        GameStatus
	Ready         map[UserId]struct{}
	Odais         []*Odai
	TimeLimit     TimeLimit // seconds
	Timeout       Timeout   // minute
	Timer         *time.Timer
	DrawCount     DrawCount
	ShowCount     ShowCount
	NextShowPhase GameNextShowPhase
	Canvas        Canvas
	BreakTimer    *time.Timer
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

type Img string

func (i Img) AddPrefix() string {
	return "data:image/png;base64," + string(i)
}

type Drawer struct {
	UserId UserId
	AreaId AreaId
}

type AreaId int

func (i AreaId) Int() int {
	return int(i)
}

type TimeLimit int

const DefaultTimeLimit = TimeLimit(30) // Default time limit is 30 seconds

type Timeout time.Time

type DrawCount int

func (d DrawCount) Int() int {
	return int(d)
}

type ShowCount int

func (s ShowCount) Int() int {
	return int(s)
}

type GameNextShowPhase int

const (
	GameShowPhaseOdai GameNextShowPhase = iota
	GameShowPhaseCanvas
	GameShowPhaseAnswer
	GameShowPhaseEnd
)

type Canvas struct {
	BoardName string // TODO: なんならenum 優先度低
	AllArea   int
}

func InitGame() *Game {
	return &Game{
		Status:        GameStatusRoom,
		Ready:         make(map[UserId]struct{}),
		Odais:         make([]*Odai, 0, 100),
		TimeLimit:     DefaultTimeLimit,
		Timeout:       Timeout(time.Now()),
		Timer:         time.NewTimer(0),
		DrawCount:     0,
		ShowCount:     0,
		NextShowPhase: 0,
		Canvas: Canvas{
			BoardName: "5x5",
			AllArea:   25,
		},
		BreakTimer: time.NewTimer(time.Minute * 15),
	}
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

func (g *Game) ResetReady() {
	g.Ready = make(map[UserId]struct{})
}

func (g *Game) ResetImgUpdated() {
	for _, v := range g.Odais {
		v.ImgUpdated = false
	}
}
