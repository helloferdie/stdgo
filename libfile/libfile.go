package libfile

import (
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "image/jpeg" // Load img JPEG
	_ "image/png"  // Load img PNG
	"mime/multipart"
	"strings"

	"github.com/dchest/uniuri"
	"github.com/gabriel-vasile/mimetype"
	"github.com/helloferdie/stdgo/libnumber"
	"github.com/helloferdie/stdgo/libresponse"
	"github.com/helloferdie/stdgo/libslice"
)

// Config -
type Config struct {
	Var                string
	VarLocale          string
	CheckExtension     bool
	AllowExtension     []string
	CheckFileSize      bool
	MinFileSize        int64
	MaxFileSize        int64
	CheckImageSize     bool
	ImageWidthSize     int64
	MinImageWidthSize  int64
	MaxImageWidthSize  int64
	ImageHeightSize    int64
	MinImageHeightSize int64
	MaxImageHeightSize int64
	ImageAspectRatio   []float64
}

// CheckFile -
func CheckFile(fh *multipart.FileHeader, cfg Config) (bool, *libresponse.Default) {
	// Default config
	if cfg.Var == "" {
		cfg.Var = "file"
		cfg.VarLocale = "general.var_file"
	}

	res := libresponse.GetDefault()

	reader, err := fh.Open()
	defer reader.Close()

	if err != nil {
		// Error open file
		res.Code = 422
		res.Message = "general.error_validation"
		res.Error = "general.error_validation_open_file_var"
		res.ErrorVar = []interface{}{cfg.VarLocale}
		res.Data = map[string]interface{}{
			cfg.Var: map[string]interface{}{
				"error":     "general.error_validation_open_file",
				"error_var": []interface{}{},
			},
		}
		return false, res
	}

	if cfg.CheckExtension {
		// Check file extension
		mime, err := mimetype.DetectReader(reader)
		_, inSlice := libslice.Contains(mime.Extension(), cfg.AllowExtension)
		if err != nil || !inSlice {
			allowStr := "!" + strings.Join(cfg.AllowExtension[:], ", ")
			res.Code = 422
			res.Message = "general.error_validation"
			res.Error = "general.error_validation_upload_extension"
			res.ErrorVar = []interface{}{allowStr}
			res.Data = map[string]interface{}{
				cfg.Var: map[string]interface{}{
					"error":     "general.error_validation_upload_extension",
					"error_var": []interface{}{allowStr},
				},
			}
			return false, res
		}
	}

	if cfg.CheckFileSize {
		// Check file size
		if fh.Size < cfg.MinFileSize {
			sizeStr := "!" + libnumber.SuffixByte(float64(cfg.MinFileSize))
			res.Code = 422
			res.Message = "general.error_validation"
			res.Error = "general.error_validation_file_size_min"
			res.ErrorVar = []interface{}{sizeStr}
			res.Data = map[string]interface{}{
				cfg.Var: map[string]interface{}{
					"error":     "general.error_validation_file_size_min",
					"error_var": []interface{}{sizeStr},
				},
			}

			return false, res
		}

		if fh.Size > cfg.MaxFileSize {
			sizeStr := "!" + libnumber.SuffixByte(float64(cfg.MaxFileSize))
			res.Code = 422
			res.Message = "general.error_validation"
			res.Error = "general.error_validation_file_size_max"
			res.ErrorVar = []interface{}{sizeStr}
			res.Data = map[string]interface{}{
				cfg.Var: map[string]interface{}{
					"error":     "general.error_validation_file_size_max",
					"error_var": []interface{}{sizeStr},
				},
			}
			return false, res
		}
	}

	if cfg.CheckImageSize {
		reader, err := fh.Open()
		defer reader.Close()

		image, _, err := image.DecodeConfig(reader)
		if err != nil {
			// Try read image file
			res.Code = 422
			res.Message = "general.error_validation"
			res.Error = "general.error_validation_open_image_file_var"
			res.ErrorVar = []interface{}{cfg.VarLocale}
			res.Data = map[string]interface{}{
				cfg.Var: map[string]interface{}{
					"error":     "general.error_validation_open_file_image",
					"error_var": []interface{}{},
				},
			}
			return false, res
		}

		w := int64(image.Width)
		h := int64(image.Height)

		// Check image width
		if cfg.ImageWidthSize != 0 {
			sizeStr := fmt.Sprintf("!%v px", cfg.ImageWidthSize)
			res.Code = 422
			res.Message = "general.error_validation"
			res.Error = "general.error_validation_image_width"
			res.ErrorVar = []interface{}{sizeStr}
			res.Data = map[string]interface{}{
				cfg.Var: map[string]interface{}{
					"error":     "general.error_validation_image_width",
					"error_var": []interface{}{sizeStr},
				},
			}
			return false, res
		}
		if cfg.MinImageWidthSize != 0 && w < cfg.MinImageWidthSize {
			sizeStr := fmt.Sprintf("!%v px", cfg.MinImageWidthSize)
			res.Code = 422
			res.Message = "general.error_validation"
			res.Error = "general.error_validation_image_width_min"
			res.ErrorVar = []interface{}{sizeStr}
			res.Data = map[string]interface{}{
				cfg.Var: map[string]interface{}{
					"error":     "general.error_validation_image_width_min",
					"error_var": []interface{}{sizeStr},
				},
			}
			return false, res
		}
		if cfg.MaxImageWidthSize != 0 && w > cfg.MaxImageWidthSize {
			sizeStr := fmt.Sprintf("!%v px", cfg.MaxImageWidthSize)
			res.Code = 422
			res.Message = "general.error_validation"
			res.Error = "general.error_validation_image_width_max"
			res.ErrorVar = []interface{}{sizeStr}
			res.Data = map[string]interface{}{
				cfg.Var: map[string]interface{}{
					"error":     "general.error_validation_image_width_max",
					"error_var": []interface{}{sizeStr},
				},
			}
			return false, res
		}

		// Check image height
		if cfg.ImageHeightSize != 0 {
			sizeStr := fmt.Sprintf("!%v px", cfg.ImageHeightSize)
			res.Code = 422
			res.Message = "general.error_validation"
			res.Error = "general.error_validation_image_height"
			res.ErrorVar = []interface{}{sizeStr}
			res.Data = map[string]interface{}{
				cfg.Var: map[string]interface{}{
					"error":     "general.error_validation_image_height",
					"error_var": []interface{}{sizeStr},
				},
			}
			return false, res
		}
		if cfg.MinImageHeightSize != 0 && w < cfg.MinImageHeightSize {
			sizeStr := fmt.Sprintf("!%v px", cfg.MinImageHeightSize)
			res.Code = 422
			res.Message = "general.error_validation"
			res.Error = "general.error_validation_image_height_min"
			res.ErrorVar = []interface{}{sizeStr}
			res.Data = map[string]interface{}{
				cfg.Var: map[string]interface{}{
					"error":     "general.error_validation_image_height_min",
					"error_var": []interface{}{sizeStr},
				},
			}
			return false, res
		}
		if cfg.MaxImageHeightSize != 0 && w > cfg.MaxImageHeightSize {
			sizeStr := fmt.Sprintf("!%v px", cfg.MaxImageHeightSize)
			res.Code = 422
			res.Message = "general.error_validation"
			res.Error = "general.error_validation_image_height_max"
			res.ErrorVar = []interface{}{sizeStr}
			res.Data = map[string]interface{}{
				cfg.Var: map[string]interface{}{
					"error":     "general.error_validation_image_height_max",
					"error_var": []interface{}{sizeStr},
				},
			}
			return false, res
		}

		// Check image aspect ratio
		if len(cfg.ImageAspectRatio) > 0 {
			ratio := float64(w) / float64(h)
			for _, rt := range cfg.ImageAspectRatio {
				if rt != ratio {
					res.Code = 422
					res.Message = "general.error_validation"
					res.Error = "general.error_validation_image_ratio"
					res.ErrorVar = []interface{}{}
					res.Data = map[string]interface{}{
						cfg.Var: map[string]interface{}{
							"error":     "general.error_validation_image_ratio",
							"error_var": []interface{}{},
						},
					}
					return false, res
				}
			}
		}
	}
	return true, nil
}

// GenerateTemporaryFile -
func GenerateTemporaryFile(fh *multipart.FileHeader, name string) (bool, string, string) {
	// Open file
	src, err := fh.Open()
	defer src.Close()
	if err != nil {
		return false, "", ""
	}

	// Create empty file
	filename := uniuri.NewLen(8) + "_" + name
	if name == "" {
		mime, err := mimetype.DetectReader(src)
		if err == nil {
			filename = uniuri.NewLen(24) + mime.Extension()
		}
		src.Close()
		src, _ = fh.Open()
		defer src.Close()
	}
	path := DirectoryCreateTmp()
	dst, err := os.Create(path + "/" + filename)
	if err != nil {
		return false, "", ""
	}
	defer dst.Close()

	// Write empty file
	if _, err = io.Copy(dst, src); err != nil {
		return false, "", ""
	}
	return true, path, filename
}

// GenerateTemporaryFileRaw - Generate file no random filename
func GenerateTemporaryFileRaw(fh *multipart.FileHeader, name string) (bool, string, string) {
	// Open file
	src, err := fh.Open()
	defer src.Close()
	if err != nil {
		return false, "", ""
	}

	// Create empty file
	path := DirectoryCreateTmp()
	dst, err := os.Create(path + "/" + name)
	if err != nil {
		return false, "", ""
	}
	defer dst.Close()

	// Write empty file
	if _, err = io.Copy(dst, src); err != nil {
		return false, "", ""
	}
	return true, path, name
}

// DirectoryCreateTmp -
func DirectoryCreateTmp() string {
	path := os.Getenv("dir_root") + "/tmp"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0777)
	}
	path += "/" + strconv.FormatInt(time.Now().UnixNano()/1000000, 10)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0777)
	}
	return path
}

// HouseKeepingTmp - Clear TMP folder file
func HouseKeepingTmp() {
	path := os.Getenv("dir_root") + "/tmp"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0777)
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	expired := time.Now().UTC().Add(time.Hour * -6)
	for _, file := range files {
		if expired.After(file.ModTime().UTC()) {
			if file.IsDir() {
				os.RemoveAll(path + "/" + file.Name())
			} else {
				os.Remove(path + "/" + file.Name())
			}
		}
	}
}

// DownloadFile -
func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
