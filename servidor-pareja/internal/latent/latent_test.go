package latent

import (
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
	// TODO: usa el constructor real. Ejemplos posibles:
	//   return sesion.NuevoTokens("clave-de-prueba")
	//   return sesion.NewTokens([]byte("clave-de-prueba"))
	panic("completar con el constructor de sesion.Tokens")
}

// tokenPara emite un token válido para el rol e ID indicados.
func tokenPara(tk *sesion.Tokens, rol string, id uint) string {
	// TODO: usa el método real. Ejemplo posible:
	//   token, _ := tk.Emitir(id, rol)
	//   return token
	panic("completar con el método que emite tokens")
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
	(&Manejador{DB: db}).Rutas(r, tk)
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