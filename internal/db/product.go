package db

import (
	"context"
	"fmt"
	"log"

	"github.com/nathanialw/ecommerce/internal/models"
)

// InsertProduct inserts a new product row into the products table.
func InsertProduct(title, author, description string, variants []models.Variant) error {
	// First, insert into the products table
	sql := `
		INSERT INTO products (title, author, description)
		VALUES ($1, $2, $3) RETURNING id
	`
	var productID int
	err := db.QueryRow(ctx, sql, title, author, description).Scan(&productID)
	if err != nil {
		log.Printf("Failed to insert product: %v\n", err)
		return err
	}

	// Then insert the variants into the product_variants table
	for _, v := range variants {
		sqlVariant := `
			INSERT INTO product_variants (product_id, color, stock, price, image_path)
			VALUES ($1, $2, $3, $4, $5)
		`
		_, err := db.Exec(ctx, sqlVariant, productID, v.Color, v.Stock, v.Cents, v.ImagePath)
		if err != nil {
			log.Printf("Failed to insert variant: %v\n", err)
			return err
		}
	}
	log.Printf("Inserted product: %s by %s with %d variants\n", title, author, len(variants))
	return nil
}

func SearchProducts(query string) ([]models.Product, error) {
	rows, err := db.Query(ctx, `
		SELECT b.id, b.title, b.author, b.description,
		       v.color, v.stock, v.price, v.image_path
		FROM products b
		LEFT JOIN product_variants v ON b.id = v.product_id
		WHERE b.title % $1 OR b.author % $1
		ORDER BY GREATEST(similarity(b.title, $1), similarity(b.author, $1)) DESC,
		         b.id, v.color
		LIMIT 20
	`, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	var currentProduct *models.Product

	for rows.Next() {
		var b models.Product
		var v models.Variant

		var color *string
		var stock *int
		var price *int64
		var imagePath *string

		err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Description,
			&color, &stock, &price, &imagePath)
		if err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}

		if currentProduct == nil || currentProduct.ID != b.ID {
			if currentProduct != nil {
				products = append(products, *currentProduct)
			}
			currentProduct = &b
			currentProduct.Variants = []models.Variant{}
		}

		if color != nil {
			v.Color = *color
		}
		if stock != nil {
			v.Stock = *stock
		}
		if price != nil {
			v.Cents = *price
			v.Price = float64(*price) / 100.0
		}
		if imagePath != nil {
			v.ImagePath = *imagePath
		}

		if v.Color != "" {
			currentProduct.Variants = append(currentProduct.Variants, v)
		}
	}

	if currentProduct != nil {
		products = append(products, *currentProduct)
	}

	return products, rows.Err()
}

func GetProductByID(id int) (*models.Product, error) {
	// Initialize product
	var b models.Product

	// Query for product details
	err := db.QueryRow(context.Background(), `
		SELECT id, title, author, description
		FROM products
		WHERE id = $1
	`, id).Scan(&b.ID, &b.Title, &b.Author, &b.Description)
	if err != nil {
		// Handle error if product is not found
		return nil, fmt.Errorf("error fetching product: %v", err)
	}

	// Get variants using the separate function
	variants, err := GetVariantsByProductID(id)
	if err != nil {
		return nil, err
	}

	// Assign the variants to the product
	b.Variants = variants

	// Return the product with variants
	return &b, nil
}

func UpdateProductByID(id int, title, author, desc string) error {
	// Update the product information
	queryProduct := `
		UPDATE products 
		SET title=$1, author=$2, description=$3
		WHERE id=$4
	`
	_, err := db.Exec(ctx, queryProduct, title, author, desc, id)
	if err != nil {
		log.Printf("Failed to update product info: %v\n", err)
		return err
	}

	return nil
}

func GetAllProducts() ([]models.Product, error) {
	// Query to join product with product_variants
	rows, err := db.Query(context.Background(), `
		SELECT b.id, b.title, b.author, b.description, 
		       v.color, v.stock, v.price, v.image_path
		FROM products b
		LEFT JOIN product_variants v ON b.id = v.product_id
		ORDER BY b.id, v.color
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	var currentProduct *models.Product

	// Iterate through query results
	for rows.Next() {
		var b models.Product
		var v models.Variant

		// Use pointers for nullable fields
		var color *string
		var stock *int
		var price *int64
		var imagePath *string

		// Scan the values into the appropriate variables, including pointers
		err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Description,
			&color, &stock, &price, &imagePath)
		if err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}

		// Check if we need to create a new product or if we're still processing the same product
		if currentProduct == nil || currentProduct.ID != b.ID {
			// Add the previous product to the list, if any
			if currentProduct != nil {
				products = append(products, *currentProduct)
			}

			// Create a new product and set it as current
			currentProduct = &b
			currentProduct.Variants = []models.Variant{} // Initialize the variants array
		}

		// Only append the variant if it's not NULL (i.e., a valid variant)
		if color != nil {
			v.Color = *color
		}
		if stock != nil {
			v.Stock = *stock
		}
		if price != nil {
			v.Cents = *price
			v.Price = float64(*price) / 100.0
		}
		if imagePath != nil {
			v.ImagePath = *imagePath
		}

		// Append the variant to the product's variants array if it's valid
		if v.Color != "" {
			currentProduct.Variants = append(currentProduct.Variants, v)
		}
	}

	// Add the last product to the list
	if currentProduct != nil {
		products = append(products, *currentProduct)
	}

	// for i := 0; i < len(products); i++ {
	// 	println(products[i].Title)
	// }

	return products, rows.Err()
}

func DeleteProduct(id int) {
	// Before deleting the product:
	// product, err := db.GetProductByID(productID)
	// if err == nil && product.ImagePath != "" {
	// 	os.Remove("static/img/" + product.ImagePath)
	// }
	DeleteProductEntry(id)
	DeleteVariantEntries(id)
}

func InsertProductReturningID(title, author, description string) (int, error) {
	var id int
	sql := `
		INSERT INTO products (title, author, description)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	err := db.QueryRow(ctx, sql, title, author, description).Scan(&id)
	if err != nil {
		log.Printf("InsertProduct error: %v\n", err)
		return 0, err
	}
	return id, nil
}

func DeleteProductEntry(id int) {
	_, err := db.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		log.Printf("error deleting product: %v\n", err)
	}
}
