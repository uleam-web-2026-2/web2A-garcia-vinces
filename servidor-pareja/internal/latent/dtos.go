package latent

import "time"

// EventoEntrada contiene únicamente los campos que un anfitrión puede enviar
// al crear o editar un evento.
type EventoEntrada struct {
	Nombre             string `json:"nombre"`
	DisparoPorInvitado int    `json:"disparo_por_invitado"`
	MaxInvitado        int    `json:"max_invitado"`
}

func (e EventoEntrada) comoNuevo() EventoNuevo {
	return EventoNuevo{Nombre: e.Nombre, DisparoPorInvitado: e.DisparoPorInvitado, MaxInvitado: e.MaxInvitado}
}

func (e EventoEntrada) comoEdicion() EventoEdicion {
	return EventoEdicion{Nombre: e.Nombre, DisparoPorInvitado: e.DisparoPorInvitado, MaxInvitado: e.MaxInvitado}
}

// EventoSalida es la representación HTTP del evento; la entidad de dominio
// conserva únicamente las etiquetas que necesita GORM.
type EventoSalida struct {
	ID                 uint       `json:"id"`
	Nombre             string     `json:"nombre"`
	CodigoAcceso       string     `json:"codigo_acceso"`
	DisparoPorInvitado int        `json:"disparo_por_invitado"`
	MaxInvitado        int        `json:"max_invitado"`
	Estado             string     `json:"estado"`
	FechaRevelado      *time.Time `json:"fecha_revelado"`
	UsuarioID          uint       `json:"usuario_id"`
	Invitados          []Invitado `json:"invitados"`
}

func nuevaEventoSalida(e Evento) EventoSalida {
	return EventoSalida{
		ID: e.ID, Nombre: e.Nombre, CodigoAcceso: e.CodigoAcceso,
		DisparoPorInvitado: e.DisparoPorInvitado, MaxInvitado: e.MaxInvitado,
		Estado: e.Estado, FechaRevelado: e.FechaRevelado, UsuarioID: e.UsuarioID,
		Invitados: e.Invitados,
	}
}

func listaEventosSalida(eventos []Evento) []EventoSalida {
	if eventos == nil {
		return nil
	}
	salidas := make([]EventoSalida, len(eventos))
	for i, evento := range eventos {
		salidas[i] = nuevaEventoSalida(evento)
	}
	return salidas
}
