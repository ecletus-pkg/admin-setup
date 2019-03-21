package admin_setup

import (
	"strings"

	"github.com/moisespsena-go/iolr"

	"github.com/aghape-pkg/site-setup"
	"github.com/aghape-pkg/user"
	"github.com/aghape/auth"
	"github.com/aghape/media/oss"
	"github.com/aghape/notification"
	"github.com/aghape/plug"
	"github.com/moisespsena-go/aorm"
	"github.com/moisespsena/go-default-logger"
	"github.com/moisespsena/go-error-wrap"
	"github.com/moisespsena/go-path-helpers"
)

var log = defaultlogger.NewLogger(path_helpers.GetCalledDir())

type Plugin struct {
	plug.EventDispatcher
	AuthKey, NotificationKey string
}

func (p *Plugin) RequireOptions() []string {
	return []string{p.AuthKey, p.NotificationKey}
}

func (p *Plugin) OnRegister() {
	site_setup.OnRegister(p, func(e *site_setup.SiteSetupEvent) {
		e.SetupCMD.Flags().StringP("admin-email", "E", "", "E-mail for admin user")
		e.SetupCMD.Flags().StringP("admin-password", "P", "", "The Password. Use BLANK for generated password.")
	})

	site_setup.OnSetup(p, func(e *site_setup.SiteSetupEvent) (err error) {
		site := e.Site
		var adminUser user.User
		db := oss.IgnoreCallback(site.GetSystemDB().DB)
		err = db.First(&adminUser, "name = ?", "admin").Error
		if err != nil && aorm.IsRecordNotFoundError(err) {
			err = nil
			log.Info("Create System Administrator user")
			var (
				Auth         = e.Options().GetInterface(p.AuthKey).(*auth.Auth)
				Notification = e.Options().GetInterface(p.NotificationKey).(*notification.Notification)
			)
			return user.CreateAdminUserIfNotExists(site, Auth, Notification, func() (string, error) {
				adminEmail, err := e.SetupCMD.Flags().GetString("admin-email")
				if err != nil {
					return "", errwrap.Wrap(err, "Get admin-email flag")
				}
				for adminEmail == "" {
					adminEmail, err = iolr.STDMessageLR.ReadS("Enter the email address for admin user")
					if err != nil {
						return "", errwrap.Wrap(err, "Get admin-email from STDIN")
					}
					if !strings.Contains(adminEmail, "@") {
						log.Errorf("The %q isn't valid mail address. Try now.", adminEmail)
						adminEmail = ""
						continue
					}
					break
				}
				return adminEmail, nil
			}, func() (string, error) {
				value, err := e.SetupCMD.Flags().GetString("admin-password")
				if err != nil {
					return "", errwrap.Wrap(err, "Get admin-password flag")
				}
				return value, nil
			})
		}
		return nil
	})
}
