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

type CheckersBotRawFlags struct {
	TeamString            string
	SyncGatewayUrl        string
	FeedString            string
	RandomDelayBeforeMove int
}

func GetCheckersBotRawFlags() *CheckersBotRawFlags {

	checkersBotRawFlags := CheckersBotRawFlags{}

	flag.StringVar(
		&checkersBotRawFlags.TeamString,
		"team",
		"NO_DEFAULT",
		"The team, either 'RED' or 'BLUE'",
	)

	flag.StringVar(
		&checkersBotRawFlags.SyncGatewayUrl,
		"syncGatewayUrl",
		"http://localhost:4984/checkers",
		"The server URL, eg: http://foo.com:4984/checkers",
	)

	flag.StringVar(
		&checkersBotRawFlags.FeedString,
		"feed",
		"longpoll",
		"The feed type: longpoll | normal",
	)

	flag.IntVar(
		&checkersBotRawFlags.RandomDelayBeforeMove,
		"randomDelayBeforeMove",
		0,
		"The max random delay before moving in seconds.  0 to disable it",
	)

	return &checkersBotRawFlags

}

func (rawFlags *CheckersBotRawFlags) GetCheckersBotFlags() CheckersBotFlags {

	checkersBotFlags := CheckersBotFlags{}

	if rawFlags.TeamString == "BLUE" {
		checkersBotFlags.Team = BLUE_TEAM
	} else if rawFlags.TeamString == "RED" {
		checkersBotFlags.Team = RED_TEAM
	} else {
		flag.PrintDefaults()
		panic("Invalid command line args given")
	}

	if len(rawFlags.SyncGatewayUrl) == 0 {
		flag.PrintDefaults()
		panic("Invalid command line args given")
	} else {
		checkersBotFlags.SyncGatewayUrl = rawFlags.SyncGatewayUrl
	}

	if rawFlags.FeedString == "longpoll" {
		checkersBotFlags.FeedType = LONGPOLL
	} else if rawFlags.FeedString == "normal" {
		checkersBotFlags.FeedType = NORMAL
	} else {
		flag.PrintDefaults()
		panic("Invalid command line args given")
	}

	checkersBotFlags.RandomDelayBeforeMove = rawFlags.RandomDelayBeforeMove

	return checkersBotFlags

}

func ParseCmdLine() (checkersBotFlags CheckersBotFlags) {

	checkersBotRawFlags := GetCheckersBotRawFlags()

	flag.Parse()

	checkersBotFlags = checkersBotRawFlags.GetCheckersBotFlags()

	return

}
