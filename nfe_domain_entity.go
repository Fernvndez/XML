package domain

import (
	"time"

	"github.com/google/uuid"
)

// NFe representa uma Nota Fiscal Eletrônica no domínio da aplicação
type NFe struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	ChaveAcesso   string     `json:"chave_acesso" db:"chave_acesso"`
	Numero        string     `json:"numero" db:"numero"`
	Serie         string     `json:"serie" db:"serie"`
	CNPJEmitente  string     `json:"cnpj_emitente" db:"cnpj_emitente"`
	NomeEmitente  string     `json:"nome_emitente" db:"nome_emitente"`
	DataEmissao   time.Time  `json:"data_emissao" db:"data_emissao"`
	ValorTotal    float64    `json:"valor_total" db:"valor_total"`
	XMLPath       string     `json:"xml_path" db:"xml_path"`
	Status        NFeStatus  `json:"status" db:"status"`
	DataCancelamento *time.Time `json:"data_cancelamento,omitempty" db:"data_cancelamento"`
	MotivoCancelamento string  `json:"motivo_cancelamento,omitempty" db:"motivo_cancelamento"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// NFeStatus representa o status de uma NFe
type NFeStatus string

const (
	NFeStatusAutorizada  NFeStatus = "autorizada"
	NFeStatusCancelada   NFeStatus = "cancelada"
	NFeStatusDenegada    NFeStatus = "denegada"
	NFeStatusRejeitada   NFeStatus = "rejeitada"
	NFeStatusProcessando NFeStatus = "processando"
)

// IsValid verifica se o status é válido
func (s NFeStatus) IsValid() bool {
	switch s {
	case NFeStatusAutorizada, NFeStatusCancelada, NFeStatusDenegada, 
		 NFeStatusRejeitada, NFeStatusProcessando:
		return true
	}
	return false
}

// NFeFilter representa os filtros para busca de NFes
type NFeFilter struct {
	CNPJEmitente string     `json:"cnpj_emitente"`
	Status       NFeStatus  `json:"status"`
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	Page         int        `json:"page"`
	Limit        int        `json:"limit"`
}

// Validate valida os filtros
func (f *NFeFilter) Validate() error {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
	if f.Status != "" && !f.Status.IsValid() {
		return ErrInvalidStatus
	}
	return nil
}

// GetOffset retorna o offset para paginação
func (f *NFeFilter) GetOffset() int {
	return (f.Page - 1) * f.Limit
}

// NFePaginatedResponse representa uma resposta paginada de NFes
type NFePaginatedResponse struct {
	Data       []NFe      `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// Pagination representa informações de paginação
type Pagination struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

// NFeStats representa estatísticas de NFes
type NFeStats struct {
	TotalNFes    int64              `json:"total_nfes"`
	ValorTotal   float64            `json:"valor_total"`
	Periodo      Periodo            `json:"periodo"`
	PorStatus    map[NFeStatus]int64 `json:"por_status"`
}

// Periodo representa um período de datas
type Periodo struct {
	Inicio time.Time `json:"inicio"`
	Fim    time.Time `json:"fim"`
}

// SyncJob representa um job de sincronização
type SyncJob struct {
	ID        uuid.UUID       `json:"id"`
	Status    SyncJobStatus   `json:"status"`
	StartedAt time.Time       `json:"started_at"`
	EndedAt   *time.Time      `json:"ended_at,omitempty"`
	NFesFound int             `json:"nfes_found"`
	NFesError int             `json:"nfes_error"`
	Error     string          `json:"error,omitempty"`
}

// SyncJobStatus representa o status de um job de sincronização
type SyncJobStatus string

const (
	SyncJobStatusRunning   SyncJobStatus = "running"
	SyncJobStatusCompleted SyncJobStatus = "completed"
	SyncJobStatusFailed    SyncJobStatus = "failed"
)

// NFeRepository define a interface para repositório de NFes
type NFeRepository interface {
	Create(nfe *NFe) error
	Update(nfe *NFe) error
	FindByChaveAcesso(chaveAcesso string) (*NFe, error)
	FindByFilter(filter NFeFilter) ([]NFe, int64, error)
	ExistsByChaveAcesso(chaveAcesso string) (bool, error)
	GetStats(startDate, endDate time.Time) (*NFeStats, error)
}

// NFeService define a interface para serviço de NFes
type NFeService interface {
	SyncNFes() (*SyncJob, error)
	ListNFes(filter NFeFilter) (*NFePaginatedResponse, error)
	GetNFeByChave(chaveAcesso string) (*NFe, error)
	GetXMLPath(chaveAcesso string) (string, error)
	GetStats(startDate, endDate time.Time) (*NFeStats, error)
}

// SefazClient define a interface para cliente SEFAZ
type SefazClient interface {
	ConsultarNFes(cnpj string, dataInicio, dataFim time.Time) ([]string, error)
	DownloadXML(chaveAcesso string) ([]byte, error)
}