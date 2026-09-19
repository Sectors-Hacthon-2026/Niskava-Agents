package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/internal/server"
	"github.com/spf13/cobra"
)

var (
	portFlag int
	openFlag bool
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start local web workspace server (REST & SSE streaming)",
	RunE: func(cmd *cobra.Command, args []string) error {
		port := portFlag
		if port == 0 {
			port = cfg.Server.Port
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		srv, err := server.Start(ctx, port, appDB)
		if err != nil {
			return fmt.Errorf("failed to start server: %w", err)
		}

		fmt.Printf("[●] Niskava Web Server active at %s\n", srv.URL)
		fmt.Println("Sesi REST API dan SSE streaming siap diakses.")
		fmt.Println("Tekan Ctrl+C untuk menghentikan server.")

		if openFlag {
			fmt.Printf("Membuka Web Workspace di browser: %s\n", srv.URL)
			_ = server.OpenBrowser(srv.URL)
		}

		<-ctx.Done()
		fmt.Println("\nMenghentikan server daemon...")
		return nil
	},
}

func init() {
	serveCmd.Flags().IntVarP(&portFlag, "port", "p", 8080, "server port (default: 8080)")
	serveCmd.Flags().BoolVarP(&openFlag, "open", "o", false, "open web dashboard in browser automatically")
	RootCmd.AddCommand(serveCmd)
}
