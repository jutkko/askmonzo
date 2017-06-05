package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var state string
var expiryTime time.Time
var authResponse AuthResponse

type AuthResponse struct {
	AccessToken         string `json:"access_token"`
	ClientID            string `json:"client_id"`
	ExpiresIn           int    `json:"expires_in,int"`
	RefreshToken        string `json:"refresh_token"`
	TokenType           string `json:"token_type"`
	UserID              string `json:"user_id"`
	AuthExpiryTimestamp int64
}

type BearerResponse struct {
	Authenticated bool   `json:"authenticated"`
	ClientID      string `json:"client_id"`
	UserID        string `json:"user_id"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	newServer().Run(":" + port)
}

func newServer() *gin.Engine {
	router := gin.Default()

	// Set the environment variables
	clientID := getEnv("CLIENT_ID")
	clientSecret := getEnv("CLIENT_SECRET")

	router.GET("/ping", pingHandler)
	router.GET("/auth", authHandlerWrapper(clientID))
	router.GET("/auth/callback", setAuthCallbackEndpointWrapper(clientID, clientSecret))

	return router
}

func pingHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func authHandlerWrapper(clientID string) func(c *gin.Context) {
	state = getRandomString()

	return func(c *gin.Context) {
		link := url.URL{
			Scheme:   "https",
			Host:     "auth.getmondo.co.uk",
			RawQuery: "client_id=" + clientID + "&redirect_uri=" + c.Request.Host + "/auth/callback&response_type=code&state=" + state,
		}
		c.Redirect(http.StatusTemporaryRedirect, link.String())
	}
}

func setAuthCallbackEndpointWrapper(clientID, clientSecret string) func(c *gin.Context) {
	return func(c *gin.Context) {
		client := &http.Client{}
		var resp *http.Response
		var err error

		if authResponse.AuthExpiryTimestamp <= time.Now().Unix() {
			if authResponse.RefreshToken != "" {
				form := url.Values{}
				form.Add("grant_type", "refresh_token")
				form.Add("client_id", clientID)
				form.Add("client_secret", clientSecret)
				form.Add("refresh_token", authResponse.RefreshToken)

				req, err := http.NewRequest("POST", "https://api.monzo.com/oauth2/token", strings.NewReader(form.Encode()))
				if err != nil {
					panic("Failed to create a new request")
				}
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				resp, err = client.Do(req)
				if err != nil {
					panic("Failed to create a new request")
				}

				defer resp.Body.Close()
				respBody, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					panic("Failed to create a new request")
				}

				err = json.Unmarshal(respBody, &authResponse)
				if err != nil {
					panic("Failed to unmarshal the response" + err.Error())
				}
			} else {
				err := c.Request.ParseForm()
				if err != nil {
					panic("Failed to parse form")
				}

				authorizationCode := c.Request.Form.Get("code")
				monzoState := c.Request.Form.Get("state")
				if monzoState != state {
					c.JSON(404, gin.H{
						"Error": "The state does not match, what are you trying to do",
					})
					return
				}

				form := url.Values{}
				form.Add("grant_type", "authorization_code")
				form.Add("client_id", clientID)
				form.Add("client_secret", clientSecret)
				form.Add("redirect_uri", c.Request.Host+"/auth/callback")
				form.Add("code", authorizationCode)

				req, err := http.NewRequest("POST", "https://api.monzo.com/oauth2/token", strings.NewReader(form.Encode()))
				if err != nil {
					panic("Failed to create a new request")
				}
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				resp, err = client.Do(req)
				if err != nil {
					panic("Failed to create a new request")
				}

				defer resp.Body.Close()
				respBody, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					panic("Failed to create a new request")
				}

				err = json.Unmarshal(respBody, &authResponse)
				if err != nil {
					panic("Failed to unmarshal the response" + err.Error())
				}
			}

			authResponse.AuthExpiryTimestamp = time.Now().Unix() + int64(authResponse.ExpiresIn)
		}

		req, err := http.NewRequest("GET", "https://api.monzo.com/ping/whoami", nil)
		if err != nil {
			panic("Failed to create a new request")
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authResponse.AccessToken))
		req.Header.Set("Content-Type", "application/json")

		resp1, err := client.Do(req)
		if err != nil {
			panic(fmt.Sprintf("Failed to create a new request", err))
		}

		respBody, err := ioutil.ReadAll(resp1.Body)
		if err != nil {
			panic("Failed to create a new request")
		}

		var bearerResponse BearerResponse
		err = json.Unmarshal(respBody, &bearerResponse)
		if err != nil {
			panic("Failed to unmarshal the response" + err.Error())
		}

		c.JSON(resp1.StatusCode, gin.H{
			"message": "authentication successful",
		})

		// Successful exchange

	}
}

func getEnv(v string) string {
	env := os.Getenv(v)
	if env == "" {
		panic(fmt.Sprintf("No %s defined in the env", v))
	}

	return env
}

func getRandomString() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	number := r.Int63()
	return strconv.FormatInt(number, 10)
}
