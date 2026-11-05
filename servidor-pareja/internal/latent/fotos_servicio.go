package latent

import "context"

// ServicioFotos representa el caso de uso de guardar una foto sin depender
// del motor de persistencia.
type ServicioFotos struct{ repo RepositorioFotos }

func NuevoServicioFotos(repo RepositorioFotos) *ServicioFotos {
	return &ServicioFotos{repo: repo}
}

func (s *ServicioFotos) Guardar(ctx context.Context, foto *Foto) error {
	return s.repo.GuardarFoto(ctx, foto)
}
