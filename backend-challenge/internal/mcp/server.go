// Package mcp provides an MCP (Model Context Protocol) server for the API.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/oolio-group/order-management/internal/models"
	"github.com/oolio-group/order-management/internal/service"
)

// Server wraps the MCP server with order management tools.
type Server struct {
	mcpServer  *server.MCPServer
	productSvc *service.ProductService
	orderSvc   *service.OrderService
}

// NewServer creates a new MCP server with registered tools.
func NewServer(productSvc *service.ProductService, orderSvc *service.OrderService) *Server {
	s := &Server{
		mcpServer:  server.NewMCPServer("Order Food Online", "1.0.0"),
		productSvc: productSvc,
		orderSvc:   orderSvc,
	}
	s.registerTools()
	return s
}

func (s *Server) registerTools() {
	s.mcpServer.AddTool(
		mcp.NewTool("list_products",
			mcp.WithDescription("List all available food products"),
		),
		s.listProductsHandler,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("get_product",
			mcp.WithDescription("Get details for a specific product by ID"),
			mcp.WithString("productId", mcp.Required(), mcp.Description("The product ID")),
		),
		s.getProductHandler,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("place_order",
			mcp.WithDescription("Place a new food order"),
			mcp.WithString("items", mcp.Required(), mcp.Description("JSON array of {productId, quantity} objects")),
			mcp.WithString("couponCode", mcp.Description("Optional coupon/promo code")),
		),
		s.placeOrderHandler,
	)
}

func (s *Server) listProductsHandler(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	products, err := s.productSvc.ListProducts(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list products: %v", err)), nil
	}
	data, _ := json.MarshalIndent(products, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) getProductHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	productID := mcp.ParseString(request, "productId", "")
	if productID == "" {
		return mcp.NewToolResultError("productId is required"), nil
	}

	product, err := s.productSvc.GetProduct(ctx, productID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get product: %v", err)), nil
	}
	data, _ := json.MarshalIndent(product, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) placeOrderHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	itemsStr := mcp.ParseString(request, "items", "")
	if itemsStr == "" {
		return mcp.NewToolResultError("items is required (JSON array)"), nil
	}

	var items []models.OrderItem
	if err := json.Unmarshal([]byte(itemsStr), &items); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid items JSON: %v", err)), nil
	}

	req := models.OrderRequest{Items: items}
	if code := mcp.ParseString(request, "couponCode", ""); code != "" {
		req.CouponCode = code
	}

	resp, err := s.orderSvc.PlaceOrder(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("order failed: %v", err)), nil
	}
	data, _ := json.MarshalIndent(resp, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// StartSSE starts the MCP server with SSE transport.
func (s *Server) StartSSE(addr string) error {
	sseServer := server.NewSSEServer(s.mcpServer)
	slog.Info("MCP SSE server listening", slog.String("addr", addr))
	srv := &http.Server{
		Addr:         addr,
		Handler:      sseServer,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0, // SSE streams are long-lived
		IdleTimeout:  60 * time.Second,
	}
	return srv.ListenAndServe()
}
