package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// Register routes
	rt.router.GET("/", rt.getHelloWorld)
	rt.router.GET("/context", rt.wrap(rt.getContextReply))

	rt.router.POST("/session", rt.wrap(rt.doLogin))

	rt.router.GET("/users", rt.wrap(rt.searchUser))
	rt.router.PUT("/users/me/username", rt.wrap(rt.setMyUserName))
	rt.router.PUT("/users/me/pic", rt.wrap(rt.setMyPhoto))

	rt.router.GET("/conversations", rt.wrap(rt.getMyConversations))
	rt.router.POST("/conversations", rt.wrap(rt.startConversation))
	rt.router.GET("/conversations/:conversation_id", rt.wrap(rt.getConversation))

	rt.router.POST("/groups", rt.wrap(rt.createGroup))
    rt.router.GET("/groups/:group_id", rt.wrap(rt.getGroup))
	rt.router.PUT("/groups/:group_id/name", rt.wrap(rt.setGroupName))
	rt.router.PUT("/groups/:group_id/photo", rt.wrap(rt.setGroupPhoto))
	rt.router.POST("/groups/:group_id/members", rt.wrap(rt.addToGroup))
	rt.router.DELETE("/groups/:group_id/members", rt.wrap(rt.leaveGroup))

	rt.router.POST("/conversations/:conversation_id/messages/:message_id/forward", rt.wrap(rt.forwardMessage))
	rt.router.POST("/conversations/:conversation_id/messages", rt.wrap(rt.sendMessage))
	rt.router.DELETE("/conversations/:conversation_id/messages/:message_id", rt.wrap(rt.deleteMessage))

	rt.router.POST("/conversations/:conversation_id/messages/:message_id/comments", rt.wrap(rt.commentMessage))
	rt.router.DELETE("/conversations/:conversation_id/messages/:message_id/comments/:comment_id", rt.wrap(rt.uncommentMessage))
	rt.router.GET("/conversations/:conversation_id/messages/:message_id/comments", rt.wrap(rt.getComments))

	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	return rt.router
}
