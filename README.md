# Dual-Write Product API

## Ringkasan
Dual-Write Product API adalah layanan backend Golang untuk manajemen katalog produk e-commerce. Aplikasi ini menggunakan pola Dual-Write untuk menyimpan data ke:
- **MySQL** sebagai *Source of Truth* untuk transaksi dan laporan inventaris.
- **MongoDB** sebagai *Read-Optimized Store* untuk performa baca tinggi pada etalase dan pencarian.

## Arsitektur
Proyek disusun dalam lapisan terpisah:
- `config` : inisialisasi koneksi database
- `handlers` : HTTP handler dan endpoint
- `services` : logika bisnis
- `repositories` : akses data MySQL dan MongoDB
- `models` : definisi struct dan filter produk

## Dual-Write Table
| Operasi | MySQL | MongoDB | Catatan |
|---|---|---|---|
| Create | `INSERT INTO products` | `InsertOne` | MySQL ditulis terlebih dahulu, lalu MongoDB direplikasi |
| Read All | - | `Find({})` | Semua baca disajikan dari MongoDB |
| Read By ID | - | `FindOne({_id: id})` | MongoDB jadi read store 100% |
| Search | - | `Find(query)` | Filter kategori, rentang harga, dan tags |
| Update | `UPDATE products` | `UpdateOne({_id: id})` | Jika MySQL sukses, baru update MongoDB |
| Delete | `DELETE FROM products` | `DeleteOne({_id: id})` | Hapus berurutan MySQL lalu MongoDB |

## Struktur Database

### MySQL DDL: tabel `products`
```sql
CREATE TABLE IF NOT EXISTS products (
  id VARCHAR(64) PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  category VARCHAR(128),
  price DOUBLE NOT NULL,
  stock INT NOT NULL,
  tags TEXT,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);
```
> Implementasi menyimpan `tags` sebagai string terpisah dengan koma di MySQL, sedangkan MongoDB menyimpan `tags` sebagai array.

### MongoDB Skema Dokumen
Contoh dokumen `products` di MongoDB:
```json
{
  "_id": "prod-1680000000000000000",
  "name": "Nama Produk",
  "description": "Deskripsi produk...",
  "category": "Elektronik",
  "price": 199.99,
  "stock": 100,
  "tags": ["promo", "baru"],
  "created_at": "2026-04-17T10:00:00Z",
  "updated_at": "2026-04-17T10:00:00Z"
}
```

## Endpoint API
Base URL: `http://localhost:8082`

Akses `http://localhost:8082/` akan menampilkan status API dan endpoint yang tersedia.

### 1. `POST /product`
Tambah produk baru.

Body JSON:
```json
{
  "name": "Sepatu Olahraga",
  "description": "Sepatu lari nyaman",
  "category": "Fashion",
  "price": 749000,
  "stock": 50,
  "tags": ["sport", "diskon"]
}
```

### 2. `GET /products`
Ambil daftar produk. Mendukung filter query:
- `category`
- `min_price`
- `max_price`
- `tags` (dipisah koma)

Contoh query:
- `/products?category=Fashion`
- `/products?min_price=100000&max_price=1000000`
- `/products?tags=sport,diskon`

### 3. `GET /product/{id}`
Ambil detail produk berdasarkan ID.

### 4. `PUT /product/{id}`
Update produk.

Body JSON:
```json
{
  "name": "Sepatu Olahraga Premium",
  "description": "Sepatu lari nyaman, lebih ringan",
  "category": "Fashion",
  "price": 799000,
  "stock": 45,
  "tags": ["sport", "premium"]
}
```

### 5. `DELETE /product/{id}`
Hapus produk berdasarkan ID.

## Contoh cURL Testing
### Tambah produk
```bash
curl -X POST http://localhost:8082/product \
  -H "Content-Type: application/json" \
  -d '{"name":"Sepatu Olahraga","description":"Sepatu lari nyaman","category":"Fashion","price":749000,"stock":50,"tags":["sport","diskon"]}'
```

### Ambil semua produk
```bash
curl http://localhost:8082/products
```

### Ambil produk berdasarkan ID
```bash
curl http://localhost:8082/product/prod-1680000000000000000
```

### Update produk
```bash
curl -X PUT http://localhost:8082/product/prod-1680000000000000000 \
  -H "Content-Type: application/json" \
  -d '{"name":"Sepatu Olahraga Premium","description":"Sepatu lari nyaman, lebih ringan","category":"Fashion","price":799000,"stock":45,"tags":["sport","premium"]}'
```

### Hapus produk
```bash
curl -X DELETE http://localhost:8082/product/prod-1680000000000000000
```

### Search filter kategori
```bash
curl "http://localhost:8082/products?category=Fashion&min_price=500000&max_price=1000000&tags=sport,diskon"
```

## Setup Lokal
### Persyaratan
- Go 1.22+
- MySQL berjalan dengan database `dualwrite`
- MongoDB berjalan lokal pada `mongodb://localhost:27017`

### Konfigurasi environment
Anda dapat mengatur environment variables:
```bash
export MYSQL_DSN="user:password@tcp(localhost:3306)/dualwrite?parseTime=true"
export MYSQL_MAX_OPEN_CONNS=25
export MYSQL_MAX_IDLE_CONNS=10
export MONGODB_URI="mongodb://localhost:27017"
export MONGODB_DATABASE="dualwrite"
export MONGODB_COLLECTION="products"
export MONGODB_MAX_POOL_SIZE=100
```

### Menjalankan aplikasi
```bash
go run main.go
```

## MCP Setup (VS Code Copilot)
1. Pastikan VS Code sudah terhubung ke extension GitHub Copilot Chat.
2. Buka folder proyek `d:\SEM 6\Topik Khusus\matkul-topik-khusus-3` di VS Code.
3. Pastikan database lokal berjalan dan environment variables diatur.
4. Jika extension mendukung Model Context Protocol (MCP), gunakan konfigurasi lokal atau `launch.json` untuk mengakses `MYSQL_DSN` dan `MONGODB_URI`.
5. Contoh pengaturan MCP: gunakan endpoint/mode lokal yang membuka koneksi ke MySQL dan MongoDB saat bekerja dengan Copilot.

> Catatan: MCP memungkinkan agent melihat konteks basis data lokal untuk membantu generasi query dan debugging. Pastikan akses jaringan dan hak baca/tulis database disetujui.

## MongoDB Compass Query
### Filter kategori
```js
{ "category": "Fashion" }
```

### Search harga dan tags
```js
{
  "category": "Fashion",
  "price": { "$gte": 500000, "$lte": 1000000 },
  "tags": { "$all": ["sport", "diskon"] }
}
```

### Sorting harga produk naik
```js
{}
```

Gunakan fitur sort di Compass untuk `price` ascending atau descending.

## Catatan Tambahan
- Semua operasi baca menggunakan MongoDB untuk mengurangi beban MySQL.
- Operasi write mengikuti alur `MySQL -> MongoDB` untuk menjaga konsistensi data.
- Error write MongoDB setelah MySQL sukses dilog spesifik agar inkonsistensi dapat dilacak.
