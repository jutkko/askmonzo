package main

import (
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
			RawQuery: "client_id=" + clientID + "&redirect_uri=" + c.Request.Host + "/auth/callback&response_type=code",
		}
		c.Redirect(http.StatusTemporaryRedirect, link.String())
	})
}

func setAuthCallbackEndpoint(router *gin.Engine, clientID, clientSecret string) {
	router.GET("/auth/callback", func(c *gin.Context) {
		authorizationCode, exists := c.Get("authorization_code")
		if !exists {
			panic("No authorization code in the callback request")
		}

		form := url.Values{}
		form.Add("grant_type", "authorization_code")
		form.Add("client_id", clientID)
		form.Add("client_secret", clientSecret)
		form.Add("redirect_uri", "https://google.com")
		form.Add("code", authorizationCode.(string))

		req, err := http.NewRequest("POST", "https://api.monzo.com/oauth2/token", strings.NewReader(form.Encode()))
		if err != nil {
			panic("Failed to create a new request")
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error when sending request to the server")
			return
		}

		defer resp.Body.Close()
		respBody, _ := ioutil.ReadAll(resp.Body)

		fmt.Println(resp.Status)
		fmt.Println(string(respBody))
	})
}

func getEnv(v string) string {
	env := os.Getenv(v)
	if env == "" {
		panic(fmt.Sprintf("No %s defined in the env", v))
	}

	return env
}
