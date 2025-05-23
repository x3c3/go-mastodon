package mastodon

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestStreamingWSUser(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wsMock))
	defer ts.Close()

	client := NewClient(&Config{Server: ts.URL}).NewWSClient()
	ctx, cancel := context.WithCancel(context.Background())
	q, err := client.StreamingWSUser(ctx)
	if err != nil {
		t.Fatalf("should not be fail: %v", err)
	}

	wsTest(t, q, cancel)
}

func TestStreamingWSPublic(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wsMock))
	defer ts.Close()

	client := NewClient(&Config{Server: ts.URL}).NewWSClient()
	ctx, cancel := context.WithCancel(context.Background())
	q, err := client.StreamingWSPublic(ctx, false)
	if err != nil {
		t.Fatalf("should not be fail: %v", err)
	}

	wsTest(t, q, cancel)
}

func TestStreamingWSHashtag(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wsMock))
	defer ts.Close()

	client := NewClient(&Config{Server: ts.URL}).NewWSClient()
	ctx, cancel := context.WithCancel(context.Background())
	q, err := client.StreamingWSHashtag(ctx, "zzz", true)
	if err != nil {
		t.Fatalf("should not be fail: %v", err)
	}
	wsTest(t, q, cancel)

	ctx, cancel = context.WithCancel(context.Background())
	q, err = client.StreamingWSHashtag(ctx, "zzz", false)
	if err != nil {
		t.Fatalf("should not be fail: %v", err)
	}
	wsTest(t, q, cancel)
}

func TestStreamingWSList(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wsMock))
	defer ts.Close()

	client := NewClient(&Config{Server: ts.URL}).NewWSClient()
	ctx, cancel := context.WithCancel(context.Background())
	q, err := client.StreamingWSList(ctx, "123")
	if err != nil {
		t.Fatalf("should not be fail: %v", err)
	}
	wsTest(t, q, cancel)
}

func TestStreamingWSDirect(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wsMock))
	defer ts.Close()

	client := NewClient(&Config{Server: ts.URL}).NewWSClient()
	ctx, cancel := context.WithCancel(context.Background())
	q, err := client.StreamingWSDirect(ctx)
	if err != nil {
		t.Fatalf("should not be fail: %v", err)
	}
	wsTest(t, q, cancel)
}

func wsMock(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/streaming" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	u := websocket.Upgrader{}
	conn, err := u.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	err = conn.WriteMessage(websocket.TextMessage,
		[]byte(`{"event":"update","payload":"{\"content\":\"foo\"}"}`))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage,
		[]byte(`{"event":"status.update","payload":"{\"content\":\"bar\"}"}`))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage,
		[]byte(`{"event":"notification","payload":"{\"id\":123}"}`))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage,
		[]byte(`{"event":"delete","payload":1234567}`))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage,
		[]byte(`{"event":"conversation","payload":"{\"id\":819516}"}`))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage,
		[]byte(`{"event":"update","payload":"<html></html>"}`))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	time.Sleep(10 * time.Second)
}

func wsTest(t *testing.T, q chan Event, cancel func()) {
	time.AfterFunc(time.Second, func() {
		cancel()
	})
	events := []Event{}
	for e := range q {
		events = append(events, e)
	}
	if len(events) != 8 {
		t.Fatalf("result should be 8: %d", len(events))
	}
	if events[0].(*UpdateEvent).Status.Content != "foo" {
		t.Fatalf("want %q but %q", "foo", events[0].(*UpdateEvent).Status.Content)
	}
	if events[1].(*UpdateEditEvent).Status.Content != "bar" {
		t.Fatalf("want %q but %q", "bar", events[1].(*UpdateEditEvent).Status.Content)
	}
	if events[2].(*NotificationEvent).Notification.ID != "123" {
		t.Fatalf("want %q but %q", "123", events[2].(*NotificationEvent).Notification.ID)
	}
	if events[3].(*DeleteEvent).ID != "1234567" {
		t.Fatalf("want %q but %q", "1234567", events[3].(*DeleteEvent).ID)
	}
	if events[4].(*ConversationEvent).Conversation.ID != "819516" {
		t.Fatalf("want %q but %q", "819516", events[4].(*ConversationEvent).Conversation.ID)
	}
	if errorEvent, ok := events[5].(*ErrorEvent); !ok {
		t.Fatalf("should be fail: %v", errorEvent.Err)
	}
	if errorEvent, ok := events[6].(*ErrorEvent); !ok {
		t.Fatalf("should be fail: %v", errorEvent.Err)
	}
	if errorEvent, ok := events[6].(*ErrorEvent); !ok {
		t.Fatalf("should be fail: %v", errorEvent.Err)
	}
	if errorEvent, ok := events[7].(*ErrorEvent); !ok {
		t.Fatalf("should be fail: %v", errorEvent.Err)
	}
}

func TestStreamingWS(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wsMock))
	defer ts.Close()

	client := NewClient(&Config{Server: ":"}).NewWSClient()
	_, err := client.StreamingWSPublic(context.Background(), true)
	if err == nil {
		t.Fatalf("should be fail: %v", err)
	}

	client = NewClient(&Config{Server: ts.URL}).NewWSClient()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	q, err := client.StreamingWSPublic(ctx, true)
	if err != nil {
		t.Fatalf("should not be fail: %v", err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		e := <-q
		if errorEvent, ok := e.(*ErrorEvent); !ok {
			t.Errorf("should be fail: %v", errorEvent.Err)
		}
	}()
	wg.Wait()
}

func TestHandleWS(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := websocket.Upgrader{}
		conn, err := u.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		err = conn.WriteMessage(websocket.TextMessage,
			[]byte(`<html></html>`))
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		time.Sleep(10 * time.Second)
	}))
	defer ts.Close()

	q := make(chan Event)
	client := NewClient(&Config{}).NewWSClient()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		e := <-q
		if errorEvent, ok := e.(*ErrorEvent); !ok {
			t.Errorf("should be fail: %v", errorEvent.Err)
		}
	}()
	err := client.handleWS(context.Background(), ":", q)
	if err == nil {
		t.Fatalf("should be fail: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	wg.Add(1)
	go func() {
		defer wg.Done()
		e := <-q
		if errorEvent, ok := e.(*ErrorEvent); !ok {
			t.Errorf("should be fail: %v", errorEvent.Err)
		}
	}()
	err = client.handleWS(ctx, "ws://"+ts.Listener.Addr().String(), q)
	if err == nil {
		t.Fatalf("should be fail: %v", err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		e := <-q
		if errorEvent, ok := e.(*ErrorEvent); !ok {
			t.Errorf("should be fail: %v", errorEvent.Err)
		}
	}()
	client.handleWS(context.Background(), "ws://"+ts.Listener.Addr().String(), q)

	wg.Wait()
}

func TestDialRedirect(t *testing.T) {
	client := NewClient(&Config{}).NewWSClient()
	_, err := client.dialRedirect(":")
	if err == nil {
		t.Fatalf("should be fail: %v", err)
	}
}

func TestDial(t *testing.T) {
	canErr := true
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if canErr {
			canErr = false
			http.Redirect(w, r, ":", http.StatusMovedPermanently)
			return
		}

		http.Redirect(w, r, "http://www.example.com/", http.StatusMovedPermanently)
	}))
	defer ts.Close()

	client := NewClient(&Config{}).NewWSClient()
	_, _, err := client.dial(":")
	if err == nil {
		t.Fatalf("should be fail: %v", err)
	}

	_, _, err = client.dial("ws://" + ts.Listener.Addr().String())
	if err == nil {
		t.Fatalf("should be fail: %v", err)
	}

	_, rawurl, err := client.dial("ws://" + ts.Listener.Addr().String())
	if err != nil {
		t.Fatalf("should not be fail: %v", err)
	}
	if rawurl != "ws://www.example.com/" {
		t.Fatalf("want %q but %q", "ws://www.example.com/", rawurl)
	}
}

func TestChangeWebSocketScheme(t *testing.T) {
	_, err := changeWebSocketScheme(":")
	if err == nil {
		t.Fatalf("should be fail: %v", err)
	}

	u, err := changeWebSocketScheme("http://example.com/")
	if err != nil {
		t.Fatalf("should not be fail: %v", err)
	}
	if u.Scheme != "ws" {
		t.Fatalf("want %q but %q", "ws", u.Scheme)
	}

	u, err = changeWebSocketScheme("https://example.com/")
	if err != nil {
		t.Fatalf("should not be fail: %v", err)
	}
	if u.Scheme != "wss" {
		t.Fatalf("want %q but %q", "wss", u.Scheme)
	}
}
