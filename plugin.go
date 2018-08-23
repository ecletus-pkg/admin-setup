package admin_setup

import (
	"strings"

	"github.com/aghape-pkg/site-setup"
	"github.com/aghape-pkg/user"
	"github.com/aghape/auth"
	"github.com/aghape/media"
	"github.com/aghape/notification"
	"github.com/aghape/plug"
	"github.com/moisespsena/go-default-logger"
	"github.com/moisespsena/go-error-wrap"
	"github.com/moisespsena/go-ioutil"
	"github.com/moisespsena/go-path-helpers"
)

var log = defaultlogger.NewLogger(path_helpers.GetCalledDir())

type Plugin struct {
	plug.EventDispatcher
	AuthKey string
}

func (p *Plugin) RequireOptions() []string {
	return []string{p.AuthKey}
}

func (p *Plugin) OnRegister() {
	site_setup.OnRegister(p, func(e *site_setup.SiteSetupEvent) {
		e.SetupCMD.Flags().String("admin-email", "", "E-mail for admin user")
	})

	site_setup.OnSetup(p, func(e *site_setup.SiteSetupEvent) error {
		site := e.Site
		var adminUser user.User
		db := media.IgnoreCallback(site.GetSystemDB().DB)
		db.First(&adminUser, "name = ?", "admin")
		if adminUser.ID == "" {
			log.Info("Create System Administrator user")
			var (
				Auth         = e.Options().GetInterface(p.AuthKey).(*auth.Auth)
				Notification = notification.New(&notification.Config{})
			)
			return user.CreateAdminUserIfNotExists(site, Auth, Notification, func() (string, error) {
				adminEmail, err := e.SetupCMD.Flags().GetString("admin-email")
				if err != nil {
					return "", errwrap.Wrap(err, "Get admin-mail flag")
				}
				for adminEmail == "" {
					line, err := ioutil.STDStringMessageLR.Read("Enter the mail address for admin user")
					if err != nil {
						return "", errwrap.Wrap(err, "Get admin-email from STDIN")
					}
					adminEmail = string(line)
					if !strings.Contains(adminEmail, "@") {
						log.Errorf("The %q isn't valid mail address. Try now.", adminEmail)
						adminEmail = ""
						continue
					}
					break
				}
				return adminEmail, nil
			})
		}
		return nil
	})
}
