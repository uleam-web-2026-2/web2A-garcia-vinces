package latent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/Jeanvncs/myfirstgo/internal/sesion"
)

// ---------------------------------------------------------------------------
// ÚNICA parte que depende de internal/sesion/tokens.go.
// Ajusta estas dos funciones a lo que ofrezca tu paquete.
// ---------------------------------------------------------------------------

// nuevosTokens construye el emisor/validador de tokens para las pruebas.
func nuevosTokens() *sesion.Tokens {
	return sesion.NuevosTokens("clave-de-prueba-con-longitud-suficiente")
}

// tokenPara emite un token válido para el rol e ID indicados.
func tokenPara(tk *sesion.Tokens, rol string, id uint) string {
	var (
		token string
		err   error
	)
	if rol == sesion.RolInvitado {
		token, err = tk.EmitirInvitado(id, 1, "Invitado de prueba")
	} else {
		token, err = tk.Emitir(sesion.Usuario{ID: id, Nombre: "Anfitrión de prueba", Rol: rol})
	}
	if err != nil {
		panic(err)
	}
	return token
}

// ---------------------------------------------------------------------------
// Ayudantes de prueba (no dependen de tokens.go)
// ---------------------------------------------------------------------------

type servidorPrueba struct {
	DB     *gorm.DB
	router chi.Router
	tokens *sesion.Tokens
}

func nuevoServidor(t *testing.T) *servidorPrueba {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	})
	if err := db.AutoMigrate(&Usuario{}, &Evento{}, &Invitado{}, &Foto{}); err != nil {
		t.Fatal(err)
	}

	tk := nuevosTokens()
	r := chi.NewRouter()
	(&Manejador{DB: db, Tokens: tk}).Rutas(r, tk)
	return &servidorPrueba{DB: db, router: r, tokens: tk}
}

// llamar envía la petición sin token (para rutas públicas).
func llamar(s *servidorPrueba, metodo, ruta, cuerpo string) *httptest.ResponseRecorder {
	return enviar(s, "", metodo, ruta, cuerpo)
}

// llamarComo envía la petición con un token del rol y el ID indicados.
func llamarComo(s *servidorPrueba, rol string, id uint, metodo, ruta, cuerpo string) *httptest.ResponseRecorder {
	return enviar(s, tokenPara(s.tokens, rol, id), metodo, ruta, cuerpo)
}

func enviar(s *servidorPrueba, token, metodo, ruta, cuerpo string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
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

func crear(t *testing.T, db *gorm.DB, valor any) {
	t.Helper()
	if err := db.Create(valor).Error; err != nil {
		t.Fatal(err)
	}
}

// ---------------------------------------------------------------------------
// Pruebas (mismas comprobaciones que antes)
// ---------------------------------------------------------------------------

func TestEventoInvalido(t *testing.T) {
	s := nuevoServidor(t)
	rec := llamarComo(s, sesion.RolAnfitrion, 1, "POST", "/eventos",
		`{"disparo_por_invitado":3,"max_invitado":50}`)
	esperar(t, rec, http.StatusBadRequest)
}

func TestCodigoInexistente(t *testing.T) {
	s := nuevoServidor(t)
	rec := llamar(s, "POST", "/eventos/unirse",
		`{"codigo_acceso":"NO-EXISTE","nombre":"Carlos","pin":"0427"}`)
	esperar(t, rec, http.StatusNotFound)
}

func TestCuotaAgotada(t *testing.T) {
	s := nuevoServidor(t)
	ev := Evento{Nombre: "Boda", CodigoAcceso: "T-1", DisparoPorInvitado: 3, MaxInvitado: 10, Estado: EstadoActivo, UsuarioID: 1}
	crear(t, s.DB, &ev)
	inv := Invitado{Nombre: "Carlos", Pin: "0427", FotosTomadas: 3, EventoID: ev.ID}
	crear(t, s.DB, &inv)

	cuerpo := fmt.Sprintf(`{"uuid":"u-1","url_archivo":"/archivos/1.jpg","invitado_id":%d}`, inv.ID)
	rec := llamarComo(s, sesion.RolInvitado, inv.ID, "POST", "/fotos", cuerpo)
	esperar(t, rec, http.StatusForbidden)
}

func TestEstadoDesconocido(t *testing.T) {
	s := nuevoServidor(t)
	ev := Evento{Nombre: "Boda", CodigoAcceso: "T-2", DisparoPorInvitado: 3, MaxInvitado: 10, Estado: EstadoCreado, UsuarioID: 1}
	crear(t, s.DB, &ev)

	ruta := fmt.Sprintf("/eventos/%d/estado", ev.ID)
	rec := llamarComo(s, sesion.RolAnfitrion, 1, "PATCH", ruta, `{"estado":"volando"}`)
	esperar(t, rec, http.StatusUnprocessableEntity)
}

func TestTransicionProhibida(t *testing.T) {
	s := nuevoServidor(t)
	ev := Evento{Nombre: "Boda", CodigoAcceso: "T-3", DisparoPorInvitado: 3, MaxInvitado: 10, Estado: EstadoFinalizado, UsuarioID: 1}
	crear(t, s.DB, &ev)

	ruta := fmt.Sprintf("/eventos/%d/estado", ev.ID)
	rec := llamarComo(s, sesion.RolAnfitrion, 1, "PATCH", ruta, `{"estado":"activo"}`)
	esperar(t, rec, http.StatusUnprocessableEntity)

	var despues Evento
	if err := s.DB.First(&despues, ev.ID).Error; err != nil {
		t.Fatal(err)
	}
	if despues.Estado != EstadoFinalizado {
		t.Fatalf("el estado cambió a %s y debía seguir finalizado", despues.Estado)
	}
}

func TestSinTokenLasRutasProtegidasResponden401(t *testing.T) {
	s := nuevoServidor(t)
	rec := llamar(s, "GET", "/eventos", "")
	esperar(t, rec, http.StatusUnauthorized)
}

func TestEventoCRUD(t *testing.T) {
	s := nuevoServidor(t)
	creado := llamarComo(s, sesion.RolAnfitrion, 1, http.MethodPost, "/eventos",
		`{"nombre":"Cumpleaños","disparo_por_invitado":3,"max_invitado":30}`)
	esperar(t, creado, http.StatusCreated)
	var envoltura struct {
		Datos Evento `json:"datos"`
	}
	if err := json.Unmarshal(creado.Body.Bytes(), &envoltura); err != nil {
		t.Fatal(err)
	}
	id := envoltura.Datos.ID
	if id == 0 || envoltura.Datos.Estado != EstadoCreado || envoltura.Datos.CodigoAcceso == "" {
		t.Fatalf("evento creado incompleto: %+v", envoltura.Datos)
	}

	ruta := fmt.Sprintf("/eventos/%d", id)
	actualizado := llamarComo(s, sesion.RolAnfitrion, 1, http.MethodPut, ruta,
		`{"nombre":"Cumpleaños de Ana","disparo_por_invitado":4,"max_invitado":35}`)
	esperar(t, actualizado, http.StatusOK)
	verificado := llamarComo(s, sesion.RolAnfitrion, 1, http.MethodGet, ruta, "")
	esperar(t, verificado, http.StatusOK)

	eliminado := llamarComo(s, sesion.RolAnfitrion, 1, http.MethodDelete, ruta, "")
	esperar(t, eliminado, http.StatusNoContent)
	esperar(t, llamarComo(s, sesion.RolAnfitrion, 1, http.MethodGet, ruta, ""), http.StatusNotFound)
}

func TestDemoHito2RespuestasDeRolesYTransicion(t *testing.T) {
	s := nuevoServidor(t)
	if err := s.DB.AutoMigrate(&sesion.Usuario{}); err != nil {
		t.Fatal(err)
	}
	if err := sesion.Sembrar(s.DB); err != nil {
		t.Fatal(err)
	}
	usuario := Usuario{ID: 2, Nombre: "luis", Rol: sesion.RolAnfitrion}
	crear(t, s.DB, &usuario)

	finalizado := Evento{Nombre: "Boda revelada", CodigoAcceso: "EVT-DEMO1", DisparoPorInvitado: 3, MaxInvitado: 50, Estado: EstadoFinalizado, UsuarioID: 2}
	creado := Evento{Nombre: "Fiesta de demostración", CodigoAcceso: "EVT-DEMO2", DisparoPorInvitado: 3, MaxInvitado: 50, Estado: EstadoCreado, UsuarioID: 2}
	crear(t, s.DB, &finalizado)
	crear(t, s.DB, &creado)

	r := chi.NewRouter()
	r.Post("/login", (&sesion.Manejador{DB: s.DB, Tokens: s.tokens}).Login)
	(&Manejador{DB: s.DB, Tokens: s.tokens}).Rutas(r, s.tokens)
	s.router = r

	// Petición 1: ruta protegida sin token.
	esperar(t, llamar(s, "GET", "/eventos/2/fotos", ""), http.StatusUnauthorized)

	// Preparación: el invitado entra con código, nombre y PIN, y recibe un
	// token vinculado a su registro y a ese evento.
	union := llamar(s, http.MethodPost, "/eventos/unirse",
		`{"codigo_acceso":"EVT-DEMO2","nombre":"Carlos","pin":"0427"}`)
	esperar(t, union, http.StatusOK)
	var respuestaUnion struct {
		Datos struct {
			Token string `json:"token"`
		} `json:"datos"`
	}
	if err := json.Unmarshal(union.Body.Bytes(), &respuestaUnion); err != nil {
		t.Fatal(err)
	}
	invitadoToken := respuestaUnion.Datos.Token

	// Petición 2: el anfitrión inicia sesión y obtiene su token.
	luis := loginDePrueba(t, s, "luis", "luis-2026")

	// Petición 3: la solicitante puede ver la galería del evento.
	esperar(t, enviar(s, invitadoToken, "GET", "/eventos/2/fotos", ""), http.StatusOK)

	// Petición 4: la solicitante no puede administrar estados.
	esperar(t, enviar(s, invitadoToken, "PATCH", "/eventos/2", `{"estado":"activo"}`), http.StatusForbidden)

	// Petición 5: el anfitrión puede iniciar un evento creado.
	esperar(t, enviar(s, luis, "PATCH", "/eventos/2", `{"estado":"activo"}`), http.StatusOK)

	// Petición 6: un evento revelado no puede reactivarse.
	esperar(t, enviar(s, luis, "PATCH", "/eventos/1", `{"estado":"activo"}`), http.StatusUnprocessableEntity)
}

func loginDePrueba(t *testing.T, s *servidorPrueba, usuario, clave string) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(fmt.Sprintf(`{"usuario":%q,"clave":%q}`, usuario, clave)))
	res := httptest.NewRecorder()
	(&sesion.Manejador{DB: s.DB, Tokens: s.tokens}).Login(res, req)
	esperar(t, res, http.StatusOK)
	var cuerpo struct {
		Datos struct {
			Token string `json:"token"`
		} `json:"datos"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &cuerpo); err != nil {
		t.Fatal(err)
	}
	return cuerpo.Datos.Token
}
