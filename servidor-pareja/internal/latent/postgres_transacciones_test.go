package latent

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestRollbackTransaccionesPostgres es una comprobación optativa de integración.
// Solo se ejecuta cuando se entrega explícitamente una URL de PostgreSQL; se recomienda una base de pruebas.
func TestRollbackTransaccionesPostgres(t *testing.T) {
	dsn := os.Getenv("TALLER10_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("defina TALLER10_TEST_DATABASE_URL para ejecutar contra PostgreSQL")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("no se pudo abrir PostgreSQL")
	}
	conexion, err := db.DB()
	if err != nil {
		t.Fatal("no se pudo obtener la conexión PostgreSQL")
	}
	t.Cleanup(func() { _ = conexion.Close() })
	if err := conexion.PingContext(context.Background()); err != nil {
		t.Fatal("PostgreSQL no responde")
	}

	sufijo := fmt.Sprintf("%d", time.Now().UnixNano())
	evento := Evento{
		Nombre: "Verificación taller 10 " + sufijo, CodigoAcceso: "EVT-T10-" + sufijo,
		DisparoPorInvitado: 3, MaxInvitado: 10, Estado: EstadoActivo, UsuarioID: 2,
	}
	if err := db.Create(&evento).Error; err != nil {
		t.Fatal("no se pudo crear el evento temporal")
	}
	invitado := Invitado{Nombre: "Verificación " + sufijo, Pin: "0427", EventoID: evento.ID}
	t.Cleanup(func() {
		if err := db.Where("evento_id = ?", evento.ID).Delete(&Foto{}).Error; err != nil {
			t.Errorf("no se pudieron limpiar las fotos temporales: %v", err)
		}
		if invitado.ID != 0 {
			if err := db.Delete(&Invitado{}, invitado.ID).Error; err != nil {
				t.Errorf("no se pudo limpiar el invitado temporal: %v", err)
			}
		}
		if err := db.Delete(&Evento{}, evento.ID).Error; err != nil {
			t.Errorf("no se pudo limpiar el evento temporal: %v", err)
		}
		var restantes int64
		if err := db.Model(&Evento{}).Where("id = ?", evento.ID).Count(&restantes).Error; err != nil || restantes != 0 {
			t.Errorf("el evento temporal no quedó limpio; filas restantes: %d, error: %v", restantes, err)
		}
	})
	if err := db.Create(&invitado).Error; err != nil {
		t.Fatal("no se pudo crear el invitado temporal")
	}
	existente := Foto{
		UUID: "rollback-t10-" + sufijo, URLArchivo: "/archivos/ya-existente.jpg",
		Sincronizado: true, EventoID: evento.ID, InvitadoID: invitado.ID,
	}
	if err := db.Create(&existente).Error; err != nil {
		t.Fatal("no se pudo preparar el UUID duplicado")
	}

	verificarRollback := func(nombre string) {
		t.Helper()
		var actual Invitado
		if err := db.First(&actual, invitado.ID).Error; err != nil {
			t.Fatalf("%s: no se encontró el invitado temporal", nombre)
		}
		if actual.FotosTomadas != 0 {
			t.Fatalf("%s: el contador quedó en %d", nombre, actual.FotosTomadas)
		}
		var total int64
		if err := db.Model(&Foto{}).Where("invitado_id = ?", invitado.ID).Count(&total).Error; err != nil {
			t.Fatalf("%s: no se pudo contar las fotos", nombre)
		}
		if total != 1 {
			t.Fatalf("%s: quedaron %d fotos; se esperaba solo la existente", nombre, total)
		}
	}

	fotoGORM := Foto{UUID: existente.UUID, URLArchivo: "/archivos/repetida-gorm.jpg", InvitadoID: invitado.ID}
	if err := NuevoRepositorioFotosGORM(db).GuardarFoto(context.Background(), &fotoGORM); err == nil {
		t.Fatal("GORM debía fallar al insertar el UUID duplicado")
	}
	verificarRollback("GORM")

	fotoSQLC := Foto{UUID: existente.UUID, URLArchivo: "/archivos/repetida-sqlc.jpg", InvitadoID: invitado.ID}
	if err := NuevoRepositorioFotosSQLC(conexion).GuardarFoto(context.Background(), &fotoSQLC); err == nil {
		t.Fatal("SQLC debía fallar al insertar el UUID duplicado")
	}
	verificarRollback("SQLC")
}
