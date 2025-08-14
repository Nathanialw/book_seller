package db

import (
	"context"
	"fmt"
	"log"

	"github.com/nathanialw/ecommerce/pkg/models"
)

func GetVariantByID(variant_id int) (models.Variant, error) {
	var v models.Variant

	err := db.QueryRow(context.Background(), `
		SELECT id, color, stock, cents, image_path
		FROM variants
		WHERE id = $1
	`, variant_id).Scan(&v.ID, &v.Color, &v.Stock, &v.Cents, &v.ImagePath)
	v.Price = float64(v.Cents) / 100.0

	if err != nil {
		return models.Variant{}, fmt.Errorf("error fetching variant: %v", err)
	}

	return v, nil
}

// Function to get variants by product ID
func GetVariantsByProductID(product_id int) ([]models.Variant, error) {
	var variants []models.Variant

	// Query for variants associated with the product
	rows, err := db.Query(context.Background(), `
		SELECT id, color, stock, cents, image_path
		FROM variants
		WHERE product_id = $1
	`, product_id)
	if err != nil {
		// Handle error if variants can't be fetched
		return nil, fmt.Errorf("error fetching variants: %v", err)
	}
	defer rows.Close()

	// Scan each variant and append to the variants slice
	for rows.Next() {
		var v models.Variant
		err := rows.Scan(&v.ID, &v.Color, &v.Stock, &v.Cents, &v.ImagePath)
		v.Price = float64(v.Cents) / 100.0
		if err != nil {
			// Handle scanning error for variants
			return nil, fmt.Errorf("error scanning variant: %v", err)
		}
		variants = append(variants, v)
	}

	return variants, nil
}

func UpdateProductVariants(product_id int, variants []models.Variant) error {
	for _, variant := range variants {
		queryVariant := `
			UPDATE variants
			SET color=$1, stock=$2, cents=$3, image_path=$4
			WHERE product_id=$5 AND color=$1
		`
		_, err := db.Exec(ctx, queryVariant, variant.Color, variant.Stock, variant.Cents, variant.ImagePath, product_id)
		if err != nil {
			log.Printf("Failed to update variant (color: %s): %v\n", variant.Color, err)
			return err
		}
	}
	return nil
}

func UpdateProductVariantByID(variant models.Variant) error {
	query := `
		UPDATE variants
		SET color = $1, stock = $2, cents = $3, image_path = $4
		WHERE id = $5
	`
	_, err := db.Exec(ctx, query, variant.Color, variant.Stock, variant.Cents, variant.ImagePath, variant.ID)
	if err != nil {
		log.Printf("Failed to update variant (id: %d): %v\n", variant.ID, err)
		return err
	}
	return nil
}

func InsertVariant(product_id int, color string, stock int, cents int64, imagePath string) error {
	_, err := db.Exec(ctx, `
		INSERT INTO variants (product_id, color, stock, cents, image_path)
		VALUES ($1, $2, $3, $4, $5)
	`, product_id, color, stock, cents, imagePath)
	if err != nil {
		log.Printf("InsertVariant error: %v\n", err)
	}
	return err
}

func DeleteVariantEntries(product_id int) {
	_, err := db.Exec(ctx, `DELETE FROM variants WHERE product_id = $1`, product_id)
	if err != nil {
		log.Printf("error deleting variants: %v\n", err)
	}
}

func DeleteVariantEntry(variant_id int) {
	_, err := db.Exec(ctx, `DELETE FROM variants WHERE id = $1`, variant_id)
	if err != nil {
		log.Printf("error deleting variants: %v\n", err)
	}
}
