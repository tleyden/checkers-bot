package checkersbot

import (
	"encoding/json"
	"fmt"
	"github.com/couchbaselabs/logg"
	"github.com/nu7hatch/gouuid"
	"github.com/tleyden/go-couch"
	"io"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	DEFAULT_SERVER_URL = "http://localhost:4984/checkers"
	GAME_DOC_ID        = "game:checkers"
	VOTES_DOC_ID       = "votes:checkers"
)

type TeamType int

const (
	RED_TEAM = TeamType(iota)
	BLUE_TEAM
)

func (t TeamType) String() string {
	switch t {
	case RED_TEAM:
		return "RED"
	default:
		return "BLUE"
	}

}

type FeedType int

const (
	NORMAL = FeedType(iota)
	LONGPOLL
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
	isThinking      bool
	isThinkingMutex sync.Mutex
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

	// buffered channel is hackish workaround for cases where the
	// it was missing revisions from the changes feed because
	// the select staement was blocked on processing previous changes.
	changesChan := make(chan Changes, 10)

	// buffered channel is hackish workaround for essentially a deadlock
	// where the thing trying to write to the closeChan is blocked because
	// this goroutine is not reading from it, because its blocked on the
	// call to changesChan <- changes
	closeChan := make(chan bool, 10)

	handleChange := func(reader io.Reader) string {
		logg.LogTo("CHECKERSBOT", "handleChange() callback called. team %v: curSinceValue: %v.  game: %p", game.ourTeamName(), curSinceValue, game)
		logg.LogTo("CHECKERSBOT", "# of goroutines %v", runtime.NumGoroutine())
		changes := decodeChanges(reader)

		changesChan <- changes // TODO: put this in select ?
		select {
		case _ = <-closeChan:
			logg.LogTo("CHECKERSBOT", "Got msg on closeChan, returning -1. team %v: %v", game.ourTeamName(), curSinceValue)
			return "-1" // causes Changes() to return
		default:
			curSinceValue = getNextSinceValue(curSinceValue, changes)
			if game.feedType == NORMAL {
				time.Sleep(time.Second * 1)
			}
			logg.LogTo("CHECKERSBOT", "New sinceValue for team %v: %v", game.ourTeamName(), curSinceValue)
			return curSinceValue

		}
	}

	go func() {
		options := Changes{"since": curSinceValue}
		if game.feedType == LONGPOLL {
			options["feed"] = "longpoll"
		}
		game.db.Changes(handleChange, options)
		logg.LogTo("CHECKERSBOT", "game.db.Changes() finished. team %v: %v", game.ourTeamName(), curSinceValue)

	}()

	movesChan := make(chan ValidMove)

	shouldQuit := false

	for {
		select {
		case changes := <-changesChan:
			logg.LogTo("CHECKERSBOT", "Got changes from changesChan, handle it. team %v: curSinceValue: %v", game.ourTeamName(), curSinceValue)
			shouldQuit = game.handleChanges(changes, movesChan)
			logg.LogTo("CHECKERSBOT", "Done handle changes from changesChan. team %v: curSinceValue: %v", game.ourTeamName(), curSinceValue)
			if shouldQuit {
				logg.LogTo("CHECKERSBOT", "shouldQuit == true. team %v: curSinceValue: %v", game.ourTeamName(), curSinceValue)
				closeChan <- true
				logg.LogTo("CHECKERSBOT", "sent true to closeChan. team %v: curSinceValue: %v", game.ourTeamName(), curSinceValue)

				// fix attempt for crash.  my theory is that since there
				// is still a thinker running when we exit the main
				// loop, things break.
				for {
					game.isThinkingMutex.Lock()
					if game.isThinking {
						game.isThinkingMutex.Unlock()
						logg.LogTo("CHECKERSBOT", "thinker still thinking, sleep 1 second. team %v", game.ourTeamName())
						time.Sleep(1 * time.Second)
					} else {
						logg.LogTo("CHECKERSBOT", "thinker done thinking, call break. team %v", game.ourTeamName())
						game.isThinkingMutex.Unlock()
						break
					}

				}

			}
		case bestMove := <-movesChan:
			logg.LogTo("CHECKERSBOT", "%v thinker returned move, sending vote", game.ourTeamName())
			outgoingVote := game.OutgoingVoteFromMove(bestMove)
			game.PostChosenMove(outgoingVote)
			logg.LogTo("CHECKERSBOT", "%v done sending vote", game.ourTeamName())

		}

		if shouldQuit {
			logg.LogTo("CHECKERSBOT", "GAME_LOOP_FINISHED.  break out of for loop. team %v: curSinceValue: %v", game.ourTeamName(), curSinceValue)
			break
		}

	}

	logg.LogTo("CHECKERSBOT", "GAME_LOOP_FINISHED .. last line. team %v: curSinceValue: %v", game.ourTeamName(), curSinceValue)

}

// Given a list of changes, we only care if the game doc has changed.
// If it has changed, and it's our turn to make a move, then call
// the embedded Thinker to make a move or abort the game.
func (game *Game) handleChanges(changes Changes, movesChan chan ValidMove) (shouldQuit bool) {
	msg := fmt.Sprintf("Handle changes called for %v", game.ourTeamName())
	logg.LogTo("CHECKERSBOT", msg)

	shouldQuit = false
	gameDocChanged := game.hasGameDocChanged(changes)
	if gameDocChanged {
		gameState, err := game.fetchLatestGameState()
		msg := fmt.Sprintf("Fetched latest gameState. team %v.  Game state rev: %v", game.ourTeamName(), gameState.Rev)
		logg.LogTo("CHECKERSBOT", msg)

		if err != nil {
			logg.LogError(err)
			msg := fmt.Sprintf("Due to error fetching game state team %v quitting.  Game state: %v", game.ourTeamName(), gameState)
			logg.LogTo("CHECKERSBOT", msg)
			shouldQuit = true
			return
		}

		if game.finished(gameState) {
			msg := fmt.Sprintf("Game is finished. team %v.  Game state: %v", game.ourTeamName(), gameState)
			logg.LogTo("CHECKERSBOT", msg)

		}

		game.updateUserGameNumberCasLoop(gameState)
		game.gameState = gameState

		if game.thinkerWantsToQuit(gameState) {
			msg := fmt.Sprintf("Thinker wants to quit the %v game loop now.  Game state: %v", game.ourTeamName(), gameState)
			logg.LogTo("CHECKERSBOT", msg)
			logg.LogTo("CHECKERSBOT", "game #: %v team: %v", game.user.GameNumber, game.ourTeamName())
			shouldQuit = true
			return
		}

		if isOurTurn := game.isOurTurn(gameState); !isOurTurn {
			logg.LogTo("CHECKERSBOT", "It's not %v turn, ignoring changes", game.ourTeamName())
			return
		}

		game.isThinkingMutex.Lock()
		if !game.isThinking {
			logg.LogTo("CHECKERSBOT", "Call %v thinker", game.ourTeamName())
			game.isThinking = true
			go func() {
				bestMove, ok := game.thinker.Think(gameState)
				logg.LogTo("CHECKERSBOT", "%v thinker found a move", game.ourTeamName())
				game.isThinkingMutex.Lock()
				game.isThinking = false
				game.isThinkingMutex.Unlock()
				if ok {
					movesChan <- bestMove

				} else {
					logg.LogTo("CHECKERSBOT", "%v thinker returned not ok", game.ourTeamName())
				}
			}()

		} else {
			logg.LogTo("CHECKERSBOT", "Not claling %v thinker, already thinking in progress", game.ourTeamName())
		}
		game.isThinkingMutex.Unlock()

	}
	return
}

func (game Game) thinkerWantsToQuit(gameState GameState) (shouldQuit bool) {
	shouldQuit = false
	if game.finished(gameState) {
		if observer, ok := game.thinker.(Observer); ok {
			shouldQuit = observer.GameFinished(gameState)
			logg.LogTo("CHECKERSBOT", "%v team observer returned shouldQuit: %v", game.ourTeamName(), shouldQuit)
			return
		} else {
			logg.LogTo("CHECKERSBOT", "thinker is not an Observer, not calling GameFinished")
		}

	}
	return
}

func (game Game) finished(gameState GameState) bool {
	logg.LogTo("CHECKERSBOT", "game.finished() called for team %v, gameState #: %v game.gameState #: %v", game.ourTeamName(), gameState.Number, game.gameState.Number)
	gameHasWinner := (gameState.WinningTeam != -1)
	finished := gameHasWinner
	logg.LogTo("CHECKERSBOT", "game.finished() returning: %v.  team: %v", finished, game.ourTeamName())
	if finished {
		logg.LogTo("CHECKERSBOT", "game.finished() gamestate: %v.  team: %v", gameState, game.ourTeamName())
		logg.LogTo("CHECKERSBOT", "wining team: %v.  ourTeam: %v", gameState.WinningTeam.String(), game.ourTeamName())
		logg.LogTo("CHECKERSBOT", "game #: %v", gameState.Number)
	}
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
	logg.LogTo("CHECKERSBOT", "Created new user %v rev %v team %v", newId, newRevision, game.ourTeamName())

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
		logg.LogTo("CHECKERSBOT", "Unable to find existing vote doc: %v", votesId)
	}

	logg.LogTo("CHECKERSBOT", "GET votes, rev: %v", votes.Rev)

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

	logg.LogTo("CHECKERSBOT", "Post chosen move as %v: %v", game.ourTeamName(), votes)

	preMoveSleepSeconds := game.calculatePreMoveSleepSeconds()

	logg.LogTo("CHECKERSBOT", "Sleeping %v seconds", preMoveSleepSeconds)

	time.Sleep(time.Second * time.Duration(preMoveSleepSeconds))

	if len(votes.Locations) == 0 {
		logg.LogTo("CHECKERSBOT", "invalid move, ignoring: %v", votes)
	}

	var newId string
	var newRevision string
	var err error
	teamName := game.ourTeamName()

	if votes.Rev == "" {
		newId, newRevision, err = game.db.Insert(votes)
		logg.LogTo("CHECKERSBOT", "Game: %v -> Sent vote: %v as %v, Revision: %v", game.gameState.Number, teamName, newId, newRevision)
	} else {
		newRevision, err = game.db.Edit(votes)
		logg.LogTo("CHECKERSBOT", "Game: %v -> Sent vote: %v as %v, Revision: %v", game.gameState.Number, teamName, votes.Id, newRevision)
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
// it can change any time.  Wrap in a CAS (compare and swap) loop
// since it's possible to get a 409 conflict
func (game *Game) updateUserGameNumberCasLoop(gameState GameState) {

	logg.LogTo("CHECKERSBOT", " updateUserGameNumberCasLoop for team: %v", game.ourTeamName())

	gameNumberChanged := (game.gameState.Number != gameState.Number)
	if !gameNumberChanged {
		logg.LogTo("CHECKERSBOT", "Game number has not changed (%v == %v), doing nothing", game.gameState.Number, gameState.Number)
		return
	} else {
		logg.LogTo("CHECKERSBOT", "Game number has changed (%v != %v)", game.gameState.Number, gameState.Number)
	}

	maxTries := 5
	for i := 0; i < maxTries; i++ {

		// try to do a PUT
		game.user.GameNumber = gameState.Number
		newRevision, err := game.db.Edit(game.user)
		if err != nil {
			logg.LogError(err)
			msg := "Error updating user game number to %v"
			logg.Log(msg, gameState.Number)

			if game.finished(gameState) {
				logg.LogTo("CHECKERSBOT", "game is finished, shouldn't team %v have already quit?", game.ourTeamName())
			}

			// do a GET to get the latest user doc
			fetchedUser, fetchedUserErr := game.fetchLatestUserDoc()

			if fetchedUserErr != nil {
				logg.LogError(fetchedUserErr)
				logg.LogTo("CHECKERSBOT", "error fetching latest user doc")
			} else {
				// update the game number to the value we want
				fetchedUser.GameNumber = gameState.Number
				game.user = fetchedUser

			}

		} else {
			logg.LogTo("CHECKERSBOT", "updated game #: %v team: %v", game.user.GameNumber, game.ourTeamName())
			logg.LogTo("CHECKERSBOT", "user update, rev: %v", newRevision)
			return
		}

	}
	logg.LogPanic("Failed to update user game number in %v tries", maxTries)

}

func (game Game) opponentTeamId() TeamType {
	switch game.ourTeamId {
	case RED_TEAM:
		return BLUE_TEAM
	default:
		return RED_TEAM
	}
}

func (game Game) ourTeamName() string {
	switch game.ourTeamId {
	case RED_TEAM:
		return "RED"
	default:
		return "BLUE"
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
			logg.LogTo("CHECKERSBOT", "Game doc changedRev: %v team %v", changedRev, game.ourTeamName())
			if game.lastGameDocRev == "" || changedRev != game.lastGameDocRev {
				gameDocChanged = true
				game.lastGameDocRev = changedRev
				logg.LogTo("CHECKERSBOT", "Game changed, set new changeRev to: %v team: %v", changedRev, game.ourTeamName())

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

func (game Game) fetchLatestUserDoc() (user User, err error) {
	userFetched := &User{}
	err = game.db.Retrieve(game.user.Id, userFetched)
	if err == nil {
		user = *userFetched
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

func (game *Game) Turn() int {
	return game.gameState.Turn
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
