package daemon

import (
	"net/http"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/ory/x/flagx"
	"github.com/ory/x/healthx"

	"github.com/gorilla/context"
	"github.com/spf13/cobra"
	"github.com/urfave/negroni"

	"github.com/ory/graceful"
	"github.com/ory/x/metricsx"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/logout"
	"github.com/ory/kratos/selfservice/flow/profile"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/selfservice/strategy/password"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

func servePublic(d driver.Driver, wg *sync.WaitGroup, cmd *cobra.Command, args []string) {
	defer wg.Done()

	c := d.Configuration()
	l := d.Logger()
	n := negroni.New()
	r := d.Registry()

	router := x.NewRouterPublic()
	r.LoginHandler().RegisterPublicRoutes(router)
	r.RegistrationHandler().RegisterPublicRoutes(router)
	r.LogoutHandler().RegisterPublicRoutes(router)
	r.ProfileManagementHandler().RegisterPublicRoutes(router)
	r.LoginStrategies().RegisterPublicRoutes(router)
	r.RegistrationStrategies().RegisterPublicRoutes(router)
	r.SessionHandler().RegisterPublicRoutes(router)
	r.SelfServiceErrorHandler().RegisterPublicRoutes(router)
	r.SchemaHandler().RegisterPublicRoutes(router)
	r.HealthHandler().SetRoutes(router.Router, false)

	n.Use(NewNegroniLoggerMiddleware(l.(*logrus.Logger), "public#"+c.SelfPublicURL().String()))
	n.Use(sqa(cmd, d))

	r.WithCSRFHandler(x.NewCSRFHandler(
		router,
		r.Writer(),
		l,
		c.SelfPublicURL().Path,
		c.SelfPublicURL().Hostname(),
		!flagx.MustGetBool(cmd, "dev"),
	))
	n.UseHandler(
		r.CSRFHandler(),
	)
	server := graceful.WithDefaults(&http.Server{
		Addr:    c.PublicListenOn(),
		Handler: context.ClearHandler(n),
	})

	l.Printf("Starting the public httpd on: %s", server.Addr)
	if err := graceful.Graceful(server.ListenAndServe, server.Shutdown); err != nil {
		l.Fatalln("Failed to gracefully shutdown public httpd")
	}
	l.Println("Public httpd was shutdown gracefully")
}

func serveAdmin(d driver.Driver, wg *sync.WaitGroup, cmd *cobra.Command, args []string) {
	defer wg.Done()

	c := d.Configuration()
	l := d.Logger()
	n := negroni.New()
	r := d.Registry()

	router := x.NewRouterAdmin()
	r.RegistrationHandler().RegisterAdminRoutes(router)
	r.LoginHandler().RegisterAdminRoutes(router)
	r.ProfileManagementHandler().RegisterAdminRoutes(router)
	r.IdentityHandler().RegisterAdminRoutes(router)
	r.SessionHandler().RegisterAdminRoutes(router)
	r.HealthHandler().SetRoutes(router.Router, true)
	r.SelfServiceErrorHandler().RegisterAdminRoutes(router)

	n.Use(NewNegroniLoggerMiddleware(l.(*logrus.Logger), "admin#"+c.SelfAdminURL().String()))
	n.Use(sqa(cmd, d))

	n.UseHandler(router)
	server := graceful.WithDefaults(&http.Server{
		Addr:    c.AdminListenOn(),
		Handler: context.ClearHandler(n),
	})

	l.Printf("Starting the admin httpd on: %s", server.Addr)
	if err := graceful.Graceful(server.ListenAndServe, server.Shutdown); err != nil {
		l.Fatalln("Failed to gracefully shutdown admin httpd")
	}
	l.Println("Admin httpd was shutdown gracefully")
}

func sqa(cmd *cobra.Command, d driver.Driver) *metricsx.Service {
	// Creates only one instance
	return metricsx.New(
		cmd,
		d.Logger(),
		&metricsx.Options{
			Service: "ory-kratos",
			ClusterID: metricsx.Hash(
				strings.Join([]string{
					d.Configuration().DSN(),
					d.Configuration().SelfPublicURL().String(),
					d.Configuration().SelfAdminURL().String(),
				}, "|"),
			),
			IsDevelopment: flagx.MustGetBool(cmd, "dev"),
			WriteKey:      "qQlI6q8Q4WvkzTjKQSor4sHYOikHIvvi",
			WhitelistedPaths: []string{
				"/",
				healthx.AliveCheckPath,
				healthx.ReadyCheckPath,
				healthx.VersionPath,
				"/auth/methods/oidc/",
				password.RegistrationPath,
				password.LoginPath,
				oidc.BasePath,
				login.BrowserLoginPath,
				login.BrowserLoginRequestsPath,
				logout.BrowserLogoutPath,
				registration.BrowserRegistrationPath,
				registration.BrowserRegistrationRequestsPath,
				session.SessionsWhoamiPath,
				identity.IdentitiesPath,
				profile.PublicProfileManagementPath,
				profile.AdminBrowserProfileRequestPath,
				profile.PublicProfileManagementPath,
				profile.PublicProfileManagementUpdatePath,
				errorx.ErrorsPath,
			},
			BuildVersion: d.Registry().BuildVersion(),
			BuildHash:    d.Registry().BuildHash(),
			BuildTime:    d.Registry().BuildDate(),
		},
	)
}

func ServeAll(d driver.Driver) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		var wg sync.WaitGroup
		wg.Add(2)
		go servePublic(d, &wg, cmd, args)
		go serveAdmin(d, &wg, cmd, args)
		wg.Wait()
	}
}
