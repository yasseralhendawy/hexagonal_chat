package ginserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	chatmessage "github.com/yasseralhendawy/hexagonal_chat/internal/application/chat_message"
	userchat "github.com/yasseralhendawy/hexagonal_chat/internal/application/user_chat"
	gorillasocket "github.com/yasseralhendawy/hexagonal_chat/pkg/websocket/gorilla_socket"
)

type ChatHandler struct {
	Server *GinServer
	CMApp  *chatmessage.App
	UCApp  *userchat.App
}

func (h ChatHandler) Run(addr ...string) error {
	err := h.setupWebSocketHandlers()
	if err != nil {
		return err
	}
	h.WebSocketHandler()

	h.GetChat()

	h.GetUserHistory()
	h.CreatNewChat()
	h.AddParticipants()
	h.LeaveChat()

	return h.Server.Run(addr...)
}

func (h *ChatHandler) WebSocketHandler() {
	h.Server.Engin.GET("ws", func(ctx *gin.Context) {
		claimData, err := getClaims(ctx)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		conn, err := h.Server.websocketManager.Upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			ctx.AbortWithError(http.StatusNotFound, err)
			return
		}
		ids, err := h.UCApp.Domain.Storage.GetUserChatsIDS(claimData.UserID)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, nil)

			return
		}
		h.Server.websocketManager.Serve(conn, claimData.UserID, ids)
	})
}

func (h *ChatHandler) setupWebSocketHandlers() error {
	if h.Server.websocketManager == nil {
		return errors.New("websocket server not found")
	}
	h.Server.websocketManager.Handlers["send_message"] = func(message *gorillasocket.Message, c *gorillasocket.Client) error {
		return h.sendMessage(message, c)
	}
	h.Server.websocketManager.Handlers["edit_message"] = func(message *gorillasocket.Message, c *gorillasocket.Client) error {
		return h.editMessage(message, c)
	}

	return nil
}

func (s *ChatHandler) sendMessage(message *gorillasocket.Message, c *gorillasocket.Client) error {
	var req chatmessage.AddMessgaeRequest

	err := json.Unmarshal(message.Payload, &req)
	if err != nil {
		return err
	}
	res, err := s.CMApp.AddMesage(c.ClientID, req)
	if err != nil {
		return err
	}
	domainMessage := res.Messages[len(res.Messages)-1]
	room := c.Server.GetChat(domainMessage.ChatID)
	domainData, err := domainMessage.Marshal()
	if err != nil {
		return err
	}
	msg := gorillasocket.Message{
		Type:    "new_message",
		Payload: domainData,
	}
	data, err := msg.Marshal()
	if err != nil {
		return err
	}
	room.Broadcast <- &data
	return nil
}

func (s *ChatHandler) editMessage(message *gorillasocket.Message, c *gorillasocket.Client) error {
	var req chatmessage.EditMessgaeRequest

	err := json.Unmarshal(message.Payload, &req)
	if err != nil {
		return err
	}
	res, err := s.CMApp.EditMessageText(c.ClientID, req)
	if err != nil {
		return err
	}
	domainMessage, err := res.GetMessage(req.MessageID)
	if err != nil {
		return err
	}
	room := c.Server.GetChat(domainMessage.ChatID)
	domainData, err := domainMessage.Marshal()
	if err != nil {
		return err
	}
	msg := gorillasocket.Message{
		Type:    "edit_message",
		Payload: domainData,
	}
	data, err := msg.Marshal()
	if err != nil {
		return err
	}
	room.Broadcast <- &data
	return nil
}

func (h ChatHandler) GetChat() {
	h.Server.Engin.GET("get_chat", func(ctx *gin.Context) {
		_, err := getClaims(ctx)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		req := new(chatmessage.GetMessageRequest)

		err = ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		res, err := h.CMApp.GetChat(*req)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.JSON(http.StatusOK, res)
	})
}

func (h ChatHandler) GetUserHistory() {
	h.Server.Engin.GET("get_user_history", func(ctx *gin.Context) {
		claim, err := getClaims(ctx)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		res, err := h.UCApp.GetUserHistory(claim.UserID)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.JSON(http.StatusOK, res)
	})
}

func (h ChatHandler) CreatNewChat() {
	h.Server.Engin.POST("create_new_chat", func(ctx *gin.Context) {

		claim, err := getClaims(ctx)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		req := new(userchat.CreateChatReq)

		err = ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		res, err := h.UCApp.CreatNewChat(req, claim.UserID)
		if err != nil {
			ctx.JSON(http.StatusConflict, err.Error())
			return
		}
		err = h.addParticipantsToWebsocketRoomChat(req.Participants, res.ChatID)
		if err != nil {
			// log here
		}

		ctx.JSON(http.StatusOK, res)
	})
}

func (h ChatHandler) AddParticipants() {
	h.Server.Engin.POST("add_participants", func(ctx *gin.Context) {

		_, err := getClaims(ctx)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		req := new(userchat.AddParticipantsRequest)

		err = ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		res, err := h.UCApp.AddParticipants(req)
		if err != nil {
			ctx.JSON(http.StatusConflict, err.Error())
			return
		}

		err = h.addParticipantsToWebsocketRoomChat(req.Participants, req.ChatID)
		if err != nil {
			// Optionally log the error here
		}
		ctx.JSON(http.StatusOK, res)
	})
}

func (h ChatHandler) addParticipantsToWebsocketRoomChat(participants []string, chatId string) error {
	room := h.Server.websocketManager.GetChat(chatId)
	msg := gorillasocket.Message{
		Type:    "joined_chat",
		Payload: fmt.Append([]byte{}, participants, " joined the chat"),
	}
	data, err := msg.Marshal()
	if err != nil {
		return err
	}
	for _, uID := range participants {
		user := h.Server.websocketManager.GetClient(uID)
		if user != nil {
			user.AddChat(chatId)
		}
		room.AddParticipant(uID, user)

	}
	room.Broadcast <- &data

	return nil
}

func (h ChatHandler) LeaveChat() {
	h.Server.Engin.POST("leave_chat", func(ctx *gin.Context) {

		claim, err := getClaims(ctx)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		req := new(userchat.LeaveChatRequest)

		err = ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		res, err := h.UCApp.LeaveChat(claim.UserID, req)
		if err != nil {
			ctx.JSON(http.StatusConflict, err.Error())
			return
		}
		room := h.Server.websocketManager.GetChat(req.ChatID)
		msg := gorillasocket.Message{
			Type:    "left_chat",
			Payload: fmt.Append([]byte{}, claim.UserID, " left the chat"),
		}
		data, err := msg.Marshal()
		if err != nil {
			// Optionally log the error here
			return
		}
		room.RemoveParticipant(claim.UserID)

		room.Broadcast <- &data
		ctx.JSON(http.StatusOK, res)
	})
}
