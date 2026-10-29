package sesion

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/Jeanvncs/myfirstgo/internal/respuesta"
)

// Los dos roles de Latent. El nombre de la constante empieza con Rol;
// el texto va en minúsculas, sin espacios y sin tildes.
const (
	RolInvitado  = "invitado"  // el que toma las fotos
	RolAnfitrion = "anfitrion" // el que organiza el evento
)

// Usuario es una persona que puede iniciar sesión.
// De la contraseña se guarda un resumen (bcrypt), nunca la contraseña.
type Usuario struct {
	ID        uint   `json:"id"`
	Nombre    string `json:"nombre" gorm:"uniqueIndex"`
	Rol       string `json:"rol"`
	ClaveHash string `json:"-"`
}

// El modelo de Latent ya tiene una tabla «usuarios»: la de inicio de
// sesión se llama «cuentas» para que no se mezclen.
func (Usuario) TableName() string {
	return "cuentas"
}

// Sembrar deja dos usuarios de ejemplo, uno por rol. Corre en cada
// arranque: si cambian un rol o una contraseña aquí, al reiniciar el
// servidor el usuario queda actualizado. Son usuarios y contraseñas
// INVENTADOS, solo para probar: nunca una contraseña real.
func Sembrar(db *gorm.DB) error {
	ejemplos := []struct{ nombre, rol, clave string }{
		{"ana", RolInvitado, "ana-2026"},
		{"luis", RolAnfitrion, "luis-2026"},
	}
	for _, e := range ejemplos {
		// bcrypt calcula el resumen: eso es lo que se guarda.
		resumen, err := bcrypt.GenerateFromPassword(
			[]byte(e.clave), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		var u Usuario
		db.Where("nombre = ?", e.nombre).Limit(1).Find(&u)
		u.Nombre, u.Rol, u.ClaveHash = e.nombre, e.rol, string(resumen)
		if err := db.Save(&u).Error; err != nil {
			return err
		}
	}
	return nil
}

// Manejador tiene la única ruta de este paquete: iniciar sesión.
type Manejador struct {
	DB     *gorm.DB
	Tokens *Tokens
}

// Login: POST /login con {"usuario": "...", "clave": "..."}.
// 400 si el JSON está roto · 401 si no coincide · 200 con el token.
func (m *Manejador) Login(w http.ResponseWriter, r *http.Request) {
	var pedido struct {
		Usuario string `json:"usuario"`
		Clave   string `json:"clave"`
	}
	if err := json.NewDecoder(r.Body).Decode(&pedido); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "json_invalido",
			"el cuerpo no es JSON válido")
		return
	}
	// El mismo mensaje para «no existe» y para «contraseña incorrecta»:
	// quien prueba nombres al azar no averigua cuáles existen.
	const noCoincide = "usuario o contraseña incorrectos"
	var u Usuario
	err := m.DB.Where("nombre = ?", pedido.Usuario).Limit(1).Find(&u).Error
	if err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "error_base",
			"no se pudo leer")
		return
	}
	// bcrypt compara la contraseña que llegó con el resumen guardado.
	// Si el usuario no existe, u.ID es 0 y no hay nada que comparar.
	if u.ID == 0 || bcrypt.CompareHashAndPassword(
		[]byte(u.ClaveHash), []byte(pedido.Clave)) != nil {
		respuesta.Error(w, http.StatusUnauthorized, "credenciales_invalidas", noCoincide)
		return
	}
	token, err := m.Tokens.Emitir(u)
	if err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "error_token",
			"no se pudo emitir el token")
		return
	}
	respuesta.Exito(w, http.StatusOK, map[string]string{"token": token})
}