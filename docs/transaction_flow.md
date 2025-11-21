FLOW BARU (REAL BLOCKCHAIN STYLE)

KETIKA SEND() DIPANGGIL:
-Validasi privateKey
-Validasi balance sender
-Buat transaksi STATUS = PENDING
-TIDAK mengubah balance
-Ledger belum dibuat
-Transaksi masuk mempool (pending list)

KETIKA BLOCK DIGENERATE:
-Ambil semua transaksi PENDING
-Urutkan sesuai timestamp
-Validasi ulang saldo
-Buat ledger (sender -amount, receiver +amount)
-Update balance
-Tandai transaksi: CONFIRMED
-Buat block baru & hitung hash
-Simpan block dan block_transactions
