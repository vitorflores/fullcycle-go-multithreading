# Busca de CEP com Multithreading

Projeto em Go que consulta um CEP em duas APIs ao mesmo tempo, a [BrasilAPI](https://brasilapi.com.br) e a [ViaCEP](https://viacep.com.br), e fica só com a resposta que chegar primeiro. A mais lenta é descartada.

Se nenhuma das duas responder em até 1 segundo, o programa mostra um erro de timeout.

## Como funciona

Cada API é chamada em uma goroutine própria. Os resultados voltam por channels e o `select` decide quem ganhou a corrida. As duas requisições compartilham um `context` com limite de 1 segundo, então quando o programa termina a requisição que ficou para trás é cancelada.

Se uma das APIs falhar (CEP inexistente, erro de rede etc.), o erro aparece no terminal e o programa continua esperando a outra. Se as duas falharem, ele encerra com erro.

## Pré-requisitos

- Go instalado (versão indicada no `go.mod` ou mais recente)
- Acesso à internet

## Como rodar

Na raiz do projeto, passe o CEP como argumento, só os números:

```bash
go run main.go 01153000
```

A saída vai ser parecida com esta:

```
API: BrasilAPI
CEP: 01153000
Logradouro: Rua Vitorino Carmilo
Bairro: Barra Funda
Cidade: São Paulo
UF: SP
```

A linha `API` mostra qual serviço respondeu mais rápido. Rodando algumas vezes, pode ser que o vencedor mude.

Se preferir gerar o binário:

```bash
go build -o buscacep .
./buscacep 01153000
```

## Exemplo de timeout

Quando nenhuma API responde dentro de 1 segundo, a mensagem é esta:

```
Erro: timeout, nenhuma API respondeu em até 1 segundo
```

Um jeito simples de ver isso acontecer é rodar o programa com a internet bem lenta. Com a internet desconectada, as duas requisições falham na hora, então o que aparece são os erros de cada API e não o timeout.
