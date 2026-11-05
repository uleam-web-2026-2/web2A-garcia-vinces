package latent

import "time"

const (
	EstadoCreado     = "creado"
	EstadoActivo     = "activo"
	EstadoFinalizado = "finalizado"
	EstadoCancelado  = "cancelado"
)

// Transiciones permitidas: estado actual -> estados a los que puede pasar.
var transiciones = map[string][]string{
	EstadoCreado:     {EstadoActivo},
	EstadoActivo:     {EstadoFinalizado, EstadoCancelado},
	EstadoFinalizado: {},
	EstadoCancelado:  {},
}

type Usuario struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Nombre string `gorm:"not null" json:"nombre"`
	Rol    string `gorm:"not null;default:admin" json:"rol"`
}

type Evento struct {
	ID                 uint   `gorm:"primaryKey"`
	Nombre             string `gorm:"not null"`
	CodigoAcceso       string `gorm:"uniqueIndex;not null"`
	DisparoPorInvitado int    `gorm:"not null;default:3"`
	MaxInvitado        int    `gorm:"not null"`
	Estado             string `gorm:"not null;default:creado"`
	FechaRevelado      *time.Time
	UsuarioID          uint
	Invitados          []Invitado
}

type Invitado struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Nombre       string `gorm:"not null;uniqueIndex:idx_evento_nombre" json:"nombre"`
	Pin          string `gorm:"not null;size:4" json:"-"`
	FotosTomadas int    `gorm:"not null;default:0" json:"fotos_tomadas"`
	EventoID     uint   `gorm:"not null;uniqueIndex:idx_evento_nombre" json:"evento_id"`
}

type Foto struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	UUID         string `gorm:"uniqueIndex;not null" json:"uuid"`
	URLArchivo   string `gorm:"not null" json:"url_archivo"`
	Sincronizado bool   `gorm:"not null;default:false" json:"sincronizado"`
	EventoID     uint   `gorm:"not null" json:"evento_id"`
	InvitadoID   uint   `gorm:"not null" json:"invitado_id"`
}
