package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type CEP struct {
	Cep    string
	Rua    string
	Bairro string
	Cidade string
	Estado string
}

func (cep CEP) String() string {
	return fmt.Sprintf(
		"CEP: %s\nRua: %s\nBairro: %s\nCidade: %s\nEstado: %s\n",
		cep.Cep, cep.Rua, cep.Bairro, cep.Cidade, cep.Estado,
	)
}

type ViaCepResponse struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

type BrasilApiResponse struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
}

type BrasilApiError struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

func main() {
	ch_brasilapi := make(chan CEP)
	ch_viacep := make(chan CEP)
	cep_string := "08071072"

	go fetch_cep_brasilapi(cep_string, ch_brasilapi)
	go fetch_cep_viacep(cep_string, ch_viacep)

	select {
	case cep := <-ch_brasilapi:
		println("CEP fetched using BrasilAPI:\n")
		fmt.Println(cep)

	case cep := <-ch_viacep:
		println("CEP fetched using ViaCEP:\n")
		fmt.Println(cep)

	case <-time.After(1 * time.Second):
		println("Timeout after 1 second")
	}
}

func fetch_cep_brasilapi(cep string, ch chan<- CEP) {
	url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)

	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching cep using BrasilAPI: %s\n", err)
		return
	}

	defer close(ch)
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Error reading response from BrasilAPI: %s\n", err)
		return
	}

	err_json := BrasilApiError{}
	err = json.Unmarshal(data, &err_json)
	if err != nil {
		fmt.Printf("Error decoding response from BrasilAPI: %s\n", err)
		return
	}

	if err_json.Type == "service_error" {
		return
	}

	response_json := BrasilApiResponse{}
	err = json.Unmarshal(data, &response_json)
	if err != nil {
		fmt.Printf("Error decoding response from BrasilAPI: %s\n", err)
		return
	}

	ch <- CEP{
		Cep:    response_json.Cep,
		Rua:    response_json.Street,
		Bairro: response_json.Neighborhood,
		Cidade: response_json.City,
		Estado: response_json.State,
	}
}

func fetch_cep_viacep(cep string, ch chan<- CEP) {
	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)

	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching cep using ViaCEP: %s\n", err)
		return
	}

	defer close(ch)
	defer response.Body.Close()

	response_json := ViaCepResponse{}
	err = json.NewDecoder(response.Body).Decode(&response_json)
	if err != nil {
		fmt.Printf("Error decoding response from ViaCEP: %s\n", err)
		return
	}

	ch <- CEP{
		Cep:    response_json.Cep,
		Rua:    response_json.Logradouro,
		Bairro: response_json.Bairro,
		Cidade: response_json.Localidade,
		Estado: response_json.Uf,
	}
}
