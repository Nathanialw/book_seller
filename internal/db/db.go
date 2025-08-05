package db

import (
	"context"
	"fmt"
	"log"

	"bookmaker.ca/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool
var ctx = context.Background()

func InitDB() {
	var err error
	db, err = pgxpool.New(context.Background(), "postgres://admin:securepassword@localhost/ecommerce")
	if err != nil {
		log.Fatal("Unable to connect to database: ", err)
	}
}

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
		_, err := db.Exec(ctx, sqlVariant, productID, v.Color, v.Stock, v.Price, v.ImagePath)
		if err != nil {
			log.Printf("Failed to insert variant: %v\n", err)
			return err
		}
	}
	log.Printf("Inserted product: %s by %s with %d variants\n", title, author, len(variants))
	return nil
}

func SearchOrders(email string, orderNumber string) (models.Order, error) {
	var order models.Order

	row := db.QueryRow(ctx, `
        SELECT id, email, number
        FROM orders
        WHERE email % $1 AND number::text % $2
        LIMIT 1
    `, email, orderNumber)

	err := row.Scan(&order.ID, &order.Email, &order.Number)
	if err != nil {
		return models.Order{}, err
	}

	return order, nil
}

func SearchProducts(query string) ([]models.Product, error) {
	rows, err := db.Query(ctx, `
        SELECT id, title, author
        FROM products
        WHERE title % $1 OR author % $1
        ORDER BY GREATEST(similarity(title, $1), similarity(author, $1)) DESC
        LIMIT 20
    `, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var b models.Product
		if err := rows.Scan(&b.ID, &b.Title, &b.Author); err != nil {
			continue
		}
		products = append(products, b)
	}
	return products, rows.Err()
}

func GetVariantByID(variantID int) (models.Variant, error) {
	var v models.Variant

	err := db.QueryRow(context.Background(), `
		SELECT id, color, stock, price, image_path
		FROM product_variants
		WHERE id = $1
	`, variantID).Scan(&v.ID, &v.Color, &v.Stock, &v.Price, &v.ImagePath)

	if err != nil {
		return models.Variant{}, fmt.Errorf("error fetching variant: %v", err)
	}

	return v, nil
}

// Function to get variants by product ID
func GetVariantsByProductID(productID int) ([]models.Variant, error) {
	var variants []models.Variant

	// Query for variants associated with the product
	rows, err := db.Query(context.Background(), `
		SELECT id, color, stock, price, image_path
		FROM product_variants
		WHERE product_id = $1
	`, productID)
	if err != nil {
		// Handle error if variants can't be fetched
		return nil, fmt.Errorf("error fetching variants: %v", err)
	}
	defer rows.Close()

	// Scan each variant and append to the variants slice
	for rows.Next() {
		var v models.Variant
		err := rows.Scan(&v.ID, &v.Color, &v.Stock, &v.Price, &v.ImagePath)
		if err != nil {
			// Handle scanning error for variants
			return nil, fmt.Errorf("error scanning variant: %v", err)
		}
		variants = append(variants, v)
	}

	return variants, nil
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

func GetCache() ([]string, error) {
	rows, err := db.Query(ctx, `
        SELECT DISTINCT author
        FROM products
        ORDER BY author ASC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authors []string
	for rows.Next() {
		var author string
		if err := rows.Scan(&author); err != nil {
			log.Println("Error scanning author:", err)
			continue
		}
		authors = append(authors, author)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return authors, nil
}

func UpdateProductAndVariants(id int, title, author, desc string, variants []models.Variant) error {
	// Update the product information
	queryProductID := `
		UPDATE products 
		SET title=$1, author=$2, description=$3
		WHERE id=$4
	`
	_, err := db.Exec(ctx, queryProductID, title, author, desc, id)
	if err != nil {
		log.Printf("Failed to update product info: %v\n", err)
		return err
	}

	// Delegate variant update to separate function
	err = UpdateProductVariants(id, variants)
	if err != nil {
		return err
	}

	log.Printf("Updated product %d with %d variants\n", id, len(variants))
	return nil
}

func UpdateProductVariants(productID int, variants []models.Variant) error {
	for _, variant := range variants {
		queryVariant := `
			UPDATE product_variants
			SET color=$1, stock=$2, price=$3, image_path=$4
			WHERE product_id=$5 AND color=$1
		`
		_, err := db.Exec(ctx, queryVariant, variant.Color, variant.Stock, variant.Price, variant.ImagePath, productID)
		if err != nil {
			log.Printf("Failed to update variant (color: %s): %v\n", variant.Color, err)
			return err
		}
	}
	return nil
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

func UpdateProductVariantByID(variant models.Variant) error {
	query := `
		UPDATE product_variants
		SET color = $1, stock = $2, price = $3, image_path = $4
		WHERE id = $5
	`
	_, err := db.Exec(ctx, query, variant.Color, variant.Stock, variant.Price, variant.ImagePath, variant.ID)
	if err != nil {
		log.Printf("Failed to update variant (id: %d): %v\n", variant.ID, err)
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
		var price *float64
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
			v.Price = *price
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

func DeleteVariantEntries(id int) {
	_, err := db.Exec(ctx, `DELETE FROM product_variants WHERE product_id = $1`, id)
	if err != nil {
		log.Printf("error deleting variants: %v\n", err)
	}
}

func DeleteProductEntry(id int) {
	_, err := db.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		log.Printf("error deleting product: %v\n", err)
	}
}

func DeleteVariantEntry(id int) {
	_, err := db.Exec(ctx, `DELETE FROM product_variants WHERE id = $1`, id)
	if err != nil {
		log.Printf("error deleting variants: %v\n", err)
	}
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

func InsertVariant(productID int, color string, stock int, price float64, imagePath string) error {
	_, err := db.Exec(ctx, `
		INSERT INTO product_variants (product_id, color, stock, price, image_path)
		VALUES ($1, $2, $3, $4, $5)
	`, productID, color, stock, price, imagePath)
	if err != nil {
		log.Printf("InsertVariant error: %v\n", err)
	}
	return err
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
