package latent

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrEventoInvalido     = errors.New("datos del evento inválidos")
	ErrEventoNoEncontrado = errors.New("evento no encontrado")
	ErrEventoAjeno        = errors.New("el evento no pertenece al anfitrión")
	ErrTransicion         = errors.New("transición no permitida")
	ErrEventoBloqueado    = errors.New("el evento no se puede editar en este estado")
	ErrEventoConDatos     = errors.New("el evento tiene invitados o fotos y no se puede eliminar")
)

// ServicioEventos contiene las reglas del negocio de la entidad principal.
// No conoce HTTP ni los detalles de GORM.
type ServicioEventos struct{ repo RepositorioEventos }

func NuevoServicioEventos(repo RepositorioEventos) *ServicioEventos {
	return &ServicioEventos{repo: repo}
}

type EventoNuevo struct {
	Nombre             string
	DisparoPorInvitado int
	MaxInvitado        int
}

type EventoEdicion struct {
	Nombre             string
	DisparoPorInvitado int
	MaxInvitado        int
}

func (s *ServicioEventos) Crear(ctx context.Context, usuarioID uint, in EventoNuevo) (Evento, error) {
	in.Nombre = strings.TrimSpace(in.Nombre)
	if in.Nombre == "" || len([]rune(in.Nombre)) > 100 || in.DisparoPorInvitado < 1 || in.DisparoPorInvitado > 20 || in.MaxInvitado < 1 || in.MaxInvitado > 500 {
		return Evento{}, ErrEventoInvalido
	}
	e := Evento{
		Nombre: in.Nombre, CodigoAcceso: generarCodigo(),
		DisparoPorInvitado: in.DisparoPorInvitado, MaxInvitado: in.MaxInvitado,
		Estado: EstadoCreado, UsuarioID: usuarioID,
	}
	if err := s.repo.Crear(ctx, &e); err != nil {
		return Evento{}, err
	}
	return e, nil
}

func (s *ServicioEventos) Listar(ctx context.Context, usuarioID uint) ([]Evento, error) {
	return s.repo.ListarDe(ctx, usuarioID)
}

func (s *ServicioEventos) Obtener(ctx context.Context, usuarioID, id uint) (Evento, error) {
	e, err := s.buscar(ctx, id)
	if err != nil {
		return Evento{}, err
	}
	if e.UsuarioID != usuarioID {
		return Evento{}, ErrEventoAjeno
	}
	return e, nil
}

func (s *ServicioEventos) Resumen(ctx context.Context, id uint) (Evento, error) {
	return s.buscar(ctx, id)
}

func (s *ServicioEventos) Actualizar(ctx context.Context, usuarioID, id uint, in EventoEdicion) (Evento, error) {
	e, err := s.Obtener(ctx, usuarioID, id)
	if err != nil {
		return Evento{}, err
	}
	if e.Estado != EstadoCreado {
		return Evento{}, ErrEventoBloqueado
	}
	in.Nombre = strings.TrimSpace(in.Nombre)
	if in.Nombre == "" || len([]rune(in.Nombre)) > 100 || in.DisparoPorInvitado < 1 || in.DisparoPorInvitado > 20 || in.MaxInvitado < 1 || in.MaxInvitado > 500 {
		return Evento{}, ErrEventoInvalido
	}
	var errCount error
	var invitados int64
	if invitados, errCount = s.repo.ContarInvitados(ctx, id); errCount != nil {
		return Evento{}, errCount
	}
	if int64(in.MaxInvitado) < invitados {
		return Evento{}, ErrEventoInvalido
	}
	e.Nombre, e.DisparoPorInvitado, e.MaxInvitado = in.Nombre, in.DisparoPorInvitado, in.MaxInvitado
	if err := s.repo.Guardar(ctx, &e); err != nil {
		return Evento{}, err
	}
	return e, nil
}

func (s *ServicioEventos) CambiarEstado(ctx context.Context, usuarioID, id uint, siguiente string) (Evento, error) {
	e, err := s.Obtener(ctx, usuarioID, id)
	if err != nil {
		return Evento{}, err
	}
	if _, existe := transiciones[siguiente]; !existe {
		return Evento{}, fmt.Errorf("%w: estado desconocido", ErrTransicion)
	}
	if !puedeTransicionar(e.Estado, siguiente) {
		return Evento{}, fmt.Errorf("%w: de %s no se puede pasar a %s", ErrTransicion, e.Estado, siguiente)
	}
	e.Estado = siguiente
	if siguiente == EstadoFinalizado {
		agora := time.Now().UTC()
		e.FechaRevelado = &agora
	}
	if err := s.repo.Guardar(ctx, &e); err != nil {
		return Evento{}, err
	}
	return e, nil
}

func (s *ServicioEventos) Eliminar(ctx context.Context, usuarioID, id uint) error {
	e, err := s.Obtener(ctx, usuarioID, id)
	if err != nil {
		return err
	}
	if e.Estado != EstadoCreado && e.Estado != EstadoCancelado {
		return ErrEventoBloqueado
	}
	var invitados, fotos int64
	if invitados, err = s.repo.ContarInvitados(ctx, id); err != nil {
		return err
	}
	if fotos, err = s.repo.ContarFotos(ctx, id); err != nil {
		return err
	}
	if invitados > 0 || fotos > 0 {
		return ErrEventoConDatos
	}
	return s.repo.Eliminar(ctx, &e)
}

func (s *ServicioEventos) buscar(ctx context.Context, id uint) (Evento, error) {
	return s.repo.Buscar(ctx, id)
}

func puedeTransicionar(de, a string) bool {
	for _, permitido := range transiciones[de] {
		if permitido == a {
			return true
		}
	}
	return false
}

func generarCodigo() string {
	const alfabeto = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("EVT-%06d", time.Now().UnixNano()%1_000_000)
	}
	for i := range b {
		b[i] = alfabeto[int(b[i])%len(alfabeto)]
	}
	return "EVT-" + string(b)
}
