package latent

import "github.com/go-chi/chi/v5"

func (m *Manejador) Rutas(r chi.Router) {
	r.Post("/eventos", m.crearEvento)
	r.Post("/eventos/unirse", m.unirse)
	r.Post("/fotos", m.crearFoto)
	r.Get("/eventos/{id}/fotos", m.listarFotos)
	r.Patch("/eventos/{id}/estado", m.cambiarEstado)
	r.Get("/eventos/{id}", m.verEvento)
}