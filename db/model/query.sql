-- name: GetAdminByEmailAndPassword :one
SELECT * FROM admins
WHERE email = $1 AND password = $2
LIMIT 1;

-- name: ListWarehouses :many
SELECT 
    w.id AS warehouse_id,
    w.fullname,
    w.locname,
    w.email,
    w.phone,
    w.is_active,
    w.created_at,
    w.updated_at,
    ws.id AS setter_id,
    ws.krd,
    ws.ar
FROM warehouses w
LEFT JOIN warehousesetter ws ON w.id = ws.warehouse_id
ORDER BY w.id ASC;

-- name: InsertWharehouses :one
INSERT INTO warehouses (fullname, locname, email, password, phone, is_active)
VALUES ($1,$2,$3,$4,$5,$6 ) RETURNING id;

-- name: InsertWarehouseSetter :one
INSERT INTO warehousesetter (krd, ar, warehouse_id)
VALUES ($1, $2, $3)
RETURNING id;

-- name: UpdateWarehouseInfo :exec
UPDATE warehouses
SET
  fullname = $1,
  locname = $2,
  email = $3,
  phone = $4,
  is_active = $5,
  updated_at = NOW()
WHERE id = $6;

-- name: UpdateWarehouseSetter :exec
UPDATE warehousesetter
SET
  krd = $1,
  ar = $2
WHERE warehouse_id = $3;

-- name: DeleteWarehouse :exec
DELETE FROM warehouses
WHERE id = $1;








-- name: GetWarehouseByEmail :one
SELECT id, email, password, is_active FROM warehouses WHERE email = $1;

-- name: GetWarehouseByID :one
SELECT id, email, password, is_active FROM warehouses WHERE id = $1;


-- name: GetStoreOwnerByEmail :one
SELECT id, email, password, is_active
FROM store_owners
WHERE email = $1
LIMIT 1;

-- name: GetStoreByID :one
SELECT id, email, password, is_active FROM store_owners WHERE id = $1;

-- name: GetStoreProfileById :one
SELECT * FROM store_owners WHERE id = $1;

-- name: DeactivateStoreAccountById :exec
UPDATE store_owners
SET is_active = false
WHERE id = $1;

-- name: InsertDeliveryStore :one
INSERT INTO deliveries (
  barcode,
  store_owner_id,
  customer_phone,
  note,
  from_city,
  to_city,
  to_subcity,
  to_specific_location,
  status,
  price,
  fdelivery_fee,
  total_price,
  warehouse_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8,
  $9, $10, $11, $12, $13
)
RETURNING id;


-- name: GetStoreSetter :one
SELECT krd, ar FROM storesetter LIMIT 1;


-- name: GetStoreBalanceById :one
SELECT in_store_balance
FROM store_balances
WHERE store_owner_id = $1;

-- name: AddToInStoreBalance :exec
UPDATE store_balances
SET in_store_balance = in_store_balance + $1,
    updated_at = now()
WHERE store_owner_id = $2;


-- name: InsertDeliveryRouting :exec
INSERT INTO delivery_routing (
  delivery_id,
  setter_krd,
  setter_ar
) VALUES (
  $1, $2, $3
);


-- name: InsertDeliveryTransfer :exec
INSERT INTO delivery_transfers (
  delivery_id,
  origin_warehouse_id,
  current_warehouse_id,
  transfer_status,
  driver_id,
  transfer_note
) VALUES (
  $1, $2, $3, 'pending', NULL, NULL
);

-- name: ListDeliveriesByStoreFiltering :many
SELECT * FROM deliveries
WHERE store_owner_id = $1
  AND (COALESCE($2, '') = '' OR status = $2)
  AND (COALESCE($3, '') = '' OR barcode ILIKE '%' || $3 || '%')
  AND (COALESCE($4, '') = '' OR customer_phone ILIKE '%' || $4 || '%')
  AND (COALESCE($5, '') = '' OR to_city ILIKE '%' || $5 || '%')
  AND (COALESCE($6, '') = '' OR to_subcity ILIKE '%' || $6 || '%')
  AND (COALESCE($7, 0) = 0 OR price >= $7)
  AND (COALESCE($8, 0) = 0 OR price <= $8)
LIMIT $9 OFFSET $10;

-- name: CountDeliveriesById :one
SELECT COUNT(*) FROM deliveries WHERE store_owner_id = $1;

-- name: GetAllAds :many
SELECT * FROM ads;