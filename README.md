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

Em construção — primeira sprint dedicada (T02.01–T02.06).

| BL | Escopo | Status |
|---|---|---|
| T02.01 | Estrutura inicial do repo | em curso |
| T02.02 | Lift-and-shift do `internal/integration/reformatax` do OTAX-001 | pendente |
| T02.03 | Validators BR (NCM/CFOP/CNAE/CNPJ/CPF) | pendente |
| T02.04 | Datasets de alíquotas | bloqueado fiscal |
| T02.05 | OTAX-001 consome via Go module | pendente |
| T02.06 | ADRs + migration guide | pendente |

## Como integrar

(será preenchido conforme os módulos forem entregues)

## Stack

- Go 1.26
- Sem dependência de banco — biblioteca pura
- Datasets em JSON/Go puro embarcados

## Licença

Proprietário G2 Tecnologia.
