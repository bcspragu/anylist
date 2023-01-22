package anylist

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/bcspragu/anylist/pb"
	"github.com/google/uuid"
	"golang.org/x/net/publicsuffix"
	"golang.org/x/net/websocket"
	"google.golang.org/protobuf/proto"
)

type Client struct {
	id string

	email    string
	password string

	refreshToken string
	accessToken  string

	client *http.Client

	// Initialized on login
	signedUserID string
	userID       string
}

func FromRefreshToken(rTkn string) (*Client, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, fmt.Errorf("failed to init cookie jar: %w", err)
	}

	id := uuid.NewString()
	c := &Client{
		id:           id,
		refreshToken: rTkn,
		client: &http.Client{
			Jar:       jar,
			Transport: &roundTripper{id: id},
		},
	}

	if err := c.refresh(); err != nil {
		return nil, fmt.Errorf("failed to refresh: %w", err)
	}

	return c, nil
}

func New(email, password string) (*Client, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, fmt.Errorf("failed to init cookie jar: %w", err)
	}

	id := uuid.NewString()
	rt := &roundTripper{id: id}
	c := &Client{
		id:       id,
		email:    email,
		password: password,
		client: &http.Client{
			Jar:       jar,
			Transport: rt,
		},
	}

	if err := c.login(); err != nil {
		return nil, fmt.Errorf("failed to log in: %w", err)
	}
	rt.signedUserID = c.signedUserID

	return c, nil
}

type refreshResponse struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

func (c *Client) refresh() error {
	data := url.Values{}
	data.Set("refresh_token", c.refreshToken)

	resp, err := c.client.PostForm("https://www.anylist.com/auth/token/refresh", data)
	if err != nil {
		return fmt.Errorf("failed to get response: %w", err)
	}

	var rr refreshResponse
	if err := json.NewDecoder(resp.Body).Decode(&rr); err != nil {
		return fmt.Errorf("failed to decode refresh response body: %w", err)
	}

	c.refreshToken = rr.RefreshToken
	c.accessToken = rr.AccessToken
	return nil
}

type loginResponse struct {
	SignedUserID string `json:"signed_user_id"`
	UserID       string `json:"user_id"`
}

func (c *Client) login() error {
	data := url.Values{}
	data.Set("email", c.email)
	data.Set("password", c.password)

	resp, err := c.client.PostForm("https://www.anylist.com/data/validate-login", data)
	if err != nil {
		return fmt.Errorf("failed to get response: %w", err)
	}

	var lr loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return fmt.Errorf("failed to decode login response body: %w", err)
	}

	c.signedUserID = lr.SignedUserID
	c.userID = lr.UserID
	return nil
}

// Note: This doesn't work yet, something wrong with the way I'm making the
// request. I've had two separate websocket client libraries fail on the
// handshake.
func (c *Client) Listen(cb func(string)) error {
	var p string
	if c.accessToken != "" {
		p = fmt.Sprintf(`/data/add-user-listener?client_id=%s&access_token=%s`, c.id, c.accessToken)
	} else if c.signedUserID != "" {
		p = fmt.Sprintf(`/data/add-user-listener/%s?client_id=%s`, c.signedUserID, c.id)
	} else {
		return errors.New("neither access token nor signed user ID was set")
	}
	u := url.URL{
		Scheme: "wss",
		Host:   "www.anylist.com",
		Path:   p,
	}

	// hdr := http.Header{}
	// hdr.Add("Origin", "https://www.anylist.com")
	// hdr.Add("Cookies", toString(c.client.Jar.Cookies(&url.URL{Scheme: "https", Host: "www.anylist.com", Path: "/"})))
	ws, err := websocket.Dial(u.String(), "", "https://www.anylist.com")
	if err != nil {
		return fmt.Errorf("failed to dial WS endpoint: %w", err)
	}
	defer ws.Close()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	done := make(chan struct{})
	defer close(done)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				_, err := ws.Write([]byte("--heartbeat--"))
				if err != nil {
					log.Println("write:", err)
					return
				}
			}
		}
	}()

	var buf bytes.Buffer
	for {
		_, err := io.Copy(&buf, ws)
		if err != nil {
			return fmt.Errorf("failed to read from conn: %w", err)
		}
		cb(buf.String())
		buf.Reset()
	}

}

func (c *Client) Lists() (*pb.PBUserDataResponse, error) {
	resp, err := c.client.Post("https://www.anylist.com/data/user-data/get", "application/x-www-form-urlencoded", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load data: %w", err)
	}

	dat, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	m := &pb.PBUserDataResponse{}
	if err := proto.Unmarshal(dat, m); err != nil {
		return nil, fmt.Errorf("failed to decode proto message: %w", err)
	}

	return m, nil
}

func (c *Client) AddItem(listID string, itemName string) error {
	itemID := uuid.NewString()
	req := &pb.PBListOperationList{
		Operations: []*pb.PBListOperation{
			{
				Metadata: &pb.PBOperationMetadata{
					OperationId: uuid.NewString(),
					HandlerId:   "add-shopping-list-item",
					UserId:      c.userID,
				},
				ListId:     listID,
				ListItemId: itemID,
				ListItem: &pb.ListItem{
					Identifier:      itemID,
					ListId:          listID,
					Name:            itemName,
					Checked:         false,
					CategoryMatchId: "other",
					UserId:          c.userID,
				},
			},
		},
	}

	dat, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request message: %w", err)
	}
	data := url.Values{}
	data.Set("operations", string(dat))

	resp, err := c.client.PostForm("https://www.anylist.com/data/shopping-lists/update", data)
	if err != nil {
		return fmt.Errorf("failed to add item: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid response code %d, expected 200 OK", resp.StatusCode)
	}

	return nil
}

func (c *Client) RemoveItem(listID, itemID string) error {
	req := &pb.PBListOperationList{
		Operations: []*pb.PBListOperation{
			{
				Metadata: &pb.PBOperationMetadata{
					OperationId: uuid.NewString(),
					HandlerId:   "remove-shopping-list-item",
					UserId:      c.userID,
				},
				ListId:     listID,
				ListItemId: itemID,
				ListItem: &pb.ListItem{
					Identifier: itemID,
					ListId:     listID,
				},
			},
		},
	}

	dat, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request message: %w", err)
	}
	data := url.Values{}
	data.Set("operations", string(dat))

	resp, err := c.client.PostForm("https://www.anylist.com/data/shopping-lists/update", data)
	if err != nil {
		return fmt.Errorf("failed to remove item: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid response code %d, expected 200 OK", resp.StatusCode)
	}

	return nil
}
func (c *Client) SetChecked(listID, itemID string, checked bool) error {
	updatedValue := "y"
	if !checked {
		updatedValue = "n"
	}
	req := &pb.PBListOperationList{
		Operations: []*pb.PBListOperation{
			{
				Metadata: &pb.PBOperationMetadata{
					OperationId: uuid.NewString(),
					HandlerId:   "set-list-item-checked",
					UserId:      c.userID,
				},
				ListId:       listID,
				ListItemId:   itemID,
				UpdatedValue: updatedValue,
			},
		},
	}

	dat, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request message: %w", err)
	}
	data := url.Values{}
	data.Set("operations", string(dat))

	resp, err := c.client.PostForm("https://www.anylist.com/data/shopping-lists/update", data)
	if err != nil {
		return fmt.Errorf("failed to updated item checked: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid response code %d, expected 200 OK", resp.StatusCode)
	}

	return nil
}

func toString(cs []*http.Cookie) string {
	var out []string
	for _, c := range cs {
		out = append(out, c.String())
	}
	return strings.Join(out, ";")
}

type roundTripper struct {
	id           string
	signedUserID string
}

func (rt *roundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("X-AnyLeaf-API-Version", "3")
	r.Header.Set("X-AnyLeaf-Client-Identifier", rt.id)

	if strings.HasPrefix(r.URL.Path, "/data/") && r.URL.Path != "/data/validate-login" {
		r.Header.Set("X-AnyLeaf-Signed-User-ID", rt.signedUserID)
	}

	return http.DefaultTransport.RoundTrip(r)
}