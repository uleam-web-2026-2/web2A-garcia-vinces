package latent

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Jeanvncs/myfirstgo/internal/latent/filas"
)

// repositorioEventosSQLC sustituye solo Buscar y Guardar; las demás
// operaciones continúan delegándose al repositorio GORM incrustado.
type repositorioEventosSQLC struct {
	RepositorioEventos
	q *filas.Queries
}

var _ RepositorioEventos = (*repositorioEventosSQLC)(nil)

func NuevoRepositorioEventosSQLC(gorm RepositorioEventos, conexion *sql.DB) RepositorioEventos {
	return &repositorioEventosSQLC{RepositorioEventos: gorm, q: filas.New(conexion)}
}

func (r *repositorioEventosSQLC) Buscar(ctx context.Context, id uint) (Evento, error) {
	fila, err := r.q.BuscarEvento(ctx, int64(id))
	if errors.Is(err, sql.ErrNoRows) {
		return Evento{}, ErrEventoNoEncontrado
	}
	if err != nil {
		return Evento{}, err
	}
	evento := Evento{
		ID: uint(fila.ID), Nombre: fila.Nombre, CodigoAcceso: fila.CodigoAcceso,
		DisparoPorInvitado: int(fila.DisparoPorInvitado), MaxInvitado: int(fila.MaxInvitado),
		Estado: fila.Estado, UsuarioID: uint(fila.UsuarioID),
	}
	if fila.FechaRevelado.Valid {
		fecha := fila.FechaRevelado.Time
		evento.FechaRevelado = &fecha
	}
	return evento, nil
}

func (r *repositorioEventosSQLC) Guardar(ctx context.Context, evento *Evento) error {
	fecha := sql.NullTime{}
	if evento.FechaRevelado != nil {
		fecha = sql.NullTime{Time: *evento.FechaRevelado, Valid: true}
	}
	filas, err := r.q.GuardarEvento(ctx, filas.GuardarEventoParams{
		ID: int64(evento.ID), Nombre: evento.Nombre, CodigoAcceso: evento.CodigoAcceso,
		DisparoPorInvitado: int64(evento.DisparoPorInvitado), MaxInvitado: int64(evento.MaxInvitado),
		Estado: evento.Estado, FechaRevelado: fecha, UsuarioID: int64(evento.UsuarioID),
	})
	if err != nil {
		return err
	}
	if filas == 0 {
		return ErrEventoNoEncontrado
	}
	return nil
}
