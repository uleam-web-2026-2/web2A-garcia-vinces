package latent

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// RepositorioEventos mantiene el acceso a datos de la entidad Evento fuera de HTTP.
type RepositorioEventos interface {
	Crear(context.Context, *Evento) error
	ListarDe(context.Context, uint) ([]Evento, error)
	Buscar(context.Context, uint) (Evento, error)
	Guardar(context.Context, *Evento) error
	ContarInvitados(context.Context, uint) (int64, error)
	ContarFotos(context.Context, uint) (int64, error)
	Eliminar(context.Context, *Evento) error
}

type repositorioEventosGORM struct{ db *gorm.DB }

func NuevoRepositorioEventos(db *gorm.DB) RepositorioEventos {
	return &repositorioEventosGORM{db: db}
}

func (r *repositorioEventosGORM) Crear(ctx context.Context, e *Evento) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *repositorioEventosGORM) ListarDe(ctx context.Context, usuarioID uint) ([]Evento, error) {
	var eventos []Evento
	err := r.db.WithContext(ctx).Where("usuario_id = ?", usuarioID).Order("id DESC").Find(&eventos).Error
	if eventos == nil {
		eventos = []Evento{}
	}
	return eventos, err
}

func (r *repositorioEventosGORM) Buscar(ctx context.Context, id uint) (Evento, error) {
	var evento Evento
	err := r.db.WithContext(ctx).First(&evento, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Evento{}, ErrEventoNoEncontrado
	}
	return evento, err
}

func (r *repositorioEventosGORM) Guardar(ctx context.Context, evento *Evento) error {
	return r.db.WithContext(ctx).Save(evento).Error
}

func (r *repositorioEventosGORM) ContarInvitados(ctx context.Context, id uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&Invitado{}).Where("evento_id = ?", id).Count(&total).Error
	return total, err
}

func (r *repositorioEventosGORM) ContarFotos(ctx context.Context, id uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&Foto{}).Where("evento_id = ?", id).Count(&total).Error
	return total, err
}

func (r *repositorioEventosGORM) Eliminar(ctx context.Context, evento *Evento) error {
	return r.db.WithContext(ctx).Delete(evento).Error
}
