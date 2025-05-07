package router

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "toxkurd.com/bagpost/db/gen"
	"toxkurd.com/bagpost/dto"
	"toxkurd.com/bagpost/middleware"
	su "toxkurd.com/bagpost/util/store"
)

func RegisterStores(app *fiber.App, q *db.Queries, conn *pgxpool.Pool) {
	store := app.Group("/store")
	store.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Store Api Is Working ")
	})

	//---------- Auth -----------
	// Login and refreshing Token
	store.Post("/login", func(c fiber.Ctx) error {
		var input dto.StoreLoginInput
		// Parsing Body
		if err := c.Bind().Body(&input); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invaild Json Request"})
		}
		// Validate
		if err := validate.Struct(&input); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invaild Json Input"})
		}

		res, err := q.GetStoreOwnerByEmail(c.Context(), pgtype.Text{String: input.Email, Valid: true})
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		// check password

		if input.Password != res.Password.String {
			return c.Status(400).JSON(fiber.Map{"error": "Wrong Password"})
		}

		// Check if account is active
		if !res.IsActive.Bool {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Account is inactive"})
		}

		// Generate JWT and Refresh Token
		accessToken, err := su.GenerateToken(int(res.ID), res.Email.String, res.IsActive.Bool)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
		}

		refreshToken, err := su.GenerateRefreshToken(int(res.ID))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate refresh token"})
		}

		// Respond with tokens
		return c.JSON(fiber.Map{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		})
	})

	store.Post("/refresh", func(c fiber.Ctx) error {
		type RefreshTokenInput struct {
			RefreshToken string `json:"refresh_token"`
		}

		var input RefreshTokenInput
		if err := c.Bind().Body(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
		}

		// Parse and validate the refresh token
		storeId, err := su.ParseRefreshToken(input.RefreshToken)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired refresh token"})
		}

		// Fetch the warehouse user by ID
		storeUser, err := q.GetStoreByID(c.Context(), int32(storeId))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch Store user"})
		}

		// Check if the account is active
		if !storeUser.IsActive.Bool {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Account is inactive"})
		}

		// Generate a new access token
		accessToken, err := su.GenerateToken(int(storeUser.ID), storeUser.Email.String, storeUser.IsActive.Bool)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate new access token"})
		}

		// Generate a new refresh token (with a longer expiration time)
		refreshToken, err := su.GenerateRefreshToken(int(storeUser.ID))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate new refresh token"})
		}

		// Respond with both new access token and refresh token
		return c.JSON(fiber.Map{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		})
	})

	// ----- Profile ------
	// Get Profile Data
	store.Get("/profile", func(c fiber.Ctx) error {
		storeIDRaw := c.Locals("store_id")
		storeID, ok := storeIDRaw.(int)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid store ID in context"})
		}

		response, err := q.GetStoreProfileById(c.Context(), int32(storeID))
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Failed to get user"})
		}

		return c.JSON(response)
	}, middleware.StoresAuthMiddleware)

	store.Delete("/profile", func(c fiber.Ctx) error {
		storeIDRaw := c.Locals("store_id")
		storeID, ok := storeIDRaw.(int)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid store ID in context"})
		}

		err := q.DeactivateStoreAccountById(c.Context(), int32(storeID))
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Failed to deactivate account"})
		}

		return c.JSON(fiber.Map{"message": true})
	}, middleware.StoresAuthMiddleware)

	// ----------- deliveries ------------
	// Insert New deliveries with rollback
	store.Post("/deliveries", func(c fiber.Ctx) error {
		storeIDRaw := c.Locals("store_id")
		storeID, ok := storeIDRaw.(int)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid store ID in context"})
		}

		// 1 - Check if it's valid input
		var input dto.DeliveriesInsert
		if err := c.Bind().Body(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON input"})
		}

		// 2 - Validate input struct
		if err := validate.Struct(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Validation failed",
				"details": err.Error(),
			})
		}

		// 3 - Start transaction
		tx, err := conn.BeginTx(c.Context(), pgx.TxOptions{})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to start transaction"})
		}
		defer tx.Rollback(c.Context())

		qtx := q.WithTx(tx)

		// 3.1 - Get store warehouse ID
		storeData, err := qtx.GetStoreProfileById(c.Context(), int32(storeID))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get store profile"})
		}

		// 3.2 - Insert new delivery and get ID
		deliveryID, err := qtx.InsertDeliveryStore(c.Context(), db.InsertDeliveryStoreParams{
			Barcode:            input.Barcode,
			StoreOwnerID:       int32(storeID),
			CustomerPhone:      input.CustomerPhone,
			Note:               pgtype.Text{String: input.Note, Valid: input.Note != ""},
			FromCity:           input.FromCity,
			ToCity:             input.ToCity,
			ToSubcity:          input.ToSubCity,
			ToSpecificLocation: pgtype.Text{String: input.ToSpecificLocation, Valid: input.ToSpecificLocation != ""},
			Status:             "pending",
			Price:              input.Price,
			FdeliveryFee:       input.Fee,
			TotalPrice:         input.Total,
			WarehouseID:        storeData.WarehouseID.Int32,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to insert delivery", "details": err.Error()})
		}

		// 3.3 - Get setter for delivery routing
		setter, err := qtx.GetStoreSetter(c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch delivery setter info", "details": err.Error()})
		}

		// 3.4 - Insert delivery routing
		err = qtx.InsertDeliveryRouting(c.Context(), db.InsertDeliveryRoutingParams{
			DeliveryID: deliveryID,
			SetterKrd:  setter.Krd,
			SetterAr:   setter.Ar,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to insert delivery routing"})
		}

		// 3.5 - Insert delivery transfer
		err = qtx.InsertDeliveryTransfer(c.Context(), db.InsertDeliveryTransferParams{
			DeliveryID:         deliveryID,
			OriginWarehouseID:  storeData.WarehouseID.Int32,
			CurrentWarehouseID: storeData.WarehouseID.Int32,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to insert delivery transfer"})
		}

		// 3.6 Add total price to To In Store Balance!
		err = qtx.AddToInStoreBalance(c.Context(),
			db.AddToInStoreBalanceParams{
				InStoreBalance: pgtype.Int4{Int32: input.Total, Valid: true},
				StoreOwnerID:   pgtype.Int4{Int32: int32(storeID), Valid: true}})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to Update Balance"})
		}

		// Commit the transaction
		if err := tx.Commit(c.Context()); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to commit transaction"})
		}

		return c.JSON(fiber.Map{
			"message":     "Delivery created successfully",
			"delivery_id": deliveryID,
		})
	}, middleware.StoresAuthMiddleware)

	store.Get("/deliveries", func(c fiber.Ctx) error {
		storeIDRaw := c.Locals("store_id")
		storeID, ok := storeIDRaw.(int)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid store ID in context"})
		}

		// Get filter parameters from query
		status := c.Query("status", "")
		barcode := c.Query("barcode", "")
		phone := c.Query("phone", "")
		toCity := c.Query("to_city", "")
		toSubcity := c.Query("to_subcity", "")

		// Price range filter parameters
		minPrice := 0
		maxPrice := 0

		if c.Query("min_price", "") != "" {
			minPrice, _ = strconv.Atoi(c.Query("min_price", "0"))
		}

		if c.Query("max_price", "") != "" {
			maxPrice, _ = strconv.Atoi(c.Query("max_price", "0"))
		}

		// Parse pagination parameters
		limit, _ := strconv.Atoi(c.Query("limit", "10"))
		offset, _ := strconv.Atoi(c.Query("offset", "0"))

		// Execute the SQL query with the dynamic filters
		deliveries, err := q.ListDeliveriesByStoreFiltering(c.Context(), db.ListDeliveriesByStoreFilteringParams{
			StoreOwnerID: int32(storeID),
			Column2:      status,
			Column3:      barcode,
			Column4:      phone,
			Column5:      toCity,
			Column6:      toSubcity,
			Column7:      int32(minPrice),
			Column8:      int32(maxPrice),
			Limit:        int32(limit),
			Offset:       int32(offset),
		})

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		// Count total deliveries for pagination
		totalDeliveries, err := q.CountDeliveriesById(c.Context(), int32(storeID))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		if deliveries == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Not Found"})
		}

		// Create a response structure that includes routes
		type RouteResponse struct {
			ID        int32     `json:"id"`
			SetterKrd string    `json:"setter_krd"`
			SetterAr  string    `json:"setter_ar"`
			CreatedAt time.Time `json:"created_at"`
		}

		type DeliveryWithRoutes struct {
			db.Delivery
			Routes []RouteResponse `json:"routes"`
		}

		// Fetch routes for each delivery and build response
		var response []DeliveryWithRoutes
		for _, delivery := range deliveries {
			// Fetch routes for this delivery
			routes, err := q.GetDeliveryRoutes(c.Context(), delivery.ID)
			if err != nil {
				// Log error but continue
				log.Printf("Error fetching routes for delivery %d: %v", delivery.ID, err)
				// Add delivery without routes
				response = append(response, DeliveryWithRoutes{
					Delivery: delivery,
					Routes:   []RouteResponse{},
				})
				continue
			}

			// Convert routes to response format
			routesResponse := make([]RouteResponse, len(routes))
			for i, route := range routes {
				routesResponse[i] = RouteResponse{
					ID:        route.ID,
					SetterKrd: route.SetterKrd,
					SetterAr:  route.SetterAr,
					CreatedAt: route.CreatedAt.Time,
				}
			}

			// Add delivery with routes to response
			response = append(response, DeliveryWithRoutes{
				Delivery: delivery,
				Routes:   routesResponse,
			})
		}

		// Cast limit and offset to int64 to match the type of totalDeliveries (int64)
		totalItems := totalDeliveries
		totalPages := (totalItems + int64(limit) - 1) / int64(limit) // Calculate total pages
		currentPage := int64(offset)/int64(limit) + 1                // Calculate current page

		var nextPage, prevPage string
		if currentPage < totalPages {
			nextPage = fmt.Sprintf("/deliveries?limit=%d&offset=%d", limit, (currentPage)*int64(limit))
		} else {
			nextPage = ""
		}

		if currentPage > 1 {
			prevPage = fmt.Sprintf("/deliveries?limit=%d&offset=%d", limit, (currentPage-2)*int64(limit))
		} else {
			prevPage = ""
		}

		// Return the response with pagination metadata
		return c.JSON(fiber.Map{
			"count": totalItems,
			"data":  response,
			"pagination": fiber.Map{
				"page":        currentPage,
				"page_size":   limit,
				"total_pages": totalPages,
				"next_page":   nextPage,
				"prev_page":   prevPage,
			},
		})
	}, middleware.StoresAuthMiddleware)

	store.Get("/ads", func(c fiber.Ctx) error {
		response, err := q.GetAllAds(c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		if response == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "no responses"})
		}
		return c.JSON(response)

	})
	store.Get("/status", func(c fiber.Ctx) error {
		storeIDRaw := c.Locals("store_id")
		storeID, ok := storeIDRaw.(int)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid store ID in context"})
		}

		response, err := q.GetDeliveryStatusByStoreId(c.Context(), int32(storeID))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		if response == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "no responses"})
		}
		return c.JSON(response)

	}, middleware.StoresAuthMiddleware)

	store.Get("/routes", func(c fiber.Ctx) error {
		storeIDRaw := c.Locals("store_id")
		storeID, ok := storeIDRaw.(int)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid store ID in context"})
		}

		// get warehouse id
		warehouseId, err := q.GetWarehouseIdByStoreId(c.Context(), int32(storeID))
		if err != nil {
			return c.Status(fiber.StatusNoContent).JSON(fiber.Map{"error": err.Error()})
		}

		res, err := q.GetAllRoutesByWarehouseId(c.Context(), warehouseId)
		if err != nil {
			return c.Status(fiber.StatusNoContent).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(res)
	}, middleware.StoresAuthMiddleware)

}
