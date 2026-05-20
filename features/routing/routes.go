package routing

func RegisterRoutes(m *Manager) {
	m.Public.Handle("GET /", m.contentHandler())
}
