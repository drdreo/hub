package owe_drahn

import "gameserver/games/owe_drahn/models"

type Player struct {
	ID          string      `json:"id"`
	UserID      string      `json:"uid"`
	Name        string      `json:"username"`
	Rank        int         `json:"rank"`  // if the user was logged in, we can show their rank
	Stats       interface{} `json:"stats"` // TODO: formatted stats
	Life        int         `json:"life"`
	Points      int         `json:"points"`
	IsReady     bool        `json:"ready"`
	IsChoosing  bool        `json:"choosing"`
	IsConnected bool        `json:"connected"`
	Score       int         `json:"score"` // how often the player won/lost
}

func NewPlayer(id string, name string) *Player {
	return &Player{
		ID:      id,
		Name:    name,
		Life:    6,
		IsReady: false,
		Score:   0,
	}
}

func (p *Player) Reset() {
	p.Life = 6
	p.IsReady = false
	p.IsChoosing = false
}

func (p *Player) SetStats(stats interface{}) {
	p.Stats = stats
	// TODO calculate rank from stats.totalGames
}

func (p *Player) ToFormattedPlayer() *models.FormattedPlayer {
	return &models.FormattedPlayer{
		Life:     p.Life,
		Points:   p.Points,
		UID:      p.UserID,
		Username: p.Name,
		Rank:     p.Rank,
	}
}
