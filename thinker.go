package checkersbot

type Thinker interface {
	Think(gameState GameState) (validMove ValidMove, ok bool)
}
