package handler

import (
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"gitlab.com/falqon/inovantapp/backend/service/chat"
	"gitlab.com/falqon/inovantapp/backend/service/user/auth"
)

// ChatHandler returns the chat wbsocket handler
type ChatHandler struct {
	db           *sqlx.DB
	hub          *chat.Hub
	claimsCtxKey string
	rolesCtxKey  string
}

// Chat returns an echo handler
// @Summary socket.chat
// @Description Socket endpoint for chat
// @Accept  json
// @Produce  json
// @Router /api/connect [get]
func (handler *ChatHandler) Chat(c echo.Context) error {
	claims, err := auth.Extract(c.Get(handler.claimsCtxKey))
	if err != nil {
		return errors.Wrap(err, "Couldn't parse token")
	}

	chat.ServeWs(handler.hub, claims.UserID, c.Response(), c.Request())
	return nil
}
