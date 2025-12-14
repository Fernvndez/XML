# 🧾 Sistema de Sincronização de NFe - SEFAZ

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-316192?style=flat&logo=postgresql)](https://www.postgresql.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

Sistema profissional em Golang para sincronização automática de Notas Fiscais Eletrônicas (NFe) diretamente da SEFAZ, com armazenamento em PostgreSQL e gestão de XMLs.

## 📋 Índice

- [Características](#-características)
- [Tecnologias](#-tecnologias)
- [Arquitetura](#-arquitetura)
- [Estrutura do Projeto](#-estrutura-do-projeto)
- [Pré-requisitos](#-pré-requisitos)
- [Instalação](#-instalação)
- [Configuração](#-configuração)
- [Executando](#-executando)
- [API Endpoints](#-api-endpoints)
- [Testes](#-testes)
- [Deploy](#-deploy)

## ✨ Características

- 🔐 **Integração Segura**: Comunicação SOAP com SEFAZ usando certificado digital A1
- 📥 **Download Automático**: Sincronização agendada de XMLs de NFe
- 💾 **Persistência**: Armazenamento estruturado em PostgreSQL
- 🏗️ **Clean Architecture**: Código organizado e testável
- 📊 **Logs Estruturados**: Rastreabilidade completa de operações
- ⚡ **Performance**: Pool de conexões e processamento assíncrono
- 🔄 **Retry Logic**: Tratamento robusto de falhas da SEFAZ
- 📁 **Gestão de Arquivos**: Organização automática de XMLs

## 🛠️ Tecnologias

- **Linguagem**: Go 1.21+
- **Banco de Dados**: PostgreSQL 15+
- **Framework Web**: Chi Router
- **ORM**: SQLX
- **Migrations**: Golang Migrate
- **Logs**: Zap (Uber)
- **Certificado**: Crypto/x509
- **Agendamento**: Cron
- **Testes**: Testify, SQLMock

## 🏛️ Arquitetura

```
Clean Architecture + Repository Pattern

┌─────────────────┐
│   HTTP/REST     │  (Handlers)
└────────┬────────┘
         │
┌────────▼────────┐
│   Services      │  (Business Logic)
└────────┬────────┘
         │
┌────────▼────────┐
│  Repositories   │  (Data Access)
└────────┬────────┘
         │
┌────────▼────────┐
│   PostgreSQL    │
└─────────────────┘
```

## 📁 Estrutura do Projeto

```
nfe-sefaz-sync/
├── cmd/
│   └── api/
│       └── main.go                 # Entry point da aplicação
├── internal/
│   ├── domain/                     # Entidades de domínio
│   │   ├── nfe.go
│   │   └── errors.go
│   ├── repository/                 # Camada de dados
│   │   ├── nfe_repository.go
│   │   └── nfe_repository_test.go
│   ├── service/                    # Lógica de negócio
│   │   ├── nfe_service.go
│   │   ├── sefaz_client.go
│   │   └── scheduler.go
│   └── handler/                    # Controllers HTTP
│       ├── nfe_handler.go
│       └── health_handler.go
├── pkg/
│   ├── logger/                     # Logger configurado
│   │   └── logger.go
│   ├── database/                   # Database setup
│   │   └── postgres.go
│   └── certificate/                # Gerenciamento de certificado
│       └── loader.go
├── configs/
│   └── config.go                   # Configurações da aplicação
├── migrations/
│   ├── 000001_create_nfe_table.up.sql
│   └── 000001_create_nfe_table.down.sql
├── storage/                        # Armazenamento de XMLs (git ignored)
│   └── xmls/
├── .env.example                    # Template de variáveis de ambiente
├── .gitignore
├── go.mod
├── go.sum
├── Makefile                        # Comandos úteis
├── Dockerfile
├── docker-compose.yml
└── README.md
```

## 📦 Pré-requisitos

- Go 1.21 ou superior
- PostgreSQL 15 ou superior
- Certificado Digital A1 (.pfx) válido
- Docker e Docker Compose (opcional)

## 🚀 Instalação

### 1. Clone o repositório

```bash
git clone https://github.com/seu-usuario/nfe-sefaz-sync.git
cd nfe-sefaz-sync
```

### 2. Instale as dependências

```bash
go mod download
```

### 3. Configure o banco de dados

#### Com Docker:

```bash
docker-compose up -d postgres
```

#### Sem Docker:

```bash
# Crie o banco de dados
createdb nfe_sefaz

# Execute as migrations
make migrate-up
```

## ⚙️ Configuração

### 1. Crie o arquivo `.env`

```bash
cp .env.example .env
```

### 2. Configure as variáveis de ambiente

```env
# Server
SERVER_PORT=8080
SERVER_HOST=localhost
ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=nfe_sefaz
DB_SSLMODE=disable
DB_MAX_CONNECTIONS=25
DB_MAX_IDLE_CONNECTIONS=5

# SEFAZ
SEFAZ_AMBIENTE=homologacao  # ou "producao"
SEFAZ_UF=SP
SEFAZ_CNPJ=12345678000100
SEFAZ_CERT_PATH=./certs/certificado.pfx
SEFAZ_CERT_PASSWORD=senha_do_certificado
SEFAZ_TIMEOUT=30s

# Storage
XML_STORAGE_PATH=./storage/xmls

# Scheduler
SYNC_CRON_SCHEDULE=0 */6 * * *  # A cada 6 horas
SYNC_ENABLED=true
```

### 3. Adicione seu certificado

```bash
mkdir -p certs
# Copie seu certificado .pfx para ./certs/
```

## 🎯 Executando

### Desenvolvimento

```bash
# Executar a aplicação
make run

# Executar com hot-reload
make dev

# Executar testes
make test

# Executar com cobertura
make test-coverage
```

### Docker

```bash
# Build e start todos os serviços
docker-compose up --build

# Apenas o banco
docker-compose up -d postgres

# Logs
docker-compose logs -f api
```

## 📡 API Endpoints

### Health Check

```http
GET /health
```

**Resposta:**
```json
{
  "status": "healthy",
  "database": "connected",
  "timestamp": "2025-12-13T10:30:00Z"
}
```

### Iniciar Sincronização Manual

```http
POST /api/v1/nfe/sync
```

**Resposta:**
```json
{
  "message": "Sincronização iniciada",
  "job_id": "uuid-do-job",
  "started_at": "2025-12-13T10:30:00Z"
}
```

### Listar NFes

```http
GET /api/v1/nfe?page=1&limit=20&start_date=2025-01-01&end_date=2025-12-31
```

**Resposta:**
```json
{
  "data": [
    {
      "id": "uuid",
      "chave_acesso": "35251234567890123456789012345678901234567890",
      "numero": "000123",
      "serie": "1",
      "cnpj_emitente": "12345678000100",
      "nome_emitente": "Empresa Exemplo LTDA",
      "data_emissao": "2025-12-13T10:00:00Z",
      "valor_total": 1500.50,
      "xml_path": "/storage/xmls/2025/12/35251234567890123456789012345678901234567890.xml",
      "status": "autorizada",
      "created_at": "2025-12-13T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150
  }
}
```

### Buscar NFe por Chave

```http
GET /api/v1/nfe/{chave_acesso}
```

### Download XML

```http
GET /api/v1/nfe/{chave_acesso}/xml
```

**Resposta**: Arquivo XML para download

### Estatísticas

```http
GET /api/v1/nfe/stats?start_date=2025-01-01&end_date=2025-12-31
```

**Resposta:**
```json
{
  "total_nfes": 1500,
  "valor_total": 450000.00,
  "periodo": {
    "inicio": "2025-01-01",
    "fim": "2025-12-31"
  },
  "por_status": {
    "autorizada": 1480,
    "cancelada": 20
  }
}
```

## 🧪 Testes

```bash
# Executar todos os testes
make test

# Testes com cobertura
make test-coverage

# Testes de integração
make test-integration

# Teste específico
go test -v ./internal/service/...
```

## 🐳 Deploy

### Docker Production

```bash
# Build da imagem
docker build -t nfe-sefaz-sync:latest .

# Run em produção
docker run -d \
  --name nfe-api \
  -p 8080:8080 \
  --env-file .env.production \
  -v /path/to/certs:/app/certs:ro \
  -v /path/to/storage:/app/storage \
  nfe-sefaz-sync:latest
```

### Kubernetes

```yaml
# Ver arquivo k8s/deployment.yaml
kubectl apply -f k8s/
```

## 📝 Makefile Commands

```bash
make help          # Mostra todos os comandos disponíveis
make run           # Executa a aplicação
make build         # Compila o binário
make test          # Executa os testes
make migrate-up    # Aplica migrations
make migrate-down  # Reverte migrations
make docker-build  # Build da imagem Docker
make lint          # Executa linter
make fmt           # Formata o código
```

## 🔒 Segurança

- ✅ Certificados armazenados com permissões restritas
- ✅ Senhas nunca commitadas (use .env)
- ✅ Conexões HTTPS obrigatórias com SEFAZ
- ✅ Validação de entrada em todos os endpoints
- ✅ Rate limiting configurável
- ✅ SQL injection prevenido (prepared statements)

## 📊 Monitoramento

- Logs estruturados em JSON
- Health check endpoint
- Métricas de performance
- Alertas de erro na sincronização

## 🤝 Contribuindo

1. Fork o projeto
2. Crie uma branch (`git checkout -b feature/nova-funcionalidade`)
3. Commit suas mudanças (`git commit -m 'Adiciona nova funcionalidade'`)
4. Push para a branch (`git push origin feature/nova-funcionalidade`)
5. Abra um Pull Request

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo [LICENSE](LICENSE) para mais detalhes.

## 👤 Autor

Seu Nome - [@seu_twitter](https://twitter.com/seu_twitter)

## 🙏 Agradecimentos

- Documentação da SEFAZ
- Comunidade Go Brasil
- Contribuidores do projeto

---

⭐ Se este projeto foi útil, considere dar uma estrela no GitHub!