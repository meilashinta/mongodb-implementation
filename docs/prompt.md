
### **Prompt 1: Product Definition (docs/product.md)**
**Tugas:** Buatlah dokumen `docs/product.md` untuk proyek 'Dual-Write Product API' menggunakan Golang.
**Latar Belakang:** API ini dirancang untuk manajemen katalog produk e-commerce yang membutuhkan performa pembacaan (read) yang sangat cepat untuk etalase, namun tetap menjaga konsistensi data relasional untuk laporan inventaris.
**Tujuan Utama:** Implementasi pola *Dual-Write* di mana data produk disimpan secara bersamaan ke MongoDB (sebagai *Read-Optimized Store* untuk aplikasi client) dan MySQL (sebagai *Source of Truth* untuk data transaksional/backup).
**Fitur Utama (CRUD):**
* **Write:** Tambah produk, Update stok/harga, dan Hapus produk (Sinkronisasi sekuensial ke MySQL & MongoDB).
* **Read:** List produk, detail produk per ID, dan fitur pencarian produk.
* **Search:** Filter produk berdasarkan kategori, rentang harga, dan tags.

---

### **Prompt 2: System Architecture (docs/architecture.md)**
**Tugas:** Berdasarkan `docs/product.md`, buat dokumen `docs/architecture.md` dengan aturan:
1.  **Arsitektur:** Menggunakan *Layered Architecture*: `Router` -> `Handler` -> `Service` -> `Repository` -> `Model`.
2.  **Mekanisme Dual-Write:** Setiap operasi penulisan (Create, Update, Delete) wajib berhasil di MySQL terlebih dahulu, baru kemudian direplikasi ke MongoDB.
3.  **Read Strategy:** Operasi `GetAll`, `GetByID`, dan `Search` harus diarahkan 100% ke MongoDB untuk mengurangi beban MySQL.
4.  **Struktur Folder:** Pisahkan package `config`, `handlers`, `services`, `repositories`, dan `models`.

---

### **Prompt 3: Agent Personality & Technical Rules (agent/agent-rules.md)**
**Tugas:** Buat file `agent/agent-rules.md`. Anda adalah Senior Golang Developer.
**Aturan Wajib:**
1.  **Clean Code:** Gunakan penamaan variabel yang deskriptif dan fungsi yang ramping.
2.  **Error Handling:** Jika penulisan ke MySQL berhasil tetapi MongoDB gagal, log error secara spesifik (karena ini menyebabkan data tidak sinkron).
3.  **Efisiensi:** Gunakan *connection pooling* untuk MySQL dan MongoDB.
4.  **Documentation:** Catat keputusan teknis penting di komentar kode (misal: alasan penggunaan `Transaction` di SQL).

---

### **Prompt 4: Development Workflow (workflows/create-crud-dualwrite.md)**
**Tugas:** Buat file `workflows/create-crud-dualwrite.md` yang memuat alur standar:
1.  **Modelling:** Buat struct `Product` di `models/product.go` dengan tag JSON, BSON, dan DB.
2.  **Infrastructure:** Inisialisasi koneksi di `config/database.go`.
3.  **Repository Logic:** Di `repositories/product_repository.go`, buat logika: jalankan query SQL `INSERT/UPDATE/DELETE`, jika tidak ada error, lanjutkan ke `InsertOne/UpdateOne/DeleteOne` di MongoDB.
4.  **API Exposure:** Hubungkan handler ke router di `main.go`.

---

### **Prompt 5: Implementation Execution**
**Tugas:** Buat kerangka proyek dan implementasikan CRUD untuk entitas **'Product'**. 
**Konteks:** Ikuti instruksi di `docs/`, `agent/`, dan `workflows/`.
**Instruksi Eksekusi:** Buat REST API endpoint (`POST /product`, `GET /products`, `GET /product/{id}`, `PUT /product/{id}`, `DELETE /product/{id}`) menggunakan Golang. Pastikan logika Dual-Write di layer Repository terimplementasi dengan komentar penjelasan di setiap langkahnya.

---

### **Prompt 6: Technical Documentation (README.md)**
**Tugas:** Buat file `README.md` yang komprehensif untuk proyek ini.
**Instruksi Eksekusi:**
1.  **Dual-Write Table:** Tabel pemetaan operasi DB (MySQL vs MongoDB).
2.  **Endpoints:** Daftar lengkap API Endpoints beserta contoh body JSON-nya.
3.  **Database Schema:** DDL untuk tabel `products` di MySQL dan skema dokumen di MongoDB.
4.  **Testing cURL:** Berikan list command cURL untuk menguji setiap fitur (termasuk filter pencarian produk).
5.  **MCP Setup:** Panduan menghubungkan VS Code Copilot ke DB lokal menggunakan MCP (Model Context Protocol).
6.  **MongoDB Compass:** Contoh query untuk filter kategori dan sorting harga produk.

---

### **Prompt 7: Environment Setup (docker-compose.yml)**
**Tugas:** Bantu saya membuat file `docker-compose.yml` untuk menjalankan environment proyek ini.
**Spesifikasi:** Sertakan container untuk MySQL 8, MongoDB 6, dan satu container Adminer untuk memantau kedua database tersebut.

---

### **Prompt 8: API Testing (api.http)**
**Tugas:** Buat file `api_test.http` untuk testing.
**Isi:** Buat permintaan tes untuk:
1.  Tambah Produk baru (Electronic).
2.  Ambil Semua Produk.
3.  Update Harga Produk berdasarkan ID.
4.  Cari Produk dengan kategori "Fashion" dan rentang harga tertentu.
5.  Hapus Produk.
