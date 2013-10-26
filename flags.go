package checkersbot

import (
	"flag"
)

type CheckersBotFlags struct {
	Team                  TeamType
	SyncGatewayUrl        string
	FeedType              FeedType
	RandomDelayBeforeMove int
}

func ParseCmdLine() (checkersBotFlags CheckersBotFlags) {

	checkersBotFlags = CheckersBotFlags{}

	var teamString = flag.String(
		"team",
		"NO_DEFAULT",
		"The team, either 'RED' or 'BLUE'",
	)
	var syncGatewayUrlPtr = flag.String(
		"syncGatewayUrl",
		"http://localhost:4984/checkers",
		"The server URL, eg: http://foo.com:4984/checkers",
	)
	var feedTypeStr = flag.String(
		"feed",
		"longpoll",
		"The feed type: longpoll | normal",
	)
	var randomDelayBeforeMove = flag.Int(
		"randomDelayBeforeMove",
		0,
		"The max random delay before moving in seconds.  0 to disable it",
	)

	flag.Parse()

	if *teamString == "BLUE" {
		checkersBotFlags.Team = BLUE_TEAM
	} else if *teamString == "RED" {
		checkersBotFlags.Team = RED_TEAM
	} else {
		flag.PrintDefaults()
		panic("Invalid command line args given")
	}

	if syncGatewayUrlPtr == nil {
		flag.PrintDefaults()
		panic("Invalid command line args given")
	}

	if *feedTypeStr == "longpoll" {
		checkersBotFlags.FeedType = LONGPOLL
	} else if *feedTypeStr == "normal" {
		checkersBotFlags.FeedType = NORMAL
	}

	checkersBotFlags.RandomDelayBeforeMove = *randomDelayBeforeMove
	checkersBotFlags.SyncGatewayUrl = *syncGatewayUrlPtr
	return
}
