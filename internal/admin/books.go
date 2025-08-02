package admin

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"bookmaker.ca/internal/cache"
	"bookmaker.ca/internal/db"
	"bookmaker.ca/internal/models"
)

func UpdateBook(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Get book details
	id, _ := strconv.Atoi(r.FormValue("id"))
	title := r.FormValue("title")
	author := r.FormValue("author")
	description := r.FormValue("description")

	// Handle variants (you may want to loop through variants from the form)
	variantIds := r.Form["variant_id"]
	colors := r.Form["color"]
	stockValues := r.Form["stock"]
	priceValues := r.Form["price"]
	imageFiles := r.MultipartForm.File["variant_image"] // Handling multiple images for variants
	existingImagePaths := r.Form["existing_image_path"]

	println("updating book")

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
		// Default to existing image path from hidden field
		imagePath := existingImagePaths[i]

		// Override only if a new file was uploaded
		if i < len(imageFiles) && imageFiles[i] != nil && imageFiles[i].Filename != "" {
			file, err := imageFiles[i].Open()
			if err != nil {
				log.Printf("Failed to open file %d: %v", i, err)
			} else {
				defer file.Close()
				imagePath = imageFiles[i].Filename
				savePath := "static/img/" + imagePath

				dst, err := os.Create(savePath)
				if err != nil {
					log.Printf("Failed to create file %d: %v", i, err)
				} else {
					defer dst.Close()
					_, err = io.Copy(dst, file)
					if err != nil {
						log.Printf("Failed to write file %d: %v", i, err)
					}
				}
			}
		}

		// Create a variant for each set of values
		variant := models.Variant{
			ID:        variantID,
			Color:     color,
			Stock:     stock,
			Price:     price,
			ImagePath: imagePath,
		}

		if variantIds[i] == "new" {
			db.InsertVariant(id, variant.Color, variant.Stock, variant.Price, variant.ImagePath)
		} else {
			db.UpdateBookVariantByID(variant)
		}
	}

	// Update book details and its variants
	err = db.UpdateBookByID(id, title, author, description)
	if err != nil {
		http.Error(w, "Failed to update", http.StatusInternalServerError)
		return
	}

	// Rebuild the cache or update any other necessary data
	cache.UpdateAuthors()
}
