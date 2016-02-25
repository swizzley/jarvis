package jarvis

type ModuleCore struct {
	b *Bot
}

func (m *ModuleCore) RegisterWith(b *Bot) {
	m.b = b

}
