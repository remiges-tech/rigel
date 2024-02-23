package main

import (
	"context"
	"fmt"
	"net/http"

	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/remiges-tech/alya/service"
	"github.com/remiges-tech/alya/wscutils"
	"github.com/remiges-tech/logharbour/logharbour"
	"github.com/remiges-tech/rigel"
	"github.com/remiges-tech/rigel/etcd"
	"github.com/remiges-tech/rigel/server/configsvc"
	"github.com/remiges-tech/rigel/server/schemaserv"
	"github.com/remiges-tech/rigel/server/utils"
)

type AppConfig struct {
	EtcdHost      string `json:"etcd_host"`
	EtcdPort      string `json:"etcd_port"`
	AppServerPort string `json:"app_server_port"`
	APIPrefix     string `json:"api_prefix"`
}

// LoadConfigFromEnv updates AppConfig with values from environment variables if they exist
func LoadConfigFromEnv(appConfig *AppConfig) {
	if etcdHost := os.Getenv("ETCD_HOST"); etcdHost != "" {
		appConfig.EtcdHost = etcdHost
	}
	if etcdPort := os.Getenv("ETCD_PORT"); etcdPort != "" {
		appConfig.EtcdPort = etcdPort
	}
	if appServerPort := os.Getenv("APP_SERVER_PORT"); appServerPort != "" {
		appConfig.AppServerPort = appServerPort
	}
	if apiPrefix := os.Getenv("API_PREFIX"); apiPrefix != "" {
		appConfig.APIPrefix = apiPrefix
	}
}

func main() {
	var appConfig AppConfig

	// Override config with environment variables if they are set
	LoadConfigFromEnv(&appConfig)

	fmt.Printf("Config: %v", appConfig)

	// Logger setup
	logFile, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	fallbackWriter := logharbour.NewFallbackWriter(logFile, os.Stdout)
	lctx := logharbour.NewLoggerContext(logharbour.Info)
	l := logharbour.NewLogger(lctx, "rigel", fallbackWriter)

	// Open the error types file
	file, err := os.Open("./errortypes.yaml")
	if err != nil {
		log.Fatalf("Failed to open error types file: %v", err)
	}
	defer file.Close()
	// Load the error types
	wscutils.LoadErrorTypes(file)

	// Router
	r := gin.Default()
	// cordMiddleware() definition changes based on build flags
	// check middleware_dev.go and middleware_non_dev.go
	// use make commands to build or run when this middleware is used
	r.Use(corsMiddleware())

	// Create a new EtcdStorage instance
	etcdStorage, err := etcd.NewEtcdStorage([]string{fmt.Sprint(appConfig.EtcdHost + ":" + appConfig.EtcdPort)})

	if err != nil {
		log.Fatalf("Failed to create EtcdStorage: %v", err)
		wscutils.NewErrorResponse("Failed to create EtcdStorage")
		return
	}

	//Create a new Rigel instance
	rigelClient := rigel.NewWithStorage(etcdStorage)

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), utils.DIALTIMEOUT)
	defer cancel()

	// Get all keys from etcd
	allkeys, err := etcdStorage.GetWithPrefix(ctx, "/")
	if err != nil {
		log.Fatalf("etcd interaction failed: %v", err)
		return
	}

	// Build a Rigel STree
	rTree := utils.NewNode("")
	for k, v := range allkeys {
		rTree.AddPath(k, v)
	}

	// Services
	s := service.NewService(r).
		WithLogHarbour(l).
		WithDependency("appConfig", appConfig).
		WithDependency("rTree", rTree).
		WithDependency("etcd", etcdStorage).
		WithDependency("rigel", rigelClient)

	// routes

	apiV1Group := r.Group(appConfig.APIPrefix)

	// Config Services
	s.RegisterRouteWithGroup(apiV1Group, http.MethodGet, "/configget", configsvc.Config_get)
	s.RegisterRouteWithGroup(apiV1Group, http.MethodGet, "/configlist", configsvc.Config_list)
	s.RegisterRouteWithGroup(apiV1Group, http.MethodPost, "/configset", configsvc.Config_set)
	s.RegisterRouteWithGroup(apiV1Group, http.MethodPost, "/configupdate", configsvc.Config_update)

	// Schema Services
	s.RegisterRouteWithGroup(apiV1Group, http.MethodGet, "/getschema", schemaserv.HandleGetSchemaRequest)
	s.RegisterRouteWithGroup(apiV1Group, http.MethodGet, "/schemalist", schemaserv.HandleGetSchemaListRequest)

	r.Run(":" + appConfig.AppServerPort)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
