package admin_setup

import (
	"strings"

	"github.com/moisespsena-go/iolr"

	site_setup "github.com/ecletus-pkg/site-setup"
	"github.com/ecletus-pkg/user"
	"github.com/ecletus/auth"
	"github.com/ecletus/media/oss"
	"github.com/ecletus/notification"
	"github.com/ecletus/plug"
	"github.com/moisespsena-go/aorm"
	defaultlogger "github.com/moisespsena-go/default-logger"
	errwrap "github.com/moisespsena-go/error-wrap"
	path_helpers "github.com/moisespsena-go/path-helpers"
)

var log = defaultlogger.GetOrCreateLogger(path_helpers.GetCalledDir())

const AdminUser = "admin"

type Plugin struct {
	plug.EventDispatcher
	AuthKey, NotificationKey string
}

func (p *Plugin) RequireOptions() []string {
	keys := []string{p.AuthKey}
	if p.NotificationKey != "" {
		keys = append(keys, p.NotificationKey)
	}
	return keys
}

func (p *Plugin) OnRegister(options *plug.Options) {
	site_setup.OnRegister(p, func(e *site_setup.SiteSetupEvent) {
		e.SetupCMD.Flags().StringP("admin-email", "E", "", "E-mail for admin user")
		e.SetupCMD.Flags().StringP("admin-password", "P", "", "The Password. Use BLANK for generated password.")
	})

	site_setup.OnSetup(p, func(e *site_setup.SiteSetupEvent) (err error) {
		site := e.Site
		var adminUser user.User
		db := oss.IgnoreCallback(site.GetSystemDB().DB)
		err = db.First(&adminUser, "name = ?", user.AdminUserName).Error
		if aorm.IsRecordNotFoundError(err) {
			err = nil
			log.Info("Create System Administrator user")
			var (
				Auth         = e.Options().GetInterface(p.AuthKey).(*auth.Auth)
				Notification *notification.Notification
			)
			if p.NotificationKey != "" {
				Notification = e.Options().GetInterface(p.NotificationKey).(*notification.Notification)
			}
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
					if !strings.ContainsRune(adminEmail, '@') {
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
