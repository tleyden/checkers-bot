package checkersbot

import (
	"github.com/couchbaselabs/go.assert"
	core "github.com/tleyden/checkers-core"
	"testing"
)

func TestGetCoreLocation(t *testing.T) {
	assert.Equals(t, GetCoreLocation(1), core.NewLocation(0, 1))
	assert.Equals(t, GetCoreLocation(2), core.NewLocation(0, 3))
	assert.Equals(t, GetCoreLocation(32), core.NewLocation(7, 6))
	assert.Equals(t, GetCoreLocation(18), core.NewLocation(4, 3))
	assert.Equals(t, GetCoreLocation(19), core.NewLocation(4, 5))
	assert.Equals(t, GetCoreLocation(20), core.NewLocation(4, 7))

	assert.Equals(t, GetCoreLocation(24), core.NewLocation(5, 6))
	assert.Equals(t, GetCoreLocation(25), core.NewLocation(6, 1))
	assert.Equals(t, GetCoreLocation(26), core.NewLocation(6, 3))
	assert.Equals(t, GetCoreLocation(27), core.NewLocation(6, 5))
	assert.Equals(t, GetCoreLocation(28), core.NewLocation(6, 7))

}

func TestFindCorrespondingValidMove(t *testing.T) {

	validMove := ValidMove{
		StartLocation: 1,
		Locations:     []int{5},
	}
	validMoves := []ValidMove{validMove}
	from := core.NewLocation(0, 1)
	to := core.NewLocation(1, 0)

	move := core.NewMoveFromTo(from, to)

	matchedValidMoveIndex := FindCorrespondingValidMove(move, validMoves)

	assert.Equals(t, matchedValidMoveIndex, 0)

}

func TestEqualsCoreMove(t *testing.T) {

	// start location doesn't match
	validMove := ValidMove{
		StartLocation: 1,
		Locations:     []int{5},
	}
	from := core.NewLocation(0, 6)
	to := core.NewLocation(1, 1)
	move := core.NewMoveFromTo(from, to)
	assert.False(t, EqualsCoreMove(validMove, move))

	// end location doesn't match
	validMove = ValidMove{
		StartLocation: 1,
		Locations:     []int{5},
	}
	from = core.NewLocation(0, 1)
	to = core.NewLocation(5, 5)
	move = core.NewMoveFromTo(from, to)
	assert.False(t, EqualsCoreMove(validMove, move))

	// single jump
	validMove = ValidMove{
		StartLocation: 1,
		Locations:     []int{10},
	}
	from = core.NewLocation(0, 1)
	to = core.NewLocation(2, 3)
	move = core.NewMoveFromTo(from, to)
	assert.True(t, EqualsCoreMove(validMove, move))

	// multi jump
	validMove = ValidMove{
		StartLocation: 1,
		Locations:     []int{10, 19},
	}
	from = core.NewLocation(0, 1)
	to = core.NewLocation(4, 5)
	move = core.NewMoveFromTo(from, to)
	assert.True(t, EqualsCoreMove(validMove, move))

	// basic move match
	validMove = ValidMove{
		StartLocation: 1,
		Locations:     []int{5},
	}
	from = core.NewLocation(0, 1)
	to = core.NewLocation(1, 0)
	move = core.NewMoveFromTo(from, to)
	assert.True(t, EqualsCoreMove(validMove, move))

}
