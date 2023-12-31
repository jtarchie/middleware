package middleware_test

import (
	"fmt"
	"log/slog"

	"github.com/imroc/req/v3"
	"github.com/jtarchie/middleware"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("LoggingMiddleware", func() {
	var (
		buffer     *gbytes.Buffer
		serverHTTP *ghttp.Server
	)

	BeforeEach(func() {
		buffer = gbytes.NewBuffer()
		logger := slog.New(slog.NewJSONHandler(buffer, nil))

		router := echo.New()
		router.Use(middleware.Logger(logger))
		router.GET("/", func(c echo.Context) error {
			//nolint: wrapcheck
			return c.String(200, "abcd")
		})
		router.GET("/empty", func(c echo.Context) error {
			//nolint: wrapcheck
			return c.NoContent(200)
		})
		router.GET("/error", func(c echo.Context) error {
			return fmt.Errorf("some error message") //nolint:goerr113
		})

		serverHTTP = ghttp.NewServer()
		serverHTTP.AppendHandlers(router.ServeHTTP)
	})

	AfterEach(func() {
		serverHTTP.Close()
	})

	When("and error occurs", func() {
		It("logs the error message", func() {
			_, err := req.C().R().
				Get(fmt.Sprintf("%s/error", serverHTTP.URL()))
			Expect(err).NotTo(HaveOccurred())

			Eventually(buffer).Should(gbytes.Say(`"error":"some error message"`))
		})
	})

	When("no content is returned", func() {
		It("sets a default value", func() {
			_, err := req.C().R().
				Get(fmt.Sprintf("%s/empty", serverHTTP.URL()))
			Expect(err).NotTo(HaveOccurred())

			Eventually(buffer).Should(gbytes.Say(`"bytes_out":"0"`))
		})
	})

	When("no request ID is provided", func() {
		It("generates a request ID", func() {
			_, err := req.C().R().
				Get(serverHTTP.URL())
			Expect(err).NotTo(HaveOccurred())

			Eventually(buffer).Should(gbytes.Say(`"level":"INFO"`))
			Eventually(buffer).Should(gbytes.Say(`"msg":"http_request"`))

			Eventually(buffer).Should(gbytes.Say(`"bytes_in":"0"`))
			Eventually(buffer).Should(gbytes.Say(`"bytes_out":"4"`))
			Eventually(buffer).Should(gbytes.Say(`"id":"[0-9a-fA-F]{8}-[[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[[0-9a-fA-F]{4}-[0-9a-fA-F]{12}"`))
			Eventually(buffer).Should(gbytes.Say(`"latency_human":"\d+`))
			Eventually(buffer).Should(gbytes.Say(`"method":"GET"`))
			Eventually(buffer).Should(gbytes.Say(`"remote_ip":"\d+\.\d+\.\d+\.\d+"`))
			Eventually(buffer).Should(gbytes.Say(`"status":200`))
			Eventually(buffer).Should(gbytes.Say(`"time":"\d+-\d+-\d+T`))
		})
	})

	When("a request ID is provided", func() {
		It("uses that one", func() {
			_, err := req.C().R().
				SetHeader("X-Request-Id", "unique-id").
				Get(serverHTTP.URL())
			Expect(err).NotTo(HaveOccurred())

			Eventually(buffer).Should(gbytes.Say(`"level":"INFO"`))
			Eventually(buffer).Should(gbytes.Say(`"msg":"http_request"`))
			Eventually(buffer).Should(gbytes.Say(`"id":"unique-id"`))
			Eventually(buffer).Should(gbytes.Say(`"status":200`))
		})
	})
})
