package handler

import (
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/pkg/errors"

	m "gitlab.com/falqon/inovantapp/backend/models"
	"gitlab.com/falqon/inovantapp/backend/service/user"
)

// AuthHandler service handler authentication
type AuthHandler struct {
	signin     func(email, password string) (*user.AuthResponse, error)
	pwdReset   func(resetID, verification, password string) error
	pwdRecover func(email string) error
}

// EmailLogin returns an echo handler
// @Summary auth.login
// @Description Login with email & password
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param credentials body handler.loginForm true "Email to login with"
// @Success 200 {object} handler.loginOut
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /auth/signin [post]
func (handler *AuthHandler) EmailLogin(c echo.Context) error {
	request := loginForm{}
	err := c.Bind(&request)
	if err != nil {
		return err
	}

	err = request.Validate()
	if err != nil {
		return c.JSON(http.StatusUnauthorized, errorResponse{
			Error: generalError{
				Code:    401,
				Message: err.Error(),
			},
		})
	}
	r, err := handler.signin(request.Email, request.Password)
	if err != nil {
		return errors.Wrap(err, "Failed to sign in")
	}
	return c.JSON(http.StatusOK, loginResponse{
		dataResponse: dataResponse{Context: c.QueryParam("context")},
		Data: loginOut{
			Kind: "authToken",
			Item: authToken{
				User: r.User,
				JWT:  r.Jwt,
			},
		},
	})
}

// PasswordRecover returns an echo handler
// @Summary auth.PasswordReset
// @Description Start password recovery for account with given email
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param email body handler.passwordVerifyForm true "Email to send reset password link"
// @Success 200 {object} handler.authTokenResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /auth/password-recover [post]
func (handler *AuthHandler) PasswordRecover(c echo.Context) error {
	req := passwordVerifyForm{}
	err := c.Bind(&req)
	if err != nil {
		return err
	}
	err = handler.pwdRecover(req.Email)
	if err != nil {
		return errors.Wrap(err, "Failed to recover password")
	}
	return c.JSON(http.StatusOK, authRecoverResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: recoverResponse{
			Kind: "Email",
			Item: req.Email,
		},
	})
}

// PasswordReset returns an echo handler
// @Summary auth.PasswordReset
// @Description Reset password for account with given email
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param resetID path string true "reset id"
// @Param email body handler.passwordResetForm true "Information to reset form"
// @Success 200 {object} handler.authResetResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /auth/password-reset/:resetID [post]
func (handler *AuthHandler) PasswordReset(c echo.Context) error {
	resetID := c.Param("resetID")
	req := passwordResetForm{}
	err := c.Bind(&req)
	if err != nil {
		return err
	}
	err = handler.pwdReset(resetID, req.Verification, req.Password)
	if err != nil {
		return errors.Wrap(err, "Failed to reset password")
	}
	return c.JSON(http.StatusOK, authResetResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: recoverResponse{
			Kind: "empty",
		},
	})
}

type passwordVerifyForm struct {
	Email string `json:"email" binding:"required" example:"user@example.com"`
}

type passwordResetForm struct {
	Verification string `json:"verification" binding:"required" example:"$2y$12$F/npgjvknmHkNvDck15aeew..."`
	Password     string `json:"password" binding:"required" example:"password"`
}

type loginForm struct {
	Email    string `json:"email" binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"mypassword123"`
}

type loginResponse struct {
	dataResponse
	Data loginOut `json:"data"`
}

type loginOut struct {
	singleItemData
	Item authToken `json:"item"`
	Kind string    `json:"kind" example:"authToken"`
}

type authRecoverResponse struct {
	dataResponse
	Data recoverResponse `json:"data"`
}
type authResetResponse struct {
	dataResponse
	Data recoverResponse `json:"data"`
}
type authTokenResponse struct {
	dataResponse
	Data loginResponse `json:"data"`
}

type recoverResponse struct {
	singleItemData
	Kind string `json:"kind" example:"Email"`
	Item string `json:"item"`
}
type resetResponse struct {
	singleItemData
	Kind string `json:"kind" example:"empty"`
	Item string `json:"item" example:""`
}

type authToken struct {
	// User auth data
	User m.User `json:"user"`
	// JWT token
	JWT string `json:"jwt" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0ZXN0IjoidGVzdCJ9.MZZ7UbJRJH9hFRdBUQHpMjU4TK4XRrYP5UxcAkEHvxE."`
}

func (f *loginForm) Validate() error {
	if len(strings.TrimSpace(f.Email)) == 0 || len(strings.TrimSpace(f.Password)) == 0 {
		return errors.New("Login information cannot be empty")
	}
	return nil
}
