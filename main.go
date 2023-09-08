package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"test/domain"
	"test/state"

	// Importa o pacote UUID
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"
)

var (
	oauth2Config = oauth2.Config{
		ClientID:     "CLIENT_ID",
		ClientSecret: "CLIENT_SECRET",
		RedirectURL:  "https://b0be-201-182-186-221.ngrok-free.app/callback",
		Scopes:       []string{"user-top-read", "user-library-read"},
		Endpoint:     spotify.Endpoint,
	}

	stateStore = make(map[string]bool)
)

func main() {
	e := echo.New()

	e.GET("/login", Authorization)
	e.GET("/callback", SpotifyCallback)

	e.Start(":8001")
}

func Authorization(c echo.Context) error {
	state := state.GenerateRandomState()
	authorizationURL := oauth2Config.AuthCodeURL(state)
	stateStore[state] = true
	return c.Redirect(http.StatusFound, authorizationURL)
}

func SpotifyCallback(c echo.Context) error {
	code := c.QueryParam("code")
	state := c.QueryParam("state")

	// Verifica se o estado é válido
	if !stateStore[state] {
		log.Println("Estado inválido ou correspondência não encontrada.")
		return c.JSON(http.StatusBadRequest, "Estado inválido ou correspondência não encontrada.")
	}

	token, err := oauth2Config.Exchange(c.Request().Context(), code)
	if err != nil {
		log.Println("Erro ao obter o Token de Acesso:", err)
		return c.JSON(http.StatusInternalServerError, "Erro ao obter o Token de Acesso")
	}

	if token == nil {
		log.Println("Token de Acesso nulo.")
		return c.JSON(http.StatusInternalServerError, "Token de Acesso nulo")
	}

	if !token.Valid() {
		log.Println("Erro durante a troca do código pelo Token de Acesso.")
		return c.JSON(http.StatusInternalServerError, "Erro durante a troca do código pelo Token de Acesso")
	}

	topTracks, err := GetTopTracks(token)
	if err != nil {
		log.Println("Erro ao obter os top tracks:", err)
		return c.JSON(http.StatusInternalServerError, "Erro ao obter os top tracks")
	}

	return c.JSON(http.StatusOK, topTracks)
}

func GetTopTracks(token *oauth2.Token) ([]domain.Track, error) {
	apiURL := "https://api.spotify.com/v1/me/top/tracks"
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Erro ao obter os top tracks: %s", resp.Status)
	}

	var tracksResponse struct {
		Items []domain.Track `json:"items"`
	}

	err = json.NewDecoder(resp.Body).Decode(&tracksResponse)
	if err != nil {
		return nil, err
	}

	return tracksResponse.Items, nil
}
