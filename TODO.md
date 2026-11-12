# TODO: Fitur Cart

## Database
- [ ] Jalankan migration: `goose -dir db/migrations postgres "<dsn>" up`
- [ ] Cek tabel `users`, `carts`, `cart_items` sudah ada

## Domain
- [ ] Buat `domain/user.go`
- [ ] `CartRepository`: tambah `FindByUserId`
- [ ] `CartItemRepository`: ganti `FindAll` jadi `FindByCartId`
- [ ] `CartService.Show`: terima `userId`

## Repository
- [ ] `cart.go`: implement `FindByUserId`, cek `found`
- [ ] Buat `cart_item.go`: implement `FindByCartId`

## Service
- [ ] `NewCart` terima cart, cart item, product repository
- [ ] `Show`: ambil cart → ambil items → ambil product → map ke DTO
- [ ] Cart tidak ada → kembalikan cart kosong

## DTO
- [ ] `CartItemData`: ganti embed `ProductData` jadi field `Product`

## API
- [ ] Buat `internal/api/cart.go` (contoh: `product.go`)
- [ ] Route `GET /cart`, ambil user dari header `X-User-Id`
- [ ] Header kosong → 401

## main.go
- [ ] Daftarkan repository, service, dan api cart

## Tes
- [ ] Insert data dummy (user, product, cart, cart_item)
- [ ] Coba `GET /cart` dengan header `X-User-Id`

## Nanti
- [ ] Tambah item ke cart
- [ ] Ubah quantity / hapus item
- [ ] Auth JWT: ganti header dengan `ctx.Locals("user_id")`
