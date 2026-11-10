package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type ViaCEP struct {
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
	Erro        any    `json:"erro"`
}

type BrasilAPI struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service"`
}

type Endereco struct {
	Cep        string
	Logradouro string
	Bairro     string
	Cidade     string
	Uf         string
}

type Resultado struct {
	API      string
	Endereco Endereco
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Uso: go run main.go <cep>")
		os.Exit(1)
	}
	cep := os.Args[1]

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	c1 := make(chan Resultado, 1)
	c2 := make(chan Resultado, 1)
	erros := make(chan error, 2)

	go func() {
		endereco, err := BuscaBrasilAPI(ctx, cep)
		if err != nil {
			erros <- fmt.Errorf("BrasilAPI: %w", err)
			return
		}
		c1 <- Resultado{API: "BrasilAPI", Endereco: *endereco}
	}()

	go func() {
		endereco, err := BuscaViaCEP(ctx, cep)
		if err != nil {
			erros <- fmt.Errorf("ViaCEP: %w", err)
			return
		}
		c2 <- Resultado{API: "ViaCEP", Endereco: *endereco}
	}()

	falhas := 0
	for {
		select {
		case res := <-c1:
			Exibe(res)
			return
		case res := <-c2:
			Exibe(res)
			return
		case err := <-erros:
			if errors.Is(err, context.DeadlineExceeded) {
				continue
			}
			fmt.Fprintf(os.Stderr, "Erro ao consultar %v\n", err)
			falhas++
			if falhas == 2 {
				os.Exit(1)
			}
		case <-ctx.Done():
			fmt.Fprintln(os.Stderr, "Erro: timeout, nenhuma API respondeu em até 1 segundo")
			os.Exit(1)
		}
	}
}

func BuscaBrasilAPI(ctx context.Context, cep string) (*Endereco, error) {
	body, err := Requisicao(ctx, "https://brasilapi.com.br/api/cep/v1/"+cep)
	if err != nil {
		return nil, err
	}
	var data BrasilAPI
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	return &Endereco{
		Cep:        data.Cep,
		Logradouro: data.Street,
		Bairro:     data.Neighborhood,
		Cidade:     data.City,
		Uf:         data.State,
	}, nil
}

func BuscaViaCEP(ctx context.Context, cep string) (*Endereco, error) {
	body, err := Requisicao(ctx, "http://viacep.com.br/ws/"+cep+"/json/")
	if err != nil {
		return nil, err
	}
	var data ViaCEP
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	if data.Erro != nil {
		return nil, errors.New("CEP não encontrado")
	}
	return &Endereco{
		Cep:        data.Cep,
		Logradouro: data.Logradouro,
		Bairro:     data.Bairro,
		Cidade:     data.Localidade,
		Uf:         data.Uf,
	}, nil
}

func Requisicao(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status inesperado %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func Exibe(res Resultado) {
	fmt.Printf("API: %s\n", res.API)
	fmt.Printf("CEP: %s\n", res.Endereco.Cep)
	fmt.Printf("Logradouro: %s\n", res.Endereco.Logradouro)
	fmt.Printf("Bairro: %s\n", res.Endereco.Bairro)
	fmt.Printf("Cidade: %s\n", res.Endereco.Cidade)
	fmt.Printf("UF: %s\n", res.Endereco.Uf)
}
