package libserver

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/helloferdie/stdgo/libserver/claim"

	"github.com/helloferdie/stdgo/libresponse"
	"github.com/helloferdie/stdgo/libslice"
	"github.com/helloferdie/stdgo/libstring"
	"github.com/helloferdie/stdgo/libvalidator"
	"github.com/helloferdie/stdgo/logger"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Host -
type Host struct {
	Echo *echo.Echo
}

// Initialize - initialize standard config
func Initialize(e *echo.Echo) {
	e.HTTPErrorHandler = ErrorHandler
	e.Use(middleware.Recover())
	e.Use(logger.EchoLogger)

	e.GET("/ping", Ping)
}

// StartHTTP - Start server in HTTP
func StartHTTP(svr *echo.Echo) {
	// Start server
	go func() {
		if err := svr.Start(":" + os.Getenv("port")); err != http.ErrServerClosed {
			logger.MakeLogEntry(nil, false).Error(err)
			logger.MakeLogEntry(nil, false).Error("Fail start HTTP server")
			logger.MakeLogEntry(nil, false).Error("Shutting down the server")
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := svr.Shutdown(ctx); err != nil {
		logger.MakeLogEntry(nil, false).Error("Fail shutting down server")
		os.Exit(1)
	} else {
		logger.MakeLogEntry(nil, false).Info("Shutdown HTTP server - done")
	}
}

// StartHTTPS - Start server in HTTPS
func StartHTTPS(svr *echo.Echo, svrInternal *echo.Echo) {
	sslCertificate := os.Getenv("ssl_certificate")
	sslKey := os.Getenv("ssl_key")

	var err error
	var cert []byte
	if cert, err = ioutil.ReadFile(sslCertificate); err != nil {
		logger.MakeLogEntry(nil, false).Error("Fail to read certificate file")
		os.Exit(1)
	}

	var key []byte
	if key, err = ioutil.ReadFile(sslKey); err != nil {
		logger.MakeLogEntry(nil, false).Error("Fail to read key file")
		os.Exit(1)
	}

	s := svr.TLSServer
	s.TLSConfig = new(tls.Config)
	s.TLSConfig.MinVersion = tls.VersionTLS12
	s.TLSConfig.CurvePreferences = []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256}
	s.TLSConfig.PreferServerCipherSuites = true
	s.TLSConfig.CipherSuites = []uint16{
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	}
	s.TLSConfig.Certificates = make([]tls.Certificate, 1)
	if s.TLSConfig.Certificates[0], err = tls.X509KeyPair(cert, key); err != nil {
		logger.MakeLogEntry(nil, false).Error("Fail to match certificate with key file")
		os.Exit(1)
	}
	s.Addr = ":" + os.Getenv("ssl_port")

	// Start redirect server
	svrRedir := echo.New()
	svrRedir.Pre(middleware.HTTPSRedirect())
	go func() {
		if err := svrRedir.Start(":" + os.Getenv("port")); err != nil {
			logger.MakeLogEntry(nil, false).Error(err)
			logger.MakeLogEntry(nil, false).Error("Fail start HTTPS redirect server")
			logger.MakeLogEntry(nil, false).Error("Shutting down server")
			os.Exit(1)
		}
	}()

	// Start HTTPS server
	go func() {
		//svr
		if err := svr.StartServer(s); err != http.ErrServerClosed {
			logger.MakeLogEntry(nil, false).Error(err)
			logger.MakeLogEntry(nil, false).Error("Shutting down server")
			os.Exit(1)
		}
	}()

	// Start internal http server
	if svrInternal != nil {
		go func() {
			if err := svrInternal.Start(":" + os.Getenv("ssl_port_internal")); err != http.ErrServerClosed {
				logger.MakeLogEntry(nil, false).Error(err)
				logger.MakeLogEntry(nil, false).Error("Fail start internal server")
				logger.MakeLogEntry(nil, false).Error("Shutting down server")
				os.Exit(1)
			}
		}()
	}

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := svrRedir.Shutdown(ctx); err != nil {
		logger.MakeLogEntry(nil, false).Error("Fail shutting down HTTPS redirect server")
		os.Exit(1)
	} else {
		logger.MakeLogEntry(nil, false).Info("Shutdown HTTPS redirect server - done")
	}

	if err := svrRedir.Shutdown(ctx); err != nil {
		logger.MakeLogEntry(nil, false).Error("Fail shutting down HTTPS server")
		os.Exit(1)
	} else {
		logger.MakeLogEntry(nil, false).Info("Shutdown HTTPS server - done")
	}

	if svrInternal != nil {
		if err := svrInternal.Shutdown(ctx); err != nil {
			logger.MakeLogEntry(nil, false).Error("Fail shutting down internal server")
			os.Exit(1)
		} else {
			logger.MakeLogEntry(nil, false).Info("Shutdown internal server - done")
		}
	}
	/*
		For Sub Domain
		sslDomain := os.Getenv("ssl_domain") + ":" + os.Getenv("port")
		hosts := map[string]*Host{}
		hosts[sslDomain] = &Host{svr}

		e1 := echo.New()
		e1.Any("/*", func(c echo.Context) (err error) {
			req := c.Request()
			res := c.Response()
			host := hosts[req.Host]
			if len(hosts) == 1 {
				host = hosts[sslDomain]
			}

			if host == nil {
				res := libresponse.GetDefault()
				res.Success = false
				res.Code = 404
				return Response(c, res)
			}
			host.Echo.ServeHTTP(res, req)
			return
		})

		// Start SSL server
		go func() {
			if err := e1.StartServer(s); err != nil {
				logger.MakeLogEntry(nil, false).Error("Shutting down the server")
				os.Exit(1)
			}
		}()

		// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := e1.Shutdown(ctx); err != nil {
			logger.MakeLogEntry(nil, false).Error("Fail start SSL server")
			logger.MakeLogEntry(nil, false).Error("Fail shutting down the server")
			os.Exit(1)
		}*/
}

// Response -
func Response(c echo.Context, res *libresponse.Default) (err error) {
	locale := os.Getenv("locale")
	if locale == "1" {
		res = ResponseLocale(res, c.Request().Header.Get("Accept-Language"))

		// Remove locale variable
		databytes, _ := json.Marshal(res)
		m := map[string]interface{}{}
		json.Unmarshal(databytes, &m)

		delete(m, "message_var")
		delete(m, "error_var")
		return c.JSON(int(res.Code), m)
	}
	return c.JSON(int(res.Code), res)
}

// ResponseProxy -
func ResponseProxy(res *http.Response) error {
	// Bypass process proxy response (response)
	if res.Header.Get("X-Proxy-Locale") == "0" {
		return nil
	}

	// Process proxy response (request)
	if res.Request.Header.Get("X-Locale") != "0" {
		resJSON := libresponse.GetDefault()
		resBody, errRead := ioutil.ReadAll(res.Body)
		if errRead == nil {
			errJSON := json.Unmarshal(resBody, &resJSON)
			if errJSON != nil {
				resJSON = libresponse.GetDefault()
				resJSON.Code = 502
				resJSON.Message = "general.error_gateway"
				resJSON.Error = "general.error_response_not_valid"
			}
		} else {
			resJSON.Code = 502
			resJSON.Message = "general.error_gateway"
			resJSON.Error = "general.error_response_read"
		}
		resJSON = ResponseLocale(resJSON, res.Request.Header.Get("Accept-Language"))

		databytes, _ := json.Marshal(resJSON)
		m := map[string]interface{}{}
		json.Unmarshal(databytes, &m)

		delete(m, "message_var")
		delete(m, "error_var")

		resByte, _ := json.Marshal(m)
		resLen := int64(len(string(resByte)))
		resLenStr := strconv.FormatInt(resLen, 10)
		res.Body = ioutil.NopCloser(bytes.NewBuffer(resByte))
		res.ContentLength = resLen
		res.Header.Set("Content-Length", resLenStr)
		res.Header.Set("Content-Type", "application/json;")
	}
	return nil
}

// ResponseLocale -
func ResponseLocale(data *libresponse.Default, lang string) *libresponse.Default {
	if data.Message != "" {
		if data.Message[0:1] == "!" {
			data.MessageLocale = libstring.Ucfirst(data.Message[1:])
		} else {
			data.MessageLocale = libstring.Ucfirst(LoadLocale(data.Message, lang, data.MessageVar))
		}
	}
	if data.Error != "" {
		if data.Error[0:1] == "!" {
			data.ErrorLocale = libstring.Ucfirst(data.Error[1:])
		} else {
			data.ErrorLocale = libstring.Ucfirst(LoadLocale(data.Error, lang, data.ErrorVar))
		}
	}

	if data.Code == 422 {
		jsonString, _ := json.Marshal(data.Data)
		listData := map[string]libvalidator.VarValidationError{}
		err := json.Unmarshal(jsonString, &listData)
		if err == nil && len(listData) > 0 {
			listNewData := map[string]interface{}{}
			for k, i := range listData {
				listNewData[k] = map[string]string{
					"error":        i.Error,
					"error_locale": libstring.Ucfirst(LoadLocale(i.Error, lang, i.ErrorVar)),
				}
			}
			data.Data = listNewData
		}
	}

	return data
}

// LoadLocale -
func LoadLocale(syntax string, lang string, params []interface{}) string {
	dirLocale := os.Getenv("dir_locale")
	split := strings.Split(syntax, ".")
	if len(split) < 2 {
		return LoadLocale("general.error_localization_syntax_not_valid", lang, []interface{}{syntax})
	}

	filename := "/" + split[0] + ".json"
	jsonFile, err := os.Open(dirLocale + "/" + lang + filename)
	if err != nil {
		jsonFile, err = os.Open(dirLocale + "/en" + filename)
		if err != nil {
			return LoadLocale("general.error_localization_file_not_found", lang, []interface{}{syntax})
		}
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	var result map[string]interface{}
	json.Unmarshal([]byte(byteValue), &result)

	val, exist := result[split[1]]
	if exist {
		localeParams := []interface{}{}
		escapeSyntax := []string{
			"general.error_localization_syntax_not_valid",
			"general.error_localization_syntax_not_found",
			"general.error_localization_file_not_found",
		}

		for _, v := range params {
			t, ok := v.(string)
			if !ok {
				localeParams = append(localeParams, v)
			} else {
				if t != "" && t[0:1] == "!" {
					localeParams = append(localeParams, t[1:])
				} else {
					_, inSlice := libslice.Contains(syntax, escapeSyntax)
					if inSlice {
						localeParams = append(localeParams, t)
					} else {
						localeParams = append(localeParams, LoadLocale(t, lang, nil))
					}
				}
			}
		}
		return fmt.Sprintf(val.(string), localeParams...)
	}
	if lang != "en" && syntax == "general.error_localization_syntax_not_found" {
		lang = "en"
	}
	return LoadLocale("general.error_localization_syntax_not_found", lang, []interface{}{syntax})
}

// RequestLog -
func RequestLog(c echo.Context, stdout bool) error {
	// Read the Body content
	var bodyBytes []byte
	if c.Request().Body != nil {
		bodyBytes, _ = ioutil.ReadAll(c.Request().Body)
	}

	err := ioutil.WriteFile(os.Getenv("log_dir")+"/log_response.log", bodyBytes, 0777)
	if err == nil {
		// Print out
		if stdout {
			fmt.Println(string(bodyBytes))
		}

		// Restore the io.ReadCloser to its original state
		c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	return err
}

// FormatOutputDefault -
func FormatOutputDefault(c echo.Context) map[string]interface{} {
	m := map[string]interface{}{
		"tz":     "UTC",
		"header": nil,
	}

	loc, err := time.LoadLocation(c.Request().Header.Get("Accept-TZ"))
	if err == nil {
		m["tz"] = loc.String()
	}

	tmp := new(libresponse.Header)
	tmp.Authorization = c.Request().Header.Get("Authorization")
	tmp.AcceptLanguage = c.Request().Header.Get("Accept-Language")
	tmp.AcceptTimezone = c.Request().Header.Get("Accept-TZ")
	tmp.RealIP = GetRealIP(c)
	tmp.RequestURL = c.Request().URL.String()
	claims := claim.GetJWTClaims(c)
	if claims != nil {
		tmp.Claims = claims
		v, ok := claims["access"].(string)
		if ok {
			tmp.Access = v
			tmp.IsLogin = true
			tmp.UserID = int64(claims["user_id"].(float64))
		}
	}
	m["header"] = tmp
	return m
}

// GetRealIP -
func GetRealIP(c echo.Context) string {
	if c != nil {
		ip := c.Request().Header.Get("X-Real-Ip")
		if ip == "" {
			ip = c.Request().RemoteAddr
		}
		return ip
	}
	return ""
}
