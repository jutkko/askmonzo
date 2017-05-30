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

func main() {
	router := gin.Default()

	// Set the environment variables
	clientID := getEnv("CLIENT_ID")
	clientSecret := getEnv("CLIENT_SECRET")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	setPingEndpoint(router)
	setAuthEndpoint(router, clientID)
	setAuthCallbackEndpoint(router, clientID, clientSecret)

	router.Run(":" + port)
}

type Response struct {
	AccessToken  string `json:"access_token,string"`
	ClientID     string `json:"client_id,string"`
	ExpiresIn    int    `json:"expires_in,int"`
	RefreshToken string `json:"refresh_token,string"`
	TokenType    string `json:"token_type,string"`
	UserID       string `json:"user_id,string"`
}

func setPingEndpoint(router *gin.Engine) {
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
}

func setAuthEndpoint(router *gin.Engine, clientID string) {
	router.GET("/auth", func(c *gin.Context) {
		link := url.URL{
			Scheme:   "https",
			Host:     "auth.getmondo.co.uk",
			RawQuery: "client_id=" + clientID + "&redirect_uri=https://" + c.Request.Host + "/auth/callback&response_type=code",
		}
		c.Redirect(http.StatusTemporaryRedirect, link.String())
	})
}

func setAuthCallbackEndpoint(router *gin.Engine, clientID, clientSecret string) {
	router.GET("/auth/callback", func(c *gin.Context) {
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
		form.Add("redirect_uri", "https://"+c.Request.Host+"/auth/callback")
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

		// There is expiry time, how to deal with it?
		c.JSON(resp.StatusCode, gin.H{
			"message": string(respBody),
		})

		response := &Response{}
		json.Unmarshal(respBody, &response)

		fmt.Printf("Message: %#+v\n", response.AccessToken)
	})
}

func getEnv(v string) string {
	env := os.Getenv(v)
	if env == "" {
		panic(fmt.Sprintf("No %s defined in the env", v))
	}

	return env
}
