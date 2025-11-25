# Sistema de Temperatura por CEP com OTEL e Zipkin

Sistema distribuído em Go que consulta temperatura por CEP, implementando tracing distribuído com OpenTelemetry e Zipkin.

## Arquitetura

- **Serviço A**: Recebe e valida o CEP (porta 8080)
- **Serviço B**: Orquestra busca de localização e temperatura (porta 8081)
- **Zipkin**: Visualização de traces (porta 9411)
- **OTEL Collector**: Coleta e processa traces (porta 4318)

## Pré-requisitos

- Docker e Docker Compose
- Chave API do OpenWeatherMap (gratuita em https://openweathermap.org/api)

## Configuração

1. Clone o repositório
2. Crie um arquivo `.env` na raiz do projeto:

```env
WEATHER_API_KEY=sua_chave_aqui
```

3. Obtenha sua chave API:
   - Acesse https://openweathermap.org/api
   - Crie uma conta gratuita
   - Copie sua API key

## Como Executar

### Subir todos os serviços

```bash
docker-compose up --build
```

### Acessar interfaces

- **Serviço A**: http://localhost:8080
- **Serviço B**: http://localhost:8081
- **Zipkin UI**: http://localhost:9411

## Testando a Aplicação

### Requisição válida

```bash
curl -X POST http://localhost:8080/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "01310100"}'
```

Resposta esperada (200):
```json
{
  "city": "São Paulo",
  "temp_C": 25.5,
  "temp_F": 77.9,
  "temp_K": 298.5
}
```

### CEP inválido (formato incorreto)

```bash
curl -X POST http://localhost:8080/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "123"}'
```

Resposta esperada (422):
```json
{
  "message": "invalid zipcode"
}
```

### CEP não encontrado

```bash
curl -X POST http://localhost:8080/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "99999999"}'
```

Resposta esperada (404):
```json
{
  "message": "can not find zipcode"
}
```

## Visualizando Traces no Zipkin

1. Acesse http://localhost:9411
2. Clique em "Run Query" para ver os traces
3. Clique em um trace para ver detalhes:
   - Span do Serviço A (validação)
   - Span do Serviço B (busca CEP + clima)
   - Tempos de resposta de cada serviço

## Estrutura do Projeto

```
.
├── service-a/
│   ├── main.go
│   ├── handlers/
│   │   └── cep_handler.go
│   └── Dockerfile
├── service-b/
│   ├── main.go
│   ├── handlers/
│   │   └── weather_handler.go
│   ├── services/
│   │   ├── cep_service.go
│   │   └── weather_service.go
│   └── Dockerfile
├── docker-compose.yml
├── otel-collector-config.yaml
└── README.md
```

## Tecnologias Utilizadas

- Go 1.21
- OpenTelemetry
- Zipkin
- Docker & Docker Compose
- APIs: ViaCEP e OpenWeatherMap

## Parar os Serviços

```bash
docker-compose down
```

## Troubleshooting

### Erro de API Key
Se receber erro 401, verifique se:
- A variável `WEATHER_API_KEY` está definida no `.env`
- A chave API está ativa (pode levar alguns minutos após criação)

### Traces não aparecem no Zipkin
- Aguarde alguns segundos após a requisição
- Verifique se o OTEL Collector está rodando: `docker-compose ps`
- Verifique logs: `docker-compose logs otel-collector`

## Observações

- O serviço usa cache interno para evitar chamadas desnecessárias às APIs
- Todos os erros são rastreados com spans no Zipkin
- Os traces incluem informações de CEP e cidade consultada