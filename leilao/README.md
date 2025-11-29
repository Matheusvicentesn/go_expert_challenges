# Sistema de Leilão com Fechamento Automático

Este projeto implementa um sistema de leilões em Go com fechamento automático baseado em tempo configurável. O sistema utiliza goroutines para gerenciar o fechamento automático de leilões após um intervalo definido.

## Funcionalidades

- ✅ Criação de leilões
- ✅ Criação de lances (bids)
- ✅ **Fechamento automático de leilões** após intervalo configurável
- ✅ Validação de leilões vencidos na criação de lances
- ✅ API REST para gerenciamento

## Arquitetura

O fechamento automático é implementado através de:

1. **Função de cálculo de intervalo**: `getAuctionInterval()` lê a variável de ambiente `AUCTION_INTERVAL`
2. **Goroutine de fechamento**: Iniciada automaticamente ao criar um leilão, aguarda o intervalo e fecha o leilão
3. **Update atômico**: Verifica se o leilão ainda está `Active` antes de fechar, evitando race conditions

## Pré-requisitos

- Docker 24+ e Docker Compose 2+
- Go 1.20+ (opcional, apenas para executar testes localmente)

## Configuração

### Variáveis de Ambiente

Crie o arquivo `cmd/auction/.env` com as seguintes variáveis:

```env
MONGODB_URI=mongodb://mongodb:27017
MONGODB_DATABASE=auction
AUCTION_INTERVAL=5m
PORT=8080
```

**Importante**: `AUCTION_INTERVAL` aceita qualquer duração compatível com `time.ParseDuration` do Go:

- `30s` - 30 segundos
- `5m` - 5 minutos
- `1h` - 1 hora
- `2h30m` - 2 horas e 30 minutos

Se a variável não estiver definida ou for inválida, o sistema usa o padrão de **5 minutos**.

## Executando com Docker

### 1. Construir e iniciar os serviços

```bash
docker compose up --build
```

Isso irá:

- Construir a imagem da aplicação Go
- Iniciar o MongoDB
- Iniciar a aplicação na porta 8080

### 2. Acessar a aplicação

- **API**: `http://localhost:8080`
- **MongoDB**: `mongodb://localhost:27017`

## Endpoints da API

A API expõe os seguintes endpoints REST:

### Leilões (Auctions)

#### `POST /auction` - Criar Leilão

Cria um novo leilão. O leilão será fechado automaticamente após o intervalo configurado em `AUCTION_INTERVAL`.

**Request Body:**
```json
{
  "product_name": "Notebook Dell",
  "category": "Eletrônicos",
  "description": "Notebook Dell Inspiron 15 com 8GB RAM",
  "condition": 1
}
```

**Campos:**
- `product_name` (string, obrigatório, mínimo 1 caractere): Nome do produto
- `category` (string, obrigatório, mínimo 2 caracteres): Categoria do produto
- `description` (string, obrigatório, entre 10 e 200 caracteres): Descrição do produto
- `condition` (int, obrigatório): Condição do produto
  - `0` = Não especificado
  - `1` = Novo (New)
  - `2` = Usado (Used)
  - `3` = Recondicionado (Refurbished)

**Response:** `201 Created` (sem body)

**Exemplo:**
```bash
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "Notebook Dell",
    "category": "Eletrônicos",
    "description": "Notebook Dell Inspiron 15 com 8GB RAM",
    "condition": 1
  }'
```

#### `GET /auction` - Listar Leilões

Lista leilões com filtros opcionais.

**Query Parameters:**
- `status` (int, opcional): Status do leilão (0 = Active, 1 = Completed)
- `category` (string, opcional): Filtrar por categoria
- `productName` (string, opcional): Filtrar por nome do produto (busca parcial, case-insensitive)

**Response:** `200 OK`
```json
[
  {
    "id": "uuid-do-leilao",
    "product_name": "Notebook Dell",
    "category": "Eletrônicos",
    "description": "Notebook Dell Inspiron 15 com 8GB RAM",
    "condition": 1,
    "status": 0,
    "timestamp": "2024-01-15 10:30:00"
  }
]
```

**Exemplo:**
```bash
curl "http://localhost:8080/auction?status=0&category=Eletrônicos"
```

#### `GET /auction/:auctionId` - Buscar Leilão por ID

Busca um leilão específico pelo ID.

**Path Parameters:**
- `auctionId` (UUID, obrigatório): ID do leilão

**Response:** `200 OK`
```json
{
  "id": "uuid-do-leilao",
  "product_name": "Notebook Dell",
  "category": "Eletrônicos",
  "description": "Notebook Dell Inspiron 15 com 8GB RAM",
  "condition": 1,
  "status": 0,
  "timestamp": "2024-01-15 10:30:00"
}
```

**Exemplo:**
```bash
curl "http://localhost:8080/auction/123e4567-e89b-12d3-a456-426614174000"
```

#### `GET /auction/winner/:auctionId` - Buscar Vencedor do Leilão

Retorna informações do leilão e do lance vencedor (maior valor).

**Path Parameters:**
- `auctionId` (UUID, obrigatório): ID do leilão

**Response:** `200 OK`
```json
{
  "auction": {
    "id": "uuid-do-leilao",
    "product_name": "Notebook Dell",
    "category": "Eletrônicos",
    "description": "Notebook Dell Inspiron 15 com 8GB RAM",
    "condition": 1,
    "status": 1,
    "timestamp": "2024-01-15 10:30:00"
  },
  "bid": {
    "id": "uuid-do-lance",
    "user_id": "uuid-do-usuario",
    "auction_id": "uuid-do-leilao",
    "amount": 2500.00,
    "timestamp": "2024-01-15 10:35:00"
  }
}
```

**Exemplo:**
```bash
curl "http://localhost:8080/auction/winner/123e4567-e89b-12d3-a456-426614174000"
```

### Lances (Bids)

#### `POST /bid` - Criar Lance

Cria um novo lance em um leilão. O lance só será aceito se o leilão estiver ativo e não tiver expirado.

**Request Body:**
```json
{
  "user_id": "uuid-do-usuario",
  "auction_id": "uuid-do-leilao",
  "amount": 1500.50
}
```

**Campos:**
- `user_id` (UUID, obrigatório): ID do usuário que está fazendo o lance
- `auction_id` (UUID, obrigatório): ID do leilão
- `amount` (float, obrigatório, > 0): Valor do lance

**Response:** `201 Created` (sem body)

**Exemplo:**
```bash
curl -X POST http://localhost:8080/bid \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "123e4567-e89b-12d3-a456-426614174001",
    "auction_id": "123e4567-e89b-12d3-a456-426614174000",
    "amount": 1500.50
  }'
```

#### `GET /bid/:auctionId` - Listar Lances de um Leilão

Lista todos os lances de um leilão específico.

**Path Parameters:**
- `auctionId` (UUID, obrigatório): ID do leilão

**Response:** `200 OK`
```json
[
  {
    "id": "uuid-do-lance-1",
    "user_id": "uuid-do-usuario-1",
    "auction_id": "uuid-do-leilao",
    "amount": 1200.00,
    "timestamp": "2024-01-15 10:32:00"
  },
  {
    "id": "uuid-do-lance-2",
    "user_id": "uuid-do-usuario-2",
    "auction_id": "uuid-do-leilao",
    "amount": 1500.50,
    "timestamp": "2024-01-15 10:33:00"
  }
]
```

**Exemplo:**
```bash
curl "http://localhost:8080/bid/123e4567-e89b-12d3-a456-426614174000"
```

### Usuários (Users)

#### `GET /user/:userId` - Buscar Usuário por ID

Busca um usuário específico pelo ID.

**Path Parameters:**
- `userId` (UUID, obrigatório): ID do usuário

**Response:** `200 OK`
```json
{
  "id": "uuid-do-usuario",
  "name": "João Silva"
}
```

**Exemplo:**
```bash
curl "http://localhost:8080/user/123e4567-e89b-12d3-a456-426614174001"
```

### Códigos de Status HTTP

- `200 OK`: Requisição bem-sucedida
- `201 Created`: Recurso criado com sucesso
- `400 Bad Request`: Erro de validação ou parâmetros inválidos
- `404 Not Found`: Recurso não encontrado
- `500 Internal Server Error`: Erro interno do servidor

### Tratamento de Erros

Todos os erros retornam um formato padronizado:

```json
{
  "message": "Mensagem de erro",
  "error": "Tipo do erro",
  "code": 400,
  "causes": [
    {
      "field": "campo",
      "message": "Mensagem específica do campo"
    }
  ]
}
```

### 3. Parar os serviços

```bash
docker compose down
```

Para remover também os volumes (dados do MongoDB):

```bash
docker compose down -v
```

## Executando em Desenvolvimento Local

### 1. Instalar dependências

```bash
go mod download
```

### 2. Configurar variáveis de ambiente

Certifique-se de que o arquivo `cmd/auction/.env` existe e está configurado.

### 3. Executar a aplicação

```bash
go run cmd/auction/main.go
```

## Testes

Execute os testes automatizados:

```bash
go test ./...
```

Para executar apenas os testes do módulo de auction:

```bash
go test ./internal/infra/database/auction -v
```

### Testes Implementados

- ✅ `TestAutoCloseRoutineTriggersAfterInterval`: Valida que a goroutine dispara após o intervalo
- ✅ `TestGetAuctionInterval`: Testa o cálculo de intervalo com diferentes valores de ambiente
- ✅ `TestUpdateAuctionStatusToCompleted`: Valida a estrutura do update

## Como Funciona o Fechamento Automático

1. **Ao criar um leilão** (`CreateAuction`):

   - O leilão é inserido no MongoDB com status `Active`
   - Uma goroutine é iniciada imediatamente

2. **Na goroutine**:

   - Aguarda o intervalo definido em `AUCTION_INTERVAL`
   - Após o intervalo, verifica se o leilão ainda está `Active`
   - Atualiza o status para `Completed` apenas se ainda estiver `Active`

3. **Validação em lances**:
   - O sistema de bids já valida se o leilão está fechado ou vencido
   - Usa a mesma lógica de `Timestamp + AUCTION_INTERVAL` para calcular o tempo de término

## Estrutura do Projeto

```
.
├── cmd/auction/           # Aplicação principal
├── internal/
│   ├── entity/            # Entidades de domínio
│   ├── infra/
│   │   ├── api/           # Controllers e rotas
│   │   └── database/      # Repositórios (MongoDB)
│   └── usecase/           # Casos de uso
├── configuration/         # Configurações (logger, DB, etc)
├── docker-compose.yml      # Orquestração Docker
├── Dockerfile             # Build da aplicação
└── README.md              # Este arquivo
```

