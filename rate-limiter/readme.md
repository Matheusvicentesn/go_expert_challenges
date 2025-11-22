# 🛡️ Go Rate Limiter

Um Rate Limiter robusto e eficiente desenvolvido em **Go**, projetado para operar como um **middleware HTTP**. Ele controla o tráfego com base em **Endereço IP** ou **Token de Acesso**, utilizando **Redis** (com Lua Scripts) para garantir persistência e atomicidade em ambientes distribuídos.

---

## 🚀 Funcionalidades Chave

* **Limitação por IP:** Restringe o número de requisições por segundo para usuários não autenticados.
* **Limitação por Token:** Permite limites diferenciados (geralmente maiores) para requisições com Token (informado no header `API_KEY`).
* **Prioridade:** Configurações de limite por Token **sempre se sobrepõem** às de IP.
* **Bloqueio Temporário:** Bloqueia o emissor por um tempo configurável (`BLOCK_TIME`) após exceder o limite.
* **Atomicidade:** Uso de **Scripts Lua no Redis** para evitar condições de corrida (*Race Conditions*) em alta concorrência.
* **Strategy Pattern:** Implementação modular que permite fácil troca do Redis por outro mecanismo de persistência.

---

## ⚙️ Configuração (.env)

As configurações são definidas via variáveis de ambiente, tipicamente no arquivo `.env` na raiz do projeto.

| Variável | Padrão | Descrição |
| :--- | :--- | :--- |
| `SERVER_PORT` | `8080` | Porta onde o servidor irá rodar. |
| `REDIS_ADDR` | `redis:6379` | Endereço do servidor Redis. |
| `RATE_LIMIT_IP` | `5` | Máximo de requisições/segundo por **IP**. |
| `RATE_LIMIT_TOKEN`| `10` | Máximo de requisições/segundo por **Token**. |
| `BLOCK_TIME` | `300` | Tempo de bloqueio (em segundos) após exceder o limite (código 429). |

---

## 🐳 Como Rodar (Docker Compose)

A maneira mais fácil de iniciar o Rate Limiter e o Redis é usando o Docker Compose.

**Suba os containers (App + Redis):**
```bash
 docker-compose up --build
```
O servidor estará disponível em `http://localhost:8080`.

---

## 🧪 Como Testar

Para garantir a eficácia do Rate Limiter, você pode realizar testes manuais usando `cURL` ou utilizar o container de teste de carga (`stress-test`) incluído no `docker-compose.yml`.

### 1. Teste Manual com cURL

#### A. Teste por IP (Sem Token)

Este teste valida o limite baixo configurado para IPs (padrão: 5 requisições por segundo).

```bash
# O comando dispara 10 requisições sequenciais.
# As primeiras 5 devem retornar 200, as seguintes 429.

# ZSH (Linux/Mac)
repeat 10 curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/ | sort | uniq -c

# Bash
for i in {1..10}; do curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/; done | sort | uniq -c
```

#### B. Teste por Token

Este teste valida o limite mais alto configurado para Tokens (padrão: 10 requisições por segundo), confirmando que ele sobrescreve o limite de IP.

```bash
# O Token deve ser passado no header API_KEY.
# As primeiras 10 devem retornar 200, as seguintes 429.

# ZSH (Linux/Mac)
repeat 15 curl -H "API_KEY: meutokensecreto" -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/ | sort | uniq -c

# Bash
for i in {1..15}; do curl -H "API_KEY: meutokensecreto" -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/; done | sort | uniq -c
```

#### C. Teste via stress-test (Outro projeto)
[Link do readme](https://github.com/Matheusvicentesn/go_expert_challenges/blob/main/stress-test/readme.md)
