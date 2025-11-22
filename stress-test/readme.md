# 🚀 Load Testing CLI em Go

Sistema testes de carga para serviços web com relatório detalhado.

## Requisitos

- Docker instalado
- Go 1.21+ (opcional, se executar localmente)

## Compilação da Imagem Docker

```bash
docker build --no-cache -t loadtester:latest .
```

## Uso

### Exemplo Básico

```bash
docker run loadtest:latest --url=https://httpbin.org/status/200 --requests=1000 --concurrency=100
```

### Parâmetros

| Parâmetro | Tipo | Descrição | Exemplo |
|-----------|------|-----------|---------|
| `--url` | string | URL do serviço a testar (obrigatório) | `https://httpbin.org/status/200` |
| `--requests` | int | Número total de requisições (obrigatório) | `1000` |
| `--concurrency` | int | Chamadas simultâneas (obrigatório) | `10` |
| `--token` | string | Token enviado no header API_KEY para testes com rate-limiter (opcional) | `meusegredo` |

### Exemplos de Uso

#### Teste simples em localhost
```bash
docker run loadtest:latest \
  --url=http://localhost:8080 \
  --requests=100 \
  --concurrency=5
```

#### Teste de carga intenso
```bash
docker run loadtest:latest \
  --url=https://api.exemplo.com/users \
  --requests=10000 \
  --concurrency=50
```

#### Teste com serviço em rede Docker
```bash
# Criar rede Docker
docker network create testnet

# Executar o load tester na mesma rede
docker run --network testnet loadtest:latest \
  --url=http://seu-servico:3000/health \
  --requests=500 \
  --concurrency=20
```

#### Teste com o projeto do rate-limiter (bloqueio de IP)
```bash
 docker run --network=rate-limiter_limiter-net loadtester --url=http://app:8080 --requests=6 --concurrency=5
```

#### Teste com o projeto do rate-limiter (bloqueio via token)
```bash
 docker run --network=rate-limiter_limiter-net loadtester --url=http://app:8080 --requests=11 --concurrency=10 --token=meusegredo
```


## Saída do Relatório

```
RESULT:
Total time (s): 5.451267895s
Total requests: 10000
Status 200: 9997 (99.97%)

HTTP CODES:
  200 OK: 9997 (99.97%)
  502 Bad Gateway: 3 (0.03%)

RPS: 1834.44 req/s
```

## Desenvolvimento Local

Para executar sem Docker:

```bash
go run main.go --url=http://localhost:8080 --requests=100 --concurrency=5
```

Para compilar executável:

```bash
go build -o loadtest main.go
./loadtest --url=http://exemplo.com --requests=1000 --concurrency=10
```

## Performance

O sistema é otimizado para:
- Milhares de requisições simultâneas
- Distribuição eficiente de carga entre workers
- Mínimo overhead de sincronização
- Tempo de resposta rápido do relatório

## Tratamento de Erros

- Valida todos os parâmetros obrigatórios
- Tratamento automático de timeouts (5)
- Contabiliza erros de conexão como status 0
- Relatório com informações sobre todos os códigos retornados

