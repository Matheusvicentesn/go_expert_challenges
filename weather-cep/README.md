# Weather CEP

Weather CEP é uma aplicação em Go que coleta informações meteorológicas com base em um CEP brasileiro. Ela integra com a API ViaCEP para encontrar a localização e a API OpenWeatherMap para buscar as condições climáticas atuais.

## Funcionalidades

- **Validação de CEP**: Valida se o CEP fornecido tem o formato correto (8 dígitos).
- **Busca de Localização**: Coleta informações de cidade e estado usando ViaCEP.
- **Dados Meteorológicos**: Busca a temperatura atual (Celsius, Fahrenheit, Kelvin) para a localização.
- **Conversão de Temperatura**: Converte automaticamente as temperaturas entre Celsius, Fahrenheit e Kelvin.

## Pré-requisitos

- [Go](https://go.dev/) 1.25 ou superior
- [Docker](https://www.docker.com/) (opcional, para execução em container)
- [Chave de API OpenWeatherMap](https://openweathermap.org/api)

## Configuração

A aplicação requer uma chave de API do OpenWeatherMap. Você pode configurá-la usando uma variável de ambiente ou um arquivo `.env`.

1. Copie o arquivo de exemplo de ambiente:
   ```bash
   cp .env.example .env
   ```

2. Edite o arquivo `.env` e adicione sua chave de API do OpenWeatherMap:
   ```env
   WEATHER_API_KEY=sua_chave_api_aqui
   ```

## Instalação e Execução

### Executando Localmente

1. Clone o repositório:
   ```bash
   git clone <url-do-repositorio>
   cd weather-cep
   ```

2. Instale as dependências:
   ```bash
   go mod download
   ```

3. Execute a aplicação:
   ```bash
   go run cmd/api/main.go
   ```
   O servidor iniciará na porta `8080`.

### Executando com Docker

1. Construa a imagem Docker:
   ```bash
   docker build -t weather-cep .
   ```

2. Execute o container:
   ```bash
   docker run -p 8080:8080 -e WEATHER_API_KEY=sua_chave_api_aqui weather-cep
   ```

### Executando com Docker Compose

1. Certifique-se de que seu arquivo `.env` está configurado ou atualize o `docker-compose.yml` diretamente (não recomendado para segredos).

2. Inicie os serviços:
   ```bash
   docker-compose up --build
   ```

## Endpoints da API

### Obter Clima por CEP

Recupera informações de localização e clima para um CEP específico.

- **URL**: `/weather/{cep}`
- **Método**: `GET`
- **Parâmetros de URL**:
  - `cep`: O CEP brasileiro (8 dígitos, com ou sem hífen).

#### Resposta de Sucesso

- **Código**: `200 OK`
- **Conteúdo**:
  ```json
  {
    "location": "Praça da Sé, São Paulo - SP",
    "temperatures": {
      "temp_C": "25.0 °C",
      "temp_F": "77.0 °F",
      "temp_K": "298.1 K",
      "temp_C_feels_like": "26.5 °C",
      "temp_C_min": "23.0 °C",
      "temp_C_max": "27.0 °C"
    }
  }
  ```

#### Respostas de Erro

- **Código**: `422 Unprocessable Entity`
  - **Conteúdo**: `invalid zipcode`
  - **Motivo**: O formato do CEP é inválido (deve ter 8 dígitos).

- **Código**: `404 Not Found`
  - **Conteúdo**: `can not find zipcode`
  - **Motivo**: O CEP não foi encontrado na base de dados do ViaCEP.

- **Código**: `500 Internal Server Error`
  - **Conteúdo**: `error message`
  - **Motivo**: Erro ao buscar dados meteorológicos ou outros problemas internos.

## Testes

Para executar os testes automatizados:

```bash
go test ./tests/...
```

## Deploy Cloud Run

[Deploy](https://full-cycle-weather-cep-531494565432.us-central1.run.app/weather/04417020)


