package owe_drahn

import (
	"gameserver/games/owe_drahn/models"
)

type Player struct {
	ID          string              `json:"id"`
	UserID      string              `json:"uid"`
	Name        string              `json:"username"`
	Rank        int                 `json:"rank"` // if the user was logged in, we can show their rank
	Stats       *models.PlayerStats `json:"stats"`
	Life        int                 `json:"life"`
	IsReady     bool                `json:"ready"`
	IsChoosing  bool                `json:"choosing"`
	IsConnected bool                `json:"connected"`
	Balance     float64             `json:"balance"` // total wins/losses
}

func NewPlayer(id string, name string) *Player {
	return &Player{
		ID:      id,
		Name:    name,
		Life:    6,
		IsReady: false,
		Balance: 0,
	}
}

func (p *Player) Reset() {
	p.Life = 6
	p.IsReady = false
	p.IsChoosing = false
}

func (p *Player) SetStats(stats *models.PlayerStats) {
	p.Stats = stats
	p.Rank = calculateRank(stats.TotalGames)
}

func (p *Player) ToFormattedPlayer() *models.FormattedPlayer {
	return &models.FormattedPlayer{
		Life:     p.Life,
		UID:      p.UserID,
		Username: p.Name,
		Rank:     p.Rank,
	}
}

func calculateRank(totalGames int) int {
	return totalGames/10 + totalGames
}
