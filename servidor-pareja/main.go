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

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	if *reset {
		// primero las tablas del lado de los muchos, después las del uno
		db.Migrator().DropTable(&latent.Foto{}, &latent.Invitado{}, &latent.Evento{}, &latent.Usuario{})
	}

	err = db.Debug().AutoMigrate(&latent.Usuario{}, &latent.Evento{}, &latent.Invitado{}, &latent.Foto{})
	if err != nil {
		log.Fatal(err)
	}

	// Admin de prueba (id 1) para el Hito 1
	db.FirstOrCreate(&latent.Usuario{ID: 1}, latent.Usuario{Nombre: "Admin", Rol: "admin"})

	r := chi.NewRouter()
	// r.Use(middleware.TuPrimero, middleware.TuSegundo)  // descomenta con los nombres reales
	(&latent.Manejador{DB: db}).Rutas(r)

	log.Println("Latent escuchando en :" + puerto)
	log.Fatal(http.ListenAndServe(":"+puerto, r))
}

//$env:PUERTO = "8081"
//$env:DATABASE_URL = "postgres://UserDB:12345@localhost:5432/Latent?sslmode=disable"
//http://localhost:8081


