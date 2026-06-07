package entity

// MinerAccountAddress adalah address internal sistem yang digunakan sebagai
// penanda bahwa transaksi tersebut dihasilkan oleh sistem itu sendiri, bukan
// oleh user. Contoh penggunaan: ketika user membeli YTE dari sistem, dari
// address sistem = MinerAccountAddress; ketika user menjual YTE ke sistem,
// ke address sistem = MinerAccountAddress.
//
// Konstanta ini menjamin konsistensi penamaan address sistem di seluruh
// lapisan (services, repository, worker) dan mencegah typo literal.
const MinerAccountAddress = "MINER_ACCOUNT"

// AdminRoleSuper adalah role admin dengan akses penuh. Admin dengan role ini
// otomatis lolos pengecekan permission di hasPermission().
const AdminRoleSuper = "admin"

// AdminPermissionWildcard adalah permission yang menandakan akses tak terbatas.
const AdminPermissionWildcard = "*"

// AdminStatusActive adalah status admin yang diizinkan masuk ke sistem.
const AdminStatusActive = "active"
