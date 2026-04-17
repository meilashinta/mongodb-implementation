# Dual-Write Product API

## Ringkasan
Dual-Write Product API adalah backend Golang untuk katalog produk e-commerce. Aplikasi ini menerapkan pola dual-write:
- **MySQL** sebagai *source of truth* untuk transaksi dan data utama.
- **MongoDB** sebagai *read-optimized store* untuk query baca dan pencarian.

## Arsitektur
Struktur utama aplikasi:
- `config/` : koneksi database
- `handlers/` : HTTP handler
- `services/` : logika bisnis
- `repositories/` : akses data MySQL dan MongoDB
- `models/` : definisi produk dan filter

## Alur Dual-Write
| Operasi | MySQL | MongoDB | Keterangan |
|---|---|---|---|
| Create | `INSERT INTO products` | `InsertOne` | MySQL dulu, lalu MongoDB |
| Read All | - | `Find({})` | Semua baca dari MongoDB |
| Read By ID | - | `FindOne({_id: id})` | MongoDB sebagai sumber baca |
| Search | - | `Find(query)` | Filter kategori, harga, tags |
| Update | `UPDATE products` | `UpdateOne({_id: id})` | MySQL dulu, lalu MongoDB |
| Delete | `DELETE FROM products` | `DeleteOne({_id: id})` | Hapus MySQL lalu MongoDB |

## Skema Database
### MySQL - tabel `products`
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
> Catatan: `tags` disimpan sebagai string koma di MySQL.

### MongoDB - dokumen `products`
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

### GET `/`
Menampilkan status API:
- `message`
- `endpoints`

### POST `/product`
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

### GET `/products`
Ambil semua produk. Mendukung query:
- `category`
- `min_price`
- `max_price`
- `tags` (dipisah koma)

Contoh:
- `/products?category=Fashion`
- `/products?min_price=100000&max_price=1000000`
- `/products?tags=sport,diskon`

### GET `/product/{id}`
Ambil detail produk berdasarkan ID.

### PUT `/product/{id}`
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

### DELETE `/product/{id}`
Hapus produk berdasarkan ID.

## Contoh Testing
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

## Mulai dari GitHub
Clone repository:
```bash
git clone https://github.com/meilashinta/mongodb-implementation.git
cd mongodb-implementation
```

## Setup Lokal
### Persyaratan
- Go 1.22+
- Docker dan Docker Compose
- Port `8082` tersedia untuk API
- Port `8083` tersedia untuk Adminer

### Jalankan dengan Docker Compose
1. Buka terminal di folder proyek:
   ```bash
   cd "d:\SEM 6\Topik Khusus\matkul-topik-khusus-3"
   ```
2. Jalankan layanan:
   ```bash
   docker-compose up -d
   ```
3. Pastikan container `dualwrite_mysql`, `dualwrite_mongodb`, dan `dualwrite_adminer` berjalan.

### Login Adminer
Buka:
- `http://localhost:8083/`

Gunakan:
- System: `MySQL`
- Server: `mysql`
- Username: `appuser`
- Password: `password`
- Database: `dualwrite`

> `Server = mysql` karena Adminer berada di jaringan Docker Compose yang sama dengan MySQL.

## Jalankan Aplikasi Go
1. Pastikan Docker Compose sudah hidup:
   ```bash
   docker-compose up -d
   ```
2. Jalankan aplikasi:
   ```bash
   go mod tidy
   go run main.go
   ```
3. Akses API:
   - `http://localhost:8082/`
   - `http://localhost:8082/products`

### Hentikan Docker
```bash
docker-compose down
```

## Konfigurasi Opsional
Jika ingin jalankan Go dari host tanpa Docker untuk database lokal:
```bash
export MYSQL_DSN="appuser:password@tcp(localhost:3306)/dualwrite?parseTime=true"
export MYSQL_MAX_OPEN_CONNS=25
export MYSQL_MAX_IDLE_CONNS=10
export MONGODB_URI="mongodb://localhost:27017"
export MONGODB_DATABASE="dualwrite"
export MONGODB_COLLECTION="products"
export MONGODB_MAX_POOL_SIZE=100
```

## Catatan Tambahan
- Semua baca produk menggunakan MongoDB.
- Semua tulis akan menulis ke MySQL lalu direplikasi ke MongoDB.
- Jika MongoDB gagal setelah MySQL sukses, log akan mencatat inkonsistensi.
