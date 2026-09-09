package cmd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/vgaro/yotocli/internal/config"
	"github.com/vgaro/yotocli/pkg/yoto"
)

// loginTimeout bounds how long we wait for the user to finish signing in.
const loginTimeout = 5 * time.Minute

// callbackResult carries what the browser redirect delivered to our local server.
type callbackResult struct {
	code  string
	state string
	err   error
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Yoto",
	Long: `Signs in to your Yoto account using the OAuth2 authorization code flow
with PKCE. Your browser is opened to the Yoto login page and this CLI briefly
listens on ` + yoto.RedirectURI + ` to receive the result.

The callback URL above must be registered as an Allowed Callback URL for your
application at https://dashboard.yoto.dev/.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		clientID := config.GetClientID()
		if clientID == "" {
			printSetupInstructions()
			fmt.Print("Please enter your Client ID: ")
			fmt.Scanln(&clientID)
			if clientID == "" {
				return fmt.Errorf("client ID is required to authenticate")
			}
		}

		// Use a temporary client for auth (no token needed yet)
		client := yoto.NewClient("", clientID)

		pkce, err := yoto.NewPKCE()
		if err != nil {
			return err
		}

		state, err := yoto.NewState()
		if err != nil {
			return err
		}

		// Bind before opening the browser so we never send the user to a login
		// page whose redirect has nowhere to land.
		listener, err := net.Listen("tcp", yoto.CallbackAddr)
		if err != nil {
			return fmt.Errorf("failed to listen on %s: %w (is another 'yoto login' running?)", yoto.CallbackAddr, err)
		}
		defer listener.Close()

		resultsC := make(chan callbackResult, 1)
		server := &http.Server{Handler: callbackHandler(state, resultsC)}

		// Closing doneC after the send lets the cleanup below receive
		// unconditionally: it gets Serve's error, or nil if the select already
		// took it.
		doneC := make(chan error, 1)
		go func() {
			doneC <- server.Serve(listener)
			close(doneC)
		}()

		// Shut down on every exit path, and wait for Serve to return so we never
		// leave the goroutine or an in-flight response behind. The grace period
		// lets the browser finish receiving the "you can close this tab" page.
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			if err := server.Shutdown(shutdownCtx); err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not shut down callback server: %v\n", err)
			}

			// Shutdown makes Serve return ErrServerClosed, which is the expected
			// outcome here rather than a problem worth reporting.
			if err := <-doneC; err != nil && !errors.Is(err, http.ErrServerClosed) {
				fmt.Fprintf(os.Stderr, "warning: callback server stopped with: %v\n", err)
			}
		}()

		authURL := client.AuthorizeURLFor(pkce.Challenge, state)

		fmt.Println("Opening your browser to sign in to Yoto...")
		if err := openBrowser(authURL); err != nil {
			fmt.Println("Could not open a browser automatically.")
		}
		fmt.Printf("\nIf your browser did not open, visit this URL:\n  %s\n\n", authURL)
		fmt.Println("Waiting for you to authorize...")

		var result callbackResult
		select {
		case result = <-resultsC:
		case err := <-doneC:
			return fmt.Errorf("callback server unexpectedly exited before authorization completed: %w", err)
		case <-time.After(loginTimeout):
			return fmt.Errorf("timed out after %s waiting for authorization", loginTimeout)
		}

		if result.err != nil {
			return result.err
		}

		tokenResp, err := client.ExchangeCode(result.code, pkce.Verifier)
		if err != nil {
			return err
		}

		if tokenResp.RefreshToken == "" {
			fmt.Println("Warning: no refresh token was issued. Check that the 'offline_access' scope is enabled for your application.")
		}

		fmt.Println("Successfully authenticated!")

		// Save tokens and client ID
		config.SetToken(tokenResp.AccessToken, tokenResp.RefreshToken)
		config.SetClientID(clientID)

		if err := config.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println("Credentials saved to config file.")
		return nil
	},
}

// printSetupInstructions tells a first-time user how to register the Yoto
// application this CLI signs in to.
func printSetupInstructions() {
	fmt.Println("No Yoto Client ID found in config, so you need an application registered with Yoto.")
	// The dashboard root rather than /applications/new: it lists any application
	// you have already set up, so you can take the Client ID from that one
	// instead of registering a second.
	fmt.Println("Find an existing one, or create a new one, at: https://dashboard.yoto.dev/")
	fmt.Println("\nA new application needs these settings:")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "  Application Type\t%s\n", "Public Client")
	fmt.Fprintf(w, "  Allowed Callback URLs\t%s\n", yoto.RedirectURI)
	fmt.Fprintf(w, "  Allowed Logout URLs\t%s\n", "(leave blank)")
	fmt.Fprintf(w, "  Application Logo\t%s\n", "(leave blank)")
	fmt.Fprintf(w, "  Privacy Policy URL\t%s\n", "(leave blank)")
	fmt.Fprintf(w, "  Scopes\t%s\n", "at least the following")
	for _, scope := range yoto.APIScopes {
		fmt.Fprintf(w, "  \t  %s\n", scope)
	}
	w.Flush()

	fmt.Println("\nThe callback URL must match exactly - this CLI listens on it during login.")
	fmt.Println("Selecting extra scopes causes no problems; the list above is the minimum needed.")
	fmt.Println("\nOnce the application is created, copy its Client ID below.")
	fmt.Println()
}

// callbackHandler serves the loopback redirect, reporting the authorization
// code (or the error Yoto sent back) on resultsC.
func callbackHandler(wantState string, resultsC chan<- callbackResult) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc(yoto.CallbackPath, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		send := func(res callbackResult) {
			// Only the first callback matters; later ones (e.g. a refreshed tab)
			// must not block on an unread channel.
			select {
			case resultsC <- res:
			default:
			}
		}

		if errCode := q.Get("error"); errCode != "" {
			desc := q.Get("error_description")
			http.Error(w, fmt.Sprintf("Authorization failed: %s\n\n%s", errCode, desc), http.StatusBadRequest)
			send(callbackResult{err: fmt.Errorf("authorization failed: %s: %s", errCode, desc)})
			return
		}

		code := q.Get("code")
		if code == "" {
			http.Error(w, "Missing authorization code.", http.StatusBadRequest)
			send(callbackResult{err: fmt.Errorf("callback did not include an authorization code")})
			return
		}

		if got := q.Get("state"); got != wantState {
			http.Error(w, "State mismatch.", http.StatusBadRequest)
			send(callbackResult{err: fmt.Errorf("state mismatch in callback: this login attempt was not the one that started")})
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, "<html><body><h2>Signed in to Yoto</h2><p>You can close this tab and return to your terminal.</p></body></html>")
		send(callbackResult{code: code, state: q.Get("state")})
	})

	return mux
}

// openBrowser tries to open url in the user's default browser.
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler"}
	default:
		cmd = "xdg-open"
	}

	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
