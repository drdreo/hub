package utils

import (
	"gameserver/games/owe_drahn/models"
)

// DefaultStats returns default player statistics
func DefaultStats() models.PlayerStats {
	return models.PlayerStats{
		RolledDice:   []int{0, 0, 0, 0, 0, 0},
		Wins:         0,
		TotalGames:   0,
		PerfectRoll:  0,
		LuckiestRoll: 0,
		WorstRoll:    0,
		Rolled21:     0,
		MaxLifeLoss:  0,
	}
}

// ExtractPlayerStats extracts player stats from a db game
func ExtractPlayerStats(uid string, game models.DBGame) models.PlayerStatAggregation {
	aggregation := models.PlayerStatAggregation{
		RolledDice:   []int{0, 0, 0, 0, 0, 0},
		Won:          false,
		PerfectRoll:  0,
		LuckiestRoll: 0,
		WorstRoll:    0,
		Rolled21:     0,
		MaxLifeLoss:  0,
	}

	// aggregate all player rolls
	var playerRolls []models.Roll
	for _, roll := range game.Rolls {
		if roll.Player.UID == uid {
			playerRolls = append(playerRolls, roll)
		}
	}

	// fail safe, if player didn't roll actually
	if len(playerRolls) == 0 {
		return aggregation
	}

	// calculate if player won
	for _, player := range game.Players {
		if player.UID == uid && player.Life > 0 {
			aggregation.Won = true
			break
		}
	}

	// extract statistics of rolled dice
	for _, roll := range playerRolls {
		aggregation.RolledDice[roll.Dice-1]++

		// Perfect roll
		if roll.Dice == 6 && roll.Total == 15 {
			aggregation.PerfectRoll++
		}
		// luckiestRoll
		if roll.Dice == 3 && roll.Total == 15 {
			aggregation.LuckiestRoll++
		}
	}

	// only have to check last roll of this players' rolls, was the ending one
	lastRoll := playerRolls[len(playerRolls)-1]
	// worst roll
	if lastRoll.Dice == 6 && lastRoll.Total == 16 {
		aggregation.WorstRoll++
	}
	// rolled 21
	if lastRoll.Total == 21 {
		aggregation.Rolled21++
	}

	// lost at max life
	if lastRoll.Player.Life == 6 && lastRoll.Total > 15 {
		aggregation.MaxLifeLoss++
	}

	return aggregation
}
