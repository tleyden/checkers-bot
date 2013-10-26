package checkersbot

import (
	"encoding/json"
	"fmt"
	"github.com/couchbaselabs/logg"
	"github.com/nu7hatch/gouuid"
	"github.com/tleyden/go-couch"
	"io"
	"strings"
	"time"
)

const (
	DEFAULT_SERVER_URL = "http://localhost:4984/checkers"
	GAME_DOC_ID        = "game:checkers"
	VOTES_DOC_ID       = "votes:checkers"
)

type TeamType int

const (
	RED_TEAM  = 0
	BLUE_TEAM = 1
)

type FeedType int

const (
	NORMAL   = 0
	LONGPOLL = 1
)

type Game struct {
	thinker         Thinker
	gameState       GameState
	ourTeamId       TeamType
	db              couch.Database
	user            User
	delayBeforeMove int
	feedType        FeedType
	serverUrl       string
	lastGameDocRev  string
}

type Changes map[string]interface{}

func NewGame(ourTeamId TeamType, thinker Thinker) *Game {
	game := &Game{ourTeamId: ourTeamId, thinker: thinker}
	return game
}

// Follow the changes feed and on each change callback
// call game.handleChanges() which will drive the game
func (game *Game) GameLoop() {

	game.InitGame()

	curSinceValue := "0"

	handleChange := func(reader io.Reader) string {
		logg.LogTo("DEBUG", "handleChange called")
		changes := decodeChanges(reader)
		shouldQuit := game.handleChanges(changes)
		if shouldQuit {
			return "-1" // causes Changes() to return
		} else {
			curSinceValue = getNextSinceValue(curSinceValue, changes)
			if game.feedType == NORMAL {
				time.Sleep(time.Second * 1)
			}
			return curSinceValue
		}
	}

	options := Changes{"since": curSinceValue}
	if game.feedType == LONGPOLL {
		options["feed"] = "longpoll"
	}
	game.db.Changes(handleChange, options)

}

// Given a list of changes, we only care if the game doc has changed.
// If it has changed, and it's our turn to make a move, then call
// the embedded Thinker to make a move or abort the game.
func (game *Game) handleChanges(changes Changes) (shouldQuit bool) {
	shouldQuit = false
	gameDocChanged := game.hasGameDocChanged(changes)
	if gameDocChanged {
		gameState, err := game.fetchLatestGameState()
		if err != nil {
			logg.LogError(err)
			shouldQuit = true
			return
		}

		game.updateUserGameNumber(gameState)
		game.gameState = gameState

		if game.thinkerWantsToQuit(gameState) {
			msg := "Thinker wants to quit the game loop now"
			logg.LogTo("DEBUG", msg)
			shouldQuit = true
			return
		}

		if isOurTurn := game.isOurTurn(gameState); !isOurTurn {
			logg.LogTo("DEBUG", "It's not our turn, ignoring changes")
			return
		}

		bestMove, ok := game.thinker.Think(gameState)
		if ok {
			outgoingVote := game.OutgoingVoteFromMove(bestMove)
			game.PostChosenMove(outgoingVote)
		}

	}
	return
}

func (game Game) thinkerWantsToQuit(gameState GameState) (shouldQuit bool) {
	shouldQuit = false
	if game.finished(gameState) {
		if observer, ok := game.thinker.(Observer); ok {
			shouldQuit = observer.GameFinished(gameState)
			logg.LogTo("DEBUG", "observer returned shouldQuit: %v", shouldQuit)
			return
		} else {
			logg.LogTo("DEBUG", "thinker is not an Observer, not calling GameFinished")
		}

	}
	return
}

func (game Game) finished(gameState GameState) bool {
	logg.LogTo("DEBUG", "game.finished() called")
	gameHasWinner := (gameState.WinningTeam != -1)
	finished := gameHasWinner
	logg.LogTo("DEBUG", "game.finished() returning: %v", finished)
	return finished
}

func (game *Game) InitGame() {
	game.InitDbConnection()
	game.CreateRemoteUser()
}

func (game *Game) CreateRemoteUser() {

	u4, err := uuid.NewV4()
	if err != nil {
		logg.LogPanic("Error generating uuid", err)
	}

	user := &User{
		Id:     fmt.Sprintf("user:%s", u4),
		TeamId: game.ourTeamId,
	}
	newId, newRevision, err := game.db.Insert(user)
	logg.LogTo("DEBUG", "Created new user %v rev %v", newId, newRevision)

	user.Rev = newRevision
	game.user = *user

}

func (game *Game) InitDbConnection() {
	serverUrl := game.ServerUrl()
	db, error := couch.Connect(serverUrl)
	if error != nil {
		logg.LogPanic("Error connecting to %v: %v", serverUrl, error)
	}
	game.db = db
}

func (game *Game) ServerUrl() string {
	serverUrl := DEFAULT_SERVER_URL
	if game.serverUrl != "" {
		serverUrl = game.serverUrl
	}
	return serverUrl
}

func (game *Game) SetServerUrl(serverUrl string) {
	game.serverUrl = serverUrl
}

func (game *Game) SetFeedType(feedType FeedType) {
	game.feedType = feedType
}

// Given a validmove (as chosen by the Thinker), create an "Outgoing Vote" that
// can be passed to the server.  NOTE: the struct OutgoingVotes needs to be
// renamed from plural to singular
func (game *Game) OutgoingVoteFromMove(validMove ValidMove) (votes *OutgoingVotes) {

	votes = &OutgoingVotes{}
	votesId := fmt.Sprintf("vote:%s", game.user.Id)

	err := game.db.Retrieve(votesId, votes)
	if err != nil {
		logg.LogTo("DEBUG", "Unable to find existing vote doc: %v", votesId)
	}

	logg.LogTo("DEBUG", "GET votes, rev: %v", votes.Rev)

	votes.Id = votesId
	votes.Turn = game.gameState.Turn
	votes.PieceId = validMove.PieceId
	votes.TeamId = game.ourTeamId
	votes.GameId = game.gameState.Number

	locations := make([]int, len(validMove.Locations)+1)
	locations[0] = validMove.StartLocation
	copy(locations[1:], validMove.Locations)

	votes.Locations = locations
	return
}

func (game *Game) PostChosenMove(votes *OutgoingVotes) {

	logg.LogTo("DEBUG", "post chosen move: %v", votes)

	preMoveSleepSeconds := game.calculatePreMoveSleepSeconds()

	logg.LogTo("MAIN", "Sleeping %v seconds", preMoveSleepSeconds)

	time.Sleep(time.Second * time.Duration(preMoveSleepSeconds))

	if len(votes.Locations) == 0 {
		logg.LogTo("DEBUG", "invalid move, ignoring: %v", votes)
	}

	var newId string
	var newRevision string
	var err error

	if votes.Rev == "" {
		newId, newRevision, err = game.db.Insert(votes)
		logg.LogTo("MAIN", "Game: %v -> Sent vote: %v, Revision: %v", game.gameState.Number, newId, newRevision)
	} else {
		newRevision, err = game.db.Edit(votes)
		logg.LogTo("MAIN", "Game: %v -> Sent vote: %v, Revision: %v", game.gameState.Number, votes.Id, newRevision)
	}

	if err != nil {
		logg.LogError(err)
		return
	}

}

func (game *Game) SetDelayBeforeMove(delayBeforeMove int) {
	game.delayBeforeMove = delayBeforeMove
}

// Update the game.user object so it has the current game number.
// It does it every time we get a new gamestate document, since
// it can change any time.
func (game *Game) updateUserGameNumber(gameState GameState) {
	gameNumberChanged := (game.gameState.Number != gameState.Number)
	if gameNumberChanged {
		// TODO: getting 409 conflicts here, need to
		// do a CAS loop
		game.user.GameNumber = gameState.Number
		newRevision, err := game.db.Edit(game.user)
		if err != nil {
			logg.LogError(err)
			return
		}
		logg.LogTo("DEBUG", "user update, rev: %v", newRevision)
	}

}

func (game Game) opponentTeamId() int {
	switch game.ourTeamId {
	case RED_TEAM:
		return BLUE_TEAM
	default:
		return RED_TEAM
	}
}

func (game Game) isOurTurn(gameState GameState) bool {
	return gameState.ActiveTeam == game.ourTeamId
}

func (game *Game) hasGameDocChanged(changes Changes) bool {
	gameDocChanged := false
	changeResultsRaw := changes["results"]
	if changeResultsRaw == nil {
		return gameDocChanged
	}
	changeResults := changeResultsRaw.([]interface{})
	for _, changeResultRaw := range changeResults {
		changeResult := changeResultRaw.(map[string]interface{})
		docIdRaw := changeResult["id"]
		docId := docIdRaw.(string)
		if strings.Contains(docId, GAME_DOC_ID) {
			changedRev := getChangedRev(changeResult)
			logg.LogTo("DEBUG", "Game doc changedRev: %v", changedRev)
			if game.lastGameDocRev == "" || changedRev != game.lastGameDocRev {
				gameDocChanged = true
				game.lastGameDocRev = changedRev
				logg.LogTo("DEBUG", "Game changed, set new changeRev to: %v", changedRev)

			}
		}
	}
	return gameDocChanged
}

func (game Game) fetchLatestGameState() (gameState GameState, err error) {
	gameStateFetched := &GameState{}

	// TODO: fix this hack
	// Hack alert!  what is a cleaner way to deal with
	// the issue where the json sometimes contains a winningTeam
	// int field?  How do I distinguish between an actual 0
	// vs a null/missing value?  One way: use a pointer
	gameStateFetched.WinningTeam = -1

	err = game.db.Retrieve(GAME_DOC_ID, gameStateFetched)
	if err == nil {
		gameState = *gameStateFetched
	}
	return
}

func decodeChanges(reader io.Reader) Changes {
	changes := make(Changes)
	decoder := json.NewDecoder(reader)
	decoder.Decode(&changes)
	return changes
}

func getNextSinceValue(curSinceValue string, changes Changes) string {
	lastSeq := changes["last_seq"]
	if lastSeq != nil {
		lastSeqAsString := lastSeq.(string)
		if len(lastSeqAsString) > 0 {
			return lastSeqAsString
		}
	}

	return curSinceValue
}

func (game *Game) calculatePreMoveSleepSeconds() (delay float64) {
	delay = 0
	if game.delayBeforeMove > 0 {
		delay = randomInRange(float64(0), float64(game.delayBeforeMove))
	}
	return
}

// Wait until the game number increments
func (game *Game) WaitForNextGame() {

	curSinceValue := "0"

	handleChange := func(reader io.Reader) string {
		changes := decodeChanges(reader)
		shouldQuit := game.handleChangesWaitForNextGame(changes)
		if shouldQuit {
			return "-1" // causes Changes() to return
		} else {
			curSinceValue = getNextSinceValue(curSinceValue, changes)
			time.Sleep(time.Second * 5)
			return curSinceValue
		}

	}

	options := Changes{"since": curSinceValue}
	game.db.Changes(handleChange, options)

}

// Follow the changes feed and wait until the game number
// increments
func (game *Game) handleChangesWaitForNextGame(changes Changes) (shouldQuit bool) {
	shouldQuit = false
	gameDocChanged := game.hasGameDocChanged(changes)
	if gameDocChanged {
		gameState, err := game.fetchLatestGameState()
		if err != nil {
			logg.LogError(err)
			return
		}
		if gameState.Number != game.gameState.Number {
			// game number changed, we're done
			shouldQuit = true
		}
		game.gameState = gameState
	}
	return
}

// Given a "change result", eg, a single row in the _changes feed result,
// figure out the revision for that row.
// json example:
// {"seq":"*:78942","id":"foo","changes":[{"rev":"2-44abc"}]}
func getChangedRev(changeResult map[string]interface{}) string {
	// clean up this garbage and replace with structs ..
	changesElement := changeResult["changes"].([]interface{})
	firstChangesElement := changesElement[0].(map[string]interface{})
	return firstChangesElement["rev"].(string)
}
