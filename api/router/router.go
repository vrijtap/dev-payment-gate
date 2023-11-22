package router

import (
	"dev-payment-gate/api/handler"
	"net/http"

	"github.com/gorilla/mux"
)

// securityMiddleware is a middleware function that enhances the security of the web server
// by setting various security headers in the HTTP response. These headers help protect
// against common web vulnerabilities and improve the overall security posture of the
// web application.
func securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Feature-Policy", `
			accelerometer 'none';
			ambient-light-sensor 'none';
			autoplay 'none';
			camera 'none';
			document-domain 'none';
			encrypted-media 'none';
			execution-while-not-rendered 'none';
			execution-while-out-of-viewport 'none';
			fullscreen 'none';
			geolocation 'none';
			gyroscope 'none';
			magnetometer 'none';
			microphone 'none';
			midi 'none';
			picture-in-picture 'none';
			publickey-credentials-get 'none';
			screen-wake-lock 'none';
			sync-xhr 'none';
			usb 'none';
			vr 'none';
			wake-lock 'none';
			xr-spatial-tracking 'none';
		`)
		next.ServeHTTP(w, r)
	})
}

// Router creates a mux router to redirect requests to the correct handler
func Router() *mux.Router {
	// Create a new router
	router := mux.NewRouter()

	// Implement security headers
	router.Use(securityMiddleware)

	// Implement a static file server for the "/static/" path
	fileServer := http.FileServer(http.Dir("web/static/"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fileServer))

	// Implement routes
	router.HandleFunc("/transaction", handler.CreateTransaction).Methods(http.MethodPost)
	router.HandleFunc("/transaction/{transaction_id}", handler.PostTransaction).Methods(http.MethodPost)
	router.HandleFunc("/transaction/{transaction_id}", handler.GetTransactionHTML).Methods(http.MethodGet)
	router.HandleFunc("/transaction/js/{transaction_id}", handler.GetTransactionJS).Methods(http.MethodGet)

	// Custom NotFoundHandler for undefined routes
	router.NotFoundHandler = http.HandlerFunc(handler.NotAvailable)

	return router
}
