package services

import (
	"context"
	"fmt"
	"go-grpc/cmd/helpers"
	pagingPb "go-grpc/pb/pagination"
	productPb "go-grpc/pb/product"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type ProductService struct {
	productPb.UnimplementedProductServiceServer
	DB *gorm.DB
}

func (p *ProductService) GetProducts(ctx context.Context, pageParam *productPb.Page) (*productPb.Products, error) {
	var page int64 = 1
	if pageParam.GetPage() != 0 {
		page = pageParam.GetPage()
	}

	var pagination *pagingPb.Pagination = &pagingPb.Pagination{}
	var products []*productPb.Product

	sql := p.DB.Table("products as p").
		Joins("LEFT JOIN categories AS c ON c.id = p.category_id").
		Select("p.id", "p.name", "p.price", "p.stock", "c.id as category_id", "c.name as category_name")

	// memanggil fungsi Pagination
	offset, limit := helpers.Pagination(sql, page, pagination)

	// Menjalankan query dengan offset dan limit
	rows, err := sql.Offset(int(offset)).Limit(int(limit)).Rows()
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to execute query: %v", err.Error()))
	}
	defer rows.Close()

	// iterate over rows
	for rows.Next() {
		var product productPb.Product
		var category productPb.Category

		// scan rows to product and category
		if err := rows.Scan(
			&product.Id,
			&product.Name,
			&product.Price,
			&product.Stock,
			&category.Id,
			&category.Name,
		); err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to scan row: %v", err.Error()))
		}

		// Menambahkan kategori ke dalam produk
		product.Category = &category

		// Menambahkan produk ke list produk
		products = append(products, &product)
	}

	// jika tidak ada produk
	if len(products) == 0 {
		return &productPb.Products{
			Pagination: pagination,
			Data:       nil,
		}, nil
	}

	// Membuat dan mengembalikan response
	response := &productPb.Products{
		Pagination: pagination,
		Data:       products,
	}

	return response, nil
}

func (p *ProductService) GetProduct(ctx context.Context, id *productPb.Id) (*productPb.Product, error) {
	row := p.DB.Table("products as p").
		Joins("LEFT JOIN categories AS c ON c.id = p.category_id").
		Select("p.id", "p.name", "p.price", "p.stock", "c.id as category_id", "c.name as category_name").
		Where("p.id = ?", id.GetId()).
		Row()

	var product productPb.Product
	var category productPb.Category

	// scan rows to product and category
	if err := row.Scan(
		&product.Id,
		&product.Name,
		&product.Price,
		&product.Stock,
		&category.Id,
		&category.Name,
	); err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Failed to scan row: %v", err.Error()))
	}

	// Menambahkan kategori ke dalam produk
	product.Category = &category

	return &product, nil
}

func (p *ProductService) CreateProduct(ctx context.Context, productData *productPb.Product) (*productPb.Id, error) {

	var Response productPb.Id

	err := p.DB.Transaction(func(tx *gorm.DB) error {
		category := productPb.Category{
			Id:   0,
			Name: productData.GetCategory().GetName(),
		}

		if err := tx.Table("categories").
			Where("LOWER(name) = ?", category.GetName()).
			FirstOrCreate(&category).Error; err != nil {
			return fmt.Errorf("failed to query or create category: %v", err)
		}

		product := struct {
			Id          uint64
			Name        string
			Price       float64
			Stock       uint32
			Category_id uint64
		}{
			Id:          productData.GetId(),
			Name:        productData.GetName(),
			Price:       productData.GetPrice(),
			Stock:       productData.GetStock(),
			Category_id: category.GetId(),
		}

		if err := tx.Table("products").Create(&product).Error; err != nil {
			return err
		}

		Response.Id = product.Id
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to create product: %v", err.Error()))
	}

	return &Response, nil
}

func (p *ProductService) UpdateProduct(ctx context.Context, productData *productPb.Product) (*productPb.Status, error) {
	var Response productPb.Status

	err := p.DB.Transaction(func(tx *gorm.DB) error {
		category := productPb.Category{
			Id:   0,
			Name: productData.GetCategory().GetName(),
		}

		if err := tx.Table("categories").
			Where("LOWER(name) = ?", category.GetName()).
			FirstOrCreate(&category).Error; err != nil {
			return fmt.Errorf("failed to query or create category: %v", err)
		}

		product := struct {
			Id          uint64
			Name        string
			Price       float64
			Stock       uint32
			Category_id uint64
		}{
			Id:          productData.GetId(),
			Name:        productData.GetName(),
			Price:       productData.GetPrice(),
			Stock:       productData.GetStock(),
			Category_id: category.GetId(),
		}

		if err := tx.Table("products").Where("id = ?", product.Id).Updates(&product).Error; err != nil {
			return err
		}

		Response.Status = 1
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to update product: %v", err.Error()))
	}

	return &Response, nil
}

func (p ProductService) DeleteProduct(ctx context.Context, id *productPb.Id) (*productPb.Status, error) {
	var Response productPb.Status

	if err := p.DB.Table("products").
		Where("id = ?", id.GetId()).
		Delete(nil).
		Error; err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete product: %v", err.Error()))
	}

	Response.Status = 1
	return &Response, nil
}
