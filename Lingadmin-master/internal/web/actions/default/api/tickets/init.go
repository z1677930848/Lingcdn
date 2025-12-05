package tickets

import (
	"github.com/iwind/TeaGo"
)

func init() {
	TeaGo.BeforeStart(func(server *TeaGo.Server) {
		server.Get("/api/tickets", new(ListAction))
		server.Get("/api/tickets/detail", new(DetailAction))
		server.Get("/api/tickets/categories", new(CategoriesAction))
		server.Post("/api/tickets/log", new(CreateLogAction))
	})
}
