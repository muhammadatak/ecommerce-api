# Review: go-ecommerce-api

Dicek terhadap commit `d837536` ditambah perubahan yang belum di-commit (checkout + transaksi). `go build` dan `go vet` lolos. Checkout juga sudah diuji langsung lewat API (lihat baris "Checkout: transaksi").

## ✅ Sudah selesai

| Area | Yang sudah benar |
|---|---|
| Struktur | domain → repository → service → api, dependency diinject lewat interface. `CartService` memanggil `ProductRepository`, bukan `ProductService`. |
| Auth | Login dan register. Email dinormalisasi (`ToLower` + `TrimSpace`) di keduanya. Pesan gagal login sama untuk email dan password (`ErrAuth`), jadi tidak bisa dipakai menebak email terdaftar. Password `min=8,max=72` (batas bcrypt). |
| JWT | `JWT_EXP`/`JWT_KEY` fail-fast di `loader.go`. `getUserId` di `api/helper.go`. `UserId` di DTO `json:"-"`, jadi tidak bisa dipalsukan client. Gagal sign → 500. |
| Cart: lihat | Cart kosong → `[]`. Produk diambil sekali dengan `IN (...)` (tanpa N+1), dan return lebih awal kalau cart tidak punya item. |
| Cart: tambah item | Cart dibuat otomatis kalau belum ada. Upsert dengan `ON CONFLICT (cart_id, product_id)`. Cek produk ada, `ACTIVE`, dan stok kumulatif (`req.Quantity + item.Quantity`). |
| Cart: ubah/hapus item | Authorization lewat `FindByIdAndCartId(id, cart.Id)`, jadi item milik orang lain → 404. User tanpa cart → 404 di PATCH dan DELETE. PATCH memvalidasi `quantity` `gt=0`, mengecek stok dan status produk, lalu hanya menulis `quantity` + `updated_at`. |
| Checkout: skema | `order_items` menyimpan salinan `product_name` dan `price`, jadi order lama tidak berubah kalau produk diedit. CHECK constraint untuk status, harga, dan quantity. Index `orders.user_id` siap untuk riwayat order. `product_id ON DELETE RESTRICT` mencegah produk yang sudah pernah dipesan terhapus. |
| Checkout: alur | Cart diambil dari user di token, bukan dari input. Stok dan status dicek per produk, stok dikurangi, item cart dihapus, total dihitung dari harga saat checkout. Response 201. |
| Checkout: transaksi | `domain.TxRepository` diinject ke `orderService` lewat `repository.NewTransactor`, `RunInTransaction` sudah dihapus, jadi lupa inject = compile error. Semua repository memakai `getExecutor(ctx, x.db)`. Diuji sebelumnya: rollback saat stok produk kedua kurang, dan 5 checkout bersamaan untuk stok 1 → tepat satu 201. `FindByIds` sekarang `ORDER BY id` sebelum `FOR UPDATE`. |
| Checkout: alur baru | Cart item dipetakan sekali ke `map[productId]` (tanpa N+1). Total dihitung dulu, order disimpan sekali dengan total yang benar, order items disimpan satu query, cart dikosongkan sekali. `len(products) != len(items)` → 404. Cart kosong → `ErrEmptyCartItem` → 400. Sisa debug `fmt.Print*` sudah dihapus. |
| Error → HTTP | Sentinel error di `domain/errors.go`, dipetakan di `errorStatus`: not found 404, auth/token 401, stok/inactive 400, email duplikat 409. Body JSON rusak → 400. |
| Validasi | `TranslateTag` punya pesan untuk `required`, `min`, `max`, `gt`, `uuid`, `email`. |
| Lain-lain | `GET /products` kosong → `[]`, `log.Fatal(app.Listen(...))`. |

## ⏳ Masih terbuka

### 🔴 1. Checkout baru selalu gagal, dan cart tidak pernah dikosongkan

Dicek dengan `ToSQL()` goqu (DB lokal sedang mati, jadi belum diuji lewat API):

- **`SaveBatch` → selalu error, jadi setiap checkout 500.** `Rows(items)` dengan `items *[]domain.OrderItem` ditolak goqu: `unsupported insert must be map, goqu.Record, or struct type got: *[]domain.OrderItem`. Transaksi di-rollback, jadi data aman, tapi checkout tidak pernah berhasil. Ubah parameternya jadi `items []domain.OrderItem` (slice sudah berupa referensi, pointer ke slice tidak perlu). Dengan slice biasa, goqu menghasilkan satu `INSERT ... VALUES (...), (...)`.
- **`DeleteByCartId` memfilter kolom yang salah.** `Where(goqu.C("id").Eq(cartId))` menghasilkan `DELETE FROM cart_items WHERE id = '<cart id>'`. Tidak ada cart item yang id-nya sama dengan id cart, jadi tidak ada yang terhapus dan tidak ada error. Setelah bug `SaveBatch` diperbaiki, checkout akan sukses tapi cart tetap berisi, dan user bisa checkout ulang cart yang sama. Ganti jadi `goqu.C("cart_id")`.

### 🟡 2. Checkout: sisa masalah

- **`POST /orders` masih wajib punya body.** Field `cart_id` sudah dihapus, tapi handler masih memanggil `Bind().Body` dan `Validate`. Tanpa body → 400 `unexpected end of JSON input`. Hapus keduanya dari `orderApi.Checkout`.
- **Kode mati dari cara lama.** `orderRepository.WithTransaction` tidak dipakai lagi. `OrderRepository.Update` dan `OrderItemRepository.Save`/`Delete` juga bisa dihapus. Komentar `// WithTransaction implements [domain.Transactor]` di `repository/tx.go` sudah tidak cocok dengan nama `TxRepository`. `OrderRepository.FindById` masih `panic("unimplemented")`.
- **Update stok menulis semua kolom.** `productRepository.Update` memakai `Set(product)`. Untuk checkout cukup `UPDATE products SET stock = stock - ? WHERE id = ? AND stock >= ?`, lalu `RowsAffected() == 0` → `ErrInsufficientStock`.
- **`FOR UPDATE` masih ada di `FindByIds`** yang juga dipakai lihat cart, jadi `GET /carts` ikut mengunci produk (singkat, karena di luar transaksi langsung autocommit, tapi tetap menunggu kalau ada checkout yang sedang memegang kunci). Pisahkan jadi `FindByIdsForUpdate`.
- **Response 201 tanpa body.** Client tidak tahu order mana yang dibuat. Kembalikan minimal `id`, `status`, `total_price`.
- Status produk dicek setelah stok. Cek status dulu.

### 🟡 3. `:id` yang bukan UUID

`productApi.Show` sekarang membalas 400 JSON. Empat tempat lain (`UpdateItemQuantity`, `DeleteItem`, `productApi.Update`, `Delete`) masih `return err` → 500 teks. Samakan semuanya; 404 lebih umum dipakai untuk id yang tidak valid di path (resource dengan id itu memang tidak ada):

```go
id, err := uuid.Parse(ctx.Params("id"))
if err != nil {
    return ctx.Status(http.StatusNotFound).JSON(dto.CreateResponseError(domain.ErrNotFound.Error()))
}
```

Karena polanya sama di lima tempat, cocok dijadikan helper `parseId(ctx)` di `api/helper.go`.

### 🟡 4. Register bersamaan → 500

Dua request dengan email yang sama bisa sama-sama lolos `FindByEmail`. Yang kedua ditolak `UNIQUE` di DB, lalu jadi 500. Tangkap error unique violation (kode Postgres `23505`) di `userRepository.Save` dan kembalikan `ErrDupeEmail`. Kecil: `strings.EqualFold(user.Email, normalizedEmail)` cukup diganti `user.Id != ""`, karena kalau baris ketemu, emailnya pasti sama.

### 🟡 5. Error DB bocor ke client

Untuk 500, `err.Error()` dari Postgres (nama tabel, constraint) tetap dikirim apa adanya. Di cabang `default` pada `errorStatus`, log error aslinya dan kirim pesan generik seperti `"internal server error"`.

### 🟡 6. Validasi product belum ada (`dto/product_data.go`)

`CreateProductReq` dan `UpdateProductReq` tidak punya tag `validate`, dan `util.Validate` tidak dipanggil di `api/product.go`. `stock: -5` atau `status: "apaaja"` menjadi 500 (kena CHECK constraint DB) atau malah tersimpan. Contoh: `validate:"required,gte=0"` untuk stok, `validate:"omitempty,oneof=ACTIVE INACTIVE"` untuk status.

### 🟡 7. Kebersihan kode

- `util.Validate`: `validator.New()` dibuat ulang setiap panggilan, jadikan variabel package-level. `err.(validator.ValidationErrors)` bisa panic kalau errornya `InvalidValidationError`, pakai `errors.As`. `case "unique"` di `TranslateTag` tidak dipakai tag mana pun.
- Dead code: `service/user.go`, `dto/user_data.go`, `UserService` di `domain/user.go`, dan `halo.md` yang kosong.
- `authRepos` di `main.go` sebenarnya berisi service.
- Komentar nyasar: `// DeleteItem implements ...` di atas `NewCart`, `// FindByCartId implements ...` di atas `NewCartItem`.
- `godotenv.Load()` memanggil `log.Fatal` kalau `.env` tidak ada. Di production konfigurasi biasanya hanya dari env var, jadi cukup log warning.

### Nit

- Register membuat resource baru, jadi 201 lebih tepat daripada 200.
- Di PATCH, status produk dicek setelah stok. Produk inactive yang stoknya kurang dibalas "Insufficient stock". Cek status dulu, seperti di `AddItem`.
- Route `/carts/item` → `/cart/items` (satu user satu cart, dan resource-nya kumpulan item).
- Konstanta `ProductStockInactive`/`ProductStockActive` padahal isinya status, lebih jelas `ProductStatus...`.
- `GET /products` biasanya publik, tapi sekarang di belakang JWT.

## Langkah berikutnya

Urutan: bug checkout (1) → sisa checkout (2) → `:id` UUID (3) → validasi product (6) → riwayat order → pagination → tes → role admin dan upload gambar.

### Fitur inti

1. **Checkout dan order.** Skema, alur, dan transaksi sudah ada. Tinggal dua bug di 1 lalu merapikan (lihat 2).
2. **Riwayat order.** `GET /orders` dan `GET /orders/:id`, hanya order milik user tersebut. Pola authorization-nya sama dengan cart item: filter `user_id` sejak query pertama. `GET /orders/:id` sekalian mengembalikan order items.
3. **Role admin.** Sekarang user mana pun yang login bisa membuat dan menghapus produk.
4. **Upload gambar produk.** `.env` sudah punya `STORAGE_PATH` dan `SERVER_ASSET_URL`, folder `storage/` juga ada, tapi belum dipakai.

### Kualitas API

5. **Pagination dan filter** di `GET /products` (`?page=&limit=&q=&status=`). Sekarang seluruh tabel diambil.
6. **Middleware Fiber**: `recover` (supaya panic tidak mematikan server), `logger`, `cors`.
7. **Graceful shutdown**: tangkap SIGINT/SIGTERM, lalu `app.Shutdown()` dan `db.Close()`.
8. **Connection pool**: `db.SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`.

### Engineering

9. **Tes.** Belum ada file `_test.go`. Mulai dari unit test `orderService.Checkout` dan `cartService` dengan mock repository: cart kosong, stok kurang di produk kedua (pastikan rollback), produk inactive, item milik user lain.
10. **README**: setup `.env`, migration goose, menjalankan server, daftar endpoint. Lebih bagus lagi dengan file `.http` atau koleksi Bruno.
11. **Seed data** user dan produk untuk development.
12. **Docker Compose** untuk Postgres.
