package checkersbot

import (
	"encoding/json"
	"fmt"
	"github.com/couchbaselabs/logg"
	core "github.com/tleyden/checkers-core"
)

// data structure that corresponds to the checkers:game json doc
type GameState struct {
	Id           string        `json:"_id"`
	Rev          string        `json:"_rev"`
	Teams        []Team        `json:"teams"`
	ActiveTeam   TeamType      `json:"activeTeam"`
	WinningTeam  TeamType      `json:"winningTeam"`
	Number       int           `json:"number"`
	Turn         int           `json:"turn"`
	MoveInterval int           `json:"moveInterval"`
	Moves        []MoveHistory `json:"moves"`
}

func NewGameStateFromString(jsonString string) GameState {
	gameState := &GameState{}

	// TODO: fix this hack
	// Hack alert!  what is a cleaner way to deal with
	// the issue where the json sometimes contains a winningTeam
	// int field?  How do I distinguish between an actual 0
	// vs a null/missing value?
	gameState.WinningTeam = -1

	jsonBytes := []byte(jsonString)
	err := json.Unmarshal(jsonBytes, gameState)
	if err != nil {
		logg.LogError(err)
	}
	return *gameState
}

func (gamestate GameState) Export() core.Board {
	board := core.NewEmptyBoard()
	for teamIndex, team := range gamestate.Teams {
		for _, piece := range team.Pieces {

			loc := GetCoreLocation(piece.Location)
			row := loc.Row()
			col := loc.Col()

			if piece.Captured == true {
				continue
			}

			switch {
			case teamIndex == 0:
				switch {
				case piece.King:
					board[row][col] = core.BLACK_KING
				default:
					board[row][col] = core.BLACK
				}
			case teamIndex == 1:
				switch {
				case piece.King:
					board[row][col] = core.RED_KING
				default:
					board[row][col] = core.RED
				}

			}

		}
	}
	return board
}

type Piece struct {

	// the locations are numbered from 1 to 32 where 1
	// represents the top-left dark square for the red team,
	// and 32 represents the bottom-right dark square for blue team.
	Location   int         `json:"location"`
	King       bool        `json:"king"`
	Captured   bool        `json:"captured"`
	ValidMoves []ValidMove `json:"validMoves"`
	PieceId    int
}

type Team struct {
	Score            int     `json:"score"`
	ParticipantCount int     `json:"participantCount"`
	Pieces           []Piece `json:"pieces"`
}

type ValidMove struct {

	// in the case of a normal move or single jump, this will contain a single
	// location which contains the location value of the move destination.
	// (the move starting point is contained in the outer Piece struct)
	// in the case of a double+ jump however, this will contain an array of
	// the locations - the first location will be the first jump landing spot,
	// the second location will be the second jump landing spot, etc..
	// so for example, in this position: http://cl.ly/image/3k470u1P0G3M
	// the piece location will be 24, and the locations will be: "locations":[15,6]
	// which means 24->15,15->6
	Locations     []int     `json:"locations"`
	Captures      []Capture `json:"captures"`
	King          bool      `json:"king"`
	PieceId       int
	StartLocation int
}

type MoveHistory struct {
	Piece     int      `json:"piece"`
	Team      TeamType `json:"team"`
	Locations []int    `json:"locations"`
}

type Capture struct {
	TeamID  int `json:"team"`
	PieceId int `json:"piece"`
}

func (t Team) AllValidMoves() (validMoves []ValidMove) {
	validMoves = make([]ValidMove, 0)

	for pieceIndex, piece := range t.Pieces {
		for _, validMove := range piece.ValidMoves {
			// enhance the validMove from some information
			// from the piece
			validMove.StartLocation = piece.Location
			validMove.PieceId = pieceIndex

			validMoves = append(validMoves, validMove)
		}
	}
	return
}

func (validMove ValidMove) EndLocation() int {

	lastIndex := len(validMove.Locations) - 1
	return validMove.Locations[lastIndex]

}

func (validMove ValidMove) String() string {

	return fmt.Sprintf("%v -> %v", validMove.StartLocation, validMove.Locations)

}
