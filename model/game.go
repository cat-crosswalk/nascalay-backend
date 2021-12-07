package model

import "time"

type Game struct {
	Status    GameStatus
	Ready     []UserId
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