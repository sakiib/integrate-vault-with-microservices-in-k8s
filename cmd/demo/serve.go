package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sakiib/integrate-vault-with-micriservices-in-k8s/api"
	"github.com/sakiib/integrate-vault-with-micriservices-in-k8s/conn"

	"github.com/go-chi/chi"
	"github.com/sakiib/integrate-vault-with-micriservices-in-k8s/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start API server",
	Long:  `Start the HTTP API server of popup service`,
	Run:   serve,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		config.Init()

		// connect database
		if err := conn.ConnectDB(); err != nil {
			return fmt.Errorf("Cant't connect database: %v", err)
		}

		return nil
	},
}

func init() {
	serveCmd.PersistentFlags().IntP("http_port", "p", 8080, "port on which the server will listen for http")
	viper.BindPFlag("app.http_port", serveCmd.PersistentFlags().Lookup("http_port"))
	rootCmd.AddCommand(serveCmd)
}

func serve(_ *cobra.Command, _ []string) {
	if err := conn.ConnectDB(); err != nil {
		log.Fatal(err)
	}

	// free the connection when the server is closed
	defer conn.CloseDB()

	appCfg := config.App()

	// app intrruption signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT)

	r := chi.NewMux()
	r.Mount("/", api.Router())
	srv := &http.Server{
		ReadTimeout:  appCfg.ReadTimeout,
		WriteTimeout: appCfg.WriteTimeout,
		IdleTimeout:  appCfg.IdleTimeout,
		Addr:         ":" + strconv.Itoa(appCfg.HTTPPort),
		Handler:      r,
	}
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()

	go func() {
		log.Println("HTTP: Listening on port " + strconv.Itoa(appCfg.HTTPPort))
		log.Fatal(srv.ListenAndServe())
	}()

	<-stop

	log.Println("Shutting down server...")
	log.Println("Server shutteddown gracefully")
}
