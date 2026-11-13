# Review: go-ecommerce-api

`go build ./...` dan `go vet ./...` lolos tanpa error.

Yang sudah bagus: lapisan domain → repository → service → api terpisah, dependency diinject lewat interface, `UserId` di request diambil dari JWT (`json:"-"`) sehingga client tidak bisa memalsukannya, dan PATCH memakai field pointer untuk partial update.

## Status progres (update terakhir)

Dicek terhadap commit `9ba22bc` ditambah perubahan yang belum di-commit. `go build` dan `go vet` lolos.

### ✅ Sudah dikerjakan

| # | Poin | Bukti |
|---|---|---|
| 2 | Upsert item | `cart_item.go` pakai `OnConflict(goqu.DoUpdate(...))`. SQL yang dihasilkan sudah dicek dan benar. |
| 3 | Validasi (sebagian) | `quantity` punya `validate:"required,gt=0"`, `util.Validate` dipanggil di handler dan mengembalikan 400. Service mengecek produk inactive dan stok. |
| 4 | Konfigurasi JWT fail-fast | `loader.go` memanggil `log.Fatal` kalau `JWT_EXP` = 0 atau `JWT_KEY` kosong. |
| 7 | User enumeration | Kedua kasus mengembalikan `"email atau password salah"`. |
| 8 | N+1 query | `FindByIds` dengan `IN (...)` ditambah lookup lewat map. |
| Bag. 2 | Gabung ke CartService | `service/item.go` dihapus, `CartItemService` hilang, logika pindah ke `CartService.Create`. |
| A | `IN ()` saat cart kosong | `Show` sekarang return lebih awal kalau `len(cartItems) == 0`. |
| B | Cek stok kumulatif | `FindByCartAndProduct` mengambil quantity yang sudah ada di cart, lalu yang dicek `req.Quantity + item.Quantity > product.Stock`. Item yang belum ada menghasilkan `Quantity` 0, jadi tetap benar. |
| 1 | Auto-create cart | `AddItem` membuat `domain.Cart` (id dari `uuid.NewString()`) lalu memanggil `cartRepository.Save` kalau user belum punya cart. `cart.Id` terisi dan dipakai untuk item. |

> Catatan belajar untuk B: urutan cek-lalu-upsert masih bisa kena race condition kalau dua request datang bersamaan. Keduanya membaca quantity lama, keduanya lolos cek, lalu total quantity melebihi stok. Untuk cart ini masih bisa diterima karena stok yang sebenarnya dikunci saat checkout, bukan saat item masuk cart.

### ⏳ Masih terbuka

- **Sisa #1 / #10.** `FindAll` (`repository/cart.go` baris 24–26) sudah tidak ada di interface, jadi itu dead code yang isinya `panic`. `Delete` masih ada di interface dan masih `panic`. Hapus keduanya.
- **#5 baru mulai.** `domain/errors.go` sudah dibuat tapi masih kosong, dan semua error masih memakai `errors.New` lalu dikembalikan sebagai 500.
- **#3 sisa.** `product_id` belum punya `validate:"required,uuid"`. DTO product (create/update) belum divalidasi sama sekali.
- **#6, #9–#14** belum disentuh: error DB bocor ke client, ekstraksi JWT masih duplikat, `panic` di `FindAll`/`Save`/`Delete`, `null` vs `[]` di `GET /products`, `_ = app.Listen`, `validator.New()` dibuat per panggilan.
- **Nit:** pesan `"Product stock not cukupz "` (typo dan campur bahasa), route `/carts/item` (saran: `/cart/items`), konstanta `ProductStockInactive` padahal isinya status, file `halo.md` kosong.

---

## Bagian 1: Temuan (urut dari yang paling berdampak)

### 🔴 Bug yang mematahkan fitur

**1. `POST /carts` tidak bisa dipakai user baru** (`internal/service/item.go`)

```go
if cart.Id == "" {
    return errors.New("product not found")   // pesan salah, dan cart tidak pernah dibuat
}
```

User yang belum punya baris di tabel `carts` tidak akan pernah bisa menambah item, dan `cartRepository.Save` masih `panic("unimplemented")`. Solusinya: kalau cart belum ada, buat dulu.

**2. Menambah produk yang sama dua kali menghasilkan 500.** Migration punya `UNIQUE (cart_id, product_id)`, sedangkan `Save` selalu INSERT. Pakai upsert supaya quantity bertambah:

```go
c.db.Insert("cart_items").Rows(item).
    OnConflict(goqu.DoUpdate("cart_id, product_id",
        goqu.Record{"quantity": goqu.L("cart_items.quantity + EXCLUDED.quantity"),
                    "updated_at": goqu.L("NOW()")}))
```

**3. Validasi input belum ada sama sekali.** `util.Validate` sudah ditulis tapi tidak dipanggil di mana pun, dan DTO belum punya tag `validate`. Akibatnya:

- `quantity: 0` atau negatif ditolak oleh CHECK constraint DB, lalu jadi 500.
- `quantity` lebih besar dari `stock` lolos.
- Produk `INACTIVE` tetap bisa masuk cart.
- `status: "apaaja"` di PATCH product jadi 500, `stock: -5` juga.

Tambahkan tag seperti `validate:"required,gt=0"` dan panggil `util.Validate` di handler. Cek stok dan status di service.

**4. Konfigurasi JWT tidak fail-fast** (`internal/config/loader.go`)

```go
expInt, _ := strconv.Atoi(os.Getenv("JWT_EXP"))
```

- `JWT_EXP` kosong → nilainya 0 → token langsung kedaluwarsa begitu dibuat.
- `JWT_KEY` kosong → token ditandatangani dengan key kosong → siapa pun bisa memalsukannya.

Hentikan aplikasi dengan `log.Fatal` kalau salah satunya tidak valid.

### 🟠 HTTP status dan error handling

**5. Hampir semua error dikembalikan sebagai 500.**

- Gagal bind body → seharusnya 400
- Produk tidak ditemukan → seharusnya 404
- Password salah → seharusnya 401

Pola yang umum: sentinel error di domain, lalu dipetakan di handler.

```go
// domain/errors.go
var ErrNotFound = errors.New("data tidak ditemukan")

// api
if errors.Is(err, domain.ErrNotFound) {
    return ctx.Status(http.StatusNotFound).JSON(...)
}
```

Dengan cara ini pengecekan `user.Id == ""` juga bisa diganti dengan nilai `found` dari `ScanStructContext`, yang sekarang dibuang dengan `_`.

**6. Pesan error DB bocor ke client.** `err.Error()` dari Postgres (nama constraint, tabel) dikirim apa adanya. Log error aslinya di server, lalu kirim pesan generik ke client.

**7. Login membuka celah user enumeration.** Pesan `"user not found"` dan `"paassword salah"` (ada typo juga) dibedakan, sehingga penyerang bisa tahu email mana yang terdaftar. Gabungkan jadi satu: `"email atau password salah"`.

### 🟡 Desain dan kualitas kode

**8. N+1 query di `cartService.Show`.** Ada satu `FindById` per item cart. Ganti dengan satu query JOIN `cart_items` ke `products`, atau `WHERE id IN (...)`.

**9. Ekstraksi userId dari JWT disalin di dua handler** (`Show` dan `AddItem` di `internal/api/cart.go`). Pindahkan ke helper `userIDFrom(ctx) (string, bool)`, atau middleware yang menyimpan nilainya di `ctx.Locals("user_id")`.

**10. Ada `panic("unimplemented")` tapi tidak ada recover middleware.** Fiber tidak memasang recover secara default, jadi kalau method itu terpanggil, servernya ikut mati. Hapus method yang tidak dipakai dari interface (`CartRepository.FindAll`, `Delete`), atau pasang `recover.New()`.

**11. Format respons tidak konsisten.**

- `GET /products` mengembalikan `"data": null` kalau kosong (`var productData []`), sementara cart memakai `make(...)` sehingga hasilnya `[]`. Samakan dengan `make`.
- Create product mengembalikan 200 dengan body, AddItem mengembalikan 201 tanpa body. Pilih satu konvensi, misalnya 201 + data.

**12. Otorisasi belum ada.** User mana pun yang sudah login bisa membuat dan menghapus produk. Tahap berikutnya: role admin. `GET /products` biasanya juga dibuat publik.

**13. Error dari `app.Listen` diabaikan** (`_ = app.Listen(...)`). Kalau port sudah dipakai, aplikasi keluar tanpa pesan apa pun. Pakai `log.Fatal(app.Listen(...))`.

**14. Hal kecil:**

- `validator.New()` dibuat ulang setiap kali `Validate` dipanggil. Objek ini mahal karena menyimpan cache, jadi jadikan variabel package-level.
- Type assertion `err.(validator.ValidationErrors)` bisa panic kalau errornya `InvalidValidationError`. Pakai `errors.As`.
- `service/user.go` beserta `dto.UserData` tidak dipakai (dead code), hapus saja.
- Variabel `authRepos` di `main.go` sebenarnya berisi service, jadi namanya menyesatkan.
- Belum ada endpoint register, jadi user harus diinsert manual ke DB.
- `TODO.md` sudah usang: semua item masih belum dicentang padahal banyak yang sudah selesai, dan masih menyebut header `X-User-Id`.
- `godotenv.Load()` memanggil `log.Fatal` kalau `.env` tidak ada, padahal di production biasanya konfigurasi datang dari env var saja. Cukup log sebagai warning.

### Saran urutan belajar berikutnya

1. Perbaiki #1–#3 dulu supaya flow cart benar-benar jalan dan aman dari input aneh.
2. Sentinel error dan pemetaan status code (#5).
3. Endpoint yang masih kurang: register, update quantity, hapus item.
4. Tes untuk service memakai mock repository. Desain berbasis interface sudah siap untuk itu.

---

## Bagian 2: Di mana seharusnya "add item" ditaruh?

Jawaban singkat: endpoint dan logikanya sebaiknya **dijadikan satu dengan cart**, bukan dipisah ke "cart item" sendiri. Pengecekan cart dan product **memang tempatnya di service**. Yang perlu diubah hanya service mana yang memegangnya.

### Kenapa digabung ke cart

Konsep **aggregate** dari DDD, yang juga jadi idiom umum di REST:

- `CartItem` tidak bisa berdiri sendiri. Dia tidak punya arti tanpa cart, dan tidak ada yang mengakses item tanpa lewat cart-nya.
- Cart adalah **aggregate root**. Semua perubahan isi cart lewat root itu, supaya aturan bisnisnya terkumpul di satu tempat: cart harus ada, satu produk satu baris, quantity tidak boleh melebihi stok.

Sekarang `cartItemService` justru tahu soal cart (mencari cart, memutuskan apa yang terjadi kalau cart tidak ada). Itu tanda batasnya salah: tanggung jawab cart bocor ke service lain. Memisah service per tabel memang terlihat rapi, tapi itu *table-driven design*, bukan *domain-driven*.

### Pembagian tanggung jawab per layer

| Layer | Tugas | Contoh di kasus add item |
|---|---|---|
| **Handler (api)** | Bentuk request dan auth | bind body, `validate:"required,gt=0"`, ambil userId dari JWT, petakan error ke status HTTP |
| **Service** | Aturan bisnis | produk ada dan `ACTIVE`, stok cukup, buat cart kalau belum ada, upsert item |
| **Repository** | Akses data saja | `FindByUserId`, `Create`, `UpsertItem`, tanpa `if` bisnis |
| **DB constraint** | Jaring pengaman terakhir | `UNIQUE`, `CHECK (quantity > 0)` |

Jangan pindahkan pengecekan ke handler (handler jadi gemuk dan logikanya tidak bisa dipakai ulang) atau ke repository (repository jadi berisi logika bisnis).

Idiom lain: `CartService` sebaiknya memanggil **`ProductRepository`**, bukan `ProductService`. Service yang saling memanggil service gampang berujung ke dependensi melingkar. Ini sudah dilakukan dengan benar.

### Bentuk yang disarankan

**Route.** Singular karena satu user hanya punya satu cart, dan cart itu sudah ditentukan oleh token:

```
GET    /cart                       -> lihat cart
POST   /cart/items                 -> tambah item (upsert quantity)
PATCH  /cart/items/:productId      -> ubah quantity
DELETE /cart/items/:productId      -> hapus item
```

Sekarang `POST /carts` secara semantik berarti "buat cart baru", bukan "tambah item".

**Domain.** Satu service, `CartItemService` dihapus:

```go
type CartService interface {
    Show(ctx context.Context, userId string) (dto.CartData, error)
    AddItem(ctx context.Context, req dto.AddCartItemRequest) error
    UpdateItem(ctx context.Context, req dto.UpdateCartItemRequest) error
    RemoveItem(ctx context.Context, userId, productId string) error
}
```

**Service:**

```go
func (s *cartService) AddItem(ctx context.Context, req dto.AddCartItemRequest) error {
    product, err := s.productRepository.FindById(ctx, req.ProductId)
    if err != nil {
        return err // ErrNotFound -> 404 di handler
    }
    if product.Status != domain.ProductStatusActive {
        return domain.ErrProductInactive
    }
    if req.Quantity > product.Stock {
        return domain.ErrInsufficientStock
    }

    cart, err := s.cartRepository.FindOrCreateByUserId(ctx, req.UserId)
    if err != nil {
        return err
    }
    return s.cartItemRepository.Upsert(ctx, cart.Id, req.ProductId, req.Quantity)
}
```

`cartItemRepository` **tetap boleh berdiri sendiri** sebagai repository terpisah, karena memang tabelnya beda. Yang digabung hanya service dan API-nya. File `internal/api/cart.go` yang berisi semua route `/cart` sudah pas.

### Catatan lanjutan: transaksi

Mencari atau membuat cart lalu upsert item seharusnya terjadi secara atomik. Dua request bersamaan dari user yang sama bisa sama-sama mencoba membuat cart. Untuk kasus ini cukup dua hal:

- `FindOrCreateByUserId` memakai `INSERT ... ON CONFLICT (user_id) DO NOTHING` lalu `SELECT`. Constraint `UNIQUE(user_id)` yang menjaga dari duplikat.
- Upsert item memakai `ON CONFLICT (cart_id, product_id)`.

Dengan dua constraint itu belum butuh transaksi eksplisit. Transaksi lintas repository baru perlu saat checkout (kurangi stok, buat order, kosongkan cart), dan itu topik yang bagus untuk dipelajari berikutnya.

---

## Bagian 3: Apa yang masih kurang

Dibandingkan API ecommerce pada umumnya, ini yang belum ada. Urutannya dari yang paling penting.

### Fitur inti

1. **Register** (`POST /register`). User sekarang harus diinsert manual dengan hash bcrypt.
2. **Kelola isi cart**: ubah quantity dan hapus item. Saat ini cart hanya bisa ditambah.
3. **Checkout dan order**. Ini inti ecommerce. Butuh tabel `orders` dan `order_items`, dan harga disalin ke `order_items` saat checkout supaya order lama tidak ikut berubah kalau harga produk naik. Prosesnya dalam **satu transaksi**: kunci stok (`SELECT ... FOR UPDATE`), kurangi stok, buat order, kosongkan cart. Di sini kamu belajar transaksi lintas repository.
4. **Riwayat order**: `GET /orders` dan `GET /orders/:id`, hanya order milik user tersebut.
5. **Role admin**: hanya admin yang boleh create, update, dan delete produk.
6. **Upload gambar produk**. `.env` sudah punya `STORAGE_PATH` dan `SERVER_ASSET_URL`, folder `storage/` juga ada, tapi belum dipakai sama sekali.

### Kualitas API

7. **Pagination dan filter** di `GET /products` (`?page=&limit=&q=&status=`). `FindAll` sekarang mengambil seluruh tabel.
8. **Validasi dan sentinel error** (lihat Bagian 1, #3 dan #5).
9. **Middleware dasar Fiber**: `recover`, `logger`, `cors`, dan `requestid` jika perlu.
10. **Graceful shutdown**: tangkap SIGINT/SIGTERM, lalu panggil `app.Shutdown()` dan `db.Close()`.
11. **Pengaturan connection pool**: `db.SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`.

### Engineering

12. **Tes**. Belum ada satu pun file `_test.go`. Mulailah dari unit test service dengan mock repository, misalnya kasus stok kurang, produk inactive, dan cart belum ada.
13. **README**: cara setup `.env`, menjalankan migration goose, menjalankan server, dan daftar endpoint. Lebih bagus lagi kalau ada koleksi Postman/Bruno atau file `.http`.
14. **Seed data** user dan produk untuk development.
15. **Docker Compose** untuk Postgres supaya setup cukup satu perintah.

### Urutan yang disarankan

Perbaiki bug di Bagian 1 → register → kelola isi cart → checkout dan order (belajar transaksi) → pagination → tes → role admin dan upload gambar.
