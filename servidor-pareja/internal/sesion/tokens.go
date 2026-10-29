package sesion

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Duracion es cuánto vale un token desde que se emite.
const Duracion = time.Hour

// Sesion es lo que el servidor sabe de quien hace la petición.
type Sesion struct {
	UsuarioID  uint
	InvitadoID uint
	EventoID   uint
	Nombre     string
	Rol        string
}

// datos viaja firmado en el JWT; su contenido no está cifrado.
type datos struct {
	Nombre     string `json:"nombre"`
	Rol        string `json:"rol"`
	InvitadoID uint   `json:"invitado_id,omitempty"`
	EventoID   uint   `json:"evento_id,omitempty"`
	jwt.RegisteredClaims
}

// Tokens emite y revisa tokens con el secreto configurado fuera del código.
type Tokens struct{ secreto []byte }

func NuevosTokens(secreto string) *Tokens { return &Tokens{secreto: []byte(secreto)} }

// Emitir firma un token para una cuenta que ya comprobó sus credenciales.
func (t *Tokens) Emitir(u Usuario) (string, error) {
	ahora := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, datos{
		Nombre: u.Nombre,
		Rol:    u.Rol,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: strconv.Itoa(int(u.ID)), IssuedAt: jwt.NewNumericDate(ahora),
			ExpiresAt: jwt.NewNumericDate(ahora.Add(Duracion)),
		},
	})
	return token.SignedString(t.secreto)
}

// EmitirInvitado crea una sesión ligada a una persona y un único evento.
func (t *Tokens) EmitirInvitado(id, eventoID uint, nombre string) (string, error) {
	ahora := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, datos{
		Nombre: nombre, Rol: RolInvitado, InvitadoID: id, EventoID: eventoID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: strconv.Itoa(int(id)), IssuedAt: jwt.NewNumericDate(ahora),
			ExpiresAt: jwt.NewNumericDate(ahora.Add(Duracion)),
		},
	})
	return token.SignedString(t.secreto)
}

// ErrVencido distingue un token válido que expiró de uno inválido.
var ErrVencido = errors.New("el token está vencido")

// Leer revisa la firma y la fecha de vencimiento antes de devolver la sesión.
func (t *Tokens) Leer(texto string) (Sesion, error) {
	var d datos
	_, err := jwt.ParseWithClaims(texto, &d,
		func(*jwt.Token) (any, error) { return t.secreto, nil },
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
	)
	if errors.Is(err, jwt.ErrTokenExpired) {
		return Sesion{}, ErrVencido
	}
	if err != nil {
		return Sesion{}, err
	}
	id, err := strconv.Atoi(d.Subject)
	if err != nil {
		return Sesion{}, err
	}
	return Sesion{UsuarioID: uint(id), InvitadoID: d.InvitadoID, EventoID: d.EventoID, Nombre: d.Nombre, Rol: d.Rol}, nil
}
