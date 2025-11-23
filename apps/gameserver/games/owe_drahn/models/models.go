package models

import (
	"time"
)

type DBUser struct {
	UID      string      `firestore:"uid"`
	Stats    PlayerStats `firestore:"stats"`
	Username string      `firestore:"username"`
}

// PlayerStats represents a player's game statistics
// This replaces the TypeScript PlayerStats interface
type PlayerStats struct {
	RolledDice   []int `json:"rolledDice" firestore:"rolledDice"` // dice value rolled 1-6 e.g. [123,33,100,300,100, 99], 123x1, 33x2, 100x3, 300x4, 100x5, 99x6
	Wins         int   `json:"wins" firestore:"wins"`
	TotalGames   int   `json:"totalGames" firestore:"totalGames"`
	PerfectRoll  int   `json:"perfectRoll" firestore:"perfectRoll"`   // rolling from 9 to 15
	LuckiestRoll int   `json:"luckiestRoll" firestore:"luckiestRoll"` // rolling 3 at 15
	WorstRoll    int   `json:"worstRoll" firestore:"worstRoll"`       // rolling a 6 at 10
	Rolled21     int   `json:"rolled21" firestore:"rolled21"`         // rolling 6 at 15
	MaxLifeLoss  int   `json:"maxLifeLoss" firestore:"maxLifeLoss"`   // losing with 6 life left
}

// PlayerStatAggregation omits wins and totalGames from PlayerStats and adds a won field
type PlayerStatAggregation struct {
	RolledDice   []int `json:"rolledDice"`
	Won          bool  `json:"won"`
	PerfectRoll  int   `json:"perfectRoll"`
	LuckiestRoll int   `json:"luckiestRoll"`
	WorstRoll    int   `json:"worstRoll"`
	Rolled21     int   `json:"rolled21"`
	MaxLifeLoss  int   `json:"maxLifeLoss"`
}

// DBGame represents a game ready to be stored in the database
type DBGame struct {
	Players    []*FormattedPlayer `firestore:"players"`
	Rolls      []Roll             `firestore:"rolls"`
	StartedAt  time.Time          `firestore:"startedAt"`
	FinishedAt time.Time          `firestore:"finishedAt"`
}

type FormattedPlayer struct {
	Life     int    `json:"life" firestore:"life"`
	UID      string `json:"uid" firestore:"uid"`
	Username string `json:"username" firestore:"username"`
	Rank     int    `json:"rank" firestore:"rank"`
}

type Roll struct {
	Player *FormattedPlayer `json:"player"`
	Dice   int              `json:"dice"`
	Total  int              `json:"total"`
}

type BetStatus int

const (
	BetStatusPending BetStatus = iota
	BetStatusAccepted
	BetStatusDeclined
	BetStatusResolved
)

// SideBet tracks the specific wager between two players
type SideBet struct {
	ID             string    `json:"id"`
	ChallengerID   string    `json:"challengerId"`
	ChallengerName string    `json:"challengerName"`
	OpponentID     string    `json:"opponentId"`
	OpponentName   string    `json:"opponentName"`
	Amount         float64   `json:"amount"`
	Status         BetStatus `json:"status"`
}
