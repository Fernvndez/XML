package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nfe-sefaz-sync/internal/domain"
)

func setupMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	return sqlxDB, mock
}

func TestCreate(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewNFeRepository(db)

	nfe := &domain.NFe{
		ID:           uuid.New(),
		ChaveAcesso:  "35251234567890123456789012345678901234567890",
		Numero:       "000123",
		Serie:        "1",
		CNPJEmitente: "12345678000100",
		NomeEmitente: "Empresa Teste LTDA",
		DataEmissao:  time.Now(),
		ValorTotal:   1500.50,
		XMLPath:      "/storage/xmls/2025/12/35251234567890123456789012345678901234567890.xml",
		Status:       domain.NFeStatusAutorizada,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	mock.ExpectExec("INSERT INTO nfes").
		WithArgs(
			nfe.ID,
			nfe.ChaveAcesso,
			nfe.Numero,
			nfe.Serie,
			nfe.CNPJEmitente,
			nfe.NomeEmitente,
			nfe.DataEmissao,
			nfe.ValorTotal,
			nfe.XMLPath,
			nfe.Status,
			nfe.CreatedAt,
			nfe.UpdatedAt,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(nfe)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByChaveAcesso_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewNFeRepository(db)

	chaveAcesso := "35251234567890123456789012345678901234567890"
	expectedNFe := &domain.NFe{
		ID:           uuid.New(),
		ChaveAcesso:  chaveAcesso,
		Numero:       "000123",
		Serie:        "1",
		CNPJEmitente: "12345678000100",
		NomeEmitente: "Empresa Teste LTDA",
		DataEmissao:  time.Now(),
		ValorTotal:   1500.50,
		XMLPath:      "/storage/xmls/2025/12/35251234567890123456789012345678901234567890.xml",
		Status:       domain.NFeStatusAutorizada,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	rows := sqlmock.NewRows([]string{
		"id", "chave_acesso", "numero", "serie", "cnpj_emitente",
		"nome_emitente", "data_emissao", "valor_total", "xml_path",
		"status", "data_cancelamento", "motivo_cancelamento",
		"created_at", "updated_at",
	}).AddRow(
		expectedNFe.ID,
		expectedNFe.ChaveAcesso,
		expectedNFe.Numero,
		expectedNFe.Serie,
		expectedNFe.CNPJEmitente,
		expectedNFe.NomeEmitente,
		expectedNFe.DataEmissao,
		expectedNFe.ValorTotal,
		expectedNFe.XMLPath,
		expectedNFe.Status,
		nil,
		"",
		expectedNFe.CreatedAt,
		expectedNFe.UpdatedAt,
	)

	mock.ExpectQuery("SELECT (.+) FROM nfes WHERE chave_acesso").
		WithArgs(chaveAcesso).
		WillReturnRows(rows)

	nfe, err := repo.FindByChaveAcesso(chaveAcesso)
	assert.NoError(t, err)
	assert.NotNil(t, nfe)
	assert.Equal(t, chaveAcesso, nfe.ChaveAcesso)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByChaveAcesso_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewNFeRepository(db)

	chaveAcesso := "35251234567890123456789012345678901234567890"

	mock.ExpectQuery("SELECT (.+) FROM nfes WHERE chave_acesso").
		WithArgs(chaveAcesso).
		WillReturnError(sql.ErrNoRows)

	nfe, err := repo.FindByChaveAcesso(chaveAcesso)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrNFeNotFound, err)
	assert.Nil(t, nfe)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExistsByChaveAcesso(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewNFeRepository(db)

	chaveAcesso := "35251234567890123456789012345678901234567890"

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(chaveAcesso).
		WillReturnRows(rows)

	exists, err := repo.ExistsByChaveAcesso(chaveAcesso)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByFilter(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewNFeRepository(db)

	filter := domain.NFeFilter{
		Page:  1,
		Limit: 20,
	}

	// Mock count query
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(countRows)

	// Mock select query
	rows := sqlmock.NewRows([]string{
		"id", "chave_acesso", "numero", "serie", "cnpj_emitente",
		"nome_emitente", "data_emissao", "valor_total", "xml_path",
		"status", "data_cancelamento", "motivo_cancelamento",
		"created_at", "updated_at",
	}).AddRow(
		uuid.New(),
		"35251234567890123456789012345678901234567890",
		"000123",
		"1",
		"12345678000100",
		"Empresa Teste LTDA",
		time.Now(),
		1500.50,
		"/storage/xmls/2025/12/35251234567890123456789012345678901234567890.xml",
		domain.NFeStatusAutorizada,
		nil,
		"",
		time.Now(),
		time.Now(),
	)

	mock.ExpectQuery("SELECT (.+) FROM nfes (.+) ORDER BY data_emissao DESC").
		WillReturnRows(rows)

	nfes, total, err := repo.FindByFilter(filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, nfes, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}