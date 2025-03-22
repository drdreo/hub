# Dice Game (Farkle-inspired)

A push-your-luck dice game inspired by Kingdom Come: Deliverance 2's dice game and Farkle.

## Game Rules

### Basic Rules
- Players start with six dice
- Each turn, players can:
  1. Roll all available dice
  2. Set aside any scoring combinations
  3. Continue rolling remaining dice
  4. End their turn to bank points
- If a roll contains no scoring combinations, the player loses all points accumulated in that turn
- First player to reach the target score (10,000 points) wins

### Scoring Combinations

#### Basic Combinations
- One "1" - 100 points
- One "5" - 50 points
- Three of a kind:
  - Three "1s" - 1,000 points
  - Three "2s" - 200 points
  - Three "3s" - 300 points
  - Three "4s" - 400 points
  - Three "5s" - 500 points
  - Three "6s" - 600 points

#### Runs
- Run of "1-5" - 500 points
- Run of "2-6" - 750 points
- Run of "1-6" - 1,500 points

#### Bonus Scoring
- Each additional die beyond three of a kind doubles the score
  - Example: Four "2s" = 400 points, Five "2s" = 800 points

### Special Features
- The Devil's Head (special die) functions as a joker
- Future implementation will include special dice with unique properties

## Technical Implementation

### Game State
The game maintains the following state:
- Player scores and turn order
- Current dice roll
- Set aside dice
- Current turn score
- Round score

### Actions
Players can perform the following actions:
1. `roll` - Roll all available dice
2. `select` - Select a dice for setting aside
3. `set_aside` - Set aside selected dice for scoring
4. `end_turn` - End current turn and bank points

### Message Format
```json
{
    "type": "action_type",
    "playerId": "player_id",
    "diceIndex": [0, 1, 2]  // Optional, used for select and set_aside action
}
```

## Future Enhancements
1. Special dice with unique properties
2. Shop system for purchasing special dice
3. Achievement system
4. Statistics tracking
5. Tournament mode 