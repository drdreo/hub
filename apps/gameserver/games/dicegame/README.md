# Dice Game (Farkle-inspired)

A push-your-luck dice game inspired by Kingdom Come: Deliverance 2's dice game and Farkle.

## Game Rules

### Basic Rules

-   Players start with six dice
-   Each turn, players can:
    1. Roll all available dice
    2. Set aside any scoring combinations
    3. Continue rolling remaining dice
    4. End their turn to bank points
-   If a roll contains no scoring combinations, the player loses all points accumulated in that turn
-   First player to reach the target score (10,000 points) wins

### Scoring Combinations

#### Basic Combinations

-   One "1" - 100 points
-   One "5" - 50 points
-   Three of a kind:
    -   Three "1s" - 1,000 points
    -   Three "2s" - 200 points
    -   Three "3s" - 300 points
    -   Three "4s" - 400 points
    -   Three "5s" - 500 points
    -   Three "6s" - 600 points

#### Runs

-   Run of "1-5" - 500 points
-   Run of "2-6" - 750 points
-   Run of "1-6" - 1,500 points

#### Bonus Scoring

-   Each additional die beyond three of a kind doubles the score
    -   Example: Four "2s" = 400 points, Five "2s" = 800 points

### Special Features

-   The Devil's Head (special die) functions as a joker
-   Future implementation will include special dice with unique properties

## Technical Implementation

### Game State

The game maintains the following state:

-   Player scores and turn order
-   Current dice roll
-   Set aside dice
-   Current turn score
-   Round score

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
    "diceIndex": [0, 1, 2]
    // Optional, used for select and set_aside action
}
```

## Future Enhancements

1. Special dice with unique properties
2. Shop system for purchasing special dice
3. Achievement system
4. Statistics tracking
5. Tournament mode
6. Badges

### Rigged Dice

Extracted data
from [reddit](https://www.reddit.com/r/kingdomcome/comments/1iv8b27/kcd2_anybody_else_get_way_too_into_farkle_and/)

| Name           | Side 1 | Side 2 | Side 3 | Side 4 | Side 5 | Side 6 |
| -------------- | ------ | ------ | ------ | ------ | ------ | ------ |
| Misfortune     | 4.5    | 27.7   | 27.7   | 27.7   | 27.7   | 4.5    |
| Even           | 13.4   | 53.4   | 13.4   | 26.7   | 6.7    | 26.7   |
| Odd            | 26.7   | 6.7    | 26.7   | 6.7    | 26.7   | 6.7    |
| Favourable     | 33.3   | 0      | 5.6    | 5.6    | 33.3   | 22.2   |
| Holy Trinity 1 | 18.2   | 22.7   | 45.4   | 4.5    | 4.5    | 4.5    |
| Holy Trinity 2 | 18.2   | 22.7   | 45.4   | 4.5    | 4.5    | 4.5    |
| Holy Trinity 3 | 18.2   | 22.7   | 45.4   | 4.5    | 4.5    | 4.5    |
| Lousy          | 10     | 15     | 10     | 15     | 35     | 15     |
| Lu             | 13     | 13     | 13     | 13     | 13     | 34.8   |
| Lucky          | 27.3   | 4.5    | 9.1    | 13.6   | 18.2   | 27.3   |
| Ordinary       | 16.7   | 16.7   | 16.7   | 16.7   | 16.7   | 16.7   |
| Saint Ant      | 0      | 0      | 100    | 0      | 0      | 0      |
| Unbalanced     | 25     | 33.3   | 8.3    | 8.3    | 16.7   | 8.3    |
| Wagoner        | 5.6    | 27.8   | 33.3   | 11.1   | 11.1   | 11.1   |

### Badges

#### Tin Badges

|           Badge Name           | Description                                                                                  |
| :----------------------------: | :------------------------------------------------------------------------------------------- |
|      Doppelganger's Badge      | Doubles the points of your last throw. Can be used once per game                             |
|       Badge of Headstart       | You gain a small point headstart at the start of the game                                    |
|        Badge of Defence        | Cancels the effects of your opponent's tin badges                                            |
|        Badge of Fortune        | Allows you to roll one die again. Can be used once per game                                  |
|         Badge of Might         | Allows you to add one extra die to your throw. Can be used once per game                     |
|     Badge of Transmutation     | After your throw, change a die of your choosing to a 3. Can be used once per game            |
| Carpenter's Badge of Advantage | The combination of 3+5 now counts as a new formation, called the Cut. Can be used repeatedly |
|        Warlord's badge         | You gain 25% more points for this turn. Can be used once per game                            |
|     Badge of Resurrection      | After an unlucky throw, allows you to throw again. Can be used once per game                 |

#### Silver Badges

|            Badge Name            | Description                                                                                                          |
| :------------------------------: | :------------------------------------------------------------------------------------------------------------------- |
|       Doppelganger's Badge       | Doubles the points of your last throw. Can be used twice per game                                                    |
|        Badge of Headstart        | You gain a moderate point headstart at the start of the game                                                         |
|         Badge of Defence         | Cancels the effect of your opponent's silver badges                                                                  |
|          Swap-Out Badge          | After your throw, you can roll a die of your choosing again. Can be used once per game                               |
|         Badge of Fortune         | You can roll up to two dice again. Can be used once per game                                                         |
|          Badge of Might          | Allows you to add an extra die to your throw. Can be used twice per game                                             |
|      Badge of Tranmutation       | After your throw, change a die of your choosing to a 5. Can be used once per game                                    |
| Executioner's Badge of Advantage | The combination of 4+5+6 now counts as a new formation, called the Gallows. Can be used repeatedly                   |
|         Warlord's Badge          | Gain 50% more points this turn. Can be used once per game                                                            |
|      Badge of Resurrection       | After an unlucky throw, allows you to throw again. Can be used twice per game                                        |
|           King's Badge           | The badge of the rightful king of the birds allows you to add an extra die to your throw. Can be used twice per game |

#### Gold Badge

|         Badge Name          | Description                                                                                                  |
| :-------------------------: | :----------------------------------------------------------------------------------------------------------- |
|     Doppelganger Badge      | Doubles the points scored from your last throw. Can be used thrice per game                                  |
|     Badge of Headstart      | You gain a large point headstart at the start of the game                                                    |
|      Badge of Defence       | Cancels the effect of your opponent's gold badges                                                            |
|       Swap-Out Badge        | After your throw, you can throw two dice of the same value again. Can be used once per game                  |
|      Badge of Fortune       | You can roll up to three dice again. Can be used once per game                                               |
|       Badge of Might        | Allows you to add an extra die to your throw. Can be used once per game                                      |
|   Badge of Transmutation    | After your throw, change a die of your choosing to a 1. Can be used once per game                            |
| Priest's Badge of Advantage | The combination of 1+3+5 now counts as a new formation, called the Eye. Can be used repeatedly               |
|       Warlord's Badge       | Gain double points for this turn. Can be used once per game                                                  |
|    Badge of Resurrection    | After an unlucky throw, allows you to throw again. Can be used thrice per game                               |
|    Gold Emperor's Badge     | Triples the points gained for the formation 1+1+1. Can be used repeatedly                                    |
|     Gold Wedding Badge      | A memento of Agnes and Olda's big day. Allows you to throw up to three dice again. Can be used once per game |
