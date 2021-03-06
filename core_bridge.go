package checkersbot

import (
	_ "github.com/couchbaselabs/logg"
	core "github.com/tleyden/checkers-core"
)

func CorrespondingValidMoveIndex(move core.Move, allValidMoves []ValidMove) (found bool, index int) {
	for i, validMove := range allValidMoves {
		if EqualsCoreMove(validMove, move) {
			found = true
			index = i
			return
		}
	}
	found = false
	index = -1
	return

}

func EqualsCoreMove(validMove ValidMove, move core.Move) bool {

	// translate valid move start location to core location
	validMoveStart := GetCoreLocation(validMove.StartLocation)

	// are start locations equal?
	if !validMoveStart.Equals(move.From()) {
		return false
	}

	// translate valid move end location to core location
	validMoveEnd := GetCoreLocation(validMove.EndLocation())

	// are end locations equal?
	if !validMoveEnd.Equals(move.To()) {
		return false
	}

	return true

}

func GetCorePlayer(teamId TeamType) core.Player {
	switch {
	case teamId == 0:
		return core.BLACK_PLAYER
	default:
		return core.RED_PLAYER
	}
}

func GetCoreLocation(location int) core.Location {

	if location == 1 {
		return core.NewLocation(0, 1)
	}
	if location == 2 {
		return core.NewLocation(0, 3)
	}
	if location == 3 {
		return core.NewLocation(0, 5)
	}
	if location == 4 {
		return core.NewLocation(0, 7)
	}
	if location == 5 {
		return core.NewLocation(1, 0)
	}
	if location == 6 {
		return core.NewLocation(1, 2)
	}
	if location == 7 {
		return core.NewLocation(1, 4)
	}
	if location == 8 {
		return core.NewLocation(1, 6)
	}
	if location == 9 {
		return core.NewLocation(2, 1)
	}
	if location == 10 {
		return core.NewLocation(2, 3)
	}
	if location == 11 {
		return core.NewLocation(2, 5)
	}
	if location == 12 {
		return core.NewLocation(2, 7)
	}
	if location == 13 {
		return core.NewLocation(3, 0)
	}
	if location == 14 {
		return core.NewLocation(3, 2)
	}
	if location == 15 {
		return core.NewLocation(3, 4)
	}
	if location == 16 {
		return core.NewLocation(3, 6)
	}
	if location == 17 {
		return core.NewLocation(4, 1)
	}
	if location == 18 {
		return core.NewLocation(4, 3)
	}
	if location == 19 {
		return core.NewLocation(4, 5)
	}
	if location == 20 {
		return core.NewLocation(4, 7)
	}
	if location == 21 {
		return core.NewLocation(5, 0)
	}
	if location == 22 {
		return core.NewLocation(5, 2)
	}
	if location == 23 {
		return core.NewLocation(5, 4)
	}
	if location == 24 {
		return core.NewLocation(5, 6)
	}
	if location == 25 {
		return core.NewLocation(6, 1)
	}
	if location == 26 {
		return core.NewLocation(6, 3)
	}
	if location == 27 {
		return core.NewLocation(6, 5)
	}
	if location == 28 {
		return core.NewLocation(6, 7)
	}
	if location == 29 {
		return core.NewLocation(7, 0)
	}
	if location == 30 {
		return core.NewLocation(7, 2)
	}
	if location == 31 {
		return core.NewLocation(7, 4)
	}
	if location == 32 {
		return core.NewLocation(7, 6)
	}

	return core.NewLocation(-1, -1)

}

func ExportCoreLocation(location core.Location) int {

	if location == core.NewLocation(0, 1) {
		return 1
	}
	if location == core.NewLocation(0, 3) {
		return 2
	}
	if location == core.NewLocation(0, 5) {
		return 3
	}
	if location == core.NewLocation(0, 7) {
		return 4
	}
	if location == core.NewLocation(1, 0) {
		return 5
	}
	if location == core.NewLocation(1, 2) {
		return 6
	}
	if location == core.NewLocation(1, 4) {
		return 7
	}
	if location == core.NewLocation(1, 6) {
		return 8
	}
	if location == core.NewLocation(2, 1) {
		return 9
	}
	if location == core.NewLocation(2, 3) {
		return 10
	}
	if location == core.NewLocation(2, 5) {
		return 11
	}
	if location == core.NewLocation(2, 7) {
		return 12
	}
	if location == core.NewLocation(3, 0) {
		return 13
	}
	if location == core.NewLocation(3, 2) {
		return 14
	}
	if location == core.NewLocation(3, 4) {
		return 15
	}
	if location == core.NewLocation(3, 6) {
		return 16
	}
	if location == core.NewLocation(4, 1) {
		return 17
	}
	if location == core.NewLocation(4, 3) {
		return 18
	}
	if location == core.NewLocation(4, 5) {
		return 19
	}
	if location == core.NewLocation(4, 7) {
		return 20
	}
	if location == core.NewLocation(5, 0) {
		return 21
	}
	if location == core.NewLocation(5, 2) {
		return 22
	}
	if location == core.NewLocation(5, 4) {
		return 23
	}
	if location == core.NewLocation(5, 6) {
		return 24
	}
	if location == core.NewLocation(6, 1) {
		return 25
	}
	if location == core.NewLocation(6, 3) {
		return 26
	}
	if location == core.NewLocation(6, 5) {
		return 27
	}
	if location == core.NewLocation(6, 7) {
		return 28
	}
	if location == core.NewLocation(7, 0) {
		return 29
	}
	if location == core.NewLocation(7, 2) {
		return 30
	}
	if location == core.NewLocation(7, 4) {
		return 31
	}
	if location == core.NewLocation(7, 6) {
		return 32
	}

	return -1

}
