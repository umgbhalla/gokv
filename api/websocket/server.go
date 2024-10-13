package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/umgbhalla/gokv/internal/query"
	"github.com/umgbhalla/gokv/internal/store"
)

type Server struct {
	store    *store.Store
	query    *query.Query
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
}

func NewServer(store *store.Store, query *query.Query) *Server {
	return &Server{
		store: store,
		query: query,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients: make(map[*websocket.Conn]bool),
	}
}

func (s *Server) Start(addr string) error {
	http.HandleFunc("/ws", s.handleWebSocket)
	return http.ListenAndServe(addr, nil)
}

func (s *Server) Shutdown(ctx context.Context) error {

	for conn := range s.clients {
		conn.Close()
	}
	return nil
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	s.clients[conn] = true
	defer delete(s.clients, conn)

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			return
		}
		if messageType == websocket.TextMessage {
			s.handleMessage(conn, p)
		}
	}
}

func (s *Server) handleMessage(conn *websocket.Conn, message []byte) {
	var request map[string]interface{}
	if err := json.Unmarshal(message, &request); err != nil {
		s.sendError(conn, "Invalid JSON")
		return
	}

	action, ok := request["action"].(string)
	if !ok {
		s.sendError(conn, "Missing or invalid 'action' field")
		return
	}

	switch action {
	case "get":
		s.handleGet(conn, request["key"].(string))
	case "set":
		s.handleSet(conn, request["key"].(string), request["value"], request["ttl"])
	case "delete":
		s.handleDelete(conn, request["key"].(string))
	case "query":
		s.handleQuery(conn, request["query"].(string))
	default:
		s.sendError(conn, "Unknown action")
	}
}

func (s *Server) handleGet(conn *websocket.Conn, key string) {
	value, ok := s.store.Get(key)
	if !ok {
		s.sendError(conn, "Key not found")
		return
	}
	s.sendResponse(conn, map[string]interface{}{"action": "get", "key": key, "value": value})
}

func (s *Server) handleSet(conn *websocket.Conn, key string, value interface{}, ttl interface{}) {
	var duration time.Duration
	if ttl != nil {
		if ttlFloat, ok := ttl.(float64); ok {
			duration = time.Duration(ttlFloat) * time.Second
		}
	}

	if err := s.store.Set(key, value, duration); err != nil {
		s.sendError(conn, "Error setting value")
		return
	}
	s.sendResponse(conn, map[string]interface{}{"action": "set", "key": key, "status": "ok"})
}

func (s *Server) handleDelete(conn *websocket.Conn, key string) {
	if err := s.store.Delete(key); err != nil {
		s.sendError(conn, "Error deleting key")
		return
	}
	s.sendResponse(conn, map[string]interface{}{"action": "delete", "key": key, "status": "ok"})
}

func (s *Server) handleQuery(conn *websocket.Conn, queryString string) {
	result, err := s.query.Execute(queryString)
	if err != nil {
		s.sendError(conn, "Error executing query")
		return
	}
	s.sendResponse(conn, map[string]interface{}{"action": "query", "result": result})
}

func (s *Server) sendError(conn *websocket.Conn, message string) {
	s.sendResponse(conn, map[string]interface{}{"error": message})
}

func (s *Server) sendResponse(conn *websocket.Conn, response interface{}) {
	if err := conn.WriteJSON(response); err != nil {
		log.Println("WebSocket write error:", err)
	}
}
