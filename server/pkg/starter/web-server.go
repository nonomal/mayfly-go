package starter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mayfly-go/pkg/gox"
	"mayfly-go/pkg/i18n"
	"mayfly-go/pkg/logx"
	"mayfly-go/pkg/middleware"
	"mayfly-go/pkg/req"
	"mayfly-go/pkg/utils/collx"
	"mayfly-go/pkg/utils/jsonx"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func runWebServer(ctx context.Context, serverConfig ServerConf, options *Options) error {
	// 设置gin日志输出器
	logOut := logx.GetConfig().GetLogOut()
	gin.DefaultErrorWriter = logOut
	gin.DefaultWriter = logOut

	gin.SetMode(serverConfig.Model)

	// i18n配置
	i18n.SetLang(serverConfig.Lang)

	var router = gin.New()
	// 最大请求体大小限制 100M
	router.MaxMultipartMemory = 100 << 20
	// 初始化接口路由
	initRouter(router, req.RouterConfig{ContextPath: serverConfig.ContextPath})
	// 设置静态资源
	setStatic(router, serverConfig, options.StaticRouter)
	if options != nil && options.OnRoutesReady != nil {
		options.OnRoutesReady(router)
	}

	// 是否允许跨域
	if serverConfig.Cors {
		router.Use(middleware.Cors())
	}

	srv := http.Server{
		Addr: serverConfig.GetPort(),
		// 注册路由
		Handler: router,
	}

	go func() {
		defer gox.Recover()
		<-ctx.Done()
		logx.Info("Shutdown HTTP Server ...")
		timeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := srv.Shutdown(timeout)
		if err != nil {
			logx.Errorf("Failed to Shutdown HTTP Server: %v", err)
		}

		Terminate()
	}()

	logx.Infof("Listening and serving HTTP on %s", srv.Addr+serverConfig.ContextPath)
	var err error
	if serverConfig.TLS.Enable {
		err = srv.ListenAndServeTLS(serverConfig.TLS.CertFile, serverConfig.TLS.KeyFile)
	} else {
		err = srv.ListenAndServe()
	}

	if errors.Is(err, http.ErrServerClosed) {
		logx.Info("HTTP Server Shutdown")
		return nil
	}

	if err != nil {
		logx.Errorf("Failed to Start HTTP Server: %v", err)
	}

	return err
}

func setStatic(router *gin.Engine, serverConfig ServerConf, staticRouter *StaticRouter) {
	contextPath := serverConfig.ContextPath

	// 设置静态资源
	for _, scs := range serverConfig.Statics {
		router.StaticFS(scs.RelativePath, http.Dir(scs.Root))
	}
	// 设置静态文件
	for _, sfs := range serverConfig.StaticFiles {
		router.StaticFile(sfs.RelativePath, sfs.Filepath)
	}

	if staticRouter == nil {
		return
	}

	fileServer := http.FileServer(http.FS(staticRouter.Fs))
	handler := WrapStaticHandler(http.StripPrefix(contextPath, fileServer))
	for _, p := range staticRouter.Paths {
		router.GET(contextPath+p, handler)
	}

	// Vue History 模式支持
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 排除 API
		if strings.HasPrefix(path, contextPath+"/api/") {
			c.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"msg":  fmt.Sprintf("not found '%s:%s'", c.Request.Method, c.Request.URL.Path)},
			)
			return
		}

		// 尝试读取文件
		filePath := strings.TrimPrefix(path, contextPath)
		if filePath == "" || !isStaticResource(path) {
			filePath = "/index.html"
		}

		file, err := http.FS(staticRouter.Fs).Open(filePath)
		if err != nil {
			// 文件不存在，返回 index.html
			file, err = http.FS(staticRouter.Fs).Open("/index.html")
			if err != nil {
				c.Status(http.StatusNotFound)
				return
			}
		}
		defer file.Close()
		c.Status(http.StatusOK)

		if strings.HasSuffix(filePath, "index.html") {
			data, err := io.ReadAll(file)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}

			// 注入配置项到 index.html 中
			configJSON := jsonx.ToStr(collx.M{
				"CTX_PATH": contextPath,
			})

			injectScript := fmt.Sprintf(
				`<script>window.__APP_CONFIG__ = %s;</script>`,
				string(configJSON),
			)

			// 替换<app-config />
			html := strings.Replace(
				string(data),
				"<app-config />",
				injectScript+"\n",
				1,
			)
			// 替换base path为 contextPath
			if contextPath != "" {
				html = strings.Replace(html, `<base href="/" />`, `<base href="`+contextPath+`/" />`, 1)
			}

			c.Header("Content-Type", "text/html; charset=utf-8")
			c.Writer.Write([]byte(html))
			return
		}

		http.ServeContent(c.Writer, c.Request, filePath, time.Now(), file)
	})
}

func WrapStaticHandler(h http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Cache-Control", `public, max-age=31536000`)
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// isStaticResource 判断是否为静态资源文件
func isStaticResource(path string) bool {
	staticExtensions := []string{
		".js", ".css", ".png", ".jpg", ".jpeg", ".gif",
		".svg", ".ico", ".woff", ".woff2", ".ttf", ".eot",
		".json", ".xml", ".txt",
	}

	for _, ext := range staticExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}
