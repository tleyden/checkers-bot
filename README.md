
[![Build Status](https://drone.io/github.com/tleyden/checkers-bot/status.png)](https://drone.io/github.com/tleyden/checkers-bot/latest)

A checkers bot framework which makes it easy to build a checkers bot which can connect to a [Checkers Overlord](https://github.com/apage43/checkers-overlord).  This was built as a demonstration app for the [Couchbase Lite](http://www.couchbase.com/communities/couchbase-lite) NoSQL Mobile database.

Here is a screenshot of what the [Checkers-iOS](https://github.com/couchbaselabs/Checkers-iOS) app looks like.  

![screenshot](http://cl.ly/image/1w423h062S1d/Screen%20Shot%202013-09-25%20at%2012.46.27%20AM.png)

# Big Picture

![architecture png](http://cl.ly/image/051o132q3K06/Screen%20Shot%202013-10-08%20at%2010.28.43%20PM.png)

# Bot implementations

* [Random Bot](https://github.com/tleyden/checkers-bot-random)
* [Checkerlution](https://github.com/tleyden/checkerlution)

# Sample code for implementing a bot

```
type RandomThinker struct {
        ourTeamId int
}

func (r RandomThinker) Think(gameState cbot.GameState) (bestMove cbot.ValidMove, ok bool) {

        ok = true
        ourTeam := gameState.Teams[r.ourTeamId]
        allValidMoves := ourTeam.AllValidMoves()
        if len(allValidMoves) > 0 {
                randomValidMoveIndex := cbot.RandomIntInRange(0, len(allValidMoves))
                bestMove = allValidMoves[randomValidMoveIndex]
        } else {
                ok = false
        }

        return
}

func (r RandomThinker) GameFinished(gameState cbot.GameState) (shouldQuit bool) {
        shouldQuit = false
        return
}

func main() {
	thinker := &RandomThinker{}
	thinker.ourTeamId = cbot.RED_TEAM
	game := cbot.NewGame(thinker.ourTeamId, thinker)
	game.SetServerUrl("http://localhost:4984/checkers")
	game.GameLoop()
}


```

# Build your own Checkers Bot

Make a copy of the [Random Bot](https://github.com/tleyden/checkers-bot-random) and use that as a starting point for building your own checkers bot.



