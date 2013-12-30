package checkersbot

import (
	"github.com/couchbaselabs/go.assert"
	core "github.com/tleyden/checkers-core"
	"testing"
)

func TestGetCoreLocation(t *testing.T) {
	assert.Equals(t, getCoreLocation(1), core.NewLocation(0, 1))
	assert.Equals(t, getCoreLocation(2), core.NewLocation(0, 3))
	assert.Equals(t, getCoreLocation(32), core.NewLocation(7, 6))
	assert.Equals(t, getCoreLocation(18), core.NewLocation(4, 3))
	assert.Equals(t, getCoreLocation(19), core.NewLocation(4, 5))
	assert.Equals(t, getCoreLocation(20), core.NewLocation(4, 7))

	assert.Equals(t, getCoreLocation(24), core.NewLocation(5, 6))
	assert.Equals(t, getCoreLocation(25), core.NewLocation(6, 1))
	assert.Equals(t, getCoreLocation(26), core.NewLocation(6, 3))
	assert.Equals(t, getCoreLocation(27), core.NewLocation(6, 5))
	assert.Equals(t, getCoreLocation(28), core.NewLocation(6, 7))

}
