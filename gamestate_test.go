package checkersbot

import (
	"github.com/couchbaselabs/go.assert"
	"github.com/couchbaselabs/logg"
	core "github.com/tleyden/checkers-core"
	"testing"
)

func init() {
	logg.LogKeys["TEST"] = true
	logg.LogKeys["DEBUG"] = true

}

func TestParseGameState(t *testing.T) {

	jsonString := SampleJson()
	gameState := NewGameStateFromString(jsonString)

	assert.Equals(t, len(gameState.Teams), 2)

	for _, team := range gameState.Teams {
		assert.True(t, len(team.Pieces) > 0)
	}

}

func TestExport(t *testing.T) {

	// parse json into gameState
	jsonString := SampleJson()
	gameState := NewGameStateFromString(jsonString)

	// export to core.Board
	board := gameState.Export()

	logg.LogTo("TEST", "board: %v", board.CompactString(true))

	// make assertions about core.board
	assert.True(t, board.PieceAt(core.NewLocation(0, 1)) == core.BLACK_KING)
	assert.True(t, board.PieceAt(core.NewLocation(1, 0)) == core.BLACK)
	assert.True(t, board.PieceAt(core.NewLocation(3, 0)) == core.EMPTY)
	assert.True(t, board.PieceAt(core.NewLocation(3, 1)) == core.EMPTY)
	assert.True(t, board.PieceAt(core.NewLocation(5, 0)) == core.RED)
	assert.True(t, board.PieceAt(core.NewLocation(6, 6)) == core.EMPTY)

	// piece = board.PieceAt(core.NewLocation(7, 7))
	// assert.Equals(t, piece, core.RED)

}

/*
{
   "applicationUrl":"http://www.couchbase.com/checkers",
   "applicationName":"Couchbase Checkers",
   "revotingAllowed":false,
   "highlightPiecesWithMoves":true,
   "number":1,
   "startTime":"2013-08-26T16:05:30Z",
   "moveDeadline":"2013-08-26T16:05:45Z",
   "turn":1,
   "activeTeam":0,
   "winningTeam":0,
   "moves":[

   ],
   "teams":[
      {
         "participantCount":117983,
         "score":11,
         "pieces":[
            {
               "location":1,
               "king":true
            },
            {
               "location":2
            },
            {
               "location":3
            },
            {
               "location":4
            },
            {
               "location":5
            },
            {
               "location":6
            },
            {
               "location":7,
               "validMoves":[
                  {
                     "locations":[
                        11
                     ],
                     "captures":[
                        {
                           "team":1,
                           "piece":11
                        }
                     ],
                     "king":true
                  }
               ]
            },
            {
               "location":8,
               "validMoves":[
                  {
                     "locations":[
                        11
                     ],
                     "captures":[
                        {
                           "team":1,
                           "piece":8
                        },
                        {
                           "team":1,
                           "piece":9
                        },
                        {
                           "team":1,
                           "piece":10
                        }
                     ]
                  },
                  {
                     "locations":[
                        11,
                        15
                     ]
                  }
               ]
            },
            {
               "location":9,
               "validMoves":[
                  {
                     "locations":[
                        13
                     ]
                  },
                  {
                     "locations":[
                        14
                     ]
                  }
               ]
            },
            {
               "location":10,
               "validMoves":[
                  {
                     "locations":[
                        14
                     ]
                  },
                  {
                     "locations":[
                        15
                     ]
                  }
               ]
            },
            {
               "location":11,
               "captured":true
            },
            {
               "location":12,
               "king":true,
               "validMoves":[
                  {
                     "locations":[
                        16
                     ]
                  }
               ]
            }
         ]
      },
      {
         "participantCount":109217,
         "score":12,
         "pieces":[
            {
               "location":21
            },
            {
               "location":22
            },
            {
               "location":23
            },
            {
               "location":24
            },
            {
               "location":25
            },
            {
               "location":26
            },
            {
               "location":27
            },
            {
               "location":28
            },
            {
               "location":29
            },
            {
               "location":30
            },
            {
               "location":31
            },
            {
               "location":32
            }
         ]
      }
   ]
}
*/
func SampleJson() string {

	jsonString := `{"applicationUrl":"http://www.couchbase.com/checkers","applicationName":"Couchbase Checkers","revotingAllowed":false,"highlightPiecesWithMoves":true,"number":1,"startTime":"2013-08-26T16:05:30Z","moveDeadline":"2013-08-26T16:05:45Z","turn":1,"activeTeam":0,"winningTeam":0,"moves":[],"teams":[{"participantCount":117983,"score":11,"pieces":[{"location":1,"king":true},{"location":2},{"location":3},{"location":4},{"location":5},{"location":6},{"location":7,"validMoves":[{"locations":[11],"captures":[{"team":1,"piece":11}],"king":true}]},{"location":8,"validMoves":[{"locations":[11],"captures":[{"team":1,"piece":8},{"team":1,"piece":9},{"team":1,"piece":10}]},{"locations":[11,15]}]},{"location":9,"validMoves":[{"locations":[13]},{"locations":[14]}]},{"location":10,"validMoves":[{"locations":[14]},{"locations":[15]}]},{"location":11,"captured":true},{"location":12,"king":true,"validMoves":[{"locations":[16]}]}]},{"participantCount":109217,"score":12,"pieces":[{"location":21},{"location":22},{"location":23},{"location":24},{"location":25},{"location":26},{"location":27},{"location":28},{"location":29},{"location":30},{"location":31},{"location":32, "captured": true}]}]}`

	return jsonString
}
