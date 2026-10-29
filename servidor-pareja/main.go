package main

import (
	"flag"
	"log"
	"net/http"
	"os"

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
	if secreto == "" {
		log.Fatal("falta la variable de entorno CLAVE_FIRMA")
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

	// Las dos cuentas de ejemplo para iniciar sesión (una por rol)
	if err := sesion.Sembrar(db); err != nil {
		log.Fatal(err)
	}

	tokens := sesion.NuevosTokens(secreto)

	r := chi.NewRouter()
	// r.Use(middleware.TuPrimero, middleware.TuSegundo)  // descomenta con los nombres reales
	r.Post("/login", (&sesion.Manejador{DB: db, Tokens: tokens}).Login)
	(&latent.Manejador{DB: db}).Rutas(r, tokens)
	r.Handle("/*", http.FileServer(http.Dir("web")))

	log.Println("Latent escuchando en :" + puerto)
	log.Fatal(http.ListenAndServe(":"+puerto, r))
}

// Variables de entorno para correr el servidor (PowerShell):
//   $env:PUERTO = "8081"
//   $env:DATABASE_URL = "postgres://UserDB:SU_CLAVE@localhost:5432/Latent?sslmode=disable"
//   $env:CLAVE_FIRMA = "una-frase-larga-que-solo-usted-conoce"
// URL local: http://localhost:8081
