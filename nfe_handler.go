package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"nfe-sefaz-sync/internal/domain"
	"nfe-sefaz-sync/pkg/logger"
)

// NFeHandler gerencia os endpoints relacionados a NFe
type NFeHandler struct {
	service domain.NFeService
	logger  *logger.Logger
}

// NewNFeHandler cria uma nova instância do handler
func NewNFeHandler(service domain.NFeService, log *logger.Logger) *NFeHandler {
	return &NFeHandler{
		service: service,
		logger:  log,
	}
}

// RegisterRoutes registra as rotas do handler
func (h *NFeHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/nfe", func(r chi.Router) {
		r.Post("/sync", h.SyncNFes)
		r.Get("/", h.ListNFes)
		r.Get("/{chave}", h.GetNFe)
		r.Get("/{chave}/xml", h.DownloadXML)
		r.Get("/stats", h.GetStats)
	})
}

// SyncNFes inicia a sincronização de NFes
// @Summary Sincronizar NFes
// @Description Inicia a sincronização automática de NFes da SEFAZ
// @Tags NFe
// @Accept json
// @Produce json
// @Success 200 {object} domain.SyncJob
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/nfe/sync [post]
func (h *NFeHandler) SyncNFes(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Requisição de sincronização recebida")

	job, err := h.service.SyncNFes()
	if err != nil {
		h.logger.Error("Erro ao sincronizar NFes", "error", err)
		h.sendError(w, http.StatusInternalServerError, "Erro ao sincronizar NFes", err)
		return
	}

	h.sendJSON(w, http.StatusOK, job)
}

// ListNFes lista NFes com filtros e paginação
// @Summary Listar NFes
// @Description Lista NFes com filtros e paginação
// @Tags NFe
// @Accept json
// @Produce json
// @Param page query int false "Número da página" default(1)
// @Param limit query int false "Itens por página" default(20)
// @Param cnpj_emitente query string false "CNPJ do emitente"
// @Param status query string false "Status da NFe"
// @Param start_date query string false "Data início (YYYY-MM-DD)"
// @Param end_date query string false "Data fim (YYYY-MM-DD)"
// @Success 200 {object} domain.NFePaginatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/nfe [get]
func (h *NFeHandler) ListNFes(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	filter := domain.NFeFilter{
		CNPJEmitente: r.URL.Query().Get("cnpj_emitente"),
		Status:       domain.NFeStatus(r.URL.Query().Get("status")),
	}

	// Page
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			filter.Page = page
		}
	}

	// Limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	// Start date
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}

	// End date
	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filter.EndDate = &endDate
		}
	}

	// Lista as NFes
	response, err := h.service.ListNFes(filter)
	if err != nil {
		h.logger.Error("Erro ao listar NFes", "error", err)
		h.sendError(w, http.StatusInternalServerError, "Erro ao listar NFes", err)
		return
	}

	h.sendJSON(w, http.StatusOK, response)
}

// GetNFe retorna uma NFe específica pela chave de acesso
// @Summary Buscar NFe
// @Description Retorna uma NFe específica pela chave de acesso
// @Tags NFe
// @Accept json
// @Produce json
// @Param chave path string true "Chave de acesso da NFe"
// @Success 200 {object} domain.NFe
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/nfe/{chave} [get]
func (h *NFeHandler) GetNFe(w http.ResponseWriter, r *http.Request) {
	chaveAcesso := chi.URLParam(r, "chave")

	nfe, err := h.service.GetNFeByChave(chaveAcesso)
	if err != nil {
		if err == domain.ErrNFeNotFound {
			h.sendError(w, http.StatusNotFound, "NFe não encontrada", err)
			return
		}
		h.logger.Error("Erro ao buscar NFe", "chave", chaveAcesso, "error", err)
		h.sendError(w, http.StatusInternalServerError, "Erro ao buscar NFe", err)
		return
	}

	h.sendJSON(w, http.StatusOK, nfe)
}

// DownloadXML faz download do XML de uma NFe
// @Summary Download XML
// @Description Faz download do arquivo XML de uma NFe
// @Tags NFe
// @Accept json
// @Produce application/xml
// @Param chave path string true "Chave de acesso da NFe"
// @Success 200 {file} file
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/nfe/{chave}/xml [get]
func (h *NFeHandler) DownloadXML(w http.ResponseWriter, r *http.Request) {
	chaveAcesso := chi.URLParam(r, "chave")

	xmlPath, err := h.service.GetXMLPath(chaveAcesso)
	if err != nil {
		if err == domain.ErrNFeNotFound {
			h.sendError(w, http.StatusNotFound, "NFe não encontrada", err)
			return
		}
		h.logger.Error("Erro ao buscar XML", "chave", chaveAcesso, "error", err)
		h.sendError(w, http.StatusInternalServerError, "Erro ao buscar XML", err)
		return
	}

	// Lê o arquivo XML
	xmlData, err := os.ReadFile(xmlPath)
	if err != nil {
		h.logger.Error("Erro ao ler arquivo XML", "path", xmlPath, "error", err)
		h.sendError(w, http.StatusInternalServerError, "Erro ao ler XML", err)
		return
	}

	// Define headers para download
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Disposition", "attachment; filename="+chaveAcesso+".xml")
	w.Header().Set("Content-Length", strconv.Itoa(len(xmlData)))

	// Envia o arquivo
	w.WriteHeader(http.StatusOK)
	w.Write(xmlData)
}

// GetStats retorna estatísticas de NFes
// @Summary Estatísticas
// @Description Retorna estatísticas de NFes em um período
// @Tags NFe
// @Accept json
// @Produce json
// @Param start_date query string true "Data início (YYYY-MM-DD)"
// @Param end_date query string true "Data fim (YYYY-MM-DD)"
// @Success 200 {object} domain.NFeStats
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/nfe/stats [get]
func (h *NFeHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	// Parse dates
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	if startDateStr == "" || endDateStr == "" {
		h.sendError(w, http.StatusBadRequest, "start_date e end_date são obrigatórios", nil)
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Formato de data inválido para start_date", err)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Formato de data inválido para end_date", err)
		return
	}

	// Busca estatísticas
	stats, err := h.service.GetStats(startDate, endDate)
	if err != nil {
		h.logger.Error("Erro ao buscar estatísticas", "error", err)
		h.sendError(w, http.StatusInternalServerError, "Erro ao buscar estatísticas", err)
		return
	}

	h.sendJSON(w, http.StatusOK, stats)
}

// ErrorResponse representa uma resposta de erro
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// sendJSON envia uma resposta JSON
func (h *NFeHandler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// sendError envia uma resposta de erro
func (h *NFeHandler) sendError(w http.ResponseWriter, status int, message string, err error) {
	errResp := ErrorResponse{
		Message: message,
	}
	if err != nil {
		errResp.Error = err.Error()
	}
	h.sendJSON(w, status, errResp)
}