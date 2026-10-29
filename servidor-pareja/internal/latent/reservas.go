package latent

import (
	"errors"

	"gorm.io/gorm"
)

// Errores del negocio. Antes de pegar, comprueba que no existan ya con:
//   grep -rn "ErrLimiteFotos\|ErrEventoNoActivo\|ErrNoEncontrado" internal/
var (
	ErrNoEncontrado   = errors.New("no encontrado")
	ErrEventoNoActivo = errors.New("el evento no está activo")
	ErrLimiteFotos    = errors.New("se alcanzó el límite de fotos")
)

// reservarFoto suma 1 al contador del invitado solo si el evento está activo
// y todavía tiene cupo. La condición se evalúa dentro de la base, en una sola
// operación, así que dos subidas simultáneas no pueden pasarse del límite.
func reservarFoto(tx *gorm.DB, invitadoID uint) error {
	res := tx.Exec(`
		UPDATE invitados
		SET fotos_tomadas = fotos_tomadas + 1
		WHERE id = ?
		  AND fotos_tomadas < (
		      SELECT e.disparo_por_invitado FROM eventos e
		      WHERE e.id = invitados.evento_id AND e.estado = ?)`,
		invitadoID, EstadoActivo)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 1 {
		return nil
	}

	// 0 filas: averiguamos por qué para devolver el error correcto.
	var inv struct {
		FotosTomadas       int
		DisparoPorInvitado int
		Estado             string
	}
	err := tx.Raw(`
		SELECT i.fotos_tomadas, e.disparo_por_invitado, e.estado
		FROM invitados i JOIN eventos e ON e.id = i.evento_id
		WHERE i.id = ?`, invitadoID).Scan(&inv).Error
	if err != nil {
		return err
	}
	switch {
	case inv.Estado == "": // el JOIN no devolvió nada
		return ErrNoEncontrado
	case inv.Estado != EstadoActivo:
		return ErrEventoNoActivo
	default:
		return ErrLimiteFotos
	}
}

// guardarFoto reserva el cupo y crea la foto en la MISMA transacción:
// si algo falla (por ejemplo, UUID repetido), el contador no queda sumado.
func guardarFoto(db *gorm.DB, f *Foto) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := reservarFoto(tx, f.InvitadoID); err != nil {
			return err
		}
		return tx.Create(f).Error
	})
}