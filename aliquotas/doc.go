// Package aliquotas expoe datasets estruturados de aliquotas brasileiras
// (ICMS por UF, PIS/COFINS por regime tributario, ISS por municipio
// coberto) + API Go de lookup.
//
// Granularidade MVP (T02.04):
//
//   - ICMS:        media interna por UF (27 entradas). Cobre 80%+ dos
//                  casos de telecom. NCM-especifico fica pra evolucao
//                  quando aparecer caso real.
//   - PIS/COFINS:  por regime tributario (Simples, Real, Presumido).
//   - ISS:         lista finita de IBGEs cobertos (clientes piloto +
//                  capitais). Caller usa Sefin Nacional como fallback
//                  quando ok=false.
//
// Alteracoes oficiais (NT CONFAZ pra ICMS, NT SEFAZ pra PIS/COFINS, lei
// municipal pra ISS) entram via bump patch semver — ver politica em
// docs/MIGRATION_NOVO_PAIS.md.
//
// Vigencia historica (recalculo retroativo com aliquota da epoca) NAO
// eh suportada — ver "Fora do escopo MVP" no README.
package aliquotas
