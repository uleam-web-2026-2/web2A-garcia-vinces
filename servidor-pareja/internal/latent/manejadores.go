package latent

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Jeanvncs/myfirstgo/internal/sesion"

	"github.com/Jeanvncs/myfirstgo/internal/respuesta"
)

type Manejador struct {
	DB      *gorm.DB
	Tokens  *sesion.Tokens
	Eventos *ServicioEventos
}

var rePin = regexp.MustCompile(`^\d{4}$`)

type errHTTP struct {
	status int
	codigo string
	msg    string
}

func (e *errHTTP) Error() string { return e.msg }

func leerID(w http.ResponseWriter, r *http.Request) (uint, bool) {
	n, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || n <= 0 {
		respuesta.Error(w, http.StatusBadRequest, "id_invalido", "El id debe ser un numero positivo")
		return 0, false
	}
	return uint(n), true
}

// POST /eventos
func (m *Manejador) crearEvento(w http.ResponseWriter, r *http.Request) {
	ses, _ := sesion.De(r) // el guardia ya comprobó el token y el rol
	var in struct {
		Nombre             string `json:"nombre"`
		DisparoPorInvitado int    `json:"disparo_por_invitado"`
		MaxInvitado        int    `json:"max_invitado"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "json_invalido", "El cuerpo no es un JSON válido")
		return
	}
	ev, err := m.eventos().Crear(r.Context(), ses.UsuarioID, EventoNuevo{
		Nombre: in.Nombre, DisparoPorInvitado: in.DisparoPorInvitado, MaxInvitado: in.MaxInvitado,
	})
	if err != nil {
		m.responderErrorEvento(w, err)
		return
	}
	respuesta.Exito(w, http.StatusCreated, ev)
}

func (m *Manejador) eventos() *ServicioEventos {
	if m.Eventos == nil {
		m.Eventos = NuevoServicioEventos(NuevoRepositorioEventos(m.DB))
	}
	return m.Eventos
}

func (m *Manejador) responderErrorEvento(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrEventoInvalido):
		respuesta.Error(w, http.StatusBadRequest, "evento_invalido", err.Error())
	case errors.Is(err, ErrEventoNoEncontrado):
		respuesta.Error(w, http.StatusNotFound, "no_encontrado", err.Error())
	case errors.Is(err, ErrEventoAjeno):
		respuesta.Error(w, http.StatusForbidden, "no_es_tuyo", err.Error())
	case errors.Is(err, ErrTransicion):
		respuesta.Error(w, http.StatusUnprocessableEntity, "transicion_prohibida", err.Error())
	case errors.Is(err, ErrEventoBloqueado), errors.Is(err, ErrEventoConDatos):
		respuesta.Error(w, http.StatusConflict, "evento_bloqueado", err.Error())
	default:
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "No se pudo completar la operación")
	}
}

// POST /eventos/unirse
func (m *Manejador) unirse(w http.ResponseWriter, r *http.Request) {
	var in struct {
		CodigoAcceso string `json:"codigo_acceso"`
		Nombre       string `json:"nombre"`
		Pin          string `json:"pin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "json_invalido", "El cuerpo no es un JSON válido")
		return
	}
	in.Nombre = strings.TrimSpace(in.Nombre)
	if in.Nombre == "" || !rePin.MatchString(in.Pin) {
		respuesta.Error(w, http.StatusBadRequest, "datos_invalidos", "Nombre obligatorio y pin de 4 dígitos")
		return
	}

	var inv Invitado
	err := m.DB.Transaction(func(tx *gorm.DB) error {
		// Bloquea la fila del evento mientras dure la transacción,
		// así dos "unirse" simultáneos no cuentan el mismo cupo libre.
		var ev Evento
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("codigo_acceso = ?", in.CodigoAcceso).First(&ev).Error; err != nil {
			return &errHTTP{http.StatusNotFound, "codigo_inexistente", "El código de acceso no existe"}
		}
		if ev.Estado == EstadoFinalizado || ev.Estado == EstadoCancelado {
			return &errHTTP{http.StatusConflict, "evento_cerrado", "El evento ya no admite invitados"}
		}

		// ¿Ya existe con ese nombre? Reutiliza su registro (no consume cupo nuevo).
		err := tx.Where("evento_id = ? AND nombre = ?", ev.ID, in.Nombre).First(&inv).Error
		if err == nil {
			if inv.Pin != in.Pin {
				return &errHTTP{http.StatusUnauthorized, "pin_incorrecto", "PIN incorrecto"}
			}
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err // error real de BD, no "no encontrado"
		}

		var total int64
		tx.Model(&Invitado{}).Where("evento_id = ?", ev.ID).Count(&total)
		if int(total) >= ev.MaxInvitado {
			return &errHTTP{http.StatusConflict, "evento_lleno", "El evento alcanzó el máximo de invitados"}
		}

		inv = Invitado{Nombre: in.Nombre, Pin: in.Pin, EventoID: ev.ID}
		return tx.Create(&inv).Error
	})

	if err != nil {
		var eh *errHTTP
		if errors.As(err, &eh) {
			respuesta.Error(w, eh.status, eh.codigo, eh.msg)
			return
		}
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "No se pudo registrar al invitado")
		return
	}
	token := ""
	if m.Tokens != nil {
		var err error
		token, err = m.Tokens.EmitirInvitado(inv.ID, inv.EventoID, inv.Nombre)
		if err != nil {
			respuesta.Error(w, http.StatusInternalServerError, "error_token", "No se pudo crear la sesión del invitado")
			return
		}
	}
	respuesta.Exito(w, http.StatusOK, map[string]any{"invitado": inv, "token": token})
}

// POST /fotos
func (m *Manejador) crearFoto(w http.ResponseWriter, r *http.Request) {
	ses, _ := sesion.De(r)
	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
	var in struct {
		UUID       string `json:"uuid"`
		URLArchivo string `json:"url_archivo"`
		InvitadoID uint   `json:"invitado_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		var limite *http.MaxBytesError
		if errors.As(err, &limite) {
			respuesta.Error(w, http.StatusRequestEntityTooLarge, "foto_muy_grande", "La petición supera el límite de 5 MiB")
			return
		}
		respuesta.Error(w, http.StatusBadRequest, "json_invalido", "El cuerpo no es un JSON válido")
		return
	}
	if in.UUID == "" || len(in.UUID) > 80 || !archivoImagenPermitido(in.URLArchivo) || in.InvitadoID == 0 {
		respuesta.Error(w, http.StatusBadRequest, "datos_invalidos", "uuid, url_archivo e invitado_id son obligatorios")
		return
	}
	if ses.InvitadoID == 0 || ses.InvitadoID != in.InvitadoID {
		respuesta.Error(w, http.StatusForbidden, "invitado_no_coincide", "La sesión no pertenece a este invitado")
		return
	}

	// Reintento del celular con el mismo uuid: no se cuenta dos veces.
	var existente Foto
	if err := m.DB.Where("uuid = ?", in.UUID).First(&existente).Error; err == nil {
		if existente.InvitadoID != ses.InvitadoID {
			respuesta.Error(w, http.StatusForbidden, "foto_no_pertenece", "Esta foto pertenece a otra sesión")
			return
		}
		respuesta.Exito(w, http.StatusOK, existente)
		return
	}

	var foto Foto
	err := m.DB.Transaction(func(tx *gorm.DB) error {
		var inv Invitado
		if err := tx.First(&inv, in.InvitadoID).Error; err != nil {
			return &errHTTP{http.StatusNotFound, "invitado_no_encontrado", "El invitado no existe"}
		}
		var ev Evento
		if err := tx.First(&ev, inv.EventoID).Error; err != nil {
			return &errHTTP{http.StatusNotFound, "evento_no_encontrado", "El evento no existe"}
		}
		if ev.Estado != EstadoActivo {
			return &errHTTP{http.StatusForbidden, "evento_no_activo", "El evento no está activo"}
		}
		// Suma 1 solo si todavía queda cuota (atómico en el servidor).
		res := tx.Model(&Invitado{}).
			Where("id = ? AND fotos_tomadas < ?", inv.ID, ev.DisparoPorInvitado).
			UpdateColumn("fotos_tomadas", gorm.Expr("fotos_tomadas + 1"))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return &errHTTP{http.StatusForbidden, "limite_alcanzado", "Se alcanzó el límite de fotos por invitado"}
		}
		foto = Foto{UUID: in.UUID, URLArchivo: in.URLArchivo, Sincronizado: true, EventoID: ev.ID, InvitadoID: inv.ID}
		return tx.Create(&foto).Error
	})
	if err != nil {
		var eh *errHTTP
		if errors.As(err, &eh) {
			respuesta.Error(w, eh.status, eh.codigo, eh.msg)
			return
		}
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "No se pudo guardar")
		return
	}
	respuesta.Exito(w, http.StatusCreated, foto)
}

func archivoImagenPermitido(valor string) bool {
	if valor == "" || len(valor) > 4<<20 {
		return false
	}
	for _, prefijo := range []string{"data:image/jpeg;base64,", "data:image/png;base64,", "data:image/webp;base64,", "https://", "/archivos/"} {
		if strings.HasPrefix(valor, prefijo) {
			return true
		}
	}
	return false
}

// GET /eventos/{id}/fotos
func (m *Manejador) listarFotos(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}
	var ev Evento
	if err := m.DB.First(&ev, id).Error; err != nil {
		respuesta.Error(w, http.StatusNotFound, "no_encontrado", "Evento no encontrado")
		return
	}
	ses, _ := sesion.De(r)
	if ses.Rol == sesion.RolAnfitrion {
		if _, err := m.eventos().Obtener(r.Context(), ses.UsuarioID, id); err != nil {
			m.responderErrorEvento(w, err)
			return
		}
	} else if ses.Rol != sesion.RolInvitado || ses.InvitadoID == 0 || ses.EventoID != id {
		respuesta.Error(w, http.StatusForbidden, "evento_no_autorizado", "La sesión no pertenece a este evento")
		return
	}
	if ses.Rol == sesion.RolInvitado && ev.Estado != EstadoFinalizado {
		respuesta.Exito(w, http.StatusOK, []Foto{})
		return
	}
	fotos := []Foto{}
	if err := m.DB.Where("evento_id = ?", ev.ID).Order("id").Find(&fotos).Error; err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "No se pudo listar")
		return
	}
	respuesta.Exito(w, http.StatusOK, fotos)
}

// PATCH /eventos/{id}/estado
func (m *Manejador) cambiarEstado(w http.ResponseWriter, r *http.Request) {
	ses, _ := sesion.De(r) // el guardia ya comprobó el token y el rol
	id, ok := leerID(w, r)
	if !ok {
		return
	}
	var in struct {
		Estado string `json:"estado"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "json_invalido", "El cuerpo no es un JSON válido")
		return
	}
	ev, err := m.eventos().CambiarEstado(r.Context(), ses.UsuarioID, id, in.Estado)
	if err != nil {
		m.responderErrorEvento(w, err)
		return
	}
	respuesta.Exito(w, http.StatusOK, ev)
}

// GET /eventos/{id}
func (m *Manejador) verEvento(w http.ResponseWriter, r *http.Request) {
	ses, _ := sesion.De(r) // el guardia ya comprobó el token y el rol
	id, ok := leerID(w, r)
	if !ok {
		return
	}
	ev, err := m.eventos().Obtener(r.Context(), ses.UsuarioID, id)
	if err != nil {
		m.responderErrorEvento(w, err)
		return
	}
	respuesta.Exito(w, http.StatusOK, ev)
}

// actualizarEvento edita la configuración del evento antes de iniciarlo.
func (m *Manejador) actualizarEvento(w http.ResponseWriter, r *http.Request) {
	ses, _ := sesion.De(r)
	id, ok := leerID(w, r)
	if !ok {
		return
	}
	var in struct {
		Nombre             string `json:"nombre"`
		DisparoPorInvitado int    `json:"disparo_por_invitado"`
		MaxInvitado        int    `json:"max_invitado"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "json_invalido", "El cuerpo no es un JSON válido")
		return
	}
	ev, err := m.eventos().Actualizar(r.Context(), ses.UsuarioID, id, EventoEdicion{
		Nombre: in.Nombre, DisparoPorInvitado: in.DisparoPorInvitado, MaxInvitado: in.MaxInvitado,
	})
	if err != nil {
		m.responderErrorEvento(w, err)
		return
	}
	respuesta.Exito(w, http.StatusOK, ev)
}

func (m *Manejador) eliminarEvento(w http.ResponseWriter, r *http.Request) {
	ses, _ := sesion.De(r)
	id, ok := leerID(w, r)
	if !ok {
		return
	}
	if err := m.eventos().Eliminar(r.Context(), ses.UsuarioID, id); err != nil {
		m.responderErrorEvento(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /eventos — lista los eventos del anfitrión autenticado.
func (m *Manejador) listarEventos(w http.ResponseWriter, r *http.Request) {
	ses, _ := sesion.De(r)
	eventos, err := m.eventos().Listar(r.Context(), ses.UsuarioID)
	if err != nil {
		m.responderErrorEvento(w, err)
		return
	}
	respuesta.Exito(w, http.StatusOK, eventos)
}

// GET /eventos/{id}/resumen — datos mínimos que necesita cualquier asistente.
func (m *Manejador) resumenEvento(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}
	ev, err := m.eventos().Resumen(r.Context(), id)
	if err != nil {
		m.responderErrorEvento(w, err)
		return
	}
	ses, _ := sesion.De(r)
	if ses.Rol == sesion.RolAnfitrion {
		if _, err := m.eventos().Obtener(r.Context(), ses.UsuarioID, id); err != nil {
			m.responderErrorEvento(w, err)
			return
		}
	} else if ses.Rol != sesion.RolInvitado || ses.InvitadoID == 0 || ses.EventoID != id {
		respuesta.Error(w, http.StatusForbidden, "evento_no_autorizado", "La sesión no pertenece a este evento")
		return
	}
	respuesta.Exito(w, http.StatusOK, map[string]any{
		"id": ev.ID, "nombre": ev.Nombre, "estado": ev.Estado,
		"disparo_por_invitado": ev.DisparoPorInvitado,
	})
}
