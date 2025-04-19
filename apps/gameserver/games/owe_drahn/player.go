package owe_drahn

import "gameserver/internal/interfaces"

type Player struct {
	ID            string      `json:"id"`
	UserID        string      `json:"uid"`
	Name          string      `json:"username"`
	Rank          int         `json:"rank"`  // if the user was logged in, we can show their rank
	Stats         interface{} `json:"stats"` // TODO: formatted stats
	Life          int         `json:"life"`
	Points        int         `json:"points"`
	IsReady       bool        `json:"ready"`
	IsChoosing    bool        `json:"choosing"`
	IsConnected   bool        `json:"connected"`
	IsPlayersTurn bool        `json:"isPlayersTurn"`
}

func NewPlayer(id string, name string) *Player {
	return &Player{
		ID:            id,
		Name:          name,
		IsPlayersTurn: false,
		Life:          6,
		IsReady:       false,
	}
}

func (p *Player) Reset() {
	p.IsPlayersTurn = false
	p.Life = 6
	p.IsReady = false
}

func (p *Player) SetStats(stats interface{}) {
	p.Stats = stats
	// TODO calculate rank from stats.totalGames
}

func (p *Player) ToFormattedPlayer() interfaces.M {
	return interfaces.M{
		"life":     p.Life,
		"points":   p.Points,
		"uid":      p.UserID,
		"username": p.Name,
		"rank":     p.Rank,
	}
}
