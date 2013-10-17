package checkersbot

import (
	"flag"
)

func ParseCmdLine() (team int, syncGatewayUrl string, feedType FeedType) {

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

	flag.Parse()

	if *teamString == "BLUE" {
		team = BLUE_TEAM
	} else if *teamString == "RED" {
		team = RED_TEAM
	} else {
		flag.PrintDefaults()
		panic("Invalid command line args given")
	}

	if syncGatewayUrlPtr == nil {
		flag.PrintDefaults()
		panic("Invalid command line args given")
	}

	if *feedTypeStr == "longpoll" {
		feedType = LONGPOLL
	} else if *feedTypeStr == "normal" {
		feedType = NORMAL
	}

	syncGatewayUrl = *syncGatewayUrlPtr
	return
}
