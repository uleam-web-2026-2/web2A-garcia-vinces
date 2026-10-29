package latent

import (
	"github.com/go-chi/chi/v5"

	"github.com/Jeanvncs/myfirstgo/internal/sesion"
)

// Rutas registra todas las rutas de Latent y decide cuáles piden token y
// qué rol puede usar cada una (matriz ruta × rol).
func (m *Manejador) Rutas(r chi.Router, tokens *sesion.Tokens) {
	m.Tokens = tokens
	if m.Eventos == nil {
		m.Eventos = NuevoServicioEventos(NuevoRepositorioEventos(m.DB))
	}
	// Pública: el invitado entra con el código del evento (el QR), sin cuenta.
	r.Post("/eventos/unirse", m.unirse)

	// Desde aquí, TODAS piden token: sin token (o con uno inválido) = 401.
	r.Group(func(r chi.Router) {
		r.Use(tokens.Autenticar)

		// Los dos roles
		r.Get("/eventos/{id}/fotos", m.listarFotos)
		r.Get("/eventos/{id}/resumen", m.resumenEvento)

		// Solo el invitado: 403 con el token del otro rol
		r.With(sesion.SoloRol(sesion.RolInvitado)).Post("/fotos", m.crearFoto)

		// Solo el anfitrión: 403 con el token del otro rol
		r.With(sesion.SoloRol(sesion.RolAnfitrion)).Post("/eventos", m.crearEvento)
		r.With(sesion.SoloRol(sesion.RolAnfitrion)).Get("/eventos", m.listarEventos)
		r.With(sesion.SoloRol(sesion.RolAnfitrion)).Get("/eventos/{id}", m.verEvento)
		r.With(sesion.SoloRol(sesion.RolAnfitrion)).Put("/eventos/{id}", m.actualizarEvento)
		r.With(sesion.SoloRol(sesion.RolAnfitrion)).Delete("/eventos/{id}", m.eliminarEvento)
		r.With(sesion.SoloRol(sesion.RolAnfitrion)).Patch("/eventos/{id}", m.cambiarEstado)
		// Alias de compatibilidad con el contrato del Hito 1.
		r.With(sesion.SoloRol(sesion.RolAnfitrion)).Patch("/eventos/{id}/estado", m.cambiarEstado)
	})
}
