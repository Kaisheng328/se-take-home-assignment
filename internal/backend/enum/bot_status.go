package enum

// BotStatus represents the current status of a cooking bot.
type BotStatus string

const (
	Idle      BotStatus = "IDLE"
	BotActive BotStatus = "PROCESSING"
)
