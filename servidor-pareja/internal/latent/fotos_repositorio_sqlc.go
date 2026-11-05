package latent

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Jeanvncs/myfirstgo/internal/latent/filas"
)

type repositorioFotosSQLC struct {
	conexion *sql.DB
	q        *filas.Queries
}

func NuevoRepositorioFotosSQLC(conexion *sql.DB) RepositorioFotos {
	return &repositorioFotosSQLC{conexion: conexion, q: filas.New(conexion)}
}

func (r *repositorioFotosSQLC) GuardarFoto(ctx context.Context, foto *Foto) error {
	tx, err := r.conexion.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	q := r.q.WithTx(tx)

	estado, err := q.EstadoCuotaFoto(ctx, int64(foto.InvitadoID))
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNoEncontrado
	}
	if err != nil {
		return err
	}
	if estado.Estado != EstadoActivo {
		return ErrEventoNoActivo
	}
	if estado.FotosTomadas >= estado.DisparoPorInvitado {
		return ErrLimiteFotos
	}

	filasActualizadas, err := q.IncrementarCuotaFoto(ctx, int64(foto.InvitadoID))
	if err != nil {
		return err
	}
	if filasActualizadas == 0 {
		estado, err = q.EstadoCuotaFoto(ctx, int64(foto.InvitadoID))
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoEncontrado
		}
		if err != nil {
			return err
		}
		if estado.Estado != EstadoActivo {
			return ErrEventoNoActivo
		}
		return ErrLimiteFotos
	}

	foto.EventoID = uint(estado.EventoID)
	guardada, err := q.CrearFoto(ctx, filas.CrearFotoParams{
		Uuid: foto.UUID, UrlArchivo: foto.URLArchivo, Sincronizado: foto.Sincronizado,
		EventoID: int64(foto.EventoID), InvitadoID: int64(foto.InvitadoID),
	})
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	foto.ID = uint(guardada.ID)
	return nil
}
