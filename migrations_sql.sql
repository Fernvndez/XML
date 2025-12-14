-- Create extension for UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create nfes table
CREATE TABLE IF NOT EXISTS nfes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chave_acesso VARCHAR(44) UNIQUE NOT NULL,
    numero VARCHAR(20) NOT NULL,
    serie VARCHAR(10) NOT NULL,
    cnpj_emitente VARCHAR(14) NOT NULL,
    nome_emitente VARCHAR(255) NOT NULL,
    data_emissao TIMESTAMP NOT NULL,
    valor_total DECIMAL(15, 2) NOT NULL,
    xml_path VARCHAR(500) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'autorizada',
    data_cancelamento TIMESTAMP,
    motivo_cancelamento TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for better query performance
CREATE INDEX idx_nfes_chave_acesso ON nfes(chave_acesso);
CREATE INDEX idx_nfes_cnpj_emitente ON nfes(cnpj_emitente);
CREATE INDEX idx_nfes_data_emissao ON nfes(data_emissao DESC);
CREATE INDEX idx_nfes_status ON nfes(status);
CREATE INDEX idx_nfes_created_at ON nfes(created_at DESC);

-- Create composite index for common queries
CREATE INDEX idx_nfes_cnpj_data ON nfes(cnpj_emitente, data_emissao DESC);

-- Add comments for documentation
COMMENT ON TABLE nfes IS 'Tabela de Notas Fiscais Eletrônicas';
COMMENT ON COLUMN nfes.chave_acesso IS 'Chave de acesso da NFe (44 dígitos)';
COMMENT ON COLUMN nfes.cnpj_emitente IS 'CNPJ do emitente da nota fiscal';
COMMENT ON COLUMN nfes.data_emissao IS 'Data e hora de emissão da NFe';
COMMENT ON COLUMN nfes.valor_total IS 'Valor total da nota fiscal';
COMMENT ON COLUMN nfes.xml_path IS 'Caminho do arquivo XML no sistema de arquivos';
COMMENT ON COLUMN nfes.status IS 'Status da NFe: autorizada, cancelada, denegada, rejeitada, processando';