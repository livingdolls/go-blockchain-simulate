package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDomainConstants(t *testing.T) {
	// Konstanta domain dipakai di banyak tempat (SQL, business logic, label).
	// Perubahan nilai tanpa disengaja bisa menyebabkan bug sulit dilacak.
	// Test ini mencegah regresi.
	assert.Equal(t, "MINER_ACCOUNT", MinerAccountAddress,
		"MinerAccountAddress harus string literal 'MINER_ACCOUNT' (dipakai di SQL query)")

	assert.Equal(t, "admin", AdminRoleSuper,
		"AdminRoleSuper harus 'admin' (role super admin dengan akses penuh)")

	assert.Equal(t, "*", AdminPermissionWildcard,
		"AdminPermissionWildcard harus '*' (penanda akses tak terbatas)")

	assert.Equal(t, "active", AdminStatusActive,
		"AdminStatusActive harus 'active' (nilai enum di kolom status)")
}
