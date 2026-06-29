# oasis-tax-localization-br

Biblioteca de **localização fiscal brasileira** do ecossistema **Oásis** (TM Forum ODA).

Encapsula tudo que é específico do regime tributário brasileiro — alíquotas, validações de códigos fiscais (NCM/CFOP/CNAE/CNPJ/CPF), e o cliente da Calculadora oficial da Reforma Tributária do Consumo (CBS/IBS/IS) — pra que o motor fiscal (OTAX-001) fique agnóstico a país.

Equivalente conceitual: T02 no diagrama de arquitetura do Tax.

## Por quê

O motor fiscal (T01) é genérico — recebe regras, aplica DSL, produz cálculo. Tudo que é **regulação brasileira** vive aqui:

- Tabelas de alíquotas (ICMS por UF, PIS/COFINS por regime, ISS por município)
- Validators de códigos fiscais BR (NCM, CFOP, CNAE, CNPJ, CPF)
- Integração com a Calculadora da Reforma Tributária (RFB)

Próximos países (MX, AR, …) seguirão o mesmo padrão: `oasis-tax-localization-mx`, etc.

## Estrutura

```
reforma/      cliente HTTP da Calculadora da Reforma Tributária (CBS/IBS/IS)
aliquotas/    datasets BR estruturados + API de lookup
validators/   regras de validação de códigos fiscais BR
docs/         ADRs + migration guides
```

## Status

Primeira sprint concluída.

| BL | Escopo | Status |
|---|---|---|
| T02.01 | Estrutura inicial do repo | ✅ entregue |
| T02.02 | Lift-and-shift do `internal/integration/reformatax` do OTAX-001 | ✅ entregue |
| T02.03 | Validators BR (CPF/CNPJ/NCM/CFOP/CNAE) | ✅ entregue |
| T02.04 | Datasets de alíquotas (ICMS por UF, PIS/COFINS por regime, ISS por município piloto) | ✅ entregue |
| T02.05 | OTAX-001 + OTAX-002 consomem via Go module | ✅ entregue |
| T02.06 | ADR + migration guide | ✅ entregue |
| T02.07 | UF→IBGE canônico | ✅ entregue |
| T02.08 | CSTs por imposto (ICMS, PIS, COFINS) | ✅ entregue |

## Fora do escopo MVP

### Datasets de alíquotas com vigência histórica (BL-T02.09)

A biblioteca expõe **alíquotas correntes** — uma alíquota vigente por entrada de dataset. Quando o time fiscal G2 publica nova alíquota (ex.: ICMS-SP muda de 18% para 17%), o repo recebe bump patch semver (`v0.3.1 → v0.3.2`) e os consumidores atualizam `go.mod`.

A biblioteca **não** mantém histórico de mudanças de alíquota com janelas de vigência (ex.: API tipo `GetICMS(uf, data) AliquotaSet`). Razões:

1. **Sem demanda real hoje**. O fluxo de recálculo pós-REJECTED no OTAX-002 (entregue no BL-113b) reusa o `tax_calculation_id` original — o snapshot do cálculo já carrega a alíquota efetivamente aplicada. Não há caso de uso atual exigindo "qual era a alíquota em junho/2024".
2. **Compromisso de manutenção caro**. Manter histórico oficial exigiria levantamento retroativo (a partir de qual ano? 2018? 2026?) + estratégia de atualização contínua + auditoria. Esforço desproporcional ao valor entregue hoje.
3. **Cobertura imperfeita inevitável**. Mesmo se fosse implementado, datasets históricos teriam buracos (mudanças municipais de ISS são especialmente difíceis de rastrear). A aparência de "fonte oficial histórica" pode induzir uso indevido em auditoria fiscal.

**Como reabrir essa decisão**: caso apareça caso de uso concreto (ex.: cliente exige recálculo retroativo de janela X com alíquota da época), abrir issue marcada `mvp-revisit` documentando o caso. A API seria estendida pra `GetICMS(uf string, data time.Time) AliquotaSet` mantendo a versão atual como atalho.

## Como integrar

```go
import (
    "github.com/G2-Oasis/oasis-tax-localization-br/reforma"
    "github.com/G2-Oasis/oasis-tax-localization-br/validators"
)

// validar identificadores
ok := validators.IsValidCNPJ("13.332.378/0001-34")

// traduzir UF -> IBGE (chave de acesso, consStatServ, etc.)
cUF, ok := validators.UFToIBGECode("ES") // "32", true

// cliente da Calculadora da Reforma Tributária
client, err := reforma.New(mode, url, timeout)
```

Veja [`docs/MIGRATION_NOVO_PAIS.md`](./docs/MIGRATION_NOVO_PAIS.md) pra criar lib equivalente pra outros países.

## Stack

- Go 1.26
- Sem dependência de banco — biblioteca pura
- Datasets em Go puro embarcados

## Licença

Proprietário G2 Tecnologia.
