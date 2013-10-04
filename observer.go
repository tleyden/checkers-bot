package checkersbot

type Observer interface {
	GameFinished(gameState GameState) (shouldQuit bool)
}
