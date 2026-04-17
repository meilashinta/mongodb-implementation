# Sistem Arsitektur Dual-Write Product API

## Pendekatan Arsitektur
Dual-Write Product API menggunakan **Layered Architecture** yang jelas memisahkan tanggung jawab setiap lapisan untuk memudahkan pengembangan, pengujian, dan pemeliharaan.

Lapisan utama:
- `Router`: Menentukan routing HTTP dan menghubungkan endpoint ke handler.
- `Handler`: Menerima request, melakukan validasi awal, dan memanggil service yang sesuai.
- `Service`: Mengandung logika bisnis inti, mengatur alur operasi dual-write dan read strategy.
- `Repository`: Mengelola akses data ke MySQL dan MongoDB.
- `Model`: Mendefinisikan struktur data dan objek domain produk.

## Mekanisme Dual-Write
Setiap operasi penulisan pada API harus mengikuti alur sekuensial berikut:
1. `Handler` menerima permintaan create/update/delete.
2. `Service` memproses logika bisnis dan memanggil repository MySQL.
3. `Repository MySQL` menyelesaikan operasi write pertama sebagai sumber kebenaran.
4. Jika write ke MySQL berhasil, `Service` kemudian memanggil repository MongoDB untuk mereplikasi data.
5. Hanya setelah kedua write berhasil, response positif dikembalikan ke klien.

### Aturan Penulisan
- Create: Insert data produk ke MySQL, lalu replikasi ke MongoDB.
- Update: Update produk di MySQL, lalu sinkronkan perubahan ke MongoDB.
- Delete: Hapus produk di MySQL, lalu hapus entri terkait di MongoDB.
- Jika write MySQL gagal, operasi dibatalkan dan tidak dilanjutkan ke MongoDB.
- Jika write MongoDB gagal setelah MySQL sukses, sistem harus menangani kegagalan dengan logging, notifikasi, atau retry/reconciliation sesuai kebijakan.

## Strategi Baca
Semua operasi baca diarahkan 100% ke MongoDB untuk menjaga performa dan mengurangi beban MySQL:
- `GetAll`: Mengambil daftar produk dari MongoDB.
- `GetByID`: Mengambil detail produk berdasarkan ID dari MongoDB.
- `Search`: Melakukan pencarian dan filtering berdasarkan kategori, rentang harga, dan tags di MongoDB.

Dengan strategi ini, MySQL difokuskan sebagai Sumber Kebenaran (Source of Truth) untuk transaksi dan backup, sedangkan MongoDB difokuskan sebagai read-optimized store untuk etalase.

## Struktur Folder
Proyek Golang harus diorganisir ke package dengan domain tanggung jawab yang terpisah:

- `config`
  - Konfigurasi database MySQL dan MongoDB.
  - Inisialisasi koneksi dan pembacaan environment variables.

- `handlers`
  - HTTP handler untuk endpoint produk.
  - Mengonversi request/response ke bentuk domain.

- `services`
  - Logika bisnis utama dan orkestrasi Dual-Write.
  - Menyesuaikan alur untuk operasi write dan read.

- `repositories`
  - Implementasi akses data untuk MySQL dan MongoDB.
  - Mengandung fungsi CRUD data produk untuk kedua penyimpanan.

- `models`
  - Definisi struct produk dan DTO jika diperlukan.
  - Format data yang digunakan di seluruh lapisan.

## Diagram Alur Ringkas
1. Client -> `Router`
2. `Router` -> `Handler`
3. `Handler` -> `Service`
4. `Service` -> `Repository MySQL` -> `MySQL`
5. `Service` -> `Repository MongoDB` -> `MongoDB`
6. `Service` -> `Handler` -> Response ke client

## Ringkasan
Arsitektur ini memastikan konsistensi data di MySQL, sekaligus menyediakan performa baca cepat lewat MongoDB. Struktur folder yang terpisah memudahkan isolasi logika dan mempercepat pengembangan fitur CRUD, pencarian, serta proses sinkronisasi dual-write.