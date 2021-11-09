package api

import (
	"github.com/allentom/haruka"
	"github.com/gorilla/websocket"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
)

var WebsocketLogger = logrus.New().WithField("scope", "websocket")
var DefaultNotificationManager = NotificationManager{
	Conns: map[string]*NotificationConnection{},
}
var (
	EventTaskStatus = "TaskStatus"
)

type NotificationConnection struct {
	Id         string
	Uid        string
	Connection *websocket.Conn
	Logger     *logrus.Entry
	isClose    bool
}

type NotificationManager struct {
	Conns map[string]*NotificationConnection
	sync.Mutex
}

func (m *NotificationManager) addConnection(conn *websocket.Conn, uid string) *NotificationConnection {
	m.Lock()
	defer m.Unlock()
	id := xid.New().String()
	notification := &NotificationConnection{
		Connection: conn,
		Logger: WebsocketLogger.WithFields(logrus.Fields{
			"id":  id,
			"uid": uid,
		}),
		Uid: uid,
		Id:  id,
	}
	conn.SetCloseHandler(func(code int, text string) error {
		notification.isClose = true
		return nil
	})
	m.Conns[id] = notification
	return m.Conns[id]
}
func (m *NotificationManager) removeConnection(id string) {
	m.Lock()
	defer m.Unlock()
	delete(m.Conns, id)
}
func (m *NotificationManager) sendJSONToAll(data interface{}) {
	m.Lock()
	defer m.Unlock()
	for _, notificationConnection := range m.Conns {
		if notificationConnection.isClose {
			continue
		}
		err := notificationConnection.Connection.WriteJSON(data)
		if err != nil {
			notificationConnection.Logger.Error(err)
		}
	}
}
func (m *NotificationManager) sendJSONToUser(data interface{}, uid string) {
	m.Lock()
	defer m.Unlock()
	for _, notificationConnection := range m.Conns {
		if notificationConnection.Uid == uid && !notificationConnection.isClose {
			err := notificationConnection.Connection.WriteJSON(data)
			if err != nil {
				notificationConnection.Logger.Error(err)
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var notificationSocketHandler haruka.RequestHandler = func(context *haruka.Context) {

	c, err := upgrader.Upgrade(context.Writer, context.Request, nil)
	if err != nil {
		WebsocketLogger.Error(err)
		return
	}
	notifier := DefaultNotificationManager.addConnection(c, context.Param["uid"].(string))
	defer func() {
		DefaultNotificationManager.removeConnection(notifier.Id)
		c.Close()
	}()
	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, 1005, 1000) {
				notifier.Logger.Error(err)
			}
			break
		}
	}
}
