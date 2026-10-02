package latent

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type servidorPrueba struct {
	DB     *gorm.DB
	router chi.Router
}

func nuevoServidor(t *testing.T) *servidorPrueba {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&Usuario{}, &Evento{}, &Invitado{}, &Foto{}); err != nil {
		t.Fatal(err)
	}
	r := chi.NewRouter()
	(&Manejador{DB: db}).Rutas(r)
	return &servidorPrueba{DB: db, router: r}
}

func llamar(s *servidorPrueba, metodo, ruta, cuerpo string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	req.Header.Set("X-Usuario-ID", "1")
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)
	return rec
}

func esperar(t *testing.T, rec *httptest.ResponseRecorder, codigo int) {
	t.Helper()
	if rec.Code != codigo {
		t.Fatalf("esperaba %d, llegó %d: %s", codigo, rec.Code, rec.Body.String())
	}
}

func TestEventoInvalido(t *testing.T) {
	s := nuevoServidor(t)
	rec := llamar(s, "POST", "/eventos", `{"disparo_por_invitado":3,"max_invitado":50}`)
	esperar(t, rec, http.StatusBadRequest)
}

func TestCodigoInexistente(t *testing.T) {
	s := nuevoServidor(t)
	rec := llamar(s, "POST", "/eventos/unirse", `{"codigo_acceso":"NO-EXISTE","nombre":"Carlos","pin":"0427"}`)
	esperar(t, rec, http.StatusNotFound)
}

func TestCuotaAgotada(t *testing.T) {
	s := nuevoServidor(t)
	ev := Evento{Nombre: "Boda", CodigoAcceso: "T-1", DisparoPorInvitado: 3, MaxInvitado: 10, Estado: EstadoActivo, UsuarioID: 1}
	s.DB.Create(&ev)
	inv := Invitado{Nombre: "Carlos", Pin: "0427", FotosTomadas: 3, EventoID: ev.ID}
	s.DB.Create(&inv)

	rec := llamar(s, "POST", "/fotos", `{"uuid":"u-1","url_archivo":"/archivos/1.jpg","invitado_id":1}`)
	esperar(t, rec, http.StatusForbidden)
}

func TestEstadoDesconocido(t *testing.T) {
	s := nuevoServidor(t)
	ev := Evento{Nombre: "Boda", CodigoAcceso: "T-2", DisparoPorInvitado: 3, MaxInvitado: 10, Estado: EstadoCreado, UsuarioID: 1}
	s.DB.Create(&ev)

	rec := llamar(s, "PATCH", "/eventos/1/estado", `{"estado":"volando"}`)
	esperar(t, rec, http.StatusBadRequest)
}

func TestTransicionProhibida(t *testing.T) {
	s := nuevoServidor(t)
	ev := Evento{Nombre: "Boda", CodigoAcceso: "T-3", DisparoPorInvitado: 3, MaxInvitado: 10, Estado: EstadoFinalizado, UsuarioID: 1}
	s.DB.Create(&ev)

	rec := llamar(s, "PATCH", "/eventos/1/estado", `{"estado":"activo"}`)
	esperar(t, rec, http.StatusConflict)

	var despues Evento
	s.DB.First(&despues, ev.ID)
	if despues.Estado != EstadoFinalizado {
		t.Fatalf("el estado cambió a %s y debía seguir finalizado", despues.Estado)
	}
}