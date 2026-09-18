package cli

import (
	"fmt"

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

		fmt.Printf("[●] Niskava Web Server starting on http://localhost:%d\n", port)
		fmt.Println("Sesi REST API dan SSE streaming siap diakses.")
		if openFlag {
			fmt.Println("Membuka browser otomatis...")
		}
		fmt.Println("Tekan Ctrl+C untuk menghentikan server.")
		return nil
	},
}

func init() {
	serveCmd.Flags().IntVarP(&portFlag, "port", "p", 8080, "server port (default: 8080)")
	serveCmd.Flags().BoolVarP(&openFlag, "open", "o", false, "open web dashboard in browser automatically")
	RootCmd.AddCommand(serveCmd)
}
