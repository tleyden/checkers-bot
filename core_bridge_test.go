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
