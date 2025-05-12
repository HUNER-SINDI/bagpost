package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "toxkurd.com/bagpost/db/gen"
	"toxkurd.com/bagpost/dto"
	"toxkurd.com/bagpost/middleware"
	su "toxkurd.com/bagpost/util/empl"
)

func RegisterEmplRoute(app *fiber.App, q *db.Queries, conn *pgxpool.Pool) {
	empl := app.Group("/empl")

	empl.Post("/login", func(c fiber.Ctx) error {
		var input dto.EmplLoginInput

		// Parse JSON input
		if err := c.Bind().Body(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON request"})
		}

		// Validate input
		if err := validate.Struct(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
		}

		// Attempt login via SQLC
		empl, err := q.LoginEmplWithEmailAndPassword(c.Context(), db.LoginEmplWithEmailAndPasswordParams{
			Email:    pgtype.Text{String: input.Email, Valid: true},
			Password: pgtype.Text{String: input.Password, Valid: true},
		})
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
		}

		// Check if employee is active
		if !empl.IsActive.Bool {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Account is inactive"})
		}

		// Generate tokens
		accessToken, err := su.GenerateToken(int(empl.ID), empl.Email.String, empl.IsActive.Bool)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
		}

		refreshToken, err := su.GenerateRefreshToken(int(empl.ID))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate refresh token"})
		}

		// Return tokens
		return c.JSON(fiber.Map{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		})
	})
	empl.Post("/refresh", func(c fiber.Ctx) error {
		type RefreshTokenInput struct {
			RefreshToken string `json:"refresh_token"`
		}

		var input RefreshTokenInput
		if err := c.Bind().Body(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
		}

		// Parse and validate the refresh token
		emplId, err := su.ParseRefreshToken(input.RefreshToken)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}

		// Fetch the warehouse user by ID
		emplUser, err := q.GetEmplById(c.Context(), int32(emplId))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch Store user"})
		}

		// Check if the account is active
		if !emplUser.IsActive.Bool {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Account is inactive"})
		}

		// Generate a new access token
		accessToken, err := su.GenerateToken(int(emplUser.ID), emplUser.Email.String, emplUser.IsActive.Bool)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate new access token"})
		}

		// Generate a new refresh token (with a longer expiration time)
		refreshToken, err := su.GenerateRefreshToken(int(emplUser.ID))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate new refresh token"})
		}

		// Respond with both new access token and refresh token
		return c.JSON(fiber.Map{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		})
	})

	// get employe profile by id
	empl.Get("/profile", func(c fiber.Ctx) error {
		emloyeRawId := c.Locals("id")
		emplId, ok := emloyeRawId.(int)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid store ID in context"})
		}
		response, err := q.GetEmplById(c.Context(), int32(emplId))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to Get Profile By Id"})
		}
		return c.JSON(response)

	}, middleware.EmplAuthMiddleware)

	empl.Post("/transfer", func(c fiber.Ctx) error {
		var input dto.DeliveryTransferInput
		if err := c.Bind().Body(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON input"})
		}
		if err := validate.Struct(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Validation failed",
				"details": err.Error(),
			})
		}

		emloyeRawId := c.Locals("id")
		emplId, ok := emloyeRawId.(int)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid employee ID in context"})
		}

		tx, err := conn.BeginTx(c.Context(), pgx.TxOptions{})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to start transaction"})
		}
		defer tx.Rollback(c.Context())

		qtx := db.New(tx)

		// 1 - Get employee
		_, err = qtx.GetEmplById(c.Context(), int32(emplId))
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Employee not found"})
		}

		// 2 - Get delivery
		delivery, err := qtx.GetDeliveryByBarcode(c.Context(), input.Barcode)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Delivery not found"})
		}

		// 3 - Get delivery transfer
		transfer, err := qtx.GetDeliveryTransferByDeliveryID(c.Context(), delivery.ID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Delivery transfer not found"})
		}

		// 4 - Reject if driver_id already set
		if transfer.DriverID.Valid {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Delivery already has a driver"})
		}

		// 5 - Update driver_id in delivery_transfers
		err = qtx.UpdateDriverForDelivery(c.Context(), db.UpdateDriverForDeliveryParams{
			DriverID: pgtype.Int4{
				Int32: int32(emplId),
				Valid: true,
			},
			DeliveryID: delivery.ID,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update driver"})
		}

		// 6 - Update delivery status to pending
		err = qtx.UpdateDeliveryStatusToPending(c.Context(), delivery.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update delivery status"})
		}

		// 7 - Update store balance ONLY IF transfer_status == "in_store"
		if delivery.Status == "in_store" {
			err = qtx.UpdateStoreBalanceOnTransfer(c.Context(), db.UpdateStoreBalanceOnTransferParams{
				InStoreBalance: pgtype.Int4{
					Int32: int32(delivery.TotalPrice),
					Valid: true,
				},
				StoreOwnerID: pgtype.Int4{
					Int32: delivery.StoreOwnerID,
					Valid: true,
				},
			})
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update store balance"})
			}
		}

		// 8 - Insert delivery_routing
		err = qtx.InsertDeliveryRouting(c.Context(), db.InsertDeliveryRoutingParams{
			DeliveryID: delivery.ID,
			SetterKu:   input.SetterKu,
			SetterEn:   input.SetterEn,
			SetterAr:   input.SetterAr,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to insert delivery routing"})
		}

		// 9 - Create delivery_actions_employee
		_, err = qtx.CreateDeliveryActionEmployee(c.Context(), db.CreateDeliveryActionEmployeeParams{
			DeliveryID: delivery.ID,
			EmployeeID: int32(emplId),
			Price:      int32(delivery.TotalPrice),
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create delivery action"})
		}

		// 10 - Commit transaction
		if err := tx.Commit(c.Context()); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to commit transaction"})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Delivery transfer completed successfully",
		})
	}, middleware.EmplAuthMiddleware)

	empl.Get("/delivery", func(c fiber.Ctx) error {

		emloyeRawId := c.Locals("id")
		emplId, ok := emloyeRawId.(int)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid employee ID in context"})
		}
		response, err := q.GetDeliveryTransferDetailsWithEmployeePrice(c.Context(), int32(emplId))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(response)

	}, middleware.EmplAuthMiddleware)

	empl.Get("/stats", func(c fiber.Ctx) error {

		emloyeRawId := c.Locals("id")
		emplId, ok := emloyeRawId.(int)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid employee ID in context"})
		}
		response, err := q.GetDeliverySummaryByEmployee(c.Context(), int32(emplId))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(response)

	}, middleware.EmplAuthMiddleware)

}
