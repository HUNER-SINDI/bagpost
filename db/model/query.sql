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
    ws.ku,
    ws.en,
    ws.ar
FROM warehouses w
LEFT JOIN warehousesetter ws ON w.id = ws.warehouse_id
ORDER BY w.id ASC;

-- name: InsertWharehouses :one
INSERT INTO warehouses (fullname, locname, email, password, phone, is_active)
VALUES ($1,$2,$3,$4,$5,$6 ) RETURNING id;

-- name: InsertWarehouseSetter :one
INSERT INTO warehousesetter (ku, ar , en, warehouse_id)
VALUES ($1, $2, $3 , $4)
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
  ku = $1,
  en = $2,

  ar = $3
WHERE warehouse_id = $4;

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

  from_city_ku,
  from_city_en,
  from_city_ar,

  to_city_ku,
  to_city_en,
  to_city_ar,

  to_subcity_ku,
  to_subcity_en,
  to_subcity_ar,

  to_specific_location,

  status,
  price,
  fdelivery_fee,
  total_price,
  warehouse_id
) VALUES (
  $1,  $2,  $3,  $4,
  $5,  $6,  $7,
  $8,  $9,  $10,
  $11, $12, $13,
  $14,
  $15, $16, $17, $18, $19
)
RETURNING id;


-- name: GetStoreSetter :one
SELECT ku ,en, ar FROM storesetter LIMIT 1;


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
  setter_ku,
  setter_ar,
  setter_en
) VALUES (
  $1, $2, $3 , $4
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
  $1, $2, $3, 'in_store', NULL, NULL
);

-- name: ListDeliveriesByStoreFiltering :many
SELECT * FROM deliveries
WHERE store_owner_id = $1
  AND (COALESCE($2, '') = '' OR status = $2)
  AND (COALESCE($3, '') = '' OR barcode ILIKE '%' || $3 || '%')
  AND (COALESCE($4, '') = '' OR customer_phone ILIKE '%' || $4 || '%')
  AND (
    COALESCE($5, '') = '' OR
    to_city_en ILIKE '%' || $5 || '%' OR
    to_city_ar ILIKE '%' || $5 || '%' OR
    to_city_ku ILIKE '%' || $5 || '%'
  )
  AND (
    COALESCE($6, '') = '' OR
    to_subcity_en ILIKE '%' || $6 || '%' OR
    to_subcity_ar ILIKE '%' || $6 || '%' OR
    to_subcity_ku ILIKE '%' || $6 || '%'
  )
  AND (COALESCE($7, 0) = 0 OR price >= $7)
  AND (COALESCE($8, 0) = 0 OR price <= $8)
LIMIT $9 OFFSET $10;

-- name: GetDeliveryRoutes :many
SELECT * FROM delivery_routing WHERE delivery_id = $1 ORDER BY created_at ASC;
-- name: CountDeliveriesById :one
SELECT COUNT(*) FROM deliveries WHERE store_owner_id = $1;

-- name: GetAllAds :many
SELECT * FROM ads;

-- name: GetDeliveryStatusByStoreId :many
SELECT status, COUNT(*) as count
FROM deliveries
WHERE store_owner_id = $1
GROUP BY status;

-- name: GetAllRoutesByWarehouseId :many
SELECT 
  cr.id,
  cr.city_name_en,
  cr.city_name_ar,
  cr.city_name_ku,
  COALESCE(
    (
      SELECT json_agg(
        json_build_object(
          'id', sc.id,
          'subcity_name_en', sc.subcity_name_en,
          'subcity_name_ar', sc.subcity_name_ar,
          'subcity_name_ku', sc.subcity_name_ku,
          'price', sc.price
        )
      )
      FROM subcities sc
      WHERE sc.city_route_id = cr.id
    ), '[]'
  ) AS subcities
FROM city_routes cr WHERE cr.warehouse_id = $1
ORDER BY cr.id ;


-- name: GetWarehouseIdByStoreId :one 
SELECT warehouse_id FROM store_owners WHERE id = $1;


-- name: GetFromCityByStoreId :one
SELECT city_ku , city_en , city_ar FROM store_owners WHERE id = $1;


-- name: GetAllStoreBalanceById :one
SELECT * FROM store_balances WHERE id = $1;



