# Workflow: Create CRUD Dual-Write

Dokumen ini menjelaskan alur standar pengembangan fitur CRUD untuk Dual-Write Product API.

## 1. Modelling
- Buat struct `Product` di `models/product.go`.
- Sertakan tag `json`, `bson`, dan `db` untuk mendukung serialisasi request/response, penyimpanan MongoDB, dan mapping SQL.
- Contoh field utama: `ID`, `Name`, `Description`, `Category`, `Price`, `Stock`, `Tags`, `CreatedAt`, `UpdatedAt`.

## 2. Infrastructure
- Inisialisasi koneksi database di `config/database.go`.
- Konfigurasikan koneksi MySQL dan MongoDB menggunakan environment variable atau file konfigurasi.
- Pastikan connection pooling diaktifkan untuk kedua database.
- Inisialisasi client MongoDB dan connection pool MySQL sekali saat aplikasi startup.

## 3. Repository Logic
- Buat file `repositories/product_repository.go`.
- Implementasikan fungsi untuk melakukan operasi CRUD pada MySQL.
- Untuk setiap operasi write:
  1. Jalankan query SQL `INSERT`, `UPDATE`, atau `DELETE` di MySQL.
  2. Jika operasi MySQL berhasil, lanjutkan dengan operasi `InsertOne`, `UpdateOne`, atau `DeleteOne` di MongoDB.
- Simpan data produk di MySQL sebagai sumber kebenaran utama.
- Replikasi data ke MongoDB hanya setelah MySQL berhasil untuk menjaga konsistensi sekuensial.

## 4. API Exposure
- Hubungkan handler produk ke router di `main.go`.
- Pastikan `Router` memetakan endpoint CRUD ke handler yang tepat.
- Handler akan memanggil service/repository sesuai alur bisnis yang ditetapkan.
- Gunakan middleware yang diperlukan untuk logging, error handling, dan parsing JSON.

## Catatan
- Write path harus selalu `MySQL -> MongoDB`.
- Read path (`GetAll`, `GetByID`, `Search`) harus 100% menggunakan MongoDB.
- Documentasikan keputusan teknis penting, seperti alasan dual-write sekuensial.
