# Dual-Write Product API

## Deskripsi Proyek
Dual-Write Product API adalah sebuah layanan backend Golang yang dirancang untuk manajemen katalog produk e-commerce dengan fokus pada performa pembacaan tinggi sekaligus menjaga konsistensi data relasional. API ini menggunakan pola *Dual-Write* untuk menyimpan data produk secara bersamaan ke:

- **MySQL** sebagai *Source of Truth* untuk data transaksional, audit, dan laporan inventaris.
- **MongoDB** sebagai *Read-Optimized Store* untuk kebutuhan tampilan katalog dan pencarian cepat pada etalase.

## Tujuan Utama
1. Menyediakan API CRUD produk dengan konsistensi data antar kedua penyimpanan.
2. Memastikan operasi **Write** dilakukan sekuensial dan andal ke MySQL lalu MongoDB.
3. Menyediakan operasi **Read** cepat untuk daftar produk, detail produk per ID, dan pencarian/filter produk.
4. Mendukung pencarian produk berdasarkan kategori, rentang harga, dan tags.

## Arsitektur Dual-Write
1. **Write ke MySQL**
   - MySQL menyimpan data produk sebagai sumber kebenaran utama.
   - Semua perubahan transaksi produk direkam di MySQL untuk akuntabilitas dan laporan inventaris.
2. **Write ke MongoDB**
   - Setelah operasi MySQL berhasil, data yang sama di-dual-write ke MongoDB.
   - MongoDB digunakan sebagai store optimasi baca untuk mengurangi latensi pada etalase dan pencarian.
3. **Sinkronisasi Sekuensial**
   - Alur write: `Request -> MySQL -> MongoDB -> Response`
   - Jika penulisan MySQL gagal, operasi dibatalkan.
   - Jika penulisan MongoDB gagal setelah MySQL berhasil, API harus menangani kegagalan secara terkontrol, termasuk logging dan retry jika diperlukan.

## Fitur Utama
### CRUD Produk
- **Create Product**
  - Tambah produk baru ke MySQL dan MongoDB.
  - Data produk mencakup nama, deskripsi, kategori, harga, stok, dan tags.

- **Read Product**
  - **List Produk**: Ambil daftar produk dari MongoDB untuk performa baca optimal.
  - **Detail Produk per ID**: Ambil detail produk berdasarkan `product_id` dari MongoDB.
  - **Search**: Cari produk dengan filter kategori, rentang harga, dan tags.

- **Update Product**
  - Update harga, stok, deskripsi, kategori, atau tags.
  - Update dilakukan terlebih dahulu di MySQL lalu disinkronkan ke MongoDB.

- **Delete Product**
  - Hapus produk dari MySQL dan MongoDB secara sekuensial.
  - Pastikan data terhapus konsisten pada kedua store.

## Model Data Produk
- `id` (UUID atau integer)
- `name` (string)
- `description` (string)
- `category` (string)
- `price` (decimal)
- `stock` (integer)
- `tags` (array string)
- `created_at` (timestamp)
- `updated_at` (timestamp)

## Use Case
1. User menambahkan produk baru ke katalog.
2. User memperbarui harga atau stok produk.
3. User menghapus produk dari katalog.
4. Client e-commerce menampilkan daftar produk dengan latensi rendah.
5. Pengguna mencari produk berdasarkan kategori, harga, dan tags.

## Keuntungan
- **Performa baca tinggi**: MongoDB mendukung query cepat untuk etalase dan pencarian.
- **Kebenaran data**: MySQL berfungsi sebagai sumber data transaksional utama.
- **Kemudahan skalabilitas**: Pembacaan dioptimalkan tanpa mengorbankan integritas data.

## Pertimbangan Implementasi
- Tangani rollback atau retry jika penulisan MongoDB gagal setelah MySQL sukses.
- Pastikan enum kategori dan struktur tags konsisten di kedua store.
- Gunakan koneksi database yang aman dan parametrized query untuk mencegah SQL injection.
- Pertimbangkan strategi cleanup atau reconciliation bila sinkronisasi dual-write gagal.

## Catatan
- API ini cocok untuk sistem katalog e-commerce di mana latensi baca rendah sangat penting, namun data inventaris dan transaksi tetap memerlukan database relasional yang konsisten.
- Pendekatan dual-write harus disertai monitoring untuk mendeteksi inkonsistensi antara MySQL dan MongoDB.
