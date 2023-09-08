package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"test/domain"
	"test/state"

	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"
)

var (
	oauth2Config = oauth2.Config{
		ClientID:     "3348f4b437614e8b9c742c305eb9865b",
		ClientSecret: "cd3d51d52a724d18aac6d3910534420c",
		RedirectURL:  "https://b0be-201-182-186-221.ngrok-free.app/callback",
		Scopes:       []string{"user-top-read", "user-library-read", "user-read-playback-state", "user-read-playback-position", "user-read-recently-played", "user-read-currently-playing"}, // Adicione as permissões necessárias aqui
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

	log.Printf("Token de Acesso: %+v\n", token)

	topTracks, err := GetTopTracks(token, 10)
	if err != nil {
		log.Println("Erro ao obter os top tracks:", err)
		return c.JSON(http.StatusInternalServerError, "Erro ao obter os top tracks")
	}

	return c.JSON(http.StatusOK, topTracks)
}

func GetTopTracks(token *oauth2.Token, limit int) ([]domain.Track, error) {
	apiURL := fmt.Sprintf("https://api.spotify.com/v1/me/top/tracks?time_range=short_term&limit=%d", limit)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Erro ao fazer a solicitação HTTP:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Erro ao obter os top tracks. Status do HTTP:", resp.Status)
		return nil, fmt.Errorf("Erro ao obter os top tracks: %s", resp.Status)
	}

	var tracksResponse struct {
		Items []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Artists []struct {
				Name string `json:"name"`
			} `json:"artists"`
		} `json:"items"`
	}

	err = json.NewDecoder(resp.Body).Decode(&tracksResponse)
	if err != nil {
		log.Println("Erro ao decodificar a resposta JSON:", err)
		return nil, err
	}

	var tracks []domain.Track
	for _, item := range tracksResponse.Items {
		// Para simplificar, apenas pegue o primeiro nome do artista, se houver
		artistName := ""
		if len(item.Artists) > 0 {
			artistName = item.Artists[0].Name
		}

		track := domain.Track{
			ID:     item.ID,
			Name:   item.Name,
			Artist: artistName,
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

