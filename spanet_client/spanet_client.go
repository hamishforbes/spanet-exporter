package spanet_client

// Import the required packages
import (
	"bytes"         // For reading & parsing HTTP output
	"encoding/json" // For parsing JSON data
	"errors"
	"fmt"       // For printing output to terminal and string methods
	"io/ioutil" // For reading & parsing HTTP output

	// For making websocket connections
	"net/http" // For making web requests
)

const (
	apiURL = "https://api.spanet.net.au/api/"
	apiKey = "4a483b9a-8f02-4e46-8bfa-0cf5732dbbd5"
)

type SpanetSocket struct {
	Id             string `json:"id"`
	Active         string `json:"active"`
	MemberId       int    `json:"id_member"`
	SocketId       int    `json:"id_sockets"`
	MacAddr        string `json:"mac_addr"`
	MobUrl         string `json:"moburl"`
	Name           string `json:"name"`
	SpaUrl         string `json:"spaurl"`
	SignalStrength int    `json:"signalStrength"`
	Error          bool
}

type SpaNetSocketResponse struct {
	Data    map[string]interface{} `json:"data"`
	Sockets []SpanetSocket         `json:"sockets"`
	Success bool                   `json:"success"`
}

type Client struct {
	username  string
	password  string
	memberId  string
	sessionId string
}

func New(username string, password string) *Client {
	return &Client{
		username: username,
		password: password,
	}
}

func (c *Client) Login() error {
	var err error

	// First, login to API with your username and encrypted password key to see if user exists, otherwise throw error
	postBody, _ := json.Marshal(map[string]string{
		"login":    c.username,
		"api_key":  apiKey,
		"password": c.password,
	})
	responseBody := bytes.NewBuffer(postBody)

	// Make the login request
	//fmt.Println(fmt.Sprintf("Logging in to API... %s \n %s", apiURL, postBody))
	var resp *http.Response
	resp, err = http.Post(apiURL+"MemberLogin", "application/json", responseBody)
	if err != nil {
		return errors.New("Failed to send login POST")
	}

	var byt []byte
	byt, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New("Failed to read login response")
	}

	var data map[string]interface{}
	jsonerr := json.Unmarshal(byt, &data)
	if jsonerr != nil {
		return errors.New("Failed to parse login response")
	}

	//fmt.Println(data)

	if data["success"].(bool) == false {
		return errors.New(data["error"].(string))
	}

	var body = data["data"].(map[string]interface{})
	c.memberId = fmt.Sprintf("%g", body["id_member"].(float64))
	c.sessionId = body["id_session"].(string)

	//fmt.Println(fmt.Sprintf("Successfully logged into SpaNET account! member: %s, session: %s", c.memberId, c.sessionId))
	return nil
}

func (c *Client) getSockets() (SpaNetSocketResponse, error) {
	// Make the next request which will check the spas on your account
	resp, err := http.Get(apiURL + "membersockets?id_member=" + c.memberId + "&id_session=" + c.sessionId)
	byt, byterr := ioutil.ReadAll(resp.Body)

	//fmt.Println(string(byt))
	var data SpaNetSocketResponse
	jsonerr := json.Unmarshal(byt, &data)

	if err == nil && byterr == nil && jsonerr == nil && data.Success == false {
		return data, errors.New("Failed to fetch sockets")
	}

	// If you get to this section, the spa request was successful
	//fmt.Println("Successfully got list of spa's linked to SpaNET account!")

	return data, nil
}

func (c *Client) Connect(spaName string) (SpaConn, error) {
	var conn SpaConn

	sockets, err := c.getSockets()
	if err != nil {
		return conn, err
	}

	for _, socket := range sockets.Sockets {
		if socket.Name == spaName {
			conn = SpaConn{socket: socket}
		}
	}

	err = conn.Connect()
	return conn, err
}
