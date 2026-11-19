# Review: go-ecommerce-api

Dicek terhadap commit `d837536` ditambah perubahan yang belum di-commit (checkout + transaksi). `go build` dan `go vet` lolos. Checkout diuji end-to-end lewat API dan dicek langsung di tabel `orders`/`order_items`.

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
| Checkout: alur | Cart diambil dari user di token, bukan dari input. `POST /orders` tanpa body. Cart item dipetakan sekali ke `map[productId]` (tanpa N+1). Total dihitung dulu, order disimpan sekali dengan total yang benar, order items disimpan satu query (`SaveBatch`), cart dikosongkan sekali (`DeleteByCartId`). `len(products) != len(items)` → 404. Cart kosong → `ErrEmptyCartItem` → 400. Response 201 berisi `id`, `status`, `total_price`. |
| Checkout: transaksi | `domain.TxRepository` diinject ke `orderService` lewat `repository.NewTransactor`, `RunInTransaction` sudah dihapus, jadi lupa inject = compile error. Semua repository memakai `getExecutor(ctx, x.db)`. `FindByIds` selalu `ORDER BY id`, dan hanya menambahkan `FOR UPDATE` kalau dipanggil di dalam transaksi, jadi `GET /carts` tidak mengunci produk. Pengurangan stok lewat `UpdateStock`, yang hanya menulis kolom `stock`; aman memakai nilai absolut karena barisnya sudah dikunci di transaksi yang sama. |
| Checkout: diuji | (a) Produk kedua stoknya kurang → 400; stok produk pertama tetap, cart tetap 2 item, tidak ada baris `order_items` (rollback). (b) Sukses → 201 dengan body `{id, status: PENDING, total_price: 5000}`; id itu ada di tabel `orders` dengan 2 order item, stok berkurang, cart kosong. (c) Cart kosong → 400. (d) 5 user bersamaan, stok 1 → satu 201, empat 400, stok 0, satu order. (e) 6 user bersamaan dengan dua produk yang sama, urutan masuk cart dibalik → enam 201 tanpa deadlock, stok 100 → 94. (a), (b), (d) diulang setelah `FOR UPDATE` dibuat kondisional: hasil sama. |
| Error → HTTP | Sentinel error di `domain/errors.go`, dipetakan di `errorStatus`: not found 404, auth/token 401, stok/inactive 400, email duplikat 409. Body JSON rusak → 400. |
| Validasi | `TranslateTag` punya pesan untuk `required`, `min`, `max`, `gt`, `uuid`, `email`. |
| Lain-lain | `GET /products` kosong → `[]`, `log.Fatal(app.Listen(...))`. |

## ⏳ Masih terbuka

### 🟡 1. Checkout: sisa masalah

- **Kode mati dari cara lama.** `orderRepository.WithTransaction` tidak dipakai lagi. `OrderRepository.Update` dan `OrderItemRepository.Save`/`Delete` juga bisa dihapus. Komentar `// WithTransaction implements [domain.Transactor]` di `repository/tx.go` sudah tidak cocok dengan nama `TxRepository`. `OrderRepository.FindById` masih `panic("unimplemented")`.
- **Pengaman stok di level DB (opsional).** `UpdateStock` menulis `stock = <nilai baru>`. Selama checkout selalu lewat `FindByIds` di dalam transaksi, ini benar. Kalau mau pengaman kedua yang tidak bergantung pada kunci: `UPDATE products SET stock = stock - ? WHERE id = ? AND stock >= ?`, lalu `RowsAffected() == 0` → `ErrInsufficientStock`. Kecil: `UpdateStock` cukup menerima `(id string, stock int)`, tidak perlu seluruh `*Product`.
- `dto.CheckoutResponses` → `CheckoutResponse` (satu objek).
- Status produk dicek setelah stok. Cek status dulu.

### 🟡 2. `:id` yang bukan UUID

Tiga handler product (`Show`, `Update`, `Delete`) membalas 400 JSON. Dua handler cart sekarang membalas JSON, tapi lewat `errorStatus(err)`: error dari `uuid.Parse` tidak dikenal, jadi jatuh ke `default`. Diuji: `PATCH /carts/item/abc` dan `DELETE /carts/item/abc` → **500** `{"message": "invalid UUID length: 3"}`. Untuk input client yang salah, 500 berarti "server rusak", jadi tetap perlu diganti. Samakan kelimanya. 404 lebih umum dipakai untuk id yang tidak valid di path (resource dengan id itu memang tidak ada), dan pesan dari library `uuid` sebaiknya tidak dikirim ke client:

```go
id, err := uuid.Parse(ctx.Params("id"))
if err != nil {
    return ctx.Status(http.StatusNotFound).JSON(dto.CreateResponseError(domain.ErrNotFound.Error()))
}
```

Karena polanya sama di lima tempat, cocok dijadikan helper `parseId(ctx)` di `api/helper.go`.

### 🟡 3. Register bersamaan → 500

Dua request dengan email yang sama bisa sama-sama lolos `FindByEmail`. Yang kedua ditolak `UNIQUE` di DB, lalu jadi 500. Tangkap error unique violation (kode Postgres `23505`) di `userRepository.Save` dan kembalikan `ErrDupeEmail`. Kecil: `strings.EqualFold(user.Email, normalizedEmail)` cukup diganti `user.Id != ""`, karena kalau baris ketemu, emailnya pasti sama.

### 🟡 4. Error DB bocor ke client

Untuk 500, `err.Error()` dari Postgres (nama tabel, constraint) tetap dikirim apa adanya. Di cabang `default` pada `errorStatus`, log error aslinya dan kirim pesan generik seperti `"internal server error"`.

### 🟡 5. Validasi product belum ada (`dto/product_data.go`)

`CreateProductReq` dan `UpdateProductReq` tidak punya tag `validate`, dan `util.Validate` tidak dipanggil di `api/product.go`. `stock: -5` atau `status: "apaaja"` menjadi 500 (kena CHECK constraint DB) atau malah tersimpan. Contoh: `validate:"required,gte=0"` untuk stok, `validate:"omitempty,oneof=ACTIVE INACTIVE"` untuk status.

### 🟡 6. Kebersihan kode

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

Urutan: `:id` UUID (2) → validasi product (5) → riwayat order → sisa checkout (1) → pagination → tes → role admin dan upload gambar.

### Fitur inti

1. **Checkout dan order.** Selesai dan teruji, termasuk rollback dan checkout bersamaan. Sisanya perapian (lihat 1).
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
