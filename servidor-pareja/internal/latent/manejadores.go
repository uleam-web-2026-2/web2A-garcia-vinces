package latent

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Jeanvncs/myfirstgo/internal/respuesta"
)

type Manejador struct{ DB *gorm.DB }

var rePin = regexp.MustCompile(`^\d{4}$`)

type errHTTP struct {
	status int
	codigo string
	msg    string
}

func (e *errHTTP) Error() string { return e.msg }

// Identificación provisional por cabecera (el Hito 1 no pide login).
func adminID(r *http.Request) (uint, bool) {
	n, err := strconv.ParseUint(r.Header.Get("X-Usuario-ID"), 10, 64)
	return uint(n), err == nil && n > 0
}

func invitadoID(r *http.Request) (uint, bool) {
	n, err := strconv.ParseUint(r.Header.Get("X-Invitado-ID"), 10, 64)
	return uint(n), err == nil && n > 0
}

func leerID(w http.ResponseWriter, r *http.Request) (uint, bool) {
	n, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || n <= 0 {
		respuesta.Error(w, http.StatusBadRequest, "id_invalido", "El id debe ser un numero positivo")
		return 0, false
	}
	return uint(n), true
}

func generarCodigo() string {
	const alfabeto = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 4)
	rand.Read(b)
	for i := range b {
		b[i] = alfabeto[int(b[i])%len(alfabeto)]
	}
	return "EVT-" + string(b)
}

// POST /eventos
func (m *Manejador) crearEvento(w http.ResponseWriter, r *http.Request) {
	uid, ok := adminID(r)
	if !ok {
		respuesta.Error(w, http.StatusUnauthorized, "no_identificado", "Falta X-Usuario-ID")
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
	if strings.TrimSpace(in.Nombre) == "" || in.DisparoPorInvitado < 1 || in.MaxInvitado < 1 {
		respuesta.Error(w, http.StatusBadRequest, "evento_invalido",
			"nombre, disparo_por_invitado y max_invitado son obligatorios y deben ser válidos")
		return
	}
	ev := Evento{
		Nombre:             strings.TrimSpace(in.Nombre),
		CodigoAcceso:       generarCodigo(),
		DisparoPorInvitado: in.DisparoPorInvitado,
		MaxInvitado:        in.MaxInvitado,
		Estado:             EstadoCreado,
		UsuarioID:          uid,
	}
	if err := m.DB.Create(&ev).Error; err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "No se pudo guardar")
		return
	}
	respuesta.Exito(w, http.StatusCreated, ev)
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
	respuesta.Exito(w, http.StatusOK, inv)
}

// POST /fotos
func (m *Manejador) crearFoto(w http.ResponseWriter, r *http.Request) {
	var in struct {
		UUID       string `json:"uuid"`
		URLArchivo string `json:"url_archivo"`
		InvitadoID uint   `json:"invitado_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "json_invalido", "El cuerpo no es un JSON válido")
		return
	}
	if in.UUID == "" || in.URLArchivo == "" || in.InvitadoID == 0 {
		respuesta.Error(w, http.StatusBadRequest, "datos_invalidos", "uuid, url_archivo e invitado_id son obligatorios")
		return
	}

	// Reintento del celular con el mismo uuid: no se cuenta dos veces.
	var existente Foto
	if err := m.DB.Where("uuid = ?", in.UUID).First(&existente).Error; err == nil {
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

// GET /eventos/{id}/fotos
func (m *Manejador) listarFotos(w http.ResponseWriter, r *http.Request) {
	_, esAdmin := adminID(r)
	_, esInvitado := invitadoID(r)
	if !esAdmin && !esInvitado {
		respuesta.Error(w, http.StatusUnauthorized, "no_identificado", "Debes identificarte")
		return
	}
	id, ok := leerID(w, r)
	if !ok {
		return
	}
	var ev Evento
	if err := m.DB.First(&ev, id).Error; err != nil {
		respuesta.Error(w, http.StatusNotFound, "no_encontrado", "Evento no encontrado")
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
	uid, ok := adminID(r)
	if !ok {
		respuesta.Error(w, http.StatusUnauthorized, "no_identificado", "Falta X-Usuario-ID")
		return
	}
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
	if _, conocido := transiciones[in.Estado]; !conocido {
		respuesta.Error(w, http.StatusBadRequest, "estado_desconocido", "Estado desconocido")
		return
	}
	var ev Evento
	if err := m.DB.First(&ev, id).Error; err != nil {
		respuesta.Error(w, http.StatusNotFound, "no_encontrado", "Evento no encontrado")
		return
	}
	if ev.UsuarioID != uid {
		respuesta.Error(w, http.StatusForbidden, "no_es_tuyo", "El evento no es tuyo")
		return
	}
	if !slices.Contains(transiciones[ev.Estado], in.Estado) {
		respuesta.Error(w, http.StatusConflict, "transicion_prohibida",
			"Transición no permitida: "+ev.Estado+" → "+in.Estado)
		return
	}
	ev.Estado = in.Estado
	if in.Estado == EstadoFinalizado {
		ahora := time.Now()
		ev.FechaRevelado = &ahora
	}
	if err := m.DB.Save(&ev).Error; err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "No se pudo cambiar el estado")
		return
	}
	respuesta.Exito(w, http.StatusOK, ev)
}

// GET /eventos/{id}
func (m *Manejador) verEvento(w http.ResponseWriter, r *http.Request) {
	uid, ok := adminID(r)
	if !ok {
		respuesta.Error(w, http.StatusUnauthorized, "no_identificado", "Falta X-Usuario-ID")
		return
	}
	id, ok := leerID(w, r)
	if !ok {
		return
	}
	var ev Evento
	if err := m.DB.First(&ev, id).Error; err != nil {
		respuesta.Error(w, http.StatusNotFound, "no_encontrado", "Evento no encontrado")
		return
	}
	if ev.UsuarioID != uid {
		respuesta.Error(w, http.StatusForbidden, "no_es_tuyo", "El evento no es tuyo")
		return
	}
	respuesta.Exito(w, http.StatusOK, ev)
}