package api

import (
	"encoding/json"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"net/http"
	"strconv"

	"sprite_sheet_tool/internal/models"
	"sprite_sheet_tool/internal/slicer"
)

const maxUploadSize = 50 << 20 // 50 MB

func ExportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "missing image file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	configStr := r.FormValue("config")
	if configStr == "" {
		http.Error(w, "missing config field", http.StatusBadRequest)
		return
	}

	var req models.ExportRequest
	if err := json.Unmarshal([]byte(configStr), &req); err != nil {
		http.Error(w, "invalid config JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Prefix == "" {
		req.Prefix = "sprite"
	}

	zipBuf, err := slicer.Process(file, header.Filename, req)
	if err != nil {
		log.Printf("export error: %v", err)
		http.Error(w, "processing failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="export.zip"`)
	w.Write(zipBuf.Bytes())
}

func RemoveBGHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "missing image file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	tolerance := 30
	if t := r.FormValue("tolerance"); t != "" {
		if v, err := strconv.Atoi(t); err == nil && v >= 0 && v <= 255 {
			tolerance = v
		}
	}

	pngBuf, err := slicer.RemoveBackground(file, tolerance)
	if err != nil {
		log.Printf("remove-bg error: %v", err)
		http.Error(w, "processing failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(pngBuf.Bytes())
}

func ExportGIFHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "missing image file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	configStr := r.FormValue("config")
	if configStr == "" {
		http.Error(w, "missing config field", http.StatusBadRequest)
		return
	}

	var req models.GIFRequest
	if err := json.Unmarshal([]byte(configStr), &req); err != nil {
		http.Error(w, "invalid config JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Delay <= 0 {
		req.Delay = 10
	}

	gifBuf, err := slicer.GenerateGIF(file, req)
	if err != nil {
		log.Printf("export-gif error: %v", err)
		http.Error(w, "processing failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Content-Disposition", `attachment; filename="animation.gif"`)
	w.Write(gifBuf.Bytes())
}

func ResizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "missing image file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	srcImg, _, err := image.Decode(file)
	if err != nil {
		http.Error(w, "failed to decode image: "+err.Error(), http.StatusBadRequest)
		return
	}

	bounds := srcImg.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	var newWidth, newHeight int

	if widthStr := r.FormValue("width"); widthStr != "" {
		widthVal, err := strconv.Atoi(widthStr)
		if err != nil || widthVal <= 0 {
			http.Error(w, "invalid width value", http.StatusBadRequest)
			return
		}
		newWidth = widthVal
		newHeight = int(math.Round(float64(srcH) * float64(newWidth) / float64(srcW)))
	} else if pctStr := r.FormValue("percent"); pctStr != "" {
		pct, err := strconv.ParseFloat(pctStr, 64)
		if err != nil || pct <= 0 {
			http.Error(w, "invalid percent value", http.StatusBadRequest)
			return
		}
		newWidth = int(math.Round(float64(srcW) * pct / 100))
		newHeight = int(math.Round(float64(srcH) * pct / 100))
	} else {
		http.Error(w, "must provide either 'width' or 'percent' parameter", http.StatusBadRequest)
		return
	}

	if newWidth <= 0 || newHeight <= 0 {
		http.Error(w, "resulting dimensions are too small", http.StatusBadRequest)
		return
	}

	pngBuf, err := slicer.ResizeImage(srcImg, newWidth, newHeight)
	if err != nil {
		log.Printf("resize error: %v", err)
		http.Error(w, "resize failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", `attachment; filename="resized.png"`)
	w.Write(pngBuf.Bytes())
}
