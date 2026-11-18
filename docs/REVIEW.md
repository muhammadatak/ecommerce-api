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
| Checkout: transaksi | `orderRepository.WithTransaction` membuka transaksi, dan semua repository memakai `getExecutor(ctx, x.db)`, jadi query di dalam `txCtx` ikut transaksi. Diuji: (a) produk kedua stoknya kurang → 400, stok produk pertama tetap 10 dan cart masih berisi 2 item (rollback jalan); (b) 5 user checkout bersamaan untuk produk stok 1 → tepat satu 201, empat 400, stok akhir 0 (`FOR UPDATE` benar-benar mengunci). Cart kosong sudah ditolak sebelum order dibuat. |
| Error → HTTP | Sentinel error di `domain/errors.go`, dipetakan di `errorStatus`: not found 404, auth/token 401, stok/inactive 400, email duplikat 409. Body JSON rusak → 400. |
| Validasi | `TranslateTag` punya pesan untuk `required`, `min`, `max`, `gt`, `uuid`, `email`. |
| Lain-lain | `GET /products` kosong → `[]`, `log.Fatal(app.Listen(...))`. |

## ⏳ Masih terbuka

### 🟡 1. Checkout: sisa masalah

- **`RunInTransaction(ctx, repo interface{}, ...)` masih punya fallback diam-diam.** Sekarang jalan karena `orderRepository` punya `WithTransaction`, tapi kalau suatu saat service lain memanggil `RunInTransaction` dengan repo yang tidak punya method itu, kode tetap jalan tanpa transaksi dan tanpa error. Inject `domain.Transactor` langsung ke service (field `transactor domain.Transactor`), lalu hapus `RunInTransaction`. Compiler yang akan menjaga.
- **Cart kosong → 500.** Sudah ditolak sebelum order dibuat, tapi `errors.New("cart item empty")` tidak dikenal `errorStatus`. Buat sentinel `ErrCartEmpty` di `domain/errors.go` dan petakan ke 400.
- **`cart_id` di `CheckoutRequest` wajib diisi tapi tidak dipakai.** UUID acak pun diterima. Satu user satu cart, dan cart diambil dari `UserId`. Hapus field-nya (dan `Bind().Body` + `Validate` di handler). Sekarang `POST /orders` tanpa body → 400 `unexpected end of JSON input`.
- **N+1 di dalam loop.** `FindByCartAndProduct` dipanggil per produk, padahal `items` sudah diambil. Buat `map[productId]CartItem` dari `items`. Penghapusan cart item juga cukup sekali: `DELETE FROM cart_items WHERE cart_id = ?`.
- **Insert lalu update order.** Hitung total dulu (loop cek stok), baru insert order sekali dengan total yang benar, lalu insert order items (bisa satu query: `Rows(items...)`). `orderRepository.Update` jadi tidak perlu.
- **Update stok menulis semua kolom.** `productRepository.Update` memakai `Set(product)`. Untuk checkout cukup `UPDATE products SET stock = stock - ? WHERE id = ? AND stock >= ?`, lalu `RowsAffected() == 0` → `ErrInsufficientStock`.
- **`FOR UPDATE` dipasang di `FindByIds`** yang juga dipakai lihat cart. Buat method terpisah `FindByIdsForUpdate` dan tambahkan `ORDER BY id`. Tanpa urutan tetap, dua checkout dengan produk A dan B yang sama bisa mengunci dengan urutan berbeda → Postgres mendeteksi deadlock dan salah satunya gagal 500.
- **Response 201 tanpa body.** Client tidak tahu order mana yang dibuat. Kembalikan minimal `id`, `status`, `total_price`.
- **Sisa debug dan kode mati.** Lima `fmt.Printf`/`fmt.Println` di `service/order.go` (termasuk `"product = %s"` yang mencetak cart id). `err` dari `FindByIds` dicek setelah loop yang memakai hasilnya. `OrderRepository.FindById` masih `panic("unimplemented")`. `OrderItemRepository.Delete` tidak dipakai (`ON DELETE CASCADE` sudah menangani). Di `FindByIds`, cek `len(ids) == 0` sebaiknya sebelum membangun dataset.
- Status produk dicek setelah stok, sama seperti nit di PATCH. Cek status dulu.

### 🟡 2. `:id` yang bukan UUID masih 500

`uuid.Parse` sudah ditambahkan di cart dan product, tapi errornya di-`return err`. Itu jatuh ke error handler default Fiber. Diuji: `GET /products/abc` → **500** dengan body teks `invalid UUID length: 3`, bukan JSON. Balas 404 seperti handler lain:

```go
id, err := uuid.Parse(ctx.Params("id"))
if err != nil {
    return ctx.Status(http.StatusNotFound).JSON(dto.CreateResponseError(domain.ErrNotFound.Error()))
}
```

Lima tempat: `UpdateItemQuantity`, `DeleteItem`, `productApi.Show`, `Update`, `Delete`. Karena polanya sama, cocok dijadikan helper `parseId(ctx)` di `api/helper.go`.

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

Urutan: sisa checkout (1) → `:id` UUID (2) → validasi product (5) → riwayat order → pagination → tes → role admin dan upload gambar.

### Fitur inti

1. **Checkout dan order.** Skema, alur, dan transaksi sudah jalan. Tinggal merapikan (lihat 1).
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
