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

