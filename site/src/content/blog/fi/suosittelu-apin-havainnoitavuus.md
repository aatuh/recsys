---
title: "Suosittelu-APIn havainnoitavuus: mitä lokittaa ennen häiriöitä"
description: "Suosittelu-APIn havainnoitavuus request ID:ille, tenant- ja pintatiedoille, tyhjien suositusten osuudelle, latenssille, varoituksille ja palautuksen debuggaamiselle."
language: "fi"
pubDate: "2026-04-25"
translationKey: "api-observability"
tags: ["suosittelu-API", "havainnoitavuus", "operaatiot"]
---

Suosittelu-APIn havainnoitavuus on helpointa rakentaa ennen ensimmäistä tuotantohäiriötä. Kun tiimi jo debuggaa tyhjiä
vastauksia, vanhentuneita artefakteja tai selittämättömiä ranking-muutoksia, puuttuva request-konteksti muuttuu ongelman
sisäiseksi ongelmaksi.

Tavoite ei ole meluisa lokitus. Suosittelu-APIn pitäisi tuottaa riittävä rakenteinen näyttö kolmeen kysymykseen: mitä
palvelu vastaanotti, mitä päätöspolkua se käytti ja mikä artefakti- tai sääntöversio tuotti vastauksen?

## Suosittelu-APIn havainnoitavuus alkaa request-kontekstista

Jokainen tarjoiltu suositteluvastaus pitäisi voida jäljittää vakaalla request ID:llä. Saman request ID:n pitää näkyä
API-vastauksen metadatassa, exposure-tapahtumassa, arvioinnin syötteessä ja operatiivisissa lokeissa.

Hyödyllinen request-konteksti sisältää yleensä:

- request ID
- tenant, pinta ja ympäristö
- pseudonyymi käyttäjä- tai sessiotunniste, jos saatavilla
- kandidaatti- tai item-määrä kelpoisuussuodatusten jälkeen
- aktiivinen artefaktimanifestin versio
- ranking-polku, sääntöpolku tai fallback-polku
- vastauksen koko ja tyhjän vastauksen syy
- latenssiluokka ja timeout-tila

Vältä raakaa henkilötietoa operatiivisissa lokeissa.
[Tietoturvasivu](/fi/tietoturva/) kuvaa nykyisen itse ylläpidettävän ja pseudonyymeihin tunnisteisiin perustuvan linjan
tuotetasolla.

## Mittarit, jotka löytävät suosittelun virheitä

Suosittelujärjestelmät voivat epäonnistua tavoilla, joita tavallinen API-käytettävyys ei huomaa. Palvelu voi palauttaa
HTTP 200 -vastauksen, vaikka se tarjoilisi vain fallbackeja, vanhentuneita artefakteja tai tyhjiä listoja.

Seuraa mittareita, jotka liittyvät näkyvään laatuun ja operoijan toimenpiteisiin:

- tyhjien suosittelujen osuus tenantin ja pinnan mukaan
- fallback-käytön osuus
- artefaktin tuoreusikä
- p95- ja p99-latenssi endpointin mukaan
- timeout- ja degraded mode -osuus
- exposure-tapahtumien kirjoitusvirheet
- outcome-liitosten kattavuus arvioidulle liikenteelle

Näiden mittareiden pitäisi olla näkyvissä ennen julkaisua ja seurannassa jokaisen ranking- tai pipeline-julkaisun aikana.

## Debuggaa tyhjät suositukset ilman arvailua

Tyhjät suositukset heikentävät luottamusta suositteluun nopeasti. Debug-polun pitäisi olla mekaaninen: etsi request ID,
tarkista päätöspolku, varmista aktiivinen artefakti, tarkista kelpoisuussuodattimet ja vertaa tulosta odotettuun
fallback-käyttäytymiseen.

Jos palvelu osaa sanoa vain "suosituksia ei löytynyt", operoijan täytyy rekonstruoida häiriö hajanaisesta pipeline- ja
API-tilasta. Parempi tyhjän vastauksen jälki kertoo, johtuiko ongelma puuttuvista kandidaateista, vanhentuneista
featureista, tenant-rajauksesta, sääntösuodattimista, artefaktin latausvirheestä vai tarkoituksellisesta fallbackista.

Teknisissä operointidokumenteissa on erillinen
[tyhjien suositusten runbook](/documentation/technical/operations/runbooks/empty-recommendations/) ja
[palvelun valmiuden runbook](/documentation/technical/operations/runbooks/service-not-ready/).

## Säilytä havainnoitavuus palautuksen yli

Rollback ei saa poistaa näyttöä, jota tarvitaan epäonnistumisen selittämiseen. Kun manifestiosoitin, config tai
ranking-sääntö palautetaan, säilytä päätösjäljessä sekä edellinen että palautettu versio.

Tässä suosittelu-APIn havainnoitavuus liittyy arviointiin. Palautus on uskottavampi, kun tiimi voi näyttää, mikä liikenne
näki muuttuneen polun, mikä guardrail petti ja mikä vipu palautti aiemman toiminnan.

Toteutustasolla aloita [operointidokumentaatiosta](/documentation/technical/operations/) ja
[artefaktien ja pipelinejen oppaasta](/documentation/technical/artifacts-and-pipelines/). Kaupallista tai pilottikeskustelua
varten käytä [yhteyssivua](/fi/yhteys/).
