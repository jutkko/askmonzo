package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type Response struct {
	AccessToken  string `json:"access_token"`
	ClientID     string `json:"client_id"`
	ExpiresIn    int    `json:"expires_in,int"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	UserID       string `json:"user_id"`
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
	return func(c *gin.Context) {
		link := url.URL{
			Scheme:   "https",
			Host:     "auth.getmondo.co.uk",
			RawQuery: "client_id=" + clientID + "&redirect_uri=https://" + c.Request.Host + "/auth/callback&response_type=code",
		}
		c.Redirect(http.StatusTemporaryRedirect, link.String())
	}
}

func setAuthCallbackEndpointWrapper(clientID, clientSecret string) func(c *gin.Context) {
	return func(c *gin.Context) {
		err := c.Request.ParseForm()
		if err != nil {
			panic("Failed to parse form")
		}

		client := &http.Client{}

		authorizationCode := c.Request.Form.Get("code")

		form := url.Values{}
		form.Add("grant_type", "authorization_code")
		form.Add("client_id", clientID)
		form.Add("client_secret", clientSecret)
		form.Add("redirect_uri", "https://pure-oasis-86979.herokuapp.com/auth/callback")
		form.Add("code", authorizationCode)

		req, err := http.NewRequest("POST", "https://api.monzo.com/oauth2/token", strings.NewReader(form.Encode()))
		if err != nil {
			panic("Failed to create a new request")
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := client.Do(req)
		if err != nil {
			panic("Failed to create a new request")
		}

		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic("Failed to create a new request")
		}

		var response Response
		err = json.Unmarshal(respBody, &response)
		if err != nil {
			panic("Failed to unmarshal the response" + err.Error())
		}

		req, err = http.NewRequest("GET", "https://api.monzo.com/ping/whoami", nil)
		if err != nil {
			panic("Failed to create a new request")
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", response.AccessToken))
		req.Header.Set("Content-Type", "application/json")

		resp1, err := client.Do(req)
		if err != nil {
			panic(fmt.Sprintf("Failed to create a new request", err))
		}

		respBody, err = ioutil.ReadAll(resp1.Body)
		if err != nil {
			panic("Failed to create a new request")
		}

		c.JSON(resp.StatusCode, gin.H{
			"message": string(respBody),
		})
	}
}

func getEnv(v string) string {
	env := os.Getenv(v)
	if env == "" {
		panic(fmt.Sprintf("No %s defined in the env", v))
	}

	return env
}
