package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Jeanvncs/myfirstgo/internal/latent"
	"github.com/Jeanvncs/myfirstgo/internal/sesion"
)

func main() {
	reset := flag.Bool("reset", false, "borra las tablas y arranca con la base vacía")
	flag.Parse()

	puerto := os.Getenv("PUERTO")
	if puerto == "" {
		puerto = "8081"
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("falta la variable de entorno DATABASE_URL")
	}

	// El secreto con el que se firman los tokens. Nunca va escrito en el código.
	secreto := os.Getenv("CLAVE_FIRMA")
	if len(secreto) < 32 {
		log.Fatal("CLAVE_FIRMA debe tener al menos 32 caracteres")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	if *reset {
		// primero las tablas del lado de los muchos, después las del uno
		db.Migrator().DropTable(&latent.Foto{}, &latent.Invitado{}, &latent.Evento{}, &latent.Usuario{}, &sesion.Usuario{})
	}

	err = db.Debug().AutoMigrate(&latent.Usuario{}, &latent.Evento{}, &latent.Invitado{}, &latent.Foto{}, &sesion.Usuario{})
	if err != nil {
		log.Fatal(err)
	}

	// Admin de prueba (id 1) del Hito 1
	db.FirstOrCreate(&latent.Usuario{ID: 1}, latent.Usuario{Nombre: "Admin", Rol: "admin"})

	// Las credenciales de ejemplo solo se crean cuando se habilita la demo.
	if strings.EqualFold(os.Getenv("MODO_DEMO"), "true") {
		if err := sesion.Sembrar(db); err != nil {
			log.Fatal(err)
		}
	} else {
		usuario := os.Getenv("ANFITRION_USUARIO")
		clave := os.Getenv("ANFITRION_CLAVE")
		if usuario == "" || clave == "" {
			log.Fatal("configure MODO_DEMO=true o ANFITRION_USUARIO y ANFITRION_CLAVE")
		}
		if err := sesion.SembrarAnfitrion(db, usuario, clave); err != nil {
			log.Fatal(err)
		}
	}
	// Los eventos conservan una FK al usuario de dominio. En instalaciones
	// nuevas se crea la fila correspondiente a cada cuenta autenticada.
	var cuentas []sesion.Usuario
	if err := db.Find(&cuentas).Error; err != nil {
		log.Fatal(err)
	}
	for _, cuenta := range cuentas {
		usuario := latent.Usuario{ID: cuenta.ID}
		if err := db.Where("id = ?", cuenta.ID).Attrs(latent.Usuario{Nombre: cuenta.Nombre, Rol: cuenta.Rol}).FirstOrCreate(&usuario).Error; err != nil {
			log.Fatal(err)
		}
	}
	tokens := sesion.NuevosTokens(secreto)

	r := chi.NewRouter()
	// r.Use(middleware.TuPrimero, middleware.TuSegundo)  // descomenta con los nombres reales
	r.Post("/login", (&sesion.Manejador{DB: db, Tokens: tokens}).Login)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}` + "\n"))
	})
	(&latent.Manejador{DB: db, Tokens: tokens}).Rutas(r, tokens)

	log.Println("Latent escuchando en :" + puerto)
	server := &http.Server{Addr: ":" + puerto, Handler: r, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 20 * time.Second, WriteTimeout: 60 * time.Second, IdleTimeout: 60 * time.Second}
	log.Fatal(server.ListenAndServe())
}

// Variables de entorno para correr el servidor (PowerShell):
//   $env:PUERTO = "8081"
//   $env:DATABASE_URL = "postgres://UserDB:SU_CLAVE@localhost:5432/LatentDemo?sslmode=disable"
//   $env:CLAVE_FIRMA = "una-frase-larga-que-solo-usted-conoce"
// URL local: http://localhost:8081
