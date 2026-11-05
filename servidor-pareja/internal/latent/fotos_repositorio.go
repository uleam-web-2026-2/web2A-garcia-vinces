package latent

import (
	"context"
	"gorm.io/gorm"
)

// RepositorioFotos encapsula las dos escrituras que forman una sola operación:
// reservar una captura y guardar su foto.
type RepositorioFotos interface {
	GuardarFoto(context.Context, *Foto) error
}

type repositorioFotosGORM struct{ db *gorm.DB }

func NuevoRepositorioFotosGORM(db *gorm.DB) RepositorioFotos {
	return &repositorioFotosGORM{db: db}
}

func (r *repositorioFotosGORM) GuardarFoto(ctx context.Context, foto *Foto) error {
	return guardarFoto(ctx, r.db, foto)
}
