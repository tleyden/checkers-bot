package checkersbot

type Thinker interface {
	Think(gameState GameState) ValidMove
}
