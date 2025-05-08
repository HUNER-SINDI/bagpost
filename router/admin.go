package router

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "toxkurd.com/bagpost/db/gen"
	"toxkurd.com/bagpost/dto"
	"toxkurd.com/bagpost/middleware"
	utils "toxkurd.com/bagpost/util/admin"
)

var validate = validator.New()

func RegisterAdminRoute(app *fiber.App, q *db.Queries, conn *pgxpool.Pool) {
	admin := app.Group("/admin")

	admin.Post("/login", func(ctx fiber.Ctx) error {
		var adminLoginReq db.GetAdminByEmailAndPasswordParams
		err := ctx.Bind().Body(&adminLoginReq)

		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err,
			})
		}
		res, err := q.GetAdminByEmailAndPassword(ctx.Context(), adminLoginReq)

		if err != nil {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}

		token, err := utils.CreateToken(int64(res.ID), res.Email)
		if err != nil {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}

		return ctx.JSON(fiber.Map{"token": token})
	})
	admin.Post("/warehouses", func(ctx fiber.Ctx) error {

		var input dto.CreateWarehouseInput

		err := ctx.Bind().Body(&input)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON input"})
		}

		if err := validate.Struct(&input); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Validation failed",
				"details": err.Error(),
			})
		}

		// ✅ Begin transaction from real DB connection
		tx, err := conn.BeginTx(ctx.Context(), pgx.TxOptions{})
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to start transaction"})
		}
		defer tx.Rollback(ctx.Context()) // rollback if not committed

		qtx := q.WithTx(tx)

		// Insert warehouse
		warehouseParams := db.InsertWharehousesParams{
			Fullname: input.Fullname,
			Locname:  input.Locname,
			Email:    input.Email,
			Password: input.Password,
			Phone:    input.Phone,
			IsActive: pgtype.Bool{Bool: input.IsActive, Valid: true},
		}

		warehouseID, err := qtx.InsertWharehouses(ctx.Context(), warehouseParams)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to insert warehouse"})
		}

		// Insert warehouse setter
		setterParams := db.InsertWarehouseSetterParams{
			Ku:          input.Ku,
			En:          input.En,
			Ar:          input.Ar,
			WarehouseID: warehouseID,
		}

		_, err = qtx.InsertWarehouseSetter(ctx.Context(), setterParams)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to insert warehouse setter"})
		}

		// ✅ Commit transaction
		if err := tx.Commit(ctx.Context()); err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to commit transaction"})
		}

		return ctx.JSON(fiber.Map{
			"message":      "Warehouse and Setter inserted successfully",
			"warehouse_id": warehouseID,
		})
	}, middleware.AdminAuthMiddleware)

	admin.Get("/warehouses", func(ctx fiber.Ctx) error {

		res, err := q.ListWarehouses(ctx.Context())

		if err != nil {
			return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err})
		}

		if len(res) == 0 {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "00"})
		}

		return ctx.JSON(res)

	}, middleware.AdminAuthMiddleware)

	admin.Put("/warehouses/:id", func(ctx fiber.Ctx) error {
		var input dto.UpdateWarehouseInput

		// 1. Parse body
		err := ctx.Bind().Body(&input)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON input"})
		}

		// 2. Validate
		if err := validate.Struct(&input); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Validation failed",
				"details": err.Error(),
			})
		}

		// 3. Get warehouse ID from URL param
		warehouseID, err := strconv.Atoi(ctx.Params("id"))
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid warehouse ID"})
		}

		// 4. Start transaction
		tx, err := conn.BeginTx(ctx.Context(), pgx.TxOptions{})
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to start transaction"})
		}
		defer tx.Rollback(ctx.Context())

		qtx := q.WithTx(tx)

		// 5. Update warehouse
		updateParams := db.UpdateWarehouseInfoParams{
			Fullname: input.Fullname,
			Locname:  input.Locname,
			Email:    input.Email,
			Phone:    input.Phone,
			IsActive: pgtype.Bool{Bool: input.IsActive, Valid: true},
			ID:       int32(warehouseID),
		}

		err = qtx.UpdateWarehouseInfo(ctx.Context(), updateParams)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update warehouse"})
		}

		// 6. Update warehouse setter
		setterParams := db.UpdateWarehouseSetterParams{
			Ku:          input.Ku,
			En:          input.En,
			Ar:          input.Ar,
			WarehouseID: int32(warehouseID),
		}

		err = qtx.UpdateWarehouseSetter(ctx.Context(), setterParams)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update warehouse setter"})
		}

		// 7. Commit transaction
		if err := tx.Commit(ctx.Context()); err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to commit transaction"})
		}

		return ctx.JSON(fiber.Map{
			"message": "Warehouse updated successfully",
		})
	}, middleware.AdminAuthMiddleware)

	admin.Delete("/warehouses/:id", func(ctx fiber.Ctx) error {
		// 1. Get warehouse ID from URL param
		warehouseID, err := strconv.Atoi(ctx.Params("id"))
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid warehouse ID"})
		}

		// 2. Start transaction
		tx, err := conn.BeginTx(ctx.Context(), pgx.TxOptions{})
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to start transaction"})
		}
		defer tx.Rollback(ctx.Context())

		qtx := q.WithTx(tx)

		// 3. Delete only warehouse (Setter will be deleted automatically)
		err = qtx.DeleteWarehouse(ctx.Context(), int32(warehouseID))
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete warehouse"})
		}

		// 4. Commit transaction
		if err := tx.Commit(ctx.Context()); err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to commit transaction"})
		}

		return ctx.JSON(fiber.Map{
			"message": "Warehouse deleted successfully",
		})
	}, middleware.AdminAuthMiddleware)

}
