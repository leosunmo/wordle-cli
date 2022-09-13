/* -----------------------------------------------------------------------------
 * Copyright (c) Nimble Bun Works. All rights reserved.
 * This software is licensed under the MIT license.
 * See the LICENSE file for further information.
 * -------------------------------------------------------------------------- */

package save

import (
	"pkg.nimblebun.works/wordle-cli/common"
)

type Storage interface {
	Load(id string, user uint64) (*SaveFile, error)
	Save(save *SaveFile, id string, user uint64) error
}

type Statistics struct {
	GamesPlayed       int         `json:"games_played"`
	GamesWon          int         `json:"games_won"`
	GuessDistribution map[int]int `json:"guess_distribution"`
}

type SaveFile struct {
	LastGameID     int                                                                `json:"last_game_id"`
	LastGameStatus common.GameState                                                   `json:"last_game_status"`
	LastGameGrid   [common.WordleMaxGuesses][common.WordleWordLength]*common.GridItem `json:"last_game_grid"`
	Statistics     Statistics                                                         `json:"statistics"`
}

func NewSave() *SaveFile {
	return &SaveFile{
		LastGameID:     -1,
		LastGameStatus: common.GameStateRunning,
		LastGameGrid:   [common.WordleMaxGuesses][common.WordleWordLength]*common.GridItem{},
		Statistics: Statistics{
			GamesPlayed: 0,
			GamesWon:    0,
			GuessDistribution: map[int]int{
				1: 0,
				2: 0,
				3: 0,
				4: 0,
				5: 0,
				6: 0,
			},
		},
	}
}
