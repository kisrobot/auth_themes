package clean

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/kisrobot/auth"
	"github.com/kisrobot/auth/claims"
	"github.com/kisrobot/auth/providers/password"
	"github.com/kisrobot/i18n"
	"github.com/kisrobot/i18n/backends/yaml"
	"github.com/kisrobot/qor"
	"github.com/kisrobot/qor/utils"
	"github.com/kisrobot/render"
)

// ErrPasswordConfirmationNotMatch password confirmation not match error
var ErrPasswordConfirmationNotMatch = errors.New("password confirmation doesn't match password")

// New initialize clean theme
func New(config *auth.Config) *auth.Auth {
	if config == nil {
		config = &auth.Config{}
	}
	config.ViewPaths = append(config.ViewPaths, "github.com/kisrobot/auth_themes/clean/views")

	if config.DB == nil {
		fmt.Print("Please configure *gorm.DB for Auth theme clean")
	}

	if config.Render == nil {
		yamlBackend := yaml.New()
		I18n := i18n.New(yamlBackend)
		for _, gopath := range append([]string{filepath.Join(utils.AppRoot, "vendor")}, utils.GOPATH()...) {
			filePath := filepath.Join(gopath, "src", "github.com/kisrobot/auth_themes/clean/locales/en-US.yml")
			if content, err := ioutil.ReadFile(filePath); err == nil {
				translations, _ := yamlBackend.LoadYAMLContent(content)
				for _, translation := range translations {
					I18n.AddTranslation(translation)
				}
				break
			}
		}

		config.Render = render.New(&render.Config{
			FuncMapMaker: func(render *render.Render, req *http.Request, w http.ResponseWriter) template.FuncMap {
				return template.FuncMap{
					"t": func(key string, args ...interface{}) template.HTML {
						return I18n.T(utils.GetLocale(&qor.Context{Request: req}), key, args...)
					},
				}
			},
		})
	}

	Auth := auth.New(config)

	Auth.RegisterProvider(password.New(&password.Config{
		Confirmable: true,
		RegisterHandler: func(context *auth.Context) (*claims.Claims, error) {
			context.Request.ParseForm()

			if context.Request.Form.Get("confirm_password") != context.Request.Form.Get("password") {
				return nil, ErrPasswordConfirmationNotMatch
			}

			return password.DefaultRegisterHandler(context)
		},
	}))

	if Auth.Config.DB != nil {
		// Migrate Auth Identity model
		Auth.Config.DB.AutoMigrate(Auth.Config.AuthIdentityModel)
	}
	return Auth
}
