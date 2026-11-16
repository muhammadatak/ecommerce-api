# Review: go-ecommerce-api

Dicek terhadap commit `cbf0699` ditambah perubahan yang belum di-commit. `go build` dan `go vet` lolos.

## ✅ Sudah selesai

| Area | Yang sudah benar |
|---|---|
| Struktur | domain → repository → service → api, dependency diinject lewat interface. `CartService` memanggil `ProductRepository`, bukan `ProductService`. |
| Auth | Login dan register. Email dinormalisasi (`ToLower` + `TrimSpace`) di keduanya. Pesan gagal login sama untuk email dan password (`ErrAuth`), jadi tidak bisa dipakai menebak email terdaftar. Password `min=8,max=72` (batas bcrypt). |
| JWT | `JWT_EXP`/`JWT_KEY` fail-fast di `loader.go`. `getUserId` di `api/helper.go`. `UserId` di DTO `json:"-"`, jadi tidak bisa dipalsukan client. Gagal sign → 500. |
| Cart: lihat | Cart kosong → `[]`. Produk diambil sekali dengan `IN (...)` (tanpa N+1), dan return lebih awal kalau cart tidak punya item. |
| Cart: tambah item | Cart dibuat otomatis kalau belum ada. Upsert dengan `ON CONFLICT (cart_id, product_id)`. Cek produk ada, `ACTIVE`, dan stok kumulatif (`req.Quantity + item.Quantity`). |
| Cart: ubah/hapus item | Authorization lewat `FindByIdAndCartId(id, cart.Id)`, jadi item milik orang lain → 404. PATCH mengecek stok dan status produk, lalu hanya menulis `quantity` + `updated_at`. |
| Error → HTTP | Sentinel error di `domain/errors.go`, dipetakan di `errorStatus`: not found 404, auth/token 401, stok/inactive 400, email duplikat 409. Body JSON rusak → 400. |
| Lain-lain | `panic("unimplemented")` sudah dihapus, `GET /products` kosong → `[]`, `log.Fatal(app.Listen(...))`. |

> Catatan: cek stok lalu upsert masih bisa kena race condition kalau dua request datang bersamaan. Untuk cart ini masih bisa diterima, karena stok yang sebenarnya dikunci saat checkout.

## ⏳ Masih terbuka

### 🔴 1. PATCH quantity selalu ditolak 400 (`dto/item_data.go`)

```go
Quantity int `json:"quantity" validate:"gt=0,uuid"`
```

Tag `uuid` dipasang di field `int`. Angka tidak pernah lolos validasi UUID, jadi setiap PATCH (termasuk `{"quantity": 5}`) dibalas `field Quantity should be uuid`. Sudah dicek dengan memanggil `util.Validate(dto.UpdateItemQuantity{Quantity: 5})`.

Hapus `uuid` dari situ: `validate:"required,gt=0"`. Kalau tujuannya memvalidasi id item dari URL, tag di DTO tidak akan bisa: `Id` diisi dari `ctx.Params` **setelah** `Validate` dipanggil, dan punya `json:"-"`. Validasinya ada di poin 3.

### 🟠 2. DELETE item oleh user tanpa cart → 500 (`service/cart.go`)

`UpdateItemQuantity` sudah punya cek `cart.Id == ""`, sedangkan `DeleteItem` belum. Tanpa cek itu query-nya jadi `WHERE cart_id = ''`, dan Postgres menolak `''` sebagai UUID. Salin cek yang sama ke `DeleteItem`.

### 🟡 3. `:id` yang bukan UUID → 500

`PATCH /carts/item/abc`, `DELETE /carts/item/abc`, dan `/products/abc` sampai ke Postgres dan gagal dengan `invalid input syntax for type uuid`. Cek di handler:

```go
if _, err := uuid.Parse(ctx.Params("id")); err != nil {
    return ctx.Status(http.StatusNotFound).JSON(dto.CreateResponseError(domain.ErrNotFound.Error()))
}
```

### 🟡 4. `TranslateTag`: `case "gt>0"` tidak pernah cocok (`util/validation.go`)

`ActualTag()` hanya mengembalikan nama tag, yaitu `"gt"`. Parameternya ada di `fd.Param()`. Sudah dicek: `gt=0` yang gagal menghasilkan `ActualTag="gt"`. Ganti jadi `case "gt":` dan pakai `fd.Param()` seperti `min`/`max`. `case "unique"` tidak dipakai tag mana pun, boleh dihapus.

### 🟡 5. Register bersamaan → 500

Dua request dengan email yang sama bisa sama-sama lolos `FindByEmail`. Yang kedua ditolak `UNIQUE` di DB, lalu jadi 500. Tangkap error unique violation (kode Postgres `23505`) di `userRepository.Save` dan kembalikan `ErrDupeEmail`. Kecil: `strings.EqualFold(user.Email, normalizedEmail)` cukup diganti `user.Id != ""`, karena kalau baris ketemu, emailnya pasti sama.

### 🟡 6. Error DB bocor ke client

Untuk 500, `err.Error()` dari Postgres (nama tabel, constraint) tetap dikirim apa adanya. Di cabang `default` pada `errorStatus`, log error aslinya dan kirim pesan generik seperti `"internal server error"`.

### 🟡 7. Validasi product belum ada (`dto/product_data.go`)

`CreateProductReq` dan `UpdateProductReq` tidak punya tag `validate`, dan `util.Validate` tidak dipanggil di `api/product.go`. `stock: -5` atau `status: "apaaja"` menjadi 500 (kena CHECK constraint DB) atau malah tersimpan. Contoh: `validate:"required,gte=0"` untuk stok, `validate:"omitempty,oneof=ACTIVE INACTIVE"` untuk status.

### 🟡 8. Kebersihan kode

- `util.Validate`: `validator.New()` dibuat ulang setiap panggilan, jadikan variabel package-level. `err.(validator.ValidationErrors)` bisa panic kalau errornya `InvalidValidationError`, pakai `errors.As`.
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

Urutan: perbaiki 1–2 → validasi product (7) → checkout dan order → pagination → tes → role admin dan upload gambar.

### Fitur inti

1. **Checkout dan order.** Ini inti ecommerce. Butuh tabel `orders` dan `order_items`. Harga disalin ke `order_items` saat checkout supaya order lama tidak ikut berubah kalau harga produk naik. Prosesnya dalam **satu transaksi**: kunci stok (`SELECT ... FOR UPDATE`), kurangi stok, buat order, kosongkan cart. Di sini kamu belajar transaksi lintas repository.
2. **Riwayat order.** `GET /orders` dan `GET /orders/:id`, hanya order milik user tersebut. Pola authorization-nya sama dengan cart item: filter `user_id` sejak query pertama.
3. **Role admin.** Sekarang user mana pun yang login bisa membuat dan menghapus produk.
4. **Upload gambar produk.** `.env` sudah punya `STORAGE_PATH` dan `SERVER_ASSET_URL`, folder `storage/` juga ada, tapi belum dipakai.

### Kualitas API

5. **Pagination dan filter** di `GET /products` (`?page=&limit=&q=&status=`). Sekarang seluruh tabel diambil.
6. **Middleware Fiber**: `recover` (supaya panic tidak mematikan server), `logger`, `cors`.
7. **Graceful shutdown**: tangkap SIGINT/SIGTERM, lalu `app.Shutdown()` dan `db.Close()`.
8. **Connection pool**: `db.SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`.

### Engineering

9. **Tes.** Belum ada file `_test.go`. Mulai dari unit test `cartService` dengan mock repository: stok kurang, produk inactive, user tanpa cart, item milik user lain. Bug poin 1 dan 2 akan langsung ketahuan oleh tes semacam ini.
10. **README**: setup `.env`, migration goose, menjalankan server, daftar endpoint. Lebih bagus lagi dengan file `.http` atau koleksi Bruno.
11. **Seed data** user dan produk untuk development.
12. **Docker Compose** untuk Postgres.
