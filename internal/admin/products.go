package admin

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/nathanialw/ecommerce/internal/cache"
	"github.com/nathanialw/ecommerce/internal/db"
	"github.com/nathanialw/ecommerce/pkg/models"
)

func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Get product details
	id, _ := strconv.Atoi(r.FormValue("id"))
	title := r.FormValue("title")
	author := r.FormValue("author")
	description := r.FormValue("description")

	// Handle variants (you may want to loop through variants from the form)
	variantIds := r.Form["variant_id"]
	colors := r.Form["color"]
	stockValues := r.Form["stock"]
	priceValues := r.Form["price"]
	existingImagePaths := r.Form["existing_image_path"]

	for i := 0; i < len(colors); i++ {
		// Ensure all variant fields are populated
		if len(colors) != len(stockValues) || len(colors) != len(priceValues) {
			http.Error(w, "Variant fields mismatch", http.StatusBadRequest)
			return
		}

		// Ensure we have all fields for each variant

		variantID, _ := strconv.Atoi(variantIds[i])
		color := colors[i]
		stock, _ := strconv.Atoi(stockValues[i])
		price, _ := strconv.ParseFloat(priceValues[i], 64)
		cents := int64(math.Round(price * 100))
		// Default to existing image path from hidden field
		imagePath := existingImagePaths[i]

		file, handler, err := r.FormFile(fmt.Sprintf("variant_image[%d]", i))
		if err == nil {
			defer file.Close()
			imagePath = handler.Filename
			savePath := "static/img/" + imagePath

			dst, err := os.Create(savePath)
			if err != nil {
				log.Printf("Failed to create file for variant %d: %v", i, err)
			} else {
				defer dst.Close()
				_, err = io.Copy(dst, file)
				if err != nil {
					log.Printf("Failed to save image for variant %d: %v", i, err)
				}
			}
		} else {
			// No file uploaded, keep existing imagePath
		}

		// Create a variant for each set of values
		variant := models.Variant{
			ID:        variantID,
			Color:     color,
			Stock:     stock,
			Cents:     cents,
			ImagePath: imagePath,
		}

		if variantIds[i] == "new" {
			db.InsertVariant(id, variant.Color, variant.Stock, variant.Cents, variant.ImagePath)
		} else {
			db.UpdateProductVariantByID(variant)
		}
	}

	// Update product details and its variants
	err = db.UpdateProductByID(id, title, author, description)
	if err != nil {
		http.Error(w, "Failed to update", http.StatusInternalServerError)
		return
	}

	// Rebuild the cache or update any other necessary data
	cache.UpdateCache()
}
